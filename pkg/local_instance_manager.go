package pkg

import (
	"bytes"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/timex"
	"os"
	"strings"
	"time"
)

type LocalOpts struct {
	manager *InstanceManager

	UnpackDir   string
	BackupDir   string
	OverrideDir string
	ServiceMode bool
}

func NewLocalOpts(manager *InstanceManager) *LocalOpts {
	cfg := manager.aem.config.Values()

	result := &LocalOpts{manager: manager}
	result.UnpackDir = cfg.GetString("instance.local.unpack_dir")
	result.BackupDir = cfg.GetString("instance.local.backup_dir")
	result.OverrideDir = cfg.GetString("instance.local.override_dir")
	result.ServiceMode = cfg.GetBool("instance.local.service_mode")

	return result
}

func (o *LocalOpts) Initialize() error {
	// pre-validation phase (fast feedback)
	if err := o.validateUnpackDir(); err != nil {
		return err
	}
	sdk, err := o.manager.aem.vendorManager.quickstart.IsDistSDK()
	if err != nil {
		return err
	}
	if !sdk {
		_, err = o.manager.aem.vendorManager.quickstart.FindLicenseFile()
		if err != nil {
			return err
		}
	}
	for _, instance := range o.manager.Locals() {
		if err := instance.Local().CheckPassword(); err != nil {
			return err
		}
	}
	// initialization phase (heavy work)
	if _, err := o.manager.aem.baseOpts.PrepareWithChanged(); err != nil {
		return err
	}
	if _, err := o.manager.aem.vendorManager.javaManager.PrepareWithChanged(); err != nil {
		return err
	}
	if sdk {
		if _, err := o.manager.aem.vendorManager.sdk.PrepareWithChanged(); err != nil {
			return err
		}
	}
	if _, err := o.manager.aem.vendorManager.oakRun.PrepareWithChanged(); err != nil {
		return err
	}
	// post-validation phase
	for _, instance := range o.manager.Locals() {
		if err := instance.Local().CheckRecreationNeeded(); err != nil { // depends on SDK prepare
			return err
		}
	}
	return nil
}

func (o *LocalOpts) validateUnpackDir() error {
	current := pathx.Canonical(o.UnpackDir)
	if strings.Contains(current, " ") {
		return fmt.Errorf("local instance unpack dir '%s' cannot contain spaces (as shell scripts could run improperly)", current)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil // intentionally
	}
	deniedDirs := lo.Map([]string{homeDir + "/Desktop", homeDir + "/Documents"}, func(p string, _ int) string { return pathx.Canonical(p) })
	if lo.SomeBy(deniedDirs, func(d string) bool { return strings.HasPrefix(current, d) }) {
		return fmt.Errorf("local instance unpack dir '%s' cannot be located under dirs: %s", current, strings.Join(deniedDirs, ", "))
	}
	return nil
}

func NewQuickstart(manager *VendorManager) *Quickstart {
	cfg := manager.aem.config.Values()

	return &Quickstart{
		DistFile:    cfg.GetString("vendor.quickstart.dist_file"),
		LicenseFile: cfg.GetString("vendor.quickstart.license_file"),
	}
}

type Quickstart struct {
	DistFile    string
	LicenseFile string
}

func (o *Quickstart) FindDistFile() (string, error) {
	return pathx.GlobSome(o.DistFile)
}

func (o *Quickstart) FindLicenseFile() (string, error) {
	return pathx.GlobSome(o.LicenseFile)
}

func (o *Quickstart) IsDistSDK() (bool, error) {
	file, err := o.FindDistFile()
	if err != nil {
		return false, err
	}
	return pathx.Ext(file) == "zip", nil
}

func (im *InstanceManager) CreateAll() ([]Instance, error) {
	return im.Create(im.Locals())
}

