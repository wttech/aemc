package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/osgi"
	"time"
)

type OSGiComponent struct {
	manager *OSGiComponentManager

	pid string
}

func (c OSGiComponent) PID() string {
	return c.pid
}

type OSGiComponentState struct {
	data *osgi.ComponentListItem

	PID     string         `yaml:"pid" json:"pid"`
	Exists  bool           `yaml:"exists" json:"exists"`
	Details map[string]any `yaml:"details" json:"details"`
}

func (c OSGiComponent) State() (*OSGiComponentState, error) {
	data, err := c.manager.Find(c.pid)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return &OSGiComponentState{
			PID:    c.pid,
			Exists: false,
		}, nil
	}

	return &OSGiComponentState{
		data: data,

		PID:    c.pid,
		Exists: true,
		Details: map[string]any{
			"id":           data.ID,
			"pid":          data.PID,
			"name":         data.Name,
			"bundleId":     data.BundleID,
			"state":        data.State,
			"stateRaw":     data.StateRaw,
			"configurable": data.Configurable,
			"enabled":      data.Enabled(),
		},
	}, nil
}

func (s OSGiComponentState) Enabled() bool {
	return s.data.Enabled()
}

func (s OSGiComponentState) Disabled() bool {
	return !s.Enabled()
}

func (c OSGiComponent) EnableWithChanged() (bool, error) {
	state, err := c.assumeExists()
	if err != nil {
		return false, err
	}
	if state.data.Enabled() {
		return false, nil
	}
	return true, c.manager.Enable(state.data.UID())
}

func (c OSGiComponent) Enable() error {
	state, err := c.assumeExists()
	if err != nil {
		return err
	}
	err = c.manager.Enable(state.data.UID())
	if err != nil {
		return fmt.Errorf("%s > cannot enable component '%s': %w", c.manager.instance.IDColor(), c.pid, err)
	}
	return nil
}

func (c OSGiComponent) DisableWithChanged() (bool, error) {
	state, err := c.assumeExists()
	if err != nil {
		return false, err
	}
	if !state.data.Enabled() {
		return false, nil
	}
	return true, c.manager.Disable(state.data.UID())
}

func (c OSGiComponent) Disable() error {
	state, err := c.assumeExists()
	if err != nil {
		return err
	}
	err = c.manager.Disable(state.data.UID())
	if err != nil {
		return fmt.Errorf("%s > cannot disable component '%s': %w", c.manager.instance.IDColor(), c.pid, err)
	}
	return nil
}

func (c OSGiComponent) Reenable() error {
	err := c.Disable()
	if err != nil {
		return err
	}
	err = c.AwaitDisabled()
	if err != nil {
		return err
	}
	err = c.Enable()
	if err != nil {
		return err
	}
	return nil
}

func (c OSGiComponent) assumeExists() (*OSGiComponentState, error) {
	state, err := c.State()
	if err != nil {
		return state, err
	}
	if !state.Exists {
		return state, fmt.Errorf("%s > component '%s' does not exist", c.manager.instance.IDColor(), c.pid)
	}
	return state, nil
}

func (c OSGiComponent) AwaitEnabled() error {
	return c.Await("enabled", func() bool {
		state, err := c.State()
		if err != nil {
			log.Warn(err)
			return false
		}
		return state.Enabled()
	}, time.Minute*3)
}

func (c OSGiComponent) AwaitDisabled() error {
	return c.Await("disabled", func() bool {
		state, err := c.State()
		if err != nil {
			log.Warn(err)
			return false
		}
		return state.Disabled()
	}, time.Minute*1)
}

func (c OSGiComponent) Await(state string, condition func() bool, timeout time.Duration) error {
	started := time.Now()
	for {
		if condition() {
			break
		}
		if time.Now().After(started.Add(timeout)) {
			return fmt.Errorf("%s > awaiting component '%s' state '%s' reached timeout after %s", c.manager.instance.IDColor(), c.pid, state, timeout)
		}
		log.Infof("%s > awaiting component '%s' state '%s'", c.manager.instance.IDColor(), c.pid, state)
		time.Sleep(time.Second * 5)
	}
	return nil
}

func (c OSGiComponent) String() string {
	return fmt.Sprintf("component '%s'", c.pid)
}

func (c OSGiComponent) MarshalJSON() ([]byte, error) {
	state, err := c.State()
	if err != nil {
		return nil, err
	}
	return json.Marshal(state)
}

func (c OSGiComponent) MarshalYAML() (interface{}, error) {
	return c.State()
}

func (c OSGiComponent) MarshalText() string {
	state, err := c.State()
	if err != nil {
		return fmt.Sprintf("PID '%s' state cannot be read\n", c.pid)
	}
	sb := bytes.NewBufferString("")
	if state.Exists {
		sb.WriteString(fmt.Sprintf("PID '%s'\n", c.pid))
		sb.WriteString(fmtx.TblProps(state.Details))
	} else {
		sb.WriteString(fmt.Sprintf("PID '%s' cannot be found\n", c.pid))
	}
	return sb.String()
}
