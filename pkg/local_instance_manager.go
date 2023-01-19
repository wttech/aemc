package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/java"
	"os"
	"strings"
)

const (
	UnpackDir       = "aem/home/data/instance"
	BackupDir       = "aem/home/data/backup"
	LibDir          = "aem/home/lib"
	DistFile        = LibDir + "/aem-sdk-quickstart.jar"
	LicenseFile     = LibDir + "/" + LicenseFilename
	LicenseFilename = "license.properties"
)

type LocalOpts struct {
	manager *InstanceManager

	UnpackDir  string
	BackupDir  string
	JavaOpts   *java.Opts
	Quickstart *Quickstart
	Sdk        *Sdk
}

func (im *InstanceManager) NewLocalOpts(manager *InstanceManager) *LocalOpts {
	result := &LocalOpts{
		manager: manager,

		UnpackDir: UnpackDir,
		BackupDir: BackupDir,
		JavaOpts:  im.aem.javaOpts,
		Quickstart: &Quickstart{
			DistFile:    DistFile,
			LicenseFile: LicenseFile,
		},
	}
	result.Sdk = &Sdk{
		localOpts: result,
	}
	return result
}

func (o *LocalOpts) Validate() error {
	if err := o.validateUnpackDir(); err != nil {
		return err
	}
	if err := o.JavaOpts.Validate(); err != nil {
		return err
	}
	if err := o.Quickstart.Validate(); err != nil {
		return err
	}
	return nil
}

func (o *LocalOpts) validateUnpackDir() error {
	current := pathx.Abs(o.UnpackDir)
	if strings.Contains(current, " ") {
		return fmt.Errorf("local instance unpack dir '%s' cannot contain spaces (as shell scripts could run improperly)", current)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil // intentionally
	}
	deniedDirs := lo.Map([]string{homeDir + "/Desktop", homeDir + "/Documents"}, func(p string, _ int) string { return pathx.Abs(p) })
	if lo.SomeBy(deniedDirs, func(d string) bool { return strings.HasPrefix(current, d) }) {
		return fmt.Errorf("local instance unpack dir '%s' cannot be located under dirs: %s", current, strings.Join(deniedDirs, ", "))
	}
	return nil
}

func (o *LocalOpts) Jar() (string, error) {
	distFile, err := o.Quickstart.FindDistFile()
	if err != nil {
		return "", err
	}
	if IsSdkFile(distFile) {
		return o.Sdk.QuickstartJar()
	}
	return o.Quickstart.FindDistFile()
}

func IsSdkFile(path string) bool {
	return pathx.Ext(path) == "zip"
}

type Quickstart struct {
	DistFile    string
	LicenseFile string
}

func (o *Quickstart) Validate() error {
	_, err := o.FindDistFile()
	if err != nil {
		return err
	}
	_, err = o.FindLicenseFile()
	if err != nil {
		return err
	}
	return nil
}

func (o *Quickstart) FindDistFile() (string, error) {
	return pathx.GlobOne(o.DistFile)
}

func (o *Quickstart) FindLicenseFile() (string, error) {
	return pathx.GlobOne(o.LicenseFile)
}

// LocalValidate checks prerequisites needed to manage local instances
func (im *InstanceManager) LocalValidate() error {
	err := im.LocalOpts.Validate()
	if err != nil {
		return err
	}
	return nil
}

func (im *InstanceManager) CreateAll() ([]Instance, error) {
	return im.Create(im.Locals())
}

