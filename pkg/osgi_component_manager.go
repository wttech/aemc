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
		return nil, fmt.Errorf("cannot find component '%s' on instance '%s'", pid, cm.instance.ID())
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
		return nil, fmt.Errorf("cannot request component list on instance '%s': %w", cm.instance.ID(), err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("cannot request component list on instance '%s': %s", cm.instance.ID(), resp.Status())
	}
	var res osgi.ComponentList
	if err = fmtx.UnmarshalJSON(resp.RawBody(), &res); err != nil {
		return nil, fmt.Errorf("cannot parse component list from instance '%s': %w", cm.instance.ID(), err)
	}
	return &res, nil
}

func (cm *OSGiComponentManager) Enable(pid string) error {
	log.Infof("enabling component '%s' on instance '%s'", pid, cm.instance.ID())
	response, err := cm.instance.http.Request().
		SetFormData(map[string]string{"action": "enable"}).
		Post(fmt.Sprintf("%s/%s", ComponentsPath, pid))
	if err != nil {
		return fmt.Errorf("cannot enable component '%s' on instance '%s': %w", pid, cm.instance.ID(), err)
	} else if response.IsError() {
		return fmt.Errorf("cannot enable component '%s' on instance '%s': %s", pid, cm.instance.ID(), response.Status())
	}
	log.Infof("enabled component '%s' on instance '%s'", pid, cm.instance.ID())
	return nil
}

func (cm *OSGiComponentManager) Disable(pid string) error {
	log.Infof("disabling component '%s' on instance '%s'", pid, cm.instance.ID())
	response, err := cm.instance.http.Request().
		SetFormData(map[string]string{"action": "disable"}).
		Post(fmt.Sprintf("%s/%s", ComponentsPath, pid))
	if err != nil {
		return fmt.Errorf("cannot disable component '%s' on instance '%s': %w", pid, cm.instance.ID(), err)
	} else if response.IsError() {
		return fmt.Errorf("cannot disable component '%s' on instance '%s': %s", pid, cm.instance.ID(), response.Status())
	}
	log.Infof("disabled component '%s' on instance '%s'", pid, cm.instance.ID())
	return nil
}

const (
	ComponentsPath     = "/system/console/components"
	ComponentsPathJson = ComponentsPath + ".json"
)
