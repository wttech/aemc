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

	result.ProcessingMode = instance.ProcessingParallel
	result.LocalOpts = result.NewLocalOpts(result)
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

func (im *InstanceManager) OneLocal() (*LocalInstance, error) {
	instance, err := im.One()
	if err != nil {
		return nil, err
	}
	if !instance.IsLocal() {
		return nil, fmt.Errorf("the instance matching current filters is not defined as local")
	}
	return instance.Local(), nil

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

func (im *InstanceManager) NewLocalAuthor() Instance {
	i, _ := im.NewByURL(instance.URLLocalAuthor)
	return *i
}

func (im *InstanceManager) NewLocalPublish() Instance {
	i, _ := im.NewByURL(instance.URLLocalPublish)
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
	id := strings.Join(parts, instance.IDDelimiter)

	return im.New(id, url, user, password), nil
}

func (im *InstanceManager) New(id, url, user, password string) *Instance {
	res := &Instance{
		manager: im,

		id:       id,
		user:     user,
		password: password,
	}
	res.http = NewHTTP(res, url)
	res.status = NewStatus(res)
	res.repo = NewRepo(res)
	res.packageManager = NewPackageManager(res)
	res.workflowManager = NewWorkflowManager(res)
	res.osgi = NewOSGi(res)
	res.sling = NewSling(res)
	res.crypto = NewCrypto(res)

	if res.IsLocal() {
		res.local = NewLocal(res)
	}

	return res
}

func (im *InstanceManager) AwaitOne(instance Instance) error {
	return im.AwaitStartedOne(instance)
}

func (im *InstanceManager) AwaitAll() error {
	return im.AwaitStartedAll()
}

func (im *InstanceManager) Await(instances []Instance) error {
	return im.AwaitStarted(instances)
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
				if len(iCfg.StartOpts) > 0 {
					i.local.StartOpts = iCfg.StartOpts
				}
				if len(iCfg.JVMOpts) > 0 {
					i.local.JvmOpts = iCfg.JVMOpts
				}
				if len(iCfg.RunModes) > 0 {
					i.local.RunModes = iCfg.RunModes
				}
				i.local.EnvVars = iCfg.EnvVars
				i.local.SecretVars = iCfg.SecretVars
				i.local.SlingProps = iCfg.SlingProps
				if len(iCfg.Version) > 0 {
					i.local.Version = iCfg.Version
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
	instanceOpts := config.Values().Instance

	packageOpts := instanceOpts.Package
	inst.packageManager.SnapshotDeploySkipping = packageOpts.SnapshotDeploySkipping
	if packageOpts.SnapshotPatterns != nil {
		inst.packageManager.SnapshotPatterns = packageOpts.SnapshotPatterns
	}
	if packageOpts.ToggledWorkflows != nil {
		inst.packageManager.ToggledWorkflows = packageOpts.ToggledWorkflows
	}

	statusOpts := instanceOpts.Status
	inst.status.Timeout = statusOpts.Timeout

	repoOpts := instanceOpts.Repo
	inst.repo.PropertyChangeIgnored = repoOpts.PropertyChangeIgnored

	osgiOpts := instanceOpts.OSGi
	inst.osgi.shutdownDelay = osgiOpts.ShutdownDelay

	inst.osgi.bundleManager.InstallStart = osgiOpts.Bundle.Install.Start
	inst.osgi.bundleManager.InstallStartLevel = osgiOpts.Bundle.Install.StartLevel
	inst.osgi.bundleManager.InstallRefreshPackages = osgiOpts.Bundle.Install.RefreshPackages

	cryptoOpts := instanceOpts.Crypto
	inst.crypto.keyBundleSymbolicName = cryptoOpts.KeyBundleSymbolicName

	workflowOpts := instanceOpts.Workflow
	inst.workflowManager.LibRoot = workflowOpts.LibRoot
	inst.workflowManager.ConfigRoot = workflowOpts.ConfigRoot
	inst.workflowManager.ToggleRetryDelay = workflowOpts.ToggleRetryDelay
	inst.workflowManager.ToggleRetryTimeout = workflowOpts.ToggleRetryTimeout
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
	if opts.AwaitStartedTimeout.Duration > 0 {
		im.CheckOpts.AwaitStartedTimeout.Duration = opts.AwaitStartedTimeout.Duration
	}
	if opts.AwaitStoppedTimeout.Duration > 0 {
		im.CheckOpts.AwaitStoppedTimeout.Duration = opts.AwaitStoppedTimeout.Duration
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

func InstanceMsg(instances []Instance, msg string) string {
	count := len(instances)
	switch count {
	case 0:
		return fmt.Sprintf("[] > %s", msg)
	default:
		return fmt.Sprintf("%s > %s", InstanceIds(instances), msg)
	}
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
