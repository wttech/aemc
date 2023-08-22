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
	"github.com/wttech/aemc/pkg/pkg"
	"path/filepath"
	"time"
)

type PackageManager struct {
	instance *Instance

	SnapshotDeploySkipping bool
	InstallRecursive       bool
	SnapshotPatterns       []string
	ToggledWorkflows       []string
}

func NewPackageManager(res *Instance) *PackageManager {
	cv := res.manager.aem.config.Values()

	return &PackageManager{
		instance: res,

		SnapshotDeploySkipping: cv.GetBool("instance.package.snapshot_deploy_skipping"),
		InstallRecursive:       cv.GetBool("instance.package.install_recursive"),
		SnapshotPatterns:       cv.GetStringSlice("instance.package.snapshot_patterns"),
		ToggledWorkflows:       cv.GetStringSlice("instance.package.toggled_workflows"),
	}
}

func (pm *PackageManager) ByPID(pid string) (*Package, error) {
	pidConfig, err := pkg.ParsePID(pid)
	if err != nil {
		return nil, err
	}
	return pm.byPID(*pidConfig)
}

func (pm *PackageManager) ByFile(localPath string) (*Package, error) {
	pidConfig, err := pkg.ReadPIDFromZIP(localPath)
	if err != nil {
		return nil, err
	}
	return pm.byPID(*pidConfig)
}

func (pm *PackageManager) ByPath(remotePath string) (*Package, error) {
	list, err := pm.List()
	if err != nil {
		return nil, err
	}
	item, ok := lo.Find(list.List, func(item pkg.ListItem) bool { return item.Path == remotePath })
	if !ok {
		return nil, fmt.Errorf("%s > package at path '%s' does not exist", pm.instance.ID(), remotePath)
	}
	pid, err := pkg.ParsePID(item.PID)
	if err != nil {
		return nil, err
	}
	return pm.byPID(*pid)
}

func (pm *PackageManager) byPID(pidConfig pkg.PID) (*Package, error) {
	return &Package{manager: pm, PID: pidConfig}, nil
}

func (pm *PackageManager) List() (*pkg.List, error) {
	resp, err := pm.instance.http.Request().Get(ListJson)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot request package list: %w", pm.instance.ID(), err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot request package list: %s", pm.instance.ID(), resp.Status())
	}
	res := new(pkg.List)
	if err = fmtx.UnmarshalJSON(resp.RawBody(), res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse package list response: %w", pm.instance.ID(), err)
	}
	return res, nil
}

func (pm *PackageManager) Find(pid string) (*pkg.ListItem, error) {
	item, err := pm.findInternal(pid)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot find package '%s': %w", pm.instance.ID(), pid, err)
	}
	return item, nil
}

func (pm *PackageManager) findInternal(pid string) (*pkg.ListItem, error) {
	pidConfig, err := pkg.ParsePID(pid)
	if err != nil {
		return nil, err
	}
	resp, err := pm.instance.http.Request().SetQueryParam("name", pidConfig.Name).Get(ListJson)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot request package list: %w", pm.instance.ID(), err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot request package list: %s", pm.instance.ID(), resp.Status())
	}
	res := new(pkg.List)
	if err = fmtx.UnmarshalJSON(resp.RawBody(), res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse package list response: %w", pm.instance.ID(), err)
	}
	item, ok := lo.Find(res.List, func(p pkg.ListItem) bool { return p.PID == pid })
	if ok {
		return &item, nil
	}
	return nil, nil
}

func (pm *PackageManager) IsSnapshot(localPath string) bool {
	return stringsx.MatchSome(pathx.Normalize(localPath), pm.SnapshotPatterns)
}