func (im *InstanceManager) Create(instances []Instance) ([]Instance, error) {
	err := im.LocalValidate()
	if err != nil {
		return nil, err
	}

	created := []Instance{}

	distFile, err := im.LocalOpts.Quickstart.FindDistFile()
	if err != nil {
		return created, err
	}
	if IsSdkFile(distFile) {
		err := im.LocalOpts.Sdk.Prepare(distFile)
		if err != nil {
			return created, err
		}
	}
	for _, i := range instances {
		if !i.local.IsCreated() {
			log.Infof("creating instance '%s'", i.ID())
			err := i.local.Create()
			if err != nil {
				return nil, fmt.Errorf("cannot create instance '%s': %s", i.ID(), err)
			}
			log.Infof("created instance '%s'", i.ID())
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
	err := im.LocalValidate()
	if err != nil {
		return nil, err
	}

	log.Infof("checking started & out-of-date instance(s)")

	var outdated []Instance
	for _, i := range instances {
		if i.local.IsRunning() && i.local.OutOfDate() {
			outdated = append(outdated, i)

			log.Infof("instance '%s' is already started but out-of-date - stopping", i.ID())
			err := i.local.Stop()
			if err != nil {
				return nil, fmt.Errorf("cannot stop out-of-date instance '%s': %s", i.ID(), err)
			}
		}
	}

	if err := im.AwaitStopped(outdated); err != nil {
		return outdated, err
	}

	log.Infof("starting instance(s)")

	started := []Instance{}
	for _, i := range instances {
		if !i.local.IsRunning() {
			err := i.local.Start()
			if err != nil {
				return nil, fmt.Errorf("cannot start instance '%s': %s", i.ID(), err)
			}
			log.Infof("started instance '%s'", i.ID())
			started = append(started, i)
		}
	}

	if im.CheckOpts.AwaitStrict {
		if err := im.AwaitStarted(started); err != nil {
			return started, err
		}
	} else {
		if err := im.AwaitStarted(instances); err != nil {
			return instances, err
		}
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
	err := im.LocalValidate()
	if err != nil {
		return nil, err
	}

	log.Info("stopping instance(s)")

	stopped := []Instance{}
	for _, i := range instances {
		if i.local.IsRunning() {
			err := i.local.Stop()
			if err != nil {
				return nil, fmt.Errorf("cannot stop instance '%s': %s", i.ID(), err)
			}
			log.Infof("stopped instance '%s'", i.ID())
			stopped = append(stopped, i)
		}
	}

	if im.CheckOpts.AwaitStrict {
		if err := im.AwaitStopped(stopped); err != nil {
			return stopped, err
		}
		im.Clean(stopped)
	} else {
		if err := im.AwaitStopped(instances); err != nil {
			return instances, err
		}
		im.Clean(instances)
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
	err := im.LocalValidate()
	if err != nil {
		return nil, err
	}

	log.Info("killing instance(s)")

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
	// im.LocalValidate()

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

func (im *InstanceManager) AwaitStartedOne(instance Instance) error {
	return im.AwaitStarted([]Instance{instance})
}

func (im *InstanceManager) AwaitStartedAll() error {
	return im.AwaitStarted(im.All())
}

func (im *InstanceManager) AwaitStarted(instances []Instance) error {
	if len(instances) == 0 {
		return nil
	}
	log.Infof("awaiting up instance(s) '%s'", InstanceIds(instances))
	return im.Check(instances, im.CheckOpts, []Checker{
		im.CheckOpts.AwaitUpTimeout,
		im.CheckOpts.Reachable,
		im.CheckOpts.EventStable,
		im.CheckOpts.BundleStable,
		im.CheckOpts.Installer,
	})
}

func (im *InstanceManager) AwaitStoppedOne(instance Instance) error {
	return im.AwaitStopped([]Instance{instance})
}

func (im *InstanceManager) AwaitStoppedAll() error {
	return im.AwaitStopped(im.Locals())
}

func (im *InstanceManager) AwaitStopped(instances []Instance) error {
	if len(instances) == 0 {
		return nil
	}
	log.Infof("awaiting down instance(s) '%s'", InstanceIds(instances))
	return im.Check(instances, im.CheckOpts, []Checker{
		im.CheckOpts.AwaitDownTimeout,
		im.CheckOpts.StatusStopped,
		im.CheckOpts.Unreachable,
	})
}

func (im *InstanceManager) Clean(instances []Instance) ([]Instance, error) {
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

func (im *InstanceManager) configureLocalOpts(config *cfg.Config) {
	opts := config.Values().Instance.Local

	if len(opts.UnpackDir) > 0 {
		im.LocalOpts.UnpackDir = opts.UnpackDir
	}
	if len(opts.BackupDir) > 0 {
		im.LocalOpts.BackupDir = opts.BackupDir
	}
	if len(opts.Quickstart.DistFile) > 0 {
		im.LocalOpts.Quickstart.DistFile = opts.Quickstart.DistFile
	}
	if len(opts.Quickstart.LicenseFile) > 0 {
		im.LocalOpts.Quickstart.LicenseFile = opts.Quickstart.LicenseFile
	}
}
