package pkg

import (
	"bytes"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/timex"
	"github.com/wttech/aemc/pkg/java"
	"os"
	"strings"
	"time"
)

const (
	UnpackDir   = common.VarDir + "/instance"
	BackupDir   = common.VarDir + "/backup"
	OverrideDir = common.DefaultDir + "/" + common.VarDirName + "/instance"

	DistFile        = common.LibDir + "/aem-sdk-quickstart.jar"
	LicenseFile     = common.LibDir + "/" + LicenseFilename
	LicenseFilename = "license.properties"
)

type LocalOpts struct {
	manager *InstanceManager

	UnpackDir   string
	BackupDir   string
	OverrideDir string
	ToolDir     string
	JavaOpts    *java.Opts
	OakRun      *OakRun
	Quickstart  *Quickstart
	SDK         *SDK
}

func (im *InstanceManager) NewLocalOpts(manager *InstanceManager) *LocalOpts {
	result := &LocalOpts{
		manager: manager,

		UnpackDir:   UnpackDir,
		BackupDir:   BackupDir,
		OverrideDir: OverrideDir,
		ToolDir:     common.ToolDir,
		JavaOpts:    im.aem.javaOpts,
		Quickstart:  NewQuickstart(),
	}
	result.SDK = NewSDK(result)
	result.OakRun = NewOakRun(result)
	return result
}

func (o *LocalOpts) Initialize() error {
	// validate phase (fast feedback)
	if err := o.validateUnpackDir(); err != nil {
		return err
	}
	sdk, err := o.Quickstart.IsDistSDK()
	if err != nil {
		return err
	}
	if !sdk {
		_, err = o.Quickstart.FindLicenseFile()
		if err != nil {
			return err
		}
	}
	for _, instance := range o.manager.Locals() {
		if err := instance.Local().Validate(); err != nil {
			return err
		}
	}
	// preparation phase
	if err := o.manager.aem.baseOpts.Prepare(); err != nil {
		return err
	}
	if err := o.JavaOpts.Prepare(); err != nil {
		return err
	}
	if sdk {
		if err := o.SDK.Prepare(); err != nil {
			return err
		}
	}
	if err := o.OakRun.Prepare(); err != nil {
		return err
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

func (o *LocalOpts) Jar() (string, error) {
	sdk, err := o.Quickstart.IsDistSDK()
	if err != nil {
		return "", err
	}
	if sdk {
		return o.SDK.QuickstartJar()
	}
	return o.Quickstart.FindDistFile()
}

func NewQuickstart() *Quickstart {
	return &Quickstart{
		DistFile:    DistFile,
		LicenseFile: LicenseFile,
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
	log.Infof("creating instance(s) '%s'", InstanceIds(instances))
	for _, i := range instances {
		if !i.local.IsCreated() {
			err := i.local.Create()
			if err != nil {
				return nil, fmt.Errorf("cannot create instance '%s': %s", i.ID(), err)
			}
			created = append(created, i)
		}
	}
	return created, nil
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

	log.Infof("checking started & out-of-date instance(s) '%s'", InstanceIds(instances))

	var outdated []Instance
	for _, i := range instances {
		if i.local.IsRunning() && i.local.OutOfDate() {
			outdated = append(outdated, i)

			log.Infof("instance '%s' is already started but out-of-date", i.ID())
			err := i.local.Stop()
			if err != nil {
				return nil, fmt.Errorf("cannot stop out-of-date instance '%s': %s", i.ID(), err)
			}
		}
	}

	if err := im.AwaitStopped(outdated); err != nil {
		return outdated, err
	}

	log.Infof("starting instance(s) '%s'", InstanceIds(instances))

	started := []Instance{}
	for _, i := range instances {
		if !i.local.IsRunning() {
			err := i.local.Start()
			if err != nil {
				return nil, fmt.Errorf("cannot start instance '%s': %s", i.ID(), err)
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
	log.Infof("stopping instance(s) '%s'", InstanceIds(instances))
	stopped := []Instance{}
	for _, i := range instances {
		if i.local.IsRunning() {
			err := i.local.Stop()
			if err != nil {
				return nil, fmt.Errorf("cannot stop instance '%s': %s", i.ID(), err)
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
	log.Infof("killing instance(s) '%s'", InstanceIds(instances))

	killed := []Instance{}
	for _, i := range instances {
		if i.local.IsKillable() {
			err := i.local.Kill()
			if err != nil {
				log.Warnf("cannot kill instance '%s' (process not running / already killed): %s", i.ID(), err)
			} else {
				log.Infof("killed instance '%s'", i.ID())
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
	log.Infof("deleting instance(s) '%s'", InstanceIds(instances))
	deleted := []Instance{}
	for _, i := range instances {
		if i.local.IsCreated() {
			err := i.local.Delete()
			if err != nil {
				return nil, fmt.Errorf("cannot delete instance '%s': %s", i.ID(), err)
			}
			log.Infof("deleted instance '%s'", i.ID())
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
	log.Infof("cleaning instance(s) '%s'", InstanceIds(instances))
	cleaned := []Instance{}
	for _, i := range instances {
		if !i.local.IsRunning() {
			err := i.local.Clean()
			if err != nil {
				return nil, fmt.Errorf("cannot clean instance '%s': %s", i.ID(), err)
			}
			log.Infof("cleaned instance '%s'", i.ID())
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

func (im *InstanceManager) configureLocalOpts(config *cfg.Config) {
	opts := config.Values().Instance.Local

	if len(opts.UnpackDir) > 0 {
		im.LocalOpts.UnpackDir = opts.UnpackDir
	}
	if len(opts.BackupDir) > 0 {
		im.LocalOpts.BackupDir = opts.BackupDir
	}
	if len(opts.OverrideDir) > 0 {
		im.LocalOpts.OverrideDir = opts.OverrideDir
	}
	if len(opts.ToolDir) > 0 {
		im.LocalOpts.ToolDir = opts.ToolDir
	}
	if len(opts.Quickstart.DistFile) > 0 {
		im.LocalOpts.Quickstart.DistFile = opts.Quickstart.DistFile
	}
	if len(opts.Quickstart.LicenseFile) > 0 {
		im.LocalOpts.Quickstart.LicenseFile = opts.Quickstart.LicenseFile
	}
	if len(opts.OakRun.DownloadURL) > 0 {
		im.LocalOpts.OakRun.DownloadURL = opts.OakRun.DownloadURL
	}
	if len(opts.OakRun.StorePath) > 0 {
		im.LocalOpts.OakRun.StorePath = opts.OakRun.StorePath
	}
}
