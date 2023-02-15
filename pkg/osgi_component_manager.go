package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/osgi"
)

type OSGiComponentManager struct {
	instance *Instance
}

func NewComponentManager(instance *Instance) *OSGiComponentManager {
	return &OSGiComponentManager{instance: instance}
}

func (cm *OSGiComponentManager) ByPID(pid string) OSGiComponent {
	return OSGiComponent{manager: cm, pid: pid}
}

func (cm *OSGiComponentManager) Find(pid string) (*osgi.ComponentListItem, error) {
	components, err := cm.List()
	if err != nil {
		return nil, fmt.Errorf("%s > cannot find component '%s'", cm.instance.ID(), pid)
	}
	item, found := lo.Find(components.List, func(c osgi.ComponentListItem) bool { return pid == c.UID() })
	if found {
		return &item, nil
	}
	return nil, nil
}

func (cm *OSGiComponentManager) List() (*osgi.ComponentList, error) {
	resp, err := cm.instance.http.Request().Get(ComponentsPathJson)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot request component list: %w", cm.instance.ID(), err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot request component list: %s", cm.instance.ID(), resp.Status())
	}
	var res osgi.ComponentList
	if err = fmtx.UnmarshalJSON(resp.RawBody(), &res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse component list: %w", cm.instance.ID(), err)
	}
	return &res, nil
}

func (cm *OSGiComponentManager) Enable(pid string) error {
	log.Infof("%s > enabling component '%s'", cm.instance.ID(), pid)
	response, err := cm.instance.http.Request().
		SetFormData(map[string]string{"action": "enable"}).
		Post(fmt.Sprintf("%s/%s", ComponentsPath, pid))
	if err != nil {
		return fmt.Errorf("%s > cannot enable component '%s': %w", cm.instance.ID(), pid, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot enable component '%s': %s", cm.instance.ID(), pid, response.Status())
	}
	log.Infof("%s > enabled component '%s'", cm.instance.ID(), pid)
	return nil
}

func (cm *OSGiComponentManager) Disable(pid string) error {
	log.Infof("%s > disabling component '%s'", cm.instance.ID(), pid)
	response, err := cm.instance.http.Request().
		SetFormData(map[string]string{"action": "disable"}).
		Post(fmt.Sprintf("%s/%s", ComponentsPath, pid))
	if err != nil {
		return fmt.Errorf("%s > cannot disable component '%s': %w", cm.instance.ID(), pid, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot disable component '%s': %s", cm.instance.ID(), pid, response.Status())
	}
	log.Infof("%s > disabled component '%s'", cm.instance.ID(), pid)
	return nil
}

const (
	ComponentsPath     = "/system/console/components"
	ComponentsPathJson = ComponentsPath + ".json"
)
