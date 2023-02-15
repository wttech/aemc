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
		return nil, fmt.Errorf("instance '%s': cannot find bundle '%s'", bm.instance.ID(), symbolicName)
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
		return nil, fmt.Errorf("instance '%s': cannot request bundle list: %w", bm.instance.ID(), err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("instance '%s': cannot request bundle list: %s", bm.instance.ID(), resp.Status())
	}
	var res osgi.BundleList
	if err = fmtx.UnmarshalJSON(resp.RawBody(), &res); err != nil {
		return nil, fmt.Errorf("instance '%s': cannot parse bundle list: %w", bm.instance.ID(), err)
	}
	return &res, nil
}

func (bm *OSGiBundleManager) Start(id int) error {
	log.Infof("instance '%s': starting bundle '%d'", bm.instance.ID(), id)
	response, err := bm.instance.http.Request().
		SetFormData(map[string]string{"action": "start"}).
		Post(fmt.Sprintf("%s/%d", BundlesPath, id))
	if err != nil {
		return fmt.Errorf("instance '%s': cannot start bundle '%d': %w", bm.instance.ID(), id, err)
	} else if response.IsError() {
		return fmt.Errorf("instance '%s': cannot start bundle '%d': %s", bm.instance.ID(), id, response.Status())
	}
	log.Infof("instance '%s': started bundle '%d'", bm.instance.ID(), id)
	return nil
}

func (bm *OSGiBundleManager) Stop(id int) error {
	log.Infof("instance '%s': stopping bundle '%d'", bm.instance.ID(), id)
	response, err := bm.instance.http.Request().
		SetFormData(map[string]string{"action": "stop"}).
		Post(fmt.Sprintf("%s/%d", BundlesPath, id))
	if err != nil {
		return fmt.Errorf("instance '%s': cannot stop bundle '%d': %w", bm.instance.ID(), id, err)
	} else if response.IsError() {
		return fmt.Errorf("instance '%s': cannot stop bundle '%d': %s", bm.instance.ID(), id, response.Status())
	}
	log.Infof("instance '%s': stopped bundle '%d'", bm.instance.ID(), id)
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
	log.Infof("instance '%s': installing bundle '%s'", bm.instance.ID(), localPath)
	response, err := bm.instance.http.RequestFormData(map[string]any{
		"action":           "install",
		"bundlestart":      bm.InstallStart,
		"bundlestartlevel": bm.InstallStartLevel,
		"refreshPackages":  bm.InstallRefreshPackages,
	}).SetFile("bundlefile", localPath).Post(BundlesPath)
	if err != nil {
		return fmt.Errorf("instance '%s': cannot install bundle '%s': %w", bm.instance.ID(), localPath, err)
	} else if response.IsError() {
		return fmt.Errorf("instance '%s': cannot install bundle '%s': %s", bm.instance.ID(), localPath, response.Status())
	}
	log.Infof("instance '%s': installed bundle '%s'", bm.instance.ID(), localPath)
	return nil
}

func (bm *OSGiBundleManager) Uninstall(id int) error {
	log.Infof("instance '%s': uninstalling bundle '%d'", bm.instance.ID(), id)
	response, err := bm.instance.http.RequestFormData(map[string]any{"action": "uninstall"}).Post(fmt.Sprintf("%s/%d", BundlesPath, id))
	if err != nil {
		return fmt.Errorf("instance '%s': cannot uninstall bundle '%d': %w", bm.instance.ID(), id, err)
	} else if response.IsError() {
		return fmt.Errorf("instance '%s': cannot uninstall bundle '%d': %s", bm.instance.ID(), id, response.Status())
	}
	log.Infof("instance '%s': uninstalled bundle '%d'", bm.instance.ID(), id)
	return nil
}

const (
	BundlesPath     = "/system/console/bundles"
	BundlesPathJson = BundlesPath + ".json"
)
