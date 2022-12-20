package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/lox"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/instance"
	"github.com/wttech/aemc/pkg/java"
	nurl "net/url"
	"sort"
	"strings"
	"time"
)

type InstanceManager struct {
	aem *Aem

	Instances      []Instance
	LocalOpts      *LocalOpts
	CheckOpts      *CheckOpts
	ProcessingMode instance.ProcessingMode
}

const (
	RootPath    = "aem/home/instance"
	DistPath    = "aem/lib/aem-sdk-quickstart.jar"
	LicensePath = "aem/lib/license.properties"
)

func NewInstanceManager(aem *Aem) *InstanceManager {
	result := new(InstanceManager)
	result.aem = aem
	result.Instances = result.NewLocalPair()
	result.CheckOpts = result.NewCheckOpts()
	result.LocalOpts = result.NewLocalOpts()
	result.ProcessingMode = instance.ProcessingParallel

	return result
}

func (im *InstanceManager) One() (*Instance, error) {
	instances := im.All()
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instance that matches the current filters")
	}
	if len(instances) > 1 {
		return nil, fmt.Errorf("more than one instance is matching current filters")
	}
	i := instances[0]
	return &i, nil
}

func (im InstanceManager) Some() ([]Instance, error) {
	result := im.All()
	if len(result) == 0 {
		return result, fmt.Errorf("no instances defined")
	}
	return result, nil
}

func (im InstanceManager) All() []Instance {
	return im.Instances
}

func (im InstanceManager) Remotes() []Instance {
	return lo.Filter(im.All(), func(i Instance, _ int) bool { return i.IsRemote() })
}

func (im InstanceManager) Locals() []Instance {
	return lo.Filter(im.All(), func(i Instance, _ int) bool { return i.IsLocal() })
}

func (im InstanceManager) SomeLocals() ([]Instance, error) {
	result := im.Locals()
	if len(result) == 0 {
		return result, fmt.Errorf("no local instances defined")
	}
	return result, nil
}

func (im InstanceManager) Authors() []Instance {
	return lo.Filter(im.All(), func(i Instance, _ int) bool { return i.IsAuthor() })
}

