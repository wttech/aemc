package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"strings"
)

type JMX struct {
	instance *Instance
}

const (
	JMXBeanPath = "/system/sling/monitoring/mbeans"
)

func NewJMX(instance *Instance) *JMX {
	return &JMX{instance: instance}
}

func (j JMX) ReadBean(name string, out interface{}) error {
	name = strings.ReplaceAll(name, " ", "%20")
	response, err := j.instance.http.Request().Get(JMXBeanPath + "/" + name + ".json")
	if err != nil {
		return fmt.Errorf("cannot read Sling JMX Bean '%s' on instance '%s': %w", name, j.instance.ID(), err)
	} else if response.IsError() {
		return fmt.Errorf("cannot read Sling JMX Bean '%s' on instance '%s': %s", name, j.instance.ID(), response.Status())
	}
	if err := fmtx.UnmarshalJSON(response.RawBody(), &out); err != nil {
		return fmt.Errorf("cannot read Sling JMX Bean '%s' on instance '%s'; cannot parse response: %w", name, j.instance.ID(), err)
	}
	return nil
}
