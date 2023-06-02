package pkg

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/osgi"
	"golang.org/x/exp/maps"
)

type OSGiConfigManager struct {
	instance *Instance
}

func (cm *OSGiConfigManager) ByPID(pid string) OSGiConfig {
	return OSGiConfig{manager: cm, pid: pid}
}

func (cm *OSGiConfigManager) ByFactoryPID(pid string) OSGiConfig {
	factoryPid, cid := splitPID(pid)
	return OSGiConfig{manager: cm, pid: pid, fpid: factoryPid, cid: cid}
}

func splitPID(pid string) (string, string) {
	tokens := strings.SplitN(pid, "~", 2)
	if len(tokens) > 1 {
		return tokens[0], tokens[1]
	}
	if len(tokens) == 1 {
		return tokens[0], ""
	}
	return "", ""
}

func (cm *OSGiConfigManager) listPIDs() (*osgi.ConfigPIDs, error) {
	resp, err := cm.instance.http.Request().Get(ConfigMgrPath)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot request config list: %w", cm.instance.ID(), err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot request config list: %w", cm.instance.ID(), err)
	}

	htmlBytes, err := io.ReadAll(resp.RawBody())
	if err != nil {
		return nil, fmt.Errorf("%s > cannot read config list", cm.instance.ID())
	}
	html := string(htmlBytes)
	r, _ := regexp.Compile("configData = (.*);")
	pids := strings.TrimSuffix(strings.TrimPrefix(r.FindString(html), "configData = "), ";")
	if len(pids) == 0 {
		return nil, fmt.Errorf("%s > cannot find config list in HTML response", cm.instance.ID())
	}
	var res osgi.ConfigPIDs
	if err = fmtx.UnmarshalJSON(io.NopCloser(bytes.NewBufferString(pids)), &res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse config list JSON found in HTML response: %w", cm.instance.ID(), err)
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
		config := OSGiConfig{manager: cm, pid: s.PID, fpid: s.FactoryPID}
		result = append(result, config)
	}
	return result, nil
}

func (cm *OSGiConfigManager) Find(pid string) (*osgi.ConfigListItem, error) {
	resp, err := cm.instance.http.Request().Get(fmt.Sprintf("%s/%s.json", ConfigMgrPath, pid))
	if err != nil {
		return nil, fmt.Errorf("%s > cannot find config '%s': %w", cm.instance.ID(), pid, err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot find config '%s': %s", cm.instance.ID(), pid, resp.Status())
	}
	var res []osgi.ConfigListItem
	if err = fmtx.UnmarshalJSON(resp.RawBody(), &res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse config '%s': %w", cm.instance.ID(), pid, err)
	}
	if len(res) > 0 {
		return &res[0], nil
	}
	return nil, nil
}

func (cm *OSGiConfigManager) FindByFactory(fpid string, cid string) (*osgi.ConfigListItem, error) {
	pidList, err := cm.listPIDs()
	if err != nil {
		return nil, err
	}
	for _, pid := range pidList.PIDs {
		if pid.FPID == fpid {
			config, err := cm.Find(pid.ID)
			if err != nil {
				return nil, err
			}
			if config != nil && config.ConstantId() == cid {
				return config, nil
			}
		}
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

func (cm *OSGiConfigManager) Save(pid string, fpid string, props map[string]any) error {
	log.Infof("%s > saving config '%s'", cm.instance.ID(), pid)
	resp, err := cm.instance.http.RequestFormData(saveConfigProps(fpid, props)).Post(fmt.Sprintf("%s/%s", ConfigMgrPath, pid))
	if err != nil {
		return fmt.Errorf("%s > cannot save config '%s': %w", cm.instance.ID(), pid, err)
	} else if resp.IsError() {
		return fmt.Errorf("%s > cannot save config '%s': %s", cm.instance.ID(), pid, resp.Status())
	}
	log.Infof("%s > saved config '%s'", cm.instance.ID(), pid)
	return nil
}

func saveConfigProps(fpid string, props map[string]any) map[string]any {
	result := map[string]any{}
	maps.Copy(result, props)
	maps.Copy(result, map[string]any{
		"apply":  true,
		"action": "ajaxConfigManager",
		//"$location": bundleLocation, // TODO what if skipped?
		"propertylist": strings.Join(maps.Keys(result), ","),
	})
	if fpid != "" {
		maps.Copy(result, map[string]any{
			"factoryPid": fpid,
		})
	}
	return result
}

func (cm *OSGiConfigManager) Delete(pid string) error {
	log.Infof("%s > deleting config '%s'", cm.instance.ID(), pid)
	resp, err := cm.instance.http.Request().
		SetFormData(map[string]string{"delete": "1", "apply": "1"}).
		Post(fmt.Sprintf("%s/%s", ConfigMgrPath, pid))
	if err != nil {
		return fmt.Errorf("%s > cannot save config '%s': %w", cm.instance.ID(), pid, err)
	} else if resp.IsError() {
		return fmt.Errorf("%s > cannot save config '%s': %s", cm.instance.ID(), pid, resp.Status())
	}
	log.Infof("%s > deleted config '%s'", cm.instance.ID(), pid)
	return nil
}

const (
	ConfigMgrPath = "/system/console/configMgr"
)