func (im InstanceManager) Publishes() []Instance {
	return lo.Filter(im.All(), func(i Instance, _ int) bool { return i.IsPublish() })
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

	var created []Instance
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

	var started []Instance
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

	var stopped []Instance
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

func (im *InstanceManager) DeleteOne(instance Instance) (bool, error) {
	deleted, err := im.Delete([]Instance{instance})
	return len(deleted) > 0, err
}

func (im *InstanceManager) DeleteAll() ([]Instance, error) {
	return im.Delete(im.Locals())
}

func (im *InstanceManager) Delete(instances []Instance) ([]Instance, error) {
	// im.LocalValidate()

	var deleted []Instance
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

type CheckOpts struct {
	Warmup   time.Duration
	Interval time.Duration

	BundleStable   BundleStableChecker
	EventStable    EventStableChecker
	AwaitUpTimeout TimeoutChecker

	StatusStopped    StatusStoppedChecker
	AwaitDownTimeout TimeoutChecker
}

func (im *InstanceManager) NewCheckOpts() *CheckOpts {
	return &CheckOpts{
		Warmup:   time.Second * 1,
		Interval: time.Second * 5,

		BundleStable:     NewBundleStableChecker(),
		EventStable:      NewEventStableChecker(),
		AwaitUpTimeout:   NewTimeoutChecker("up", time.Minute*10),
		StatusStopped:    NewStatusStoppedChecker(),
		AwaitDownTimeout: NewTimeoutChecker("down", time.Minute*5),
	}
}

func (im *InstanceManager) Check(instances []Instance, opts *CheckOpts, checks []Checker) {
	if len(instances) == 0 {
		log.Infof("no instances to check")
		return
	}
	time.Sleep(opts.Warmup)
	for {
		if im.CheckOnce(instances, checks) {
			break
		}
		time.Sleep(opts.Interval)
	}
}

func (im *InstanceManager) CheckOnce(instances []Instance, checks []Checker) bool {
	ok := true
	for _, i := range instances {
		for _, check := range checks {
			result := check.Check(i)
			if result.abort {
				log.Fatalf("%s | %s", i.ID(), result.message)
			}
			if !result.ok {
				ok = false
			}
			if result.err != nil {
				log.Infof("cannot check instance '%s': %s", i.ID(), result.err)
			} else if len(result.message) > 0 {
				log.Infof("%s | %s", i.ID(), result.message)
			}
		}
	}
	return ok
}

func (im *InstanceManager) NewLocalAuthor() Instance {
	i, _ := im.NewByURL(URLLocalAuthor)
	return *i
}

func (im *InstanceManager) NewLocalPublish() Instance {
	i, _ := im.NewByURL(URLLocalPublish)
	return *i
}

func (im *InstanceManager) NewLocalPair() []Instance {
	return []Instance{im.NewLocalAuthor(), im.NewLocalPublish()}
}

func (im *InstanceManager) NewByURL(url string) (*Instance, error) {
	urlConfig, err := nurl.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("invalid instance URL '%s': %w", url, err)
	}

	env := locationByURL(urlConfig)
	typeName := roleByURL(urlConfig)
	classifier := classifierByURL(urlConfig)
	user, password := credentialsByURL(urlConfig)

	parts := []string{env, string(typeName)}
	if len(classifier) > 0 {
		parts = append(parts, classifier)
	}
	id := strings.Join(parts, IDDelimiter)

	return im.New(id, url, user, password), nil
}

func (im *InstanceManager) New(id, url, user, password string) *Instance {
	res := &Instance{
		manager:  im,
		id:       id,
		user:     user,
		password: password,
	}
	res.http = NewHTTP(res, url)
	res.repository = NewRepo(res)
	res.packageManager = NewPackageManager(res)
	res.osgi = NewOSGi(res)

	if res.IsLocal() {
		res.local = NewLocal(res)
	}

	return res
}

type LocalOpts struct {
	UnpackPath     string
	JavaOpts       *java.Opts
	QuickstartOpts QuickstartOpts
}

func (im *InstanceManager) NewLocalOpts() *LocalOpts {
	pathCurrent := osx.PathCurrent()
	return &LocalOpts{
		UnpackPath: pathCurrent + "/" + RootPath,
		JavaOpts:   im.aem.javaOpts,
		QuickstartOpts: QuickstartOpts{
			DistPath:    pathCurrent + "/" + DistPath,
			LicensePath: pathCurrent + "/" + LicensePath,
		},
	}
}

func (im InstanceManager) Process(instances []Instance, processor func(instance Instance) (map[string]any, error)) ([]map[string]any, error) {
	parallel := false
	if im.ProcessingMode == instance.ProcessingParallel {
		parallel = true
	} else if im.ProcessingMode == instance.ProcessingAuto {
		parallel = lo.CountBy(instances, func(instance Instance) bool { return instance.IsLocal() }) <= 1
	}
	return lox.Map(parallel, instances, processor)
}

func (o *LocalOpts) Validate() error {
	err := o.JavaOpts.Validate()
	if err != nil {
		return err
	}
	err = o.QuickstartOpts.Validate()
	if err != nil {
		return err
	}
	return nil
}

type QuickstartOpts struct {
	DistPath    string
	LicensePath string
}

func (o *QuickstartOpts) Validate() error {
	if !osx.PathExists(o.DistPath) {
		return fmt.Errorf("quickstart dist file does not exist at path '%s'; consider specifying quickstart dist file via property 'instance.local.quickstart.dist_path'", o.DistPath)
	}
	if !osx.PathExists(o.LicensePath) {
		return fmt.Errorf("quickstart license file does not exist at path '%s'; consider specifying quickstart dist file via property 'instance.local.quickstart.license_path'", o.LicensePath)
	}
	return nil
}

func (im *InstanceManager) Configure(config *cfg.Config) {
	im.configureInstances(config)
	im.configureCheckOpts(config)
	im.configureLocalOpts(config)
	im.configureRest(config)
}

func (im *InstanceManager) configureInstances(config *cfg.Config) {
	opts := config.Values().Instance
	var defined []Instance

	if len(opts.Config) > 0 {
		for ID, iCfg := range opts.Config {
			i, err := im.NewByURL(iCfg.HTTPURL)
			if err != nil {
				log.Warn(fmt.Errorf("cannot create instance from URL '%s': %w", iCfg.HTTPURL, err))
			} else {
				defined = append(defined, *i)
			}
			i.id = ID
			if len(iCfg.User) > 0 {
				i.user = iCfg.User
			}
			if len(iCfg.Password) > 0 {
				i.password = iCfg.Password
			}
			if i.IsLocal() {
				if len(iCfg.RunModes) > 0 {
					if len(iCfg.JVMOpts) > 0 {
						i.local.JvmOpts = iCfg.JVMOpts
					}
					if len(iCfg.RunModes) > 0 {
						i.local.RunModes = iCfg.RunModes
					}
					if len(iCfg.Version) > 0 {
						i.local.Version = iCfg.Version
					}
				}
			}
		}
	} else if len(opts.ConfigURL) > 0 {
		iURL, err := im.NewByURL(opts.ConfigURL)
		if err != nil {
			log.Info(fmt.Sprintf("cannot use instance with URL '%s'", opts.ConfigURL))
		} else {
			defined = append(defined, *iURL)
		}
	}

	if len(defined) == 0 {
		defined = im.NewLocalPair()
	}

	var filtered []Instance
	if len(opts.Filter.ID) > 0 {
		for _, i := range defined {
			if i.id == opts.Filter.ID {
				filtered = append(filtered, i)
				break
			}
		}
	} else {
		if opts.Filter.Author == opts.Filter.Publish {
			filtered = defined
		} else {
			if opts.Filter.Author {
				for _, i := range defined {
					if i.IsAuthor() {
						filtered = append(filtered, i)
					}
				}
			}
			if opts.Filter.Publish {
				for _, i := range defined {
					if i.IsPublish() {
						filtered = append(filtered, i)
					}
				}
			}
		}
	}

	for _, inst := range filtered {
		configureInstance(inst, config)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		return strings.Compare(filtered[i].id, filtered[j].id) < 0
	})

	im.Instances = filtered
}