func (pm *PackageManager) Build(remotePath string) error {
	log.Infof("%s > building package '%s'", pm.instance.ID(), remotePath)
	response, err := pm.instance.http.Request().Post(ServiceJsonPath + remotePath + "?cmd=build")
	if err != nil {
		return fmt.Errorf("%s > cannot build package '%s': %w", pm.instance.ID(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot build package '%s': %s", pm.instance.ID(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot build package '%s'; cannot parse response: %w", pm.instance.ID(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot build package '%s'; unexpected status: %s", pm.instance.ID(), remotePath, status.Message)
	}
	log.Infof("%s > built package '%s'", pm.instance.ID(), remotePath)
	return nil
}

func (pm *PackageManager) UploadWithChanged(localPath string) (bool, error) {
	if pm.IsSnapshot(localPath) {
		_, err := pm.Upload(localPath)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	p, err := pm.ByFile(localPath)
	if err != nil {
		return false, err
	}
	state, err := p.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		_, err = pm.Upload(localPath)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (pm *PackageManager) Upload(localPath string) (string, error) {
	log.Infof("%s > uploading package '%s'", pm.instance.ID(), localPath)
	response, err := pm.instance.http.Request().
		SetFile("package", localPath).
		SetMultipartFormData(map[string]string{"force": "true"}).
		Post(ServiceJsonPath + "/?cmd=upload")
	if err != nil {
		return "", fmt.Errorf("%s > cannot upload package '%s': %w", pm.instance.ID(), localPath, err)
	} else if response.IsError() {
		return "", fmt.Errorf("%s > cannot upload package '%s': %s", pm.instance.ID(), localPath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return "", fmt.Errorf("%s > cannot upload package '%s'; cannot parse response: %w", pm.instance.ID(), localPath, err)
	}
	if !status.Success {
		return "", fmt.Errorf("%s > cannot upload package '%s'; unexpected status: %s", pm.instance.ID(), localPath, status.Message)
	}
	log.Infof("%s > uploaded package '%s'", pm.instance.ID(), localPath)
	return status.Path, nil
}

func (pm *PackageManager) Install(remotePath string) error {
	log.Infof("%s > installing package '%s'", pm.instance.ID(), remotePath)
	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "install", "recursive": fmt.Sprintf("%v", pm.InstallRecursive)}).
		Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("%s > cannot install package '%s': %w", pm.instance.ID(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot install package '%s': '%s'", pm.instance.ID(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot install package '%s'; cannot parse response: %w", pm.instance.ID(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot install package '%s'; unexpected status: %s", pm.instance.ID(), remotePath, status.Message)
	}
	log.Infof("%s > installed package '%s'", pm.instance.ID(), remotePath)
	return nil
}

func (pm *PackageManager) DeployWithChanged(localPath string) (bool, error) {
	if pm.instance.IsLocal() && pm.IsSnapshot(localPath) { // TODO remove local check; support remote as well
		return pm.deploySnapshot(localPath)
	}
	return pm.deployRegular(localPath)
}

func (pm *PackageManager) deployRegular(localPath string) (bool, error) {
	deployed, err := pm.IsDeployed(localPath)
	if err != nil {
		return false, err
	}
	if !deployed {
		return true, pm.Deploy(localPath)
	}
	return false, nil
}

func (pm *PackageManager) deploySnapshot(localPath string) (bool, error) {
	checksum, err := filex.ChecksumFile(localPath)
	if err != nil {
		return false, err
	}
	deployed, err := pm.IsDeployed(localPath)
	if err != nil {
		return false, err
	}
	var lock = pm.deployLock(localPath, checksum)
	if deployed && pm.SnapshotDeploySkipping && lock.IsLocked() {
		lockData, err := lock.Locked()
		if err != nil {
			return false, err
		}
		if checksum == lockData.Checksum {
			log.Infof("%s > skipped deploying package '%s'", pm.instance.ID(), localPath)
			return false, nil
		}
	}
	if err := pm.Deploy(localPath); err != nil {
		return false, err
	}
	if err := lock.Lock(); err != nil {
		return false, err
	}
	return true, nil
}

func (pm *PackageManager) IsDeployed(localPath string) (bool, error) {
	p, err := pm.ByFile(localPath)
	if err != nil {
		return false, err
	}
	state, err := p.State()
	if err != nil {
		return false, err
	}
	return state.Exists && state.Data.Installed(), nil
}

func (pm *PackageManager) Deploy(localPath string) error {
	remotePath, err := pm.Upload(localPath)
	if err != nil {
		return err
	}
	return pm.instance.workflowManager.ToggleLaunchers(pm.ToggledWorkflows, func() error {
		return pm.Install(remotePath)
	})
}

func (pm *PackageManager) deployLock(file string, checksum string) osx.Lock[packageDeployLock] {
	name := filepath.Base(file)
	return osx.NewLock(fmt.Sprintf("%s/package/deploy/%s.yml", pm.instance.LockDir(), name), func() (packageDeployLock, error) {
		return packageDeployLock{Deployed: time.Now(), Checksum: checksum}, nil
	})
}

type packageDeployLock struct {
	Deployed time.Time `yaml:"deployed"`
	Checksum string    `yaml:"checksum"`
}

func (pm *PackageManager) Uninstall(remotePath string) error {
	log.Infof("%s > uninstalling package '%s'", pm.instance.ID(), remotePath)
	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "uninstall"}).
		Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("%s > cannot uninstall package '%s': %w", pm.instance.ID(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot uninstall package '%s': %s", pm.instance.ID(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot uninstall package '%s'; cannot parse response: %w", pm.instance.ID(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot uninstall package '%s'; unexpected status: %s", pm.instance.ID(), remotePath, status.Message)
	}
	log.Infof("%s > uninstalled package '%s'", pm.instance.ID(), remotePath)
	return nil
}

func (pm *PackageManager) Delete(remotePath string) error {
	log.Infof("%s > deleting package '%s'", pm.instance.ID(), remotePath)
	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "delete"}).
		Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("%s > cannot delete package '%s': %w", pm.instance.ID(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot delete package '%s': %s", pm.instance.ID(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot delete package '%s'; cannot parse response: %w", pm.instance.ID(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot delete package '%s'; unexpected status: %s", pm.instance.ID(), remotePath, status.Message)
	}
	log.Infof("%s > deleted package '%s'", pm.instance.ID(), remotePath)
	return nil
}

const (
	MgrPath         = "/crx/packmgr"
	ServicePath     = MgrPath + "/service"
	ServiceJsonPath = ServicePath + "/.json"
	ServiceHtmlPath = ServicePath + "/.html"
	ListJson        = MgrPath + "/list.jsp"
	IndexPath       = MgrPath + "/index.jsp"
)
