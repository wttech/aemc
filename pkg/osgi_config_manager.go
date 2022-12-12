package pkg

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/osgi"
	"golang.org/x/exp/maps"
	"io"
	"regexp"
	"strings"
)

type OSGiConfigManager struct {
	instance *Instance
}

func (cm *OSGiConfigManager) ByPID(pid string) OSGiConfig {
	return OSGiConfig{manager: cm, pid: pid}
}

func (cm *OSGiConfigManager) listPIDs() (*osgi.ConfigPIDs, error) {
	resp, err := cm.instance.http.Request().Get(ConfigMgrPath)
	if err != nil {
		return nil, fmt.Errorf("cannot request config list on instance '%s': %w", cm.instance.ID(), err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("cannot request config list on instance '%s': %w", cm.instance.ID(), err)
	}

	htmlBytes, err := io.ReadAll(resp.RawBody())
	if err != nil {
		return nil, fmt.Errorf("cannot read config list from instance '%s'", cm.instance.ID())
	}
	html := string(htmlBytes)
	r, _ := regexp.Compile("configData = (.*);")
	pids := strings.TrimSuffix(strings.TrimPrefix(r.FindString(html), "configData = "), ";")
	if len(pids) == 0 {
		return nil, fmt.Errorf("cannot find config list in HTML response from instance '%s'", cm.instance.ID())
	}
	var res osgi.ConfigPIDs
	if err = fmtx.UnmarshalJSON(bytes.NewBufferString(pids), &res); err != nil {
		return nil, fmt.Errorf("cannot parse bundle list from instance '%s': %w", cm.instance.ID(), err)
	}
	return &res, nil
}

func (cm *OSGiConfigManager) All() ([]OSGiConfig, error) {
	list, err := cm.FindAll()
	if err != nil {
		return nil, err
	}
	var result []OSGiConfig
	for _, s := range list.List {
		config := OSGiConfig{manager: cm, pid: s.PID}
		result = append(result, config)
	}
	return result, nil
}

func (cm *OSGiConfigManager) Find(pid string) (*osgi.ConfigListItem, error) {
	resp, err := cm.instance.http.Request().Get(fmt.Sprintf("%s/%s.json", ConfigMgrPath, pid))
	if err != nil {
		return nil, fmt.Errorf("cannot request config on instance '%s': %w", cm.instance.ID(), err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("cannot request config on instance '%s': %w", cm.instance.ID(), err)
	}
	var res []osgi.ConfigListItem
	if err = fmtx.UnmarshalJSON(resp.RawBody(), &res); err != nil {
		return nil, fmt.Errorf("cannot parse config from instance '%s': %w", cm.instance.ID(), err)
	}
	if len(res) > 0 {
		return &res[0], nil
	}
	return nil, nil
}

func (cm *OSGiConfigManager) FindAll() (*osgi.ConfigList, error) {
	pidList, err := cm.listPIDs()
	if err != nil {
		return nil, err
	}
	var result []osgi.ConfigListItem
	for _, pid := range pidList.PIDs {
		config, err := cm.Find(pid.ID)
		if err != nil {
			return nil, err
		}
		if config != nil {
			result = append(result, *config)
		}
	}
	return &osgi.ConfigList{List: result}, nil
}

func (cm *OSGiConfigManager) Save(pid string, props map[string]any) error {
	log.Infof("saving config '%s' on instance '%s'", pid, cm.instance.ID())
	resp, err := cm.instance.http.RequestFormData(saveConfigProps(props)).Post(fmt.Sprintf("%s/%s", ConfigMgrPath, pid))
	if err != nil {
		return fmt.Errorf("cannot save config '%s' on instance '%s': %w", pid, cm.instance.ID(), err)
	} else if resp.IsError() {
		return fmt.Errorf("cannot save config '%s' on instance '%s': %s", pid, cm.instance.ID(), resp.Status())
	}
	log.Infof("saved config '%s' on instance '%s'", pid, cm.instance.ID())
	return nil
}

func saveConfigProps(props map[string]any) map[string]any {
	result := map[string]any{}
	maps.Copy(result, props)
	maps.Copy(result, map[string]any{
		"apply":  true,
		"action": "ajaxConfigManager",
		//"$location": bundleLocation, // TODO what if skipped?
		"propertylist": strings.Join(maps.Keys(result), ","),
	})
	return result
}

func (cm *OSGiConfigManager) Delete(pid string) error {
	log.Infof("deleting config '%s' from instance '%s'", pid, cm.instance.ID())
	resp, err := cm.instance.http.Request().
		SetFormData(map[string]string{"delete": "1", "apply": "1"}).
		Post(fmt.Sprintf("%s/%s", ConfigMgrPath, pid))
	if err != nil {
		return fmt.Errorf("cannot save config '%s' on instance '%s': %w", pid, cm.instance.ID(), err)
	} else if resp.IsError() {
		return fmt.Errorf("cannot save config '%s' on instance '%s': %s", pid, cm.instance.ID(), resp.Status())
	}
	log.Infof("deleted config '%s' from instance '%s'", pid, cm.instance.ID())
	return nil
}

const (
	ConfigMgrPath = "/system/console/configMgr"
)
