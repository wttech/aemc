package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/lox"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"github.com/wttech/aemc/pkg/instance"
	"golang.org/x/exp/maps"
	nurl "net/url"
	"sort"
	"strings"
)

type InstanceManager struct {
	aem *AEM

	Instances []Instance
	LocalOpts *LocalOpts
	CheckOpts *CheckOpts

	AdHocURL        string
	FilterID        string
	FilterAuthors   bool
	FilterPublishes bool
	ProcessingMode  string
}

func NewInstanceManager(aem *AEM) *InstanceManager {
	result := new(InstanceManager)
	result.aem = aem

	cv := aem.config.Values()

	result.AdHocURL = cv.GetString("instance.adhoc_url")
	result.FilterID = cv.GetString("instance.filter.id")
	result.FilterAuthors = cv.GetBool("instance.filter.authors")
	result.FilterPublishes = cv.GetBool("instance.filter.publishes")
	result.ProcessingMode = cv.GetString("instance.processing_mode")

	result.LocalOpts = NewLocalOpts(result)
	result.CheckOpts = NewCheckOpts(result)

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
	result := im.newAdHocOrFromConfig()
	return im.filter(result)
}

func (im *InstanceManager) newAdHocOrFromConfig() []Instance {
	if im.AdHocURL != "" {
		iURL, err := im.NewByURL(im.AdHocURL)
		if err != nil {
			log.Fatalf("cannot create instance from ad hoc URL '%s': %s", im.AdHocURL, err)
		}
		return []Instance{*iURL}
	}
	cv := im.aem.config.Values()
	configIDs := maps.Keys(cv.GetStringMap("instance.config"))
	if len(configIDs) > 0 {
		var result []Instance
		for _, id := range configIDs {
			cv.SetDefault(fmt.Sprintf("instance.config.%s.active", id), true)
			active := cv.GetBool(fmt.Sprintf("instance.config.%s.active", id))
			if active {
				if i := im.newFromConfig(id); i != nil {
					result = append(result, *i)
				}
			}
		}
		return result
	}
	return im.NewLocalPair()
}

func (im *InstanceManager) newFromConfig(id string) *Instance {
	cv := im.aem.config.Values()

	httpURL := cv.GetString(fmt.Sprintf("instance.config.%s.http_url", id))
	i, err := im.NewByURL(httpURL)
	if err != nil {
		log.Fatalf("cannot create instance from config with ID '%s' using URL '%s': %s", id, httpURL, err)
		return nil
	}

	i.id = id

	cv.SetDefault(fmt.Sprintf("instance.config.%s.user", id), i.user)
	i.user = cv.GetString(fmt.Sprintf("instance.config.%s.user", id))

	cv.SetDefault(fmt.Sprintf("instance.config.%s.password", id), i.password)
	i.password = cv.GetString(fmt.Sprintf("instance.config.%s.password", id))

	if i.IsLocal() {
		cv.SetDefault(fmt.Sprintf("instance.config.%s.version", id), "1")
		i.local.Version = cv.GetString(fmt.Sprintf("instance.config.%s.version", id))

		i.local.StartOpts = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.start_opts", id))
		i.local.JvmOpts = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.jvm_opts", id))
		i.local.RunModes = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.run_modes", id))
		i.local.EnvVars = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.env_vars", id))
		i.local.SecretVars = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.secret_vars", id))
		i.local.SlingProps = cv.GetStringSlice(fmt.Sprintf("instance.config.%s.sling_props", id))
		i.local.UnpackDir = cv.GetString(fmt.Sprintf("instance.config.%s.unpack_dir", id))
	}
	return i
}

func (im *InstanceManager) filter(instances []Instance) []Instance {
	result := []Instance{}
	if im.FilterID != "" {
		for _, i := range instances {
			if i.id == im.FilterID {
				result = append(result, i)
				break
			}
		}
	} else {
		if im.FilterAuthors == im.FilterPublishes {
			result = instances
		} else {
			if im.FilterAuthors {
				for _, i := range instances {
					if i.IsAuthor() {
						result = append(result, i)
					}
				}
			}
			if im.FilterPublishes {
				for _, i := range instances {
					if i.IsPublish() {
						result = append(result, i)
					}
				}
			}
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return strings.Compare(result[i].id, result[j].id) < 0
	})
	return result
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
	res.ssl = NewSSL(res)

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

func InstanceIds(instances []Instance) string {
	return strings.Join(lo.Map(instances, func(i Instance, _ int) string { return i.id }), ",")
}

func InstanceMsg(instance Instance, msg any) string {
	return stringsx.AddPrefix(fmt.Sprintf("%v", msg), fmt.Sprintf("%s > ", instance.ID()))
}

func InstancesMsg(instances []Instance, msg any) string {
	count := len(instances)
	switch count {
	case 0:
		return fmt.Sprintf("[] > %v", msg)
	default:
		return fmt.Sprintf("[%s] > %v", InstanceIds(instances), msg)
	}
}

// InstanceProcess is a workaround for <https://stackoverflow.com/a/71132286/3360007> (ideally should be a method of manager)
func InstanceProcess[R any](aem *AEM, instances []Instance, processor func(instance Instance) (R, error)) ([]R, error) {
	parallel := false
	mode := aem.InstanceManager().ProcessingMode
	if mode == instance.ProcessingParallel {
		parallel = true
	} else if mode == instance.ProcessingAuto {
		parallel = lo.CountBy(instances, func(instance Instance) bool { return instance.IsLocal() }) <= 1
	}
	return lox.Map(parallel, instances, processor)
}
