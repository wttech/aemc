package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/java"
	"os"
	"strings"
	"time"
)

const (
	UnpackDir       = "aem/home/instance"
	LibDir          = "aem/home/lib"
	DistFile        = LibDir + "/aem-sdk-quickstart.jar"
	LicenseFile     = LibDir + "/" + LicenseFilename
	LicenseFilename = "license.properties"
)

type LocalOpts struct {
	manager *InstanceManager

	UnpackDir  string
	JavaOpts   *java.Opts
	Quickstart *Quickstart
	Sdk        *Sdk
}

func (im InstanceManager) NewLocalOpts(manager *InstanceManager) *LocalOpts {
	result := &LocalOpts{
		manager: manager,

		UnpackDir: UnpackDir,
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
	current := osx.PathAbs(o.UnpackDir)
	if strings.Contains(current, " ") {
		return fmt.Errorf("local instance unpack dir '%s' cannot contain spaces (as shell scripts could run improperly)", current)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil // intentionally
	}
	deniedDirs := lo.Map([]string{homeDir + "/Desktop", homeDir + "/Documents"}, func(p string, _ int) string { return osx.PathAbs(p) })
	if lo.SomeBy(deniedDirs, func(d string) bool { return strings.HasPrefix(current, d) }) {
		return fmt.Errorf("local instance unpack dir '%s' cannot be located under dirs: %s", current, strings.Join(deniedDirs, ", "))
	}
	return nil
}

func (o *LocalOpts) Jar() (string, error) {
	if o.Quickstart.IsDistFileSdk() {
		return o.Sdk.QuickstartJar()
	}
	return o.Quickstart.DistFile, nil
}

type Quickstart struct {
	DistFile    string
	LicenseFile string
}

func (o *Quickstart) Validate() error {
	if !osx.PathExists(o.DistFile) {
		return fmt.Errorf("quickstart dist file does not exist at path '%s'; consider specifying it by property 'instance.local.quickstart.dist_file'", o.DistFile)
	}
	if !osx.PathExists(o.LicenseFile) {
		return fmt.Errorf("quickstart license file does not exist at path '%s'; consider specifying it by property 'instance.local.quickstart.license_file'", o.LicenseFile)
	}
	return nil
}

func (o *Quickstart) IsDistFileSdk() bool {
	return osx.FileExt(o.DistFile) == "zip"
}

// LocalValidate checks prerequisites needed to manage local instances
func (im InstanceManager) LocalValidate() error {
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

	if im.LocalOpts.Quickstart.IsDistFileSdk() {
		err := im.LocalOpts.Sdk.Prepare(im.LocalOpts.Quickstart.DistFile)
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

	im.AwaitStopped(outdated)

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

	im.AwaitStarted(instances)

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

	im.AwaitStopped(instances)

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

func (im *InstanceManager) AwaitStartedOne(instance Instance) {
	im.AwaitStarted([]Instance{instance})
}

func (im *InstanceManager) AwaitStartedAll() {
	im.AwaitStarted(im.All())
}

// TODO add timeout and then return error
func (im *InstanceManager) AwaitStarted(instances []Instance) {
	if len(instances) == 0 {
		return
	}
	log.Infof("awaiting up instance(s) '%s'", InstanceIds(instances))
	im.Check(instances, im.CheckOpts, []Checker{
		im.CheckOpts.BundleStable,
		im.CheckOpts.EventStable,
		im.CheckOpts.AwaitUpTimeout,
	})
}

func (im *InstanceManager) AwaitStoppedOne(instance Instance) {
	im.AwaitStopped([]Instance{instance})
}

func (im *InstanceManager) AwaitStoppedAll() {
	im.AwaitStopped(im.Locals())
}

// TODO add timeout and then return error
func (im *InstanceManager) AwaitStopped(instances []Instance) {
	if len(instances) == 0 {
		return
	}
	log.Infof("awaiting down instance(s) '%s'", InstanceIds(instances))
	im.Check(instances, im.CheckOpts, []Checker{
		NewStatusStoppedChecker(),
		NewTimeoutChecker("down", time.Minute*5),
	})
}

func (im *InstanceManager) configureLocalOpts(config *cfg.Config) {
	opts := config.Values().Instance.Local

	if len(opts.UnpackDir) > 0 {
		im.LocalOpts.UnpackDir = opts.UnpackDir
	}
	if len(opts.Quickstart.DistFile) > 0 {
		im.LocalOpts.Quickstart.DistFile = opts.Quickstart.DistFile
	}
	if len(opts.Quickstart.LicenseFile) > 0 {
		im.LocalOpts.Quickstart.LicenseFile = opts.Quickstart.LicenseFile
	}
}
