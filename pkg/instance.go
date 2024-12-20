package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/instance"
	"golang.org/x/exp/maps"
	nurl "net/url"
	"strings"
	"time"
)

// Instance represents AEM instance
type Instance struct {
	manager  *InstanceManager
	id       string
	user     string
	password string

	local           *LocalInstance
	http            *HTTP
	status          *Status
	repo            *Repo
	oak             *OAK
	osgi            *OSGi
	sling           *Sling
	crypto          *Crypto
	ssl             *SSL
	gtsManager      *GTSManager
	auth            *Auth
	replication     *Replication
	packageManager  *PackageManager
	workflowManager *WorkflowManager
}

type InstanceState struct {
	ID           string   `yaml:"id" json:"id"`
	URL          string   `json:"url" json:"url"`
	AemVersion   string   `yaml:"aem_version" json:"aemVersion"`
	Attributes   []string `yaml:"attributes" json:"attributes"`
	RunModes     []string `yaml:"run_modes" json:"runModes"`
	HealthChecks []string `yaml:"health_checks" json:"healthChecks"`
}

func (i Instance) State() InstanceState {
	return InstanceState{
		ID:           i.id,
		URL:          i.http.BaseURL(),
		AemVersion:   i.AemVersion(),
		Attributes:   i.Attributes(),
		RunModes:     i.RunModes(),
		HealthChecks: i.HealthChecks(),
	}
}

func (i Instance) ID() string {
	return i.id
}

func (i Instance) User() string {
	return i.user
}

func (i Instance) Password() string {
	return i.password
}

func (i Instance) Manager() *InstanceManager {
	return i.manager
}

func (i Instance) Local() *LocalInstance {
	return i.local
}

func (i Instance) HTTP() *HTTP {
	return i.http
}

func (i Instance) Status() *Status {
	return i.status
}

func (i Instance) Repo() *Repo {
	return i.repo
}

func (i Instance) OSGI() *OSGi {
	return i.osgi
}

func (i Instance) OAK() *OAK {
	return i.oak
}

func (i Instance) Sling() *Sling {
	return i.sling
}

func (i Instance) PackageManager() *PackageManager {
	return i.packageManager
}

func (i Instance) WorkflowManager() *WorkflowManager {
	return i.workflowManager
}

func (i Instance) Crypto() *Crypto {
	return i.crypto
}

func (i Instance) SSL() *SSL {
	return i.ssl
}

func (i Instance) GTS() *GTSManager {
	return i.gtsManager
}

func (i Instance) Auth() *Auth {
	return i.auth
}

func (i Instance) Replication() *Replication {
	return i.replication
}

func (i Instance) IDInfo() IDInfo {
	parts := strings.Split(i.id, instance.IDDelimiter)
	if len(parts) == 2 {
		return IDInfo{
			Location: parts[0],
			Role:     instance.Role(parts[1]),
		}
	}
	return IDInfo{
		Location:   parts[0],
		Role:       instance.Role(parts[1]),
		Classifier: parts[2],
	}
}

func (i Instance) IDColor() string {
	id := i.ID()
	if i.IsAuthor() {
		id = color.HiMagentaString(id)
	} else {
		id = color.HiBlueString(id)
	}
	return id
}

type IDInfo struct {
	Location   string
	Role       instance.Role
	Classifier string
}

func (i Instance) IsLocal() bool {
	return i.IDInfo().Location == instance.LocationLocal
}

func (i Instance) IsRemote() bool {
	return !i.IsLocal()
}

func (i Instance) IsAuthor() bool {
	return i.IDInfo().Role == instance.RoleAuthor
}

func (i Instance) IsPublish() bool {
	return i.IDInfo().Role == instance.RolePublish
}

func (i Instance) IsAdHoc() bool {
	return i.IDInfo().Role == instance.RoleAdHoc
}

func roleByURL(config *nurl.URL) instance.Role {
	if strings.HasSuffix(config.Port(), instance.RoleAuthorPortSuffix) {
		return instance.RoleAuthor
	}
	return instance.RolePublish
}

func credentialsByURL(config *nurl.URL) (string, string) {
	user := instance.UserDefault
	pwd := instance.PasswordDefault

	urlUser := config.User.Username()
	if urlUser != "" {
		user = urlUser
	}

	urlPwd, hasPwd := config.User.Password()
	if hasPwd && urlPwd != "" {
		pwd = urlPwd
	}

	return user, pwd
}

