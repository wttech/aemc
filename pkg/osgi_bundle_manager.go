package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/osgi"
)

type OSGiBundleManager struct {
	instance *Instance

	InstallStart           bool
	InstallStartLevel      int
	InstallRefreshPackages bool
}

func NewBundleManager(instance *Instance) *OSGiBundleManager {
	return &OSGiBundleManager{
		instance: instance,

		InstallStart:           true,
		InstallStartLevel:      20,
		InstallRefreshPackages: true,
	}
}

func (bm *OSGiBundleManager) New(symbolicName string) OSGiBundle {
	return OSGiBundle{
		manager:      bm,
		symbolicName: symbolicName,
	}
}

func (bm *OSGiBundleManager) ByFile(localPath string) (*OSGiBundle, error) {
	manifest, err := osgi.ReadBundleManifest(localPath)
	if err != nil {
		return nil, err
	}
	return &OSGiBundle{manager: bm, symbolicName: manifest.SymbolicName}, nil
}

func (bm OSGiBundleManager) Find(symbolicName string) (*osgi.BundleListItem, error) {
	bundles, err := bm.List()
	if err != nil {
		return nil, fmt.Errorf("cannot find bundle '%s' on instance '%s'", symbolicName, bm.instance.ID())
	}
	item, found := lo.Find(bundles.List, func(i osgi.BundleListItem) bool { return symbolicName == i.SymbolicName })
	if found {
		return &item, nil
	}
	return nil, nil
}

func (bm *OSGiBundleManager) List() (*osgi.BundleList, error) {
	resp, err := bm.instance.http.Request().Get(BundlesPathJson)
	if err != nil {
		return nil, fmt.Errorf("cannot request bundle list on instance '%s': %w", bm.instance.ID(), err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("cannot request bundle list on instance '%s': %w", bm.instance.ID(), err)
	}
	var res osgi.BundleList
	if err = fmtx.UnmarshalJSON(resp.RawBody(), &res); err != nil {
		return nil, fmt.Errorf("cannot parse bundle list from instance '%s': %w", bm.instance.ID(), err)
	}
	return &res, nil
}

func (bm *OSGiBundleManager) Start(id int) error {
	log.Infof("starting bundle '%d' on instance '%s'", id, bm.instance.ID())
	response, err := bm.instance.http.Request().
		SetFormData(map[string]string{"action": "start"}).
		Post(fmt.Sprintf("%s/%d", BundlesPath, id))
	if err != nil {
		return fmt.Errorf("cannot start bundle '%d' on instance '%s': %w", id, bm.instance.ID(), err)
	} else if response.IsError() {
		return fmt.Errorf("cannot start bundle '%d' on instance '%s': %s", id, bm.instance.ID(), response.Status())
	}
	log.Infof("started bundle '%d' on instance '%s'", id, bm.instance.ID())
	return nil
}

func (bm *OSGiBundleManager) Stop(id int) error {
	log.Infof("stopping bundle '%d' on instance '%s'", id, bm.instance.ID())
	response, err := bm.instance.http.Request().
		SetFormData(map[string]string{"action": "stop"}).
		Post(fmt.Sprintf("%s/%d", BundlesPath, id))
	if err != nil {
		return fmt.Errorf("cannot stop bundle '%d' on instance '%s': %w", id, bm.instance.ID(), err)
	} else if response.IsError() {
		return fmt.Errorf("cannot stop bundle '%d' on instance '%s': %s", id, bm.instance.ID(), response.Status())
	}
	log.Infof("stopped bundle '%d' on instance '%s'", id, bm.instance.ID())
	return nil
}

func (bm *OSGiBundleManager) InstallWithChanged(localPath string) (bool, error) {
	manifest, err := osgi.ReadBundleManifest(localPath)
	if err != nil {
		return false, err
	}
	bundle := bm.New(manifest.SymbolicName)
	if err != nil {
		return false, nil
	}
	state, err := bundle.State()
	if err != nil {
		return false, err
	}
	if !state.Exists || state.data.Version != manifest.Version {
		err = bm.Install(localPath)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (bm *OSGiBundleManager) Install(localPath string) error {
	log.Infof("installing bundle '%s' on instance '%s'", localPath, bm.instance.ID())
	response, err := bm.instance.http.RequestFormData(map[string]any{
		"action":           "install",
		"bundlestart":      bm.InstallStart,
		"bundlestartlevel": bm.InstallStartLevel,
		"refreshPackages":  bm.InstallRefreshPackages,
	}).SetFile("bundlefile", localPath).Post(BundlesPath)
	if err != nil {
		return fmt.Errorf("cannot install bundle '%s' on instance '%s': %w", localPath, bm.instance.ID(), err)
	} else if response.IsError() {
		return fmt.Errorf("cannot install bundle '%s' on instance '%s': '%s'", localPath, bm.instance.ID(), response.Status())
	}
	log.Infof("installed bundle '%s' on instance '%s'", localPath, bm.instance.ID())
	return nil
}

func (bm *OSGiBundleManager) Uninstall(id int) error {
	log.Infof("uninstalling bundle '%d' from instance '%s'", id, bm.instance.ID())
	response, err := bm.instance.http.RequestFormData(map[string]any{"action": "uninstall"}).Post(fmt.Sprintf("%s/%d", BundlesPath, id))
	if err != nil {
		return fmt.Errorf("cannot uninstall bundle '%d' on instance '%s': %w", id, bm.instance.ID(), err)
	} else if response.IsError() {
		return fmt.Errorf("cannot uninstall bundle '%d' on instance '%s': '%s'", id, bm.instance.ID(), response.Status())
	}
	log.Infof("uninstalled bundle '%d' from instance '%s'", id, bm.instance.ID())
	return nil
}

const (
	BundlesPath     = "/system/console/bundles"
	BundlesPathJson = BundlesPath + ".json"
)