func configureInstance(inst Instance, config *cfg.Config) {
	packageOpts := config.Values().Instance.Package
	inst.packageManager.uploadForce = packageOpts.Upload.Force

	osgiOpts := config.Values().Instance.OSGi
	inst.osgi.bundleManager.InstallStart = osgiOpts.Install.Start
	inst.osgi.bundleManager.InstallStartLevel = osgiOpts.Install.StartLevel
	inst.osgi.bundleManager.InstallRefreshPackages = osgiOpts.Install.RefreshPackages
}

func (im *InstanceManager) configureLocalOpts(config *cfg.Config) {
	opts := config.Values().Instance.Local

	if len(opts.UnpackPath) > 0 {
		im.LocalOpts.UnpackPath = opts.UnpackPath
	}
	if len(opts.Quickstart.DistPath) > 0 {
		im.LocalOpts.QuickstartOpts.DistPath = opts.Quickstart.DistPath
	}
	if len(opts.Quickstart.LicensePath) > 0 {
		im.LocalOpts.QuickstartOpts.LicensePath = opts.Quickstart.LicensePath
	}
}

func (im *InstanceManager) configureCheckOpts(config *cfg.Config) {
	opts := config.Values().Instance.Check

	im.CheckOpts.Warmup = opts.Warmup
	im.CheckOpts.Interval = opts.Interval

	if opts.BundleStable.SymbolicNamesIgnored != nil {
		im.CheckOpts.BundleStable.SymbolicNamesIgnored = opts.BundleStable.SymbolicNamesIgnored
	}
	if opts.EventStable.ReceivedMaxAge > 0 {
		im.CheckOpts.EventStable.ReceivedMaxAge = opts.EventStable.ReceivedMaxAge
	}
	if opts.EventStable.TopicsUnstable != nil {
		im.CheckOpts.EventStable.TopicsUnstable = opts.EventStable.TopicsUnstable
	}
	if opts.EventStable.DetailsIgnored != nil {
		im.CheckOpts.EventStable.DetailsIgnored = opts.EventStable.DetailsIgnored
	}
	if opts.AwaitUpTimeout.Duration > 0 {
		im.CheckOpts.AwaitUpTimeout.Duration = opts.AwaitUpTimeout.Duration
	}
	if opts.AwaitDownTimeout.Duration > 0 {
		im.CheckOpts.AwaitDownTimeout.Duration = opts.AwaitDownTimeout.Duration
	}
}

func (im *InstanceManager) configureRest(config *cfg.Config) {
	im.ProcessingMode = instance.ProcessingMode(config.Values().Instance.ProcessingMode)
}

func InstanceIds(instances []Instance) string {
	return strings.Join(lo.Map(instances, func(i Instance, _ int) string { return i.id }), ",")
}
