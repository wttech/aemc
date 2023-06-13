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
	cid     string
}

func (c OSGiConfig) PID() string {
	return c.pid
}

func (c OSGiConfig) FPID() string {
	return c.fpid
}

func (c OSGiConfig) IsFactory() bool {
	return c.fpid != "" && c.cid != ""
}

type OSGiConfigState struct {
	data *osgi.ConfigListItem

	PID        string         `yaml:"pid" json:"pid"`
	Exists     bool           `yaml:"exists" json:"exists"`
	Details    map[string]any `yaml:"details" json:"details"`
	Properties map[string]any `yaml:"properties" json:"properties"`
	FactoryPID string         `yaml:"factoryPid" json:"factoryPid"`
}

func (c OSGiConfig) State() (*OSGiConfigState, error) {
	var (
		data *osgi.ConfigListItem
		err  error
	)
	if c.IsFactory() {
		data, err = c.manager.FindFactory(c.fpid, c.cid)
		if err != nil {
			return nil, err
		}
	} else {
		data, err = c.manager.Find(c.pid)
	}
	if data == nil {
		return &OSGiConfigState{
			PID:    c.pid,
			Exists: false,
		}, nil
	}
	return &OSGiConfigState{
		data:       data,
		FactoryPID: data.FactoryPID,
		PID:        data.PID,
		Exists:     true,
		Details: map[string]any{
			"title":           data.Title,
			"description":     data.Description,
			"factoryPid":      data.FactoryPID,
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

	// TODO logic with CID should be also included here as in 'SaveWithChanged'

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
		props[osgi.CidPrefix+c.cid] = osgi.CidValue
		if state.PID != (c.fpid + "~" + c.cid) {
			err = c.manager.Save(state.PID, c.fpid, props)
		} else {
			err = c.manager.Save(osgi.FPIDDummy, c.fpid, props)
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
	props[osgi.CidPrefix+c.cid] = osgi.CidValue
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
		return fmt.Errorf("%s > config '%s' cannot be deleted as it does", c.manager.instance.ID(), c.pid)
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
