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
	fpid    string
	alias   string
}

func (c OSGiConfig) PID() string {
	return c.pid
}

func (c OSGiConfig) FPID() string {
	return c.fpid
}

func (c OSGiConfig) Alias() string {
	return c.alias
}

func (c OSGiConfig) SymbolicPID() string {
	if c.pid != "" {
		return c.pid
	}
	if c.fpid != "" && c.alias != "" {
		return c.fpid + osgi.ConfigAliasSeparator + c.alias
	}
	return ""
}

type OSGiConfigState struct {
	data *osgi.ConfigListItem

	PID        string         `yaml:"pid" json:"pid"`
	FPID       string         `yaml:"fpid" json:"fpid"`
	Alias      string         `yaml:"alias" json:"alias"`
	Exists     bool           `yaml:"exists" json:"exists"`
	Details    map[string]any `yaml:"details" json:"details"`
	Properties map[string]any `yaml:"properties" json:"properties"`
}

func (c OSGiConfig) State() (*OSGiConfigState, error) {
	var (
		data *osgi.ConfigListItem
		err  error
	)
	if c.pid != "" {
		data, err = c.manager.Find(c.pid)
		if err != nil {
			return nil, err
		}
	}
	if c.fpid != "" && c.alias != "" && data == nil {
		data, err = c.manager.FindByFactory(c.fpid, c.alias)
		if err != nil {
			return nil, err
		}
	}
	if data == nil {
		return &OSGiConfigState{
			PID:    c.SymbolicPID(),
			Exists: false,
		}, nil
	}
	return &OSGiConfigState{
		data:   data,
		FPID:   data.FPID,
		PID:    data.PID,
		Alias:  data.Alias(),
		Exists: true,
		Details: map[string]any{
			"title":           data.Title,
			"description":     data.Description,
			"factoryPid":      data.FPID,
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

	return c.manager.Save(c.pid, c.fpid, propsCombined)
}

func (c OSGiConfig) SaveWithChanged(props map[string]any) (bool, error) {
	state, err := c.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		props[osgi.ConfigAliasPropPrefix+c.alias] = osgi.ConfigAliasPropValue
		if state.PID != (c.fpid + "~" + c.alias) {
			err = c.manager.Save(state.PID, c.fpid, props)
		} else {
			err = c.manager.Save(osgi.ConfigPIDPlaceholder, c.fpid, props)
		}
		if err != nil {
			return false, err
		}
		return true, nil
	}
	propsBefore := maps.Clone(state.Properties)
	if mapsx.Equal(propsBefore, props) {
		return false, nil
	}
	props[osgi.ConfigAliasPropPrefix+c.alias] = osgi.ConfigAliasPropValue
	err = c.manager.Save(state.PID, c.fpid, props)
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
		return fmt.Errorf("%s > config '%s' cannot be deleted as it does", c.manager.instance.IDColor(), c.pid)
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
		return fmt.Sprintf("PID '%s' state cannot be read\n", c.SymbolicPID())
	}
	sb := bytes.NewBufferString("")
	if state.Exists {
		sb.WriteString(fmt.Sprintf("PID '%s'\n", c.SymbolicPID()))
		sb.WriteString(fmtx.TblMap("details", "name", "value", c.detailsWithoutProperties(state.Details)))
		sb.WriteString(fmtx.TblProps(state.Properties))
	} else {
		sb.WriteString(fmt.Sprintf("PID '%s' cannot be found\n", c.SymbolicPID()))
	}
	return sb.String()
}

func (c OSGiConfig) detailsWithoutProperties(details map[string]any) map[string]any {
	result := maps.Clone(details)
	delete(result, "properties")
	return result
}