func localHosts() []string {
	return []string{"127.0.0.1", "localhost"}
}

func (i Instance) TimeLocation() *time.Location {
	loc, err := i.status.TimeLocation()
	if err != nil {
		log.Debugf("%s > cannot determine time location: %s", i.IDColor(), err)
		return time.Now().Location()
	}
	return loc
}

func (i Instance) AemVersion() string {
	// TODO try to retrieve version from filename 'aem/home/instance/[author|publish]/crx-quickstart/app/cq-quickstart-6.5.0-standalone-quickstart.jar'
	version, err := i.status.AemVersion()
	if err != nil {
		log.Debugf("%s > cannot determine AEM version: %s", i.IDColor(), err)
		return instance.AemVersionUnknown
	}
	return version
}

func (i Instance) RunModes() []string {
	runModes, err := i.status.RunModes()
	if err != nil {
		log.Debugf("%s > cannot determine run modes: %s", i.IDColor(), err)
		return []string{}
	}
	return runModes
}

func (i Instance) Now() time.Time {
	return time.Now().In(i.TimeLocation())
}

func (i Instance) Time(unixMilli int64) time.Time {
	return time.UnixMilli(unixMilli).In(i.TimeLocation())
}

func (i Instance) Attributes() []string {
	var result []string
	if i.IsLocal() {
		result = append(result, instance.AttributeLocal)
		if i.Local().IsCreated() {
			result = append(result, instance.AttributeCreated)
			status, err := i.Local().Status()
			if err == nil {
				result = append(result, status.String())
			}
			if i.Local().UpToDate() {
				result = append(result, instance.AttributeUpToDate)
			} else {
				result = append(result, instance.AttributeOutOfDate)
			}
		} else {
			result = append(result, instance.AttributeUncreated)
		}
	} else {
		result = append(result, instance.AttributeRemote)
	}
	return result
}

func (i Instance) HealthChecks() []string {
	messages := []string{}
	if !i.IsLocal() || i.Local().IsRunning() {
		checks := []Checker{
			i.manager.CheckOpts.Reachable,
			i.manager.CheckOpts.BundleStable,
			i.manager.CheckOpts.EventStable,
			i.manager.CheckOpts.Installer,
			i.manager.CheckOpts.LoginPage,
			i.manager.CheckOpts.ComponentStable,
		}
		for _, check := range checks {
			if check.Spec().Skip {
				continue
			}
			result := check.Check(i.manager.CheckContext().Value(checkContextKey{}).(CheckContext), i)
			resultText := InstanceTrim(i, result.Text())
			if resultText != "" {
				messages = append(messages, resultText)
			}
			if !result.ok && check.Spec().Mandatory {
				break
			}
		}
	}
	return messages
}

func (i Instance) String() string {
	return fmt.Sprintf("instance '%s' (url: %s, attrs: %s)", i.id, i.HTTP().BaseURL(), strings.Join(i.Attributes(), ", "))
}

func (i Instance) MarshalJSON() ([]byte, error) {
	if i.IsLocal() {
		return json.Marshal(i.Local().State())
	}
	return json.Marshal(i.State())
}

func (i Instance) MarshalYAML() (interface{}, error) {
	if i.IsLocal() {
		return i.Local().State(), nil
	}
	return i.State(), nil
}

func (i Instance) MarshalText() string {
	state := i.State()
	sb := bytes.NewBufferString("")
	sb.WriteString(fmt.Sprintf("ID '%s'\n", state.ID))
	props := map[string]any{
		"http url":      state.URL,
		"attributes":    state.Attributes,
		"aem version":   i.AemVersion(),
		"health checks": i.HealthChecks(),
		"run modes":     i.RunModes(),
	}
	if i.IsLocal() {
		l := i.Local()
		maps.Copy(props, map[string]any{
			"dir": l.Dir(),
		})
	}
	sb.WriteString(fmtx.TblProps(props))
	return sb.String()
}

func (i Instance) CacheDir() string {
	return fmt.Sprintf("%s/%s", i.manager.aem.baseOpts.CacheDir, i.ID())
}

func (i Instance) LockDir() string {
	if i.IsLocal() {
		return i.local.StateDir()
	}
	return i.CacheDir()
}