func (im *InstanceManager) Create(instances []Instance) ([]Instance, error) {
	created := []Instance{}
	if err := im.LocalOpts.Initialize(); err != nil {
		return created, err
	}
	log.Info(InstancesMsg(instances, "creating"))
	for _, i := range instances {
		if !i.local.IsCreated() {
			err := i.local.Create()
			if err != nil {
				return nil, err
			}
			created = append(created, i)
		}
	}
	return created, nil
}

func (im *InstanceManager) Import(instances []Instance) ([]Instance, error) {
	imported := []Instance{}
	log.Info(InstancesMsg(instances, "importing"))

	for _, i := range instances {
		if !i.local.IsCreated() {
			err := i.local.Import()
			if err != nil {
				return nil, err
			}
			imported = append(imported, i)
		}
	}

	running := []Instance{}
	for _, i := range imported {
		isRunning, err := i.local.IsRunningStrict()

		if err != nil {
			return nil, fmt.Errorf("%s > cannot check status of imported instance: %w", i.IDColor(), err)
		}

		if isRunning {
			running = append(running, i)
		}
	}

	if len(running) > 0 {
		log.Info(InstancesMsg(running, " imported but already running - restarting to apply configuration"))

		for _, i := range running {
			err := i.local.Restart()
			if err != nil {
				return nil, err
			}
		}

		if err := im.AwaitStarted(running); err != nil {
			return imported, err
		}
	}

	return imported, nil
}

func (im *InstanceManager) StartOne(instance Instance) (bool, error) {
	started, err := im.Start([]Instance{instance})
	return len(started) > 0, err
}

func (im *InstanceManager) StartAll() ([]Instance, error) {
	return im.Start(im.Locals())
}

func (im *InstanceManager) Start(instances []Instance) ([]Instance, error) {
	if len(instances) == 0 {
		log.Debugf("no instances to start")
		return []Instance{}, nil
	}
	if err := im.LocalOpts.Initialize(); err != nil {
		return []Instance{}, err
	}

	if !im.LocalOpts.ServiceMode {
		log.Info(InstancesMsg(instances, "checking started & out-of-date"))

		var outdated []Instance
		for _, i := range instances {
			if i.local.IsRunning() && i.local.OutOfDate() {
				outdated = append(outdated, i)

				log.Infof("%s > already started but out-of-date", i.IDColor())
				err := i.local.Stop()
				if err != nil {
					return nil, err
				}
			}
		}

		if err := im.AwaitStopped(outdated); err != nil {
			return outdated, err
		}
	}

	log.Info(InstancesMsg(instances, "starting"))

	started := []Instance{}
	for _, i := range instances {
		if !i.local.IsRunning() {
			err := i.local.Start()
			if err != nil {
				return nil, err
			}
			started = append(started, i)
		}
	}

	var awaited []Instance
	if im.CheckOpts.AwaitStrict {
		awaited = started
	} else {
		awaited = instances
	}
	if err := im.AwaitStarted(awaited); err != nil {
		return nil, err
	}

	return started, nil
}

func (im *InstanceManager) StopOne(instance Instance) (bool, error) {
	stopped, err := im.Stop([]Instance{instance})
	return len(stopped) > 0, err
}

func (im *InstanceManager) StopAll() ([]Instance, error) {
	return im.Stop(im.Locals())
}

func (im *InstanceManager) Stop(instances []Instance) ([]Instance, error) {
	if len(instances) == 0 {
		log.Debugf("no instances to stop")
		return []Instance{}, nil
	}
	log.Info(InstancesMsg(instances, "stopping"))
	stopped := []Instance{}
	for _, i := range instances {
		if i.local.IsRunning() {
			err := i.local.Stop()
			if err != nil {
				return nil, err
			}
			stopped = append(stopped, i)
		}
	}
	var awaited []Instance
	if im.CheckOpts.AwaitStrict {
		awaited = stopped
	} else {
		awaited = instances
	}
	if err := im.AwaitStopped(awaited); err != nil {
		return nil, err
	}
	_, err := im.Clean(stopped)
	if err != nil {
		return nil, err
	}
	return stopped, nil
}

