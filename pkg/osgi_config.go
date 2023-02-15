package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/mapsx"
	"github.com/wttech/aemc/pkg/osgi"
	"golang.org/x/exp/maps"
)

type OSGiConfig struct {
	manager *OSGiConfigManager
	pid     string
}

func (c OSGiConfig) Pid() string {
	return c.pid
}

type OSGiConfigState struct {
	data *osgi.ConfigListItem

	PID        string         `yaml:"pid" json:"pid"`
	Exists     bool           `yaml:"exists" json:"exists"`
	Details    map[string]any `yaml:"details" json:"details"`
	Properties map[string]any `yaml:"properties" json:"properties"`
}

func (c OSGiConfig) State() (*OSGiConfigState, error) {
	data, err := c.manager.Find(c.pid)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return &OSGiConfigState{
			PID:    c.pid,
			Exists: false,
		}, nil
	}
	return &OSGiConfigState{
		data: data,

		PID:    c.pid,
		Exists: true,
		Details: map[string]any{
			"title":           data.Title,
			"description":     data.Description,
			"bundleLocation":  data.BundleLocation,
			"serviceLocation": data.ServiceLocation,
		},
		Properties: data.PropertyValues(),
	}, nil
}

func (c OSGiConfig) Save(props map[string]any) error {
	state, err := c.State()
	if err != nil {
		return err
	}

	propsCombined := map[string]any{}
	if state.Exists {
		maps.Copy(propsCombined, state.Properties)
	}
	maps.Copy(propsCombined, props)

	return c.manager.Save(c.pid, propsCombined)
}

func (c OSGiConfig) SaveWithChanged(props map[string]any) (bool, error) {
	state, err := c.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		err = c.manager.Save(c.pid, props)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	propsBefore := maps.Clone(state.Properties)
	if mapsx.Equal(propsBefore, props) {
		return false, nil
	}
	err = c.manager.Save(c.pid, props)
	if err != nil {
		return false, err
	}
	state, err = c.State()
	if err != nil {
		return false, err
	}
	return !mapsx.Equal(propsBefore, state.Properties), nil
}

func (c OSGiConfig) Delete() error {
	state, err := c.State()
	if err != nil {
		return err
	}
	if !state.Exists {
		return fmt.Errorf("instance '%s': config '%s' cannot be deleted as it does", c.manager.instance.ID(), c.pid)
	}
	return c.manager.Delete(c.pid)
}

func (c OSGiConfig) DeleteWithChanged() (bool, error) {
	state, err := c.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		return false, nil
	}
	err = c.manager.Delete(c.pid)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c OSGiConfig) MarshalJSON() ([]byte, error) {
	state, err := c.State()
	if err != nil {
		return nil, err
	}
	return json.Marshal(state)
}

func (c OSGiConfig) MarshalYAML() (interface{}, error) {
	return c.State()
}

func (c OSGiConfig) MarshalText() string {
	state, err := c.State()
	if err != nil {
		return fmt.Sprintf("PID '%s' state cannot be read\n", c.pid)
	}
	sb := bytes.NewBufferString("")
	if state.Exists {
		sb.WriteString(fmt.Sprintf("PID '%s'\n", c.pid))
		sb.WriteString(fmtx.TblMap("details", "name", "value", c.detailsWithoutProperties(state.Details)))
		sb.WriteString(fmtx.TblProps(state.Properties))
	} else {
		sb.WriteString(fmt.Sprintf("PID '%s' cannot be found\n", c.pid))
	}
	return sb.String()
}

func (c OSGiConfig) detailsWithoutProperties(details map[string]any) map[string]any {
	result := maps.Clone(details)
	delete(result, "properties")
	return result
}
