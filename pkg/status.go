package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"github.com/wttech/aemc/pkg/instance"
	"io"
	"regexp"
	"strings"
	"time"
)

const (
	SystemPropPath          = "/system/console/status-System%20Properties.json"
	SystemPropTimezone      = "user.timezone"
	SlingPropPath           = "/system/console/status-slingprops.json"
	SlingSettingsPath       = "/system/console/status-slingsettings.json"
	SlingSettingRunModes    = "Run Modes"
	SystemProductInfoPath   = "/system/console/productinfo"
	SystemProductInfoMarker = ">Installed Products</th>"
)

var (
	aemVersionRegex = regexp.MustCompile("<td>Adobe Experience Manager \\((.*)\\)<\\/td>")
)

type Status struct {
	instance *Instance

	Timeout time.Duration
}

func NewStatus(i *Instance) *Status {
	cv := i.manager.aem.config.Values()

	return &Status{
		instance: i,

		Timeout: cv.GetDuration("instance.status.timeout"),
	}
}

func (sm Status) SystemProps() (map[string]string, error) {
	response, err := sm.instance.http.RequestWithTimeout(sm.Timeout).Get(SystemPropPath)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot read system properties", sm.instance.ID())
	}
	props, err := sm.parseProperties(response.RawBody())
	if err != nil {
		return nil, fmt.Errorf("%s > cannot parse system properties: %w", sm.instance.ID(), err)
	}
	return props, nil
}

func (sm Status) SlingProps() (map[string]string, error) {
	response, err := sm.instance.http.RequestWithTimeout(sm.Timeout).Get(SlingPropPath)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot read Sling properties", sm.instance.ID())
	}
	props, err := sm.parseProperties(response.RawBody())
	if err != nil {
		return nil, fmt.Errorf("%s > cannot parse Sling properties: %w", sm.instance.ID(), err)
	}
	return props, nil
}

func (sm Status) SlingSettings() (map[string]string, error) {
	response, err := sm.instance.http.RequestWithTimeout(sm.Timeout).Get(SlingSettingsPath)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot read Sling settings", sm.instance.ID())
	}
	props, err := sm.parseProperties(response.RawBody())
	if err != nil {
		return nil, fmt.Errorf("%s > cannot parse Sling settings: %w", sm.instance.ID(), err)
	}
	return props, nil
}

func (sm Status) parseProperties(response io.ReadCloser) (map[string]string, error) {
	responseBytes, err := io.ReadAll(response)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot parse properties: %w", sm.instance.ID(), err)
	}
	responseString := strings.ReplaceAll(string(responseBytes), "\\", "\\\\")
	var results []string
	if err = fmtx.UnmarshalJSON(io.NopCloser(strings.NewReader(responseString)), &results); err != nil {
		return nil, fmt.Errorf("%s > cannot parse properties : %w", sm.instance.ID(), err)
	}
	results = lo.Filter(results, func(r string, _ int) bool {
		return strings.Count(strings.TrimSpace(r), " = ") == 1
	})
	resultMap := lo.Associate(results, func(r string) (string, string) {
		parts := strings.Split(strings.TrimSpace(r), " = ")
		return parts[0], parts[1]
	})
	return resultMap, nil
}

func (sm Status) TimeLocation() (*time.Location, error) {
	systemProps, err := sm.SystemProps()
	if err != nil {
		return nil, err
	}
	locName, ok := systemProps[SystemPropTimezone]
	if !ok {
		return nil, fmt.Errorf("%s > system property '%s' does not exist", sm.instance.ID(), SystemPropTimezone)
	}
	timeLocation, err := time.LoadLocation(locName)
	if err != nil {
		log.Warnf("%s > cannot load time location '%s': %s", sm.instance.ID(), locName, err)
	}
	return timeLocation, nil
}

func (sm Status) RunModes() ([]string, error) {
	slingSettings, err := sm.SlingSettings()
	if err != nil {
		return nil, err
	}
	values, ok := slingSettings[SlingSettingRunModes]
	if !ok {
		return []string{}, fmt.Errorf("%s > Sling setting '%s' does not exist", sm.instance.ID(), SlingSettingRunModes)
	}
	return lo.Map(strings.Split(stringsx.Between(values, "[", "]"), ","), func(rm string, _ int) string { return strings.TrimSpace(rm) }), nil
}

func (sm Status) AemVersion() (string, error) {
	response, err := sm.instance.http.RequestWithTimeout(sm.Timeout).Get(SystemProductInfoPath)
	if err != nil {
		return instance.AemVersionUnknown, fmt.Errorf("%s > cannot read system product info", sm.instance.ID())
	}
	bytes, err := io.ReadAll(response.RawBody())
	if err != nil {
		return instance.AemVersionUnknown, fmt.Errorf("%s > cannot read system product info: %w", sm.instance.ID(), err)
	}
	html := stringsx.AfterLast(string(bytes), SystemProductInfoMarker)
	matches := aemVersionRegex.FindStringSubmatch(html)
	if matches != nil {
		return matches[1], nil
	}

	return instance.AemVersionUnknown, nil
}
