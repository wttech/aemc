package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	nurl "net/url"
	"strings"
	"time"

	"github.com/wttech/aemc/pkg/common/stringsx"
)

const (
	IDDelimiter          = "_"
	URLLocalAuthor       = "http://localhost:4502"
	URLLocalPublish      = "http://localhost:4503"
	PasswordDefault      = "admin"
	UserDefault          = "admin"
	LocationLocal        = "local"
	LocationRemote       = "remote"
	RoleAuthorPortSuffix = "02"
	ClassifierDefault    = ""
)

type Role string

const (
	RoleAuthor  Role = "author"
	RolePublish Role = "publish"
)

// Instance represents AEM instance
type Instance struct {
	manager          *InstanceManager
	systemProperties map[string]string

	local          *LocalInstance
	http           *HTTP
	repository     *Repo
	osgi           *OSGi
	packageManager *PackageManager

	id       string
	user     string
	password string
}

type InstanceState struct {
	ID         string   `yaml:"id" json:"id"`
	URL        string   `json:"url" json:"url"`
	Attributes []string `yaml:"attributes" json:"attributes"`
}

func (i Instance) State() InstanceState {
	return InstanceState{
		ID:         i.id,
		URL:        i.http.BaseURL(),
		Attributes: i.Attributes(),
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

func (i *Instance) Manager() *InstanceManager {
	return i.manager
}

func (i *Instance) Local() *LocalInstance {
	return i.local
}

func (i *Instance) HTTP() *HTTP {
	return i.http
}

func (i *Instance) Repo() *Repo {
	return i.repository
}

func (i *Instance) OSGI() *OSGi {
	return i.osgi
}

func (i *Instance) PackageManager() *PackageManager {
	return i.packageManager
}

func (i Instance) IDInfo() IDInfo {
	parts := strings.Split(i.id, IDDelimiter)
	if len(parts) == 2 {
		return IDInfo{
			Location: parts[0],
			Role:     Role(parts[1]),
		}
	}
	return IDInfo{
		Location:   parts[0],
		Role:       Role(parts[1]),
		Classifier: parts[2],
	}
}

type IDInfo struct {
	Location   string
	Role       Role
	Classifier string
}

func (i *Instance) IsLocal() bool {
	return i.IDInfo().Location == LocationLocal
}

func (i *Instance) IsRemote() bool {
	return !i.IsLocal()
}

func (i *Instance) IsAuthor() bool {
	return i.IDInfo().Role == RoleAuthor
}

func (i *Instance) IsPublish() bool {
	return i.IDInfo().Role == RolePublish
}

func locationByURL(config *nurl.URL) string {
	if stringsx.Contains(localHosts(), config.Hostname()) {
		return LocationLocal
	}
	return LocationRemote
}

func roleByURL(config *nurl.URL) Role {
	if strings.HasSuffix(config.Port(), RoleAuthorPortSuffix) {
		return RoleAuthor
	}
	return RolePublish
}

// TODO local-publish-preview etc
func classifierByURL(_ *nurl.URL) string {
	return ClassifierDefault
}

func credentialsByURL(config *nurl.URL) (string, string) {
	user := UserDefault
	pwd := PasswordDefault

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

func (i *Instance) TimeLocation() *time.Location {
	err := i.loadPropertiesOnce()
	if err != nil {
		log.Warn(err)
		return nil
	}
	loc := time.Now().Location()
	locName, ok := i.systemProperties["user.timezone"]
	if ok {
		loc, err = time.LoadLocation(locName)
		if err != nil {
			log.Warnf("cannot load time location '%s' of instance '%s': %s", locName, i.id, err)
		}
	}
	return loc
}

func (i *Instance) Now() time.Time {
	return time.Now().In(i.TimeLocation())
}

func (i *Instance) Time(unixMilli int64) time.Time {
	return time.UnixMilli(unixMilli).In(i.TimeLocation())
}

func (i *Instance) Attributes() []string {
	var result []string
	if i.IsLocal() {
		result = append(result, "local")
		if i.Local().IsCreated() {
			result = append(result, "created")
			status, err := i.Local().Status()
			if err == nil {
				result = append(result, status.String())
			}
		} else {
			result = append(result, "uncreated")
		}
	} else {
		result = append(result, "remote")
	}
	return result
}

func (i *Instance) loadPropertiesOnce() error {
	if i.systemProperties == nil {
		systemProperties, err := i.ReadSystemProperties()
		if err != nil {
			return err
		}
		i.systemProperties = systemProperties
	}
	return nil
}

func (i Instance) ReadSystemProperties() (map[string]string, error) {
	response, err := i.http.Request().Get("/system/console/status-System%20Properties.json")
	if err != nil {
		return nil, fmt.Errorf("cannot read system properties on instance '%s'", i.id)
	}
	var results []string
	if err = fmtx.UnmarshalJSON(response.RawBody(), &results); err != nil {
		return nil, fmt.Errorf("cannot parse system properties response from instance '%s': %w", i.id, err)
	}
	results = lo.Filter(results, func(r string, _ int) bool {
		return strings.Count(strings.TrimSpace(r), " = ") == 1
	})
	return lo.Associate(results, func(r string) (string, string) {
		parts := strings.Split(strings.TrimSpace(r), " = ")
		return parts[0], parts[1]
	}), nil
}

func (i *Instance) SystemProperties() map[string]string {
	return i.systemProperties
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
	var props map[string]any
	if i.IsLocal() {
		l := i.Local()
		props = map[string]any{
			"http url":   state.URL,
			"attributes": state.Attributes,
			"dir":        l.Dir(),
		}
	} else {
		props = map[string]any{
			"http url":   state.URL,
			"attributes": state.Attributes,
		}
	}
	sb.WriteString(fmtx.TblProps(props))
	return sb.String()
}
