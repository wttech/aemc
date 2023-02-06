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
	SystemPropPath        = "/system/console/status-System%20Properties.json"
	SystemPropTimezone    = "user.timezone"
	SlingPropPath         = "/system/console/status-slingprops.json"
	SlingSettingsPath     = "/system/console/status-slingsettings.json"
	SlingSettingRunModes  = "Run Modes"
	SystemProductInfoPath = "/system/console/status-productinfo.txt"
	SystemProductInfoRegex
)

var (
	aemVersionRegex = regexp.MustCompile("^ {2}Adobe Experience Manager \\((.*)\\)$")
)

type Status struct {
	instance *Instance
}

func NewStatus(res *Instance) *Status {
	return &Status{instance: res}
}

func (sm Status) SystemProps() (map[string]string, error) {
	response, err := sm.instance.http.Request().Get(SystemPropPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read system properties on instance '%s'", sm.instance.id)
	}
	var results []string
	if err = fmtx.UnmarshalJSON(response.RawBody(), &results); err != nil {
		return nil, fmt.Errorf("cannot parse system properties response from instance '%s': %w", sm.instance.id, err)
	}
	props := parseProperties(results)
	return props, nil
}

func (sm Status) SlingProps() (map[string]string, error) {
	response, err := sm.instance.http.Request().Get(SlingPropPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read Sling properties on instance '%s'", sm.instance.id)
	}
	var results []string
	if err = fmtx.UnmarshalJSON(response.RawBody(), &results); err != nil {
		return nil, fmt.Errorf("cannot parse Sling properties response from instance '%s': %w", sm.instance.id, err)
	}
	props := parseProperties(results)
	return props, nil
}

func (sm Status) SlingSettings() (map[string]string, error) {
	response, err := sm.instance.http.RequestWithTimeout(time.Millisecond * 300).Get(SlingSettingsPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read Sling settings on instance '%s'", sm.instance.id)
	}
	var results []string
	if err = fmtx.UnmarshalJSON(response.RawBody(), &results); err != nil {
		return nil, fmt.Errorf("cannot parse Sling settings response from instance '%s': %w", sm.instance.id, err)
	}
	props := parseProperties(results)
	return props, nil
}

func parseProperties(results []string) map[string]string {
	results = lo.Filter(results, func(r string, _ int) bool {
		return strings.Count(strings.TrimSpace(r), " = ") == 1
	})
	resultMap := lo.Associate(results, func(r string) (string, string) {
		parts := strings.Split(strings.TrimSpace(r), " = ")
		return parts[0], parts[1]
	})
	return resultMap
}

func (sm Status) TimeLocation() (*time.Location, error) {
	systemProps, err := sm.SystemProps()
	if err != nil {
		return nil, err
	}
	locName, ok := systemProps[SystemPropTimezone]
	if !ok {
		return nil, fmt.Errorf("system property '%s' does not exist on instance ''%s", SystemPropTimezone, sm.instance.id)
	}
	timeLocation, err := time.LoadLocation(locName)
	if err != nil {
		log.Warnf("cannot load time location '%s' of instance '%s': %s", locName, sm.instance.id, err)
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
		return []string{}, fmt.Errorf(" Sling setting '%s' does not exist on instance ''%s", SlingSettingRunModes, sm.instance.id)
	}
	return lo.Map(strings.Split(stringsx.Between(values, "[", "]"), ","), func(rm string, _ int) string { return strings.TrimSpace(rm) }), nil
}

func (sm Status) AemVersion() (string, error) {
	response, err := sm.instance.http.RequestWithTimeout(time.Millisecond * 300).Get(SystemProductInfoPath)
	if err != nil {
		return instance.AemVersionUnknown, fmt.Errorf("cannot read system product info on instance '%s'", sm.instance.id)
	}
	bytes, err := io.ReadAll(response.RawBody())
	if err != nil {
		return instance.AemVersionUnknown, fmt.Errorf("cannot read system product info on instance '%s': %w", sm.instance.id, err)
	}
	lines := string(bytes)
	for _, line := range strings.Split(lines, "\n") {
		matches := aemVersionRegex.FindStringSubmatch(strings.TrimSuffix(line, "\r"))
		if matches != nil {
			return matches[1], nil
		}
	}
	return instance.AemVersionUnknown, nil
}