func (im *InstanceManager) KillOne(instance Instance) (bool, error) {
	killed, err := im.Kill([]Instance{instance})
	return len(killed) > 0, err
}

func (im *InstanceManager) KillAll() ([]Instance, error) {
	return im.Kill(im.Locals())
}

func (im *InstanceManager) Kill(instances []Instance) ([]Instance, error) {
	log.Info(InstancesMsg(instances, "killing"))

	killed := []Instance{}
	for _, i := range instances {
		if i.local.IsKillable() {
			err := i.local.Kill()
			if err != nil {
				log.Warnf("%s > cannot kill as process not running or is already killed: %s", i.IDColor(), err)
			} else {
				killed = append(killed, i)
			}
		}
	}

	return killed, nil
}

func (im *InstanceManager) DeleteOne(instance Instance) (bool, error) {
	deleted, err := im.Delete([]Instance{instance})
	return len(deleted) > 0, err
}

func (im *InstanceManager) DeleteAll() ([]Instance, error) {
	return im.Delete(im.Locals())
}

func (im *InstanceManager) Delete(instances []Instance) ([]Instance, error) {
	if len(instances) == 0 {
		log.Debugf("no instances to delete")
		return []Instance{}, nil
	}
	log.Info(InstancesMsg(instances, "deleting"))
	deleted := []Instance{}
	for _, i := range instances {
		if i.local.IsCreated() {
			err := i.local.Delete()
			if err != nil {
				return nil, err
			}
			deleted = append(deleted, i)
		}
	}
	return deleted, nil
}

func (im *InstanceManager) Clean(instances []Instance) ([]Instance, error) {
	if len(instances) == 0 {
		log.Debugf("no instances to clean")
		return []Instance{}, nil
	}
	log.Info(InstancesMsg(instances, "cleaning"))
	cleaned := []Instance{}
	for _, i := range instances {
		if !i.local.IsRunning() {
			err := i.local.Clean()
			if err != nil {
				return nil, err
			}
			cleaned = append(cleaned, i)
		}
	}
	return cleaned, nil
}

func (im *InstanceManager) ListBackups() (*BackupList, error) {
	dir := im.LocalOpts.BackupDir
	files, err := pathx.GlobDir(dir, "*."+LocalInstanceBackupExtension)
	if err != nil {
		return nil, fmt.Errorf("cannot list instance backups in directory '%s': %w", dir, err)
	}
	return newBackupList(files), nil
}

func newBackupList(files []string) *BackupList {
	items := lo.Map(files, func(file string, _ int) BackupFile {
		stat, _ := os.Stat(file)
		return BackupFile{
			Path:     file,
			Size:     uint64(stat.Size()),
			Modified: stat.ModTime(),
		}
	})
	return &BackupList{
		Total: len(items),
		Files: items,
	}
}

type BackupList struct {
	Total int          `json:"total" yaml:"total"`
	Files []BackupFile `json:"files" yaml:"files"`
}

type BackupFile struct {
	Path     string    `json:"path" yaml:"path"`
	Size     uint64    `json:"size" yaml:"size"`
	Modified time.Time `json:"modified" yaml:"modified"`
}

func (fl BackupList) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblMap("stats", "stat", "value", map[string]any{
		"total": len(fl.Files),
		"size":  humanize.Bytes(lo.SumBy(fl.Files, func(file BackupFile) uint64 { return file.Size })),
	}))
	bs.WriteString("\n")
	bs.WriteString(fmtx.TblRows("list", false, []string{"path", "size", "modified"}, lo.Map(fl.Files, func(file BackupFile, _ int) map[string]any {
		return map[string]any{
			"path":     file.Path,
			"size":     humanize.Bytes(file.Size),
			"modified": timex.Human(file.Modified),
		}
	})))
	return bs.String()
}
