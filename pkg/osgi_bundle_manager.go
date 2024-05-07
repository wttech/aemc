package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"github.com/wttech/aemc/pkg/osgi"
	"path/filepath"
	"time"
)

type OSGiBundleManager struct {
	instance *Instance

	InstallStart            bool
	InstallStartLevel       int
	InstallRefreshPackages  bool
	SnapshotInstallSkipping bool
	SnapshotIgnored         bool
	SnapshotPatterns        []string
}

func NewBundleManager(instance *Instance) *OSGiBundleManager {
	cv := instance.manager.aem.config.Values()

	return &OSGiBundleManager{
		instance: instance,

		InstallStart:            cv.GetBool("instance.osgi.bundle.install.start"),
		InstallStartLevel:       cv.GetInt("instance.osgi.bundle.install.start_level"),
		InstallRefreshPackages:  cv.GetBool("instance.osgi.bundle.install.refresh_packages"),
		SnapshotInstallSkipping: cv.GetBool("instance.osgi.bundle.snapshot_install_skipping"),
		SnapshotIgnored:         cv.GetBool("instance.osgi.bundle.snapshot_ignored"),
		SnapshotPatterns:        cv.GetStringSlice("instance.osgi.bundle.snapshot_patterns"),
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

func (bm *OSGiBundleManager) Find(symbolicName string) (*osgi.BundleListItem, error) {
	bundles, err := bm.List()
	if err != nil {
		return nil, fmt.Errorf("%s > cannot find bundle '%s'", bm.instance.IDColor(), symbolicName)
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
		return nil, fmt.Errorf("%s > cannot request bundle list: %w", bm.instance.IDColor(), err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot request bundle list: %s", bm.instance.IDColor(), resp.Status())
	}
	var res osgi.BundleList
	if err = fmtx.UnmarshalJSON(resp.RawBody(), &res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse bundle list: %w", bm.instance.IDColor(), err)
	}
	return &res, nil
}

func (bm *OSGiBundleManager) Start(id int) error {
	log.Infof("%s > starting bundle '%d'", bm.instance.IDColor(), id)
	response, err := bm.instance.http.Request().
		SetFormData(map[string]string{"action": "start"}).
		Post(fmt.Sprintf("%s/%d", BundlesPath, id))
	if err != nil {
		return fmt.Errorf("%s > cannot start bundle '%d': %w", bm.instance.IDColor(), id, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot start bundle '%d': %s", bm.instance.IDColor(), id, response.Status())
	}
	log.Infof("%s > started bundle '%d'", bm.instance.IDColor(), id)
	return nil
}

func (bm *OSGiBundleManager) Stop(id int) error {
	log.Infof("%s > stopping bundle '%d'", bm.instance.IDColor(), id)
	response, err := bm.instance.http.Request().
		SetFormData(map[string]string{"action": "stop"}).
		Post(fmt.Sprintf("%s/%d", BundlesPath, id))
	if err != nil {
		return fmt.Errorf("%s > cannot stop bundle '%d': %w", bm.instance.IDColor(), id, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot stop bundle '%d': %s", bm.instance.IDColor(), id, response.Status())
	}
	log.Infof("%s > stopped bundle '%d'", bm.instance.IDColor(), id)
	return nil
}

func (bm *OSGiBundleManager) IsSnapshot(localPath string) bool {
	return !bm.SnapshotIgnored && stringsx.MatchSome(pathx.Normalize(localPath), bm.SnapshotPatterns)
}

func (bm *OSGiBundleManager) InstallWithChanged(localPath string) (bool, error) {
	if bm.IsSnapshot(localPath) {
		return bm.installSnapshot(localPath)
	}
	return bm.installRegular(localPath)
}

func (bm *OSGiBundleManager) installRegular(localPath string) (bool, error) {
	installed, err := bm.IsInstalled(localPath)
	if err != nil {
		return false, err
	}
	if !installed {
		return true, bm.Install(localPath)
	}
	return false, nil
}

func (bm *OSGiBundleManager) IsInstalled(localPath string) (bool, error) {
	manifest, err := osgi.ReadBundleManifest(localPath)
	if err != nil {
		return false, err
	}
	bundle := bm.New(manifest.SymbolicName)
	state, err := bundle.State()
	if err != nil {
		return false, err
	}
	return state.Exists && state.data.Version == manifest.Version, nil
}

func (bm *OSGiBundleManager) installSnapshot(localPath string) (bool, error) {
	checksum, err := filex.ChecksumFile(localPath)
	if err != nil {
		return false, err
	}
	installed, err := bm.IsInstalled(localPath)
	if err != nil {
		return false, err
	}
	var lock = bm.installLock(localPath, checksum)
	if installed && bm.SnapshotInstallSkipping && lock.IsLocked() {
		lockData, err := lock.Locked()
		if err != nil {
			return false, err
		}
		if checksum == lockData.Checksum {
			log.Infof("%s > skipped installing bundle '%s'", bm.instance.IDColor(), localPath)
			return false, nil
		}
	}
	if err := bm.Install(localPath); err != nil {
		return false, err
	}
	if err := lock.Lock(); err != nil {
		return false, err
	}
	return true, nil
}

func (bm *OSGiBundleManager) installLock(file string, checksum string) osx.Lock[osgiBundleInstallLock] {
	name := filepath.Base(file)
	return osx.NewLock(fmt.Sprintf("%s/osgi/bundle/install/%s.yml", bm.instance.LockDir(), name), func() (osgiBundleInstallLock, error) {
		return osgiBundleInstallLock{Installed: time.Now(), Checksum: checksum}, nil
	})
}

type osgiBundleInstallLock struct {
	Installed time.Time `yaml:"installed"`
	Checksum  string    `yaml:"checksum"`
}

func (bm *OSGiBundleManager) Install(localPath string) error {
	log.Infof("%s > installing bundle '%s'", bm.instance.IDColor(), localPath)
	response, err := bm.instance.http.RequestFormData(map[string]any{
		"action":           "install",
		"bundlestart":      bm.InstallStart,
		"bundlestartlevel": bm.InstallStartLevel,
		"refreshPackages":  bm.InstallRefreshPackages,
	}).SetFile("bundlefile", localPath).Post(BundlesPath)
	if err != nil {
		return fmt.Errorf("%s > cannot install bundle '%s': %w", bm.instance.IDColor(), localPath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot install bundle '%s': %s", bm.instance.IDColor(), localPath, response.Status())
	}
	log.Infof("%s > installed bundle '%s'", bm.instance.IDColor(), localPath)
	return nil
}

func (bm *OSGiBundleManager) Uninstall(id int) error {
	log.Infof("%s > uninstalling bundle '%d'", bm.instance.IDColor(), id)
	response, err := bm.instance.http.RequestFormData(map[string]any{"action": "uninstall"}).Post(fmt.Sprintf("%s/%d", BundlesPath, id))
	if err != nil {
		return fmt.Errorf("%s > cannot uninstall bundle '%d': %w", bm.instance.IDColor(), id, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot uninstall bundle '%d': %s", bm.instance.IDColor(), id, response.Status())
	}
	log.Infof("%s > uninstalled bundle '%d'", bm.instance.IDColor(), id)
	return nil
}

const (
	BundlesPath     = "/system/console/bundles"
	BundlesPathJson = BundlesPath + ".json"
)
