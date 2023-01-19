package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"io"
	"regexp"
	"strings"
	"time"
)

const (
	SystemPropPath        = "/system/console/status-System%20Properties.json"
	SystemPropTimezone    = "user.timezone"
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

func (sm Status) SystemProperties() (map[string]string, error) {
	response, err := sm.instance.http.Request().Get(SystemPropPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read system properties on instance '%s'", sm.instance.id)
	}
	var results []string
	if err = fmtx.UnmarshalJSON(response.RawBody(), &results); err != nil {
		return nil, fmt.Errorf("cannot parse system properties response from instance '%s': %w", sm.instance.id, err)
	}
	results = lo.Filter(results, func(r string, _ int) bool {
		return strings.Count(strings.TrimSpace(r), " = ") == 1
	})
	return lo.Associate(results, func(r string) (string, string) {
		parts := strings.Split(strings.TrimSpace(r), " = ")
		return parts[0], parts[1]
	}), nil
}

func (sm Status) TimeLocation() (*time.Location, error) {
	systemProperties, err := sm.SystemProperties()
	if err != nil {
		return nil, err
	}
	locName, ok := systemProperties[SystemPropTimezone]
	if !ok {
		return nil, fmt.Errorf("system property '%s' does not exist on instance ''%s", SystemPropTimezone, sm.instance.id)
	}
	timeLocation, err := time.LoadLocation(locName)
	if err != nil {
		log.Warnf("cannot load time location '%s' of instance '%s': %s", locName, sm.instance.id, err)
	}
	return timeLocation, nil
}

func (sm Status) AemVersion() (string, error) {
	response, err := sm.instance.http.Request().Get(SystemProductInfoPath)
	if err != nil {
		return AemVersionUnknown, fmt.Errorf("cannot read system product info on instance '%s'", sm.instance.id)
	}
	bytes, err := io.ReadAll(response.RawBody())
	if err != nil {
		return AemVersionUnknown, fmt.Errorf("cannot read system product info on instance '%s': %w", sm.instance.id, err)
	}
	lines := string(bytes)
	for _, line := range strings.Split(lines, "\n") {
		matches := aemVersionRegex.FindStringSubmatch(line)
		if matches != nil {
			return matches[1], nil
		}
	}
	return AemVersionUnknown, nil
}
