package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"github.com/wttech/aemc/pkg/pkg"
)

type PackageManager struct {
	instance *Instance

	DeployAvoidance  bool
	SnapshotPatterns []string
}

func NewPackageManager(res *Instance) *PackageManager {
	return &PackageManager{
		instance: res,

		DeployAvoidance:  false,
		SnapshotPatterns: []string{"**/*-SNAPSHOT.zip"},
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
		return nil, fmt.Errorf("package at path '%s' does not exist on instance '%s'", remotePath, pm.instance.ID())
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
		return nil, fmt.Errorf("cannot request package list on instance '%s': %w", pm.instance.ID(), err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("cannot request package list on instance '%s': %s", pm.instance.ID(), resp.Status())
	}
	res := new(pkg.List)
	if err = fmtx.UnmarshalJSON(resp.RawBody(), res); err != nil {
		return nil, fmt.Errorf("cannot parse package list response on instance '%s': %w", pm.instance.ID(), err)
	}
	return res, nil
}

func (pm *PackageManager) Find(pid string) (*pkg.ListItem, error) {
	item, err := pm.findInternal(pid)
	if err != nil {
		return nil, fmt.Errorf("cannot find package '%s' on instance '%s': %w", pid, pm.instance.ID(), err)
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
		return nil, fmt.Errorf("cannot request package list on instance '%s': %w", pm.instance.ID(), err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("cannot request package list on instance '%s': %s", pm.instance.ID(), resp.Status())
	}
	res := new(pkg.List)
	if err = fmtx.UnmarshalJSON(resp.RawBody(), res); err != nil {
		return nil, fmt.Errorf("cannot parse package list response: %w from instance '%s'", err, pm.instance.ID())
	}
	item, ok := lo.Find(res.List, func(p pkg.ListItem) bool { return p.PID == pid })
	if ok {
		return &item, nil
	}
	return nil, nil
}

func (pm *PackageManager) IsSnapshot(localPath string) bool {
	return stringsx.MatchSomePattern(localPath, pm.SnapshotPatterns)
}

func (pm *PackageManager) Build(remotePath string) error {
	log.Infof("building package '%s' on instance '%s'", remotePath, pm.instance.ID())
	response, err := pm.instance.http.Request().Post(ServiceJsonPath + remotePath + "?cmd=build")
	if err != nil {
		return fmt.Errorf("cannot build package '%s' on instance '%s': %w", remotePath, pm.instance.ID(), err)
	} else if response.IsError() {
		return fmt.Errorf("cannot build package '%s' on instance '%s': %s", remotePath, pm.instance.ID(), response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("cannot build package '%s' on instance '%s'; cannot parse response: %w", remotePath, pm.instance.ID(), err)
	}
	if !status.Success {
		return fmt.Errorf("cannot build package '%s' on instance '%s'; unexpected status: %s", remotePath, pm.instance.ID(), status.Message)
	}
	log.Infof("built package '%s' on instance '%s'", remotePath, pm.instance.ID())
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
	log.Infof("uploading package '%s' to instance '%s'", localPath, pm.instance.ID())
	response, err := pm.instance.http.Request().
		SetFile("package", localPath).
		SetMultipartFormData(map[string]string{"force": "true"}).
		Post(ServiceJsonPath + "/?cmd=upload")
	if err != nil {
		return "", fmt.Errorf("cannot upload package '%s' on instance '%s': %w", localPath, pm.instance.ID(), err)
	} else if response.IsError() {
		return "", fmt.Errorf("cannot upload package '%s' on instance '%s': %s", localPath, pm.instance.ID(), response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return "", fmt.Errorf("cannot upload package '%s' on instance '%s'; cannot parse response: %w", localPath, pm.instance.ID(), err)
	}
	if !status.Success {
		return "", fmt.Errorf("cannot upload package '%s' on instance '%s'; unexpected status: %s", localPath, pm.instance.ID(), status.Message)
	}
	log.Infof("uploaded package '%s' to instance '%s'", localPath, pm.instance.ID())
	return status.Path, nil
}

func (pm *PackageManager) Install(remotePath string) error {
	log.Infof("installing package '%s' on instance '%s'", remotePath, pm.instance.ID())
	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "install"}).
		Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("cannot install package '%s' on instance '%s': %w", remotePath, pm.instance.ID(), err)
	} else if response.IsError() {
		return fmt.Errorf("cannot install package '%s' on instance '%s': '%s'", remotePath, pm.instance.ID(), response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("cannot install package '%s' on instance '%s'; cannot parse response: %w", remotePath, pm.instance.ID(), err)
	}
	if !status.Success {
		return fmt.Errorf("cannot install package '%s' on instance '%s'; unexpected status: %s", remotePath, pm.instance.ID(), status.Message)
	}
	log.Infof("installed package '%s' on instance '%s'", remotePath, pm.instance.ID())
	return nil
}

func (pm *PackageManager) DeployWithChanged(localPath string) (bool, error) {
	if pm.IsSnapshot(localPath) {
		return true, pm.Deploy(localPath)
	}
	p, err := pm.ByFile(localPath)
	if err != nil {
		return false, err
	}
	state, err := p.State()
	if err != nil {
		return false, err
	}
	if !state.Exists || !state.Data.Installed() { // TODO support 'DeployAvoidance'
		return true, pm.Deploy(localPath)
	}
	return false, nil
}

func (pm *PackageManager) Deploy(localPath string) error {
	remotePath, err := pm.Upload(localPath)
	if err != nil {
		return err
	}
	if err := pm.Install(remotePath); err != nil {
		return err
	}
	return nil
}

func (pm *PackageManager) Uninstall(remotePath string) error {
	log.Infof("uninstalling package '%s' on instance '%s'", remotePath, pm.instance.ID())
	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "uninstall"}).
		Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("cannot uninstall package '%s' on instance '%s': %w", remotePath, pm.instance.ID(), err)
	} else if response.IsError() {
		return fmt.Errorf("cannot uninstall package '%s' on instance '%s': %s", remotePath, pm.instance.ID(), response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("cannot uninstall package '%s' on instance '%s'; cannot parse response: %w", remotePath, pm.instance.ID(), err)
	}
	if !status.Success {
		return fmt.Errorf("cannot uninstall package '%s' on instance '%s'; unexpected status: %s", remotePath, pm.instance.ID(), status.Message)
	}
	log.Infof("uninstalled package '%s' on instance '%s'", remotePath, pm.instance.ID())
	return nil
}

func (pm *PackageManager) Delete(remotePath string) error {
	log.Infof("deleting package '%s' from instance '%s'", remotePath, pm.instance.ID())
	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "delete"}).
		Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("cannot delete package '%s' from instance '%s': %w", remotePath, pm.instance.ID(), err)
	} else if response.IsError() {
		return fmt.Errorf("cannot delete package '%s' from instance '%s': %s", remotePath, pm.instance.ID(), response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("cannot delete package '%s' from instance '%s'; cannot parse response: %w", remotePath, pm.instance.ID(), err)
	}
	if !status.Success {
		return fmt.Errorf("cannot delete package '%s' from instance '%s'; unexpected status: %s", remotePath, pm.instance.ID(), status.Message)
	}
	log.Infof("deleted package '%s' from instance '%s'", remotePath, pm.instance.ID())
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
