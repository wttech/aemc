package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/lox"
	"github.com/wttech/aemc/pkg/instance"
	"golang.org/x/exp/maps"
	nurl "net/url"
	"sort"
	"strings"
)

type InstanceManager struct {
	aem *Aem

	Instances      []Instance
	LocalOpts      *LocalOpts
	CheckOpts      *CheckOpts
	ProcessingMode string
}

func NewInstanceManager(aem *Aem) *InstanceManager {
	result := new(InstanceManager)
	result.aem = aem

	result.ProcessingMode = aem.config.Values().GetString("instance.processing_mode")

	result.LocalOpts = result.NewLocalOpts(result)
	result.CheckOpts = result.NewCheckOpts(result)
	result.Instances = result.newInstances()

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

func (im *InstanceManager) newInstances() []Instance {
	var defined []Instance

	cv := im.aem.config.Values()
	configURL := cv.GetString("instance.config_url")
	configIDs := maps.Keys(cv.GetStringMap("instance.config"))

	if configURL != "" {
		iURL, err := im.NewByURL(configURL)
		if err != nil {
			log.Info(fmt.Sprintf("cannot use instance with URL '%s'", configURL))
		} else {
			defined = append(defined, *iURL)
		}
	} else if len(configIDs) > 0 {
		for _, id := range configIDs {
			httpURL := cv.GetString(fmt.Sprintf("instance.config.%s.http_url", id))
			i, err := im.NewByURL(httpURL)
			if err != nil {
				log.Warn(fmt.Errorf("cannot create instance from URL '%s': %w", httpURL, err))
				continue
			}
			i.id = id

			user := cv.GetString(fmt.Sprintf("instance.config.%s.user", id))
			if user != "" {
				i.user = user
			}
			password := cv.GetString(fmt.Sprintf("instance.config.%s.password", id))
			if password != "" {
				i.password = password
			}

			if i.IsLocal() {
				version := cv.GetString(fmt.Sprintf("instance.config.%s.version", id))
				if version != "" {
					i.local.Version = version
				}

				i.local.StartOpts = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.start_opts", id))
				i.local.JvmOpts = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.jvm_opts", id))
				i.local.RunModes = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.run_modes", id))
				i.local.EnvVars = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.env_vars", id))
				i.local.SecretVars = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.secret_vars", id))
				i.local.SlingProps = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.sling_props", id))
			}
			defined = append(defined, *i)
		}
	}

	if len(defined) == 0 {
		defined = im.NewLocalPair()
	}

	var filtered []Instance
	filterID := cv.GetString("instance.filter.id")

	if filterID != "" {
		for _, i := range defined {
			if i.id == filterID {
				filtered = append(filtered, i)
				break
			}
		}
	} else {
		filterAuthor := cv.GetBool("instance.filter.author")
		filterPublish := cv.GetBool("instance.filter.publish")
		if filterAuthor == filterPublish {
			filtered = defined
		} else {
			if filterAuthor {
				for _, i := range defined {
					if i.IsAuthor() {
						filtered = append(filtered, i)
					}
				}
			}
			if filterPublish {
				for _, i := range defined {
					if i.IsPublish() {
						filtered = append(filtered, i)
					}
				}
			}
		}
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		return strings.Compare(filtered[i].id, filtered[j].id) < 0
	})

	return filtered
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
