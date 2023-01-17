package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/lox"
	"github.com/wttech/aemc/pkg/instance"
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

func NewInstanceManager(aem *Aem) *InstanceManager {
	result := new(InstanceManager)
	result.aem = aem

	result.LocalOpts = result.NewLocalOpts(result)
	result.ProcessingMode = instance.ProcessingParallel
	result.CheckOpts = result.NewCheckOpts()
	result.Instances = result.NewLocalPair()

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

func (im *InstanceManager) Some() ([]Instance, error) {
	result := im.All()
	if len(result) == 0 {
		return result, fmt.Errorf("no instances defined")
	}
	return result, nil
}

func (im *InstanceManager) All() []Instance {
	return im.Instances
}

func (im *InstanceManager) Remotes() []Instance {
	return lo.Filter(im.All(), func(i Instance, _ int) bool { return i.IsRemote() })
}

func (im *InstanceManager) Locals() []Instance {
	return lo.Filter(im.All(), func(i Instance, _ int) bool { return i.IsLocal() })
}

func (im *InstanceManager) SomeLocals() ([]Instance, error) {
	result := im.Locals()
	if len(result) == 0 {
		return result, fmt.Errorf("no local instances defined")
	}
	return result, nil
}

func (im *InstanceManager) Authors() []Instance {
	return lo.Filter(im.All(), func(i Instance, _ int) bool { return i.IsAuthor() })
}

func (im *InstanceManager) Publishes() []Instance {
	return lo.Filter(im.All(), func(i Instance, _ int) bool { return i.IsPublish() })
}

type CheckOpts struct {
	Warmup        time.Duration
	Interval      time.Duration
	DoneThreshold int
	DoneNever     bool
	AwaitStrict   bool

	Reachable        ReachableHTTPChecker
	BundleStable     BundleStableChecker
	EventStable      EventStableChecker
	Installer        InstallerChecker
	AwaitUpTimeout   TimeoutChecker
	Unreachable      ReachableHTTPChecker
	StatusStopped    StatusStoppedChecker
	AwaitDownTimeout TimeoutChecker
}

func (im *InstanceManager) NewCheckOpts() *CheckOpts {
	return &CheckOpts{
		Warmup:        time.Second * 1,
		Interval:      time.Second * 5,
		DoneThreshold: 3,

		Reachable:        NewReachableChecker(true),
		BundleStable:     NewBundleStableChecker(),
		EventStable:      NewEventStableChecker(),
		AwaitUpTimeout:   NewTimeoutChecker("up", time.Minute*10),
		Installer:        NewInstallerChecker(),
		StatusStopped:    NewStatusStoppedChecker(),
		AwaitDownTimeout: NewTimeoutChecker("down", time.Minute*5),
		Unreachable:      NewReachableChecker(false),
	}
}

func (im *InstanceManager) Check(instances []Instance, opts *CheckOpts, checks []Checker) error {
	if len(instances) == 0 {
		log.Infof("no instances to check")
		return nil
	}
	time.Sleep(opts.Warmup)
	doneTimes := 0
	for {
		done, err := im.CheckOnce(instances, checks)
		if err != nil {
			return err
		}
		if done {
			if !opts.DoneNever {
				doneTimes++
				if doneTimes <= opts.DoneThreshold {
					log.Infof("instances checked (%d/%d)", doneTimes, opts.DoneThreshold)
				}
				if doneTimes == opts.DoneThreshold {
					break
				}
			}
		} else {
			doneTimes = 0
		}
		time.Sleep(opts.Interval)
	}
	return nil
}

func (im *InstanceManager) CheckOnce(instances []Instance, checks []Checker) (bool, error) {
	instanceResults, err := InstanceProcess(im.aem, instances, func(i Instance) ([]CheckResult, error) {
		var results []CheckResult
		for _, check := range checks {
			result := check.Check(i)
			results = append(results, result)
			if result.abort {
				log.Fatalf("%s | %s", i.ID(), result.message)
			}
			if result.err != nil {
				log.Infof("%s | %s", i.ID(), result.err)
			} else if len(result.message) > 0 {
				log.Infof("%s | %s", i.ID(), result.message)
			}
			if !result.ok && check.Spec().Mandatory {
				break
			}
		}
		return results, nil
	})
	if err != nil {
		return false, nil
	}
	ok := lo.EveryBy(instanceResults, func(results []CheckResult) bool {
		return lo.EveryBy(results, func(result CheckResult) bool { return result.ok })
	})
	return ok, nil
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
	res.sling = NewSling(res)

	if res.IsLocal() {
		res.local = NewLocal(res)
	}

	return res
}

func (im *InstanceManager) AwaitOne(instance Instance) {
	im.AwaitStartedOne(instance)
}

func (im *InstanceManager) AwaitAll() {
	im.AwaitStartedAll()
}

func (im *InstanceManager) Await(instances []Instance) {
	im.AwaitStarted(instances)
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
				continue
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
			defined = append(defined, *i)
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
	inst.packageManager.SnapshotDeploySkipping = packageOpts.SnapshotDeploySkipping
	if packageOpts.SnapshotPatterns != nil {
		inst.packageManager.SnapshotPatterns = packageOpts.SnapshotPatterns
	}

	osgiOpts := config.Values().Instance.OSGi
	inst.osgi.bundleManager.InstallStart = osgiOpts.Bundle.Install.Start
	inst.osgi.bundleManager.InstallStartLevel = osgiOpts.Bundle.Install.StartLevel
	inst.osgi.bundleManager.InstallRefreshPackages = osgiOpts.Bundle.Install.RefreshPackages
}

func (im *InstanceManager) configureCheckOpts(config *cfg.Config) {
	opts := config.Values().Instance.Check

	im.CheckOpts.Warmup = opts.Warmup
	im.CheckOpts.Interval = opts.Interval
	im.CheckOpts.DoneThreshold = opts.DoneThreshold

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
	im.CheckOpts.AwaitStrict = opts.AwaitStrict
	if opts.AwaitUpTimeout.Duration > 0 {
		im.CheckOpts.AwaitUpTimeout.Duration = opts.AwaitUpTimeout.Duration
	}
	if opts.AwaitDownTimeout.Duration > 0 {
		im.CheckOpts.AwaitDownTimeout.Duration = opts.AwaitDownTimeout.Duration
	}
	im.CheckOpts.Installer.State = opts.Installer.State
	im.CheckOpts.Installer.Pause = opts.Installer.Pause

	im.CheckOpts.Reachable.Timeout = opts.Reachable.Timeout
	im.CheckOpts.Unreachable.Timeout = opts.Unreachable.Timeout
}

func (im *InstanceManager) configureRest(config *cfg.Config) {
	im.ProcessingMode = instance.ProcessingMode(config.Values().Instance.ProcessingMode)
}

func InstanceIds(instances []Instance) string {
	return strings.Join(lo.Map(instances, func(i Instance, _ int) string { return i.id }), ",")
}

// InstanceProcess is a workaround for <https://stackoverflow.com/a/71132286/3360007> (ideally should be a method of manager)
func InstanceProcess[R any](aem *Aem, instances []Instance, processor func(instance Instance) (R, error)) ([]R, error) {
	parallel := false
	mode := aem.InstanceManager().ProcessingMode
	if mode == instance.ProcessingParallel {
		parallel = true
	} else if mode == instance.ProcessingAuto {
		parallel = lo.CountBy(instances, func(instance Instance) bool { return instance.IsLocal() }) <= 1
	}
	return lox.Map(parallel, instances, processor)
}
