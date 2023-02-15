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

type OSGiBundle struct {
	manager *OSGiBundleManager

	symbolicName string
}

func (b OSGiBundle) SymbolicName() string {
	return b.symbolicName
}

type OSGiBundleState struct {
	data *osgi.BundleListItem

	SymbolicName string         `yaml:"symbolic_name" json:"symbolicName"`
	Exists       bool           `yaml:"exists" json:"exists"`
	Details      map[string]any `yaml:"details" json:"details"`
}

func (b OSGiBundle) State() (*OSGiBundleState, error) {
	data, err := b.manager.Find(b.symbolicName)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return &OSGiBundleState{
			SymbolicName: b.symbolicName,
			Exists:       false,
		}, nil
	}

	return &OSGiBundleState{
		data: data,

		SymbolicName: b.symbolicName,
		Exists:       true,
		Details: map[string]any{
			"id":       data.ID,
			"state":    data.State,
			"category": data.Category,
			"fragment": data.Fragment,
			"version":  data.Version,
		},
	}, nil
}

func (s OSGiBundleState) Started() bool {
	return s.data.Stable()
}

func (s OSGiBundleState) Stopped() bool {
	return s.data.StateRaw == int(osgi.BundleStateRawResolved)
}

func (b OSGiBundle) StartWithChanged() (bool, error) {
	state, err := b.assumeExists()
	if err != nil {
		return false, err
	}
	if state.data.Stable() {
		return false, nil
	}
	return true, b.manager.Start(state.data.ID)
}

func (b OSGiBundle) Start() error {
	state, err := b.assumeExists()
	if err != nil {
		return err
	}
	err = b.manager.Start(state.data.ID)
	if err != nil {
		return fmt.Errorf("cannot start bundle '%s': %w", b.symbolicName, err)
	}
	return nil
}

func (b OSGiBundle) StopWithChanged() (bool, error) {
	state, err := b.assumeExists()
	if err != nil {
		return false, err
	}
	if !state.data.Stable() {
		return false, nil
	}
	return true, b.manager.Stop(state.data.ID)
}

func (b OSGiBundle) Stop() error {
	state, err := b.assumeExists()
	if err != nil {
		return err
	}
	err = b.manager.Stop(state.data.ID)
	if err != nil {
		return fmt.Errorf("cannot stop bundle '%s': %w", b.symbolicName, err)
	}
	return nil
}

func (b OSGiBundle) Restart() error {
	err := b.Stop()
	if err != nil {
		return err
	}
	err = b.AwaitStopped()
	if err != nil {
		return err
	}
	err = b.Start()
	if err != nil {
		return err
	}
	return nil
}

func (b OSGiBundle) Uninstall() error {
	state, err := b.assumeExists()
	if err != nil {
		return err
	}
	return b.manager.Uninstall(state.data.ID)
}

func (b OSGiBundle) UninstallWithChanged() (bool, error) {
	state, err := b.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		return false, nil
	}
	err = b.manager.Uninstall(state.data.ID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b OSGiBundle) assumeExists() (*OSGiBundleState, error) {
	state, err := b.State()
	if err != nil {
		return state, err
	}
	if !state.Exists {
		return state, fmt.Errorf("%s > bundle '%s' does not exist", b.manager.instance.ID(), b.symbolicName)
	}
	return state, nil
}

func (b OSGiBundle) AwaitStarted() error {
	return b.Await("started", func() bool {
		state, err := b.State()
		if err != nil {
			log.Warn(err)
			return false
		}
		return state.Started()
	}, time.Minute*3)
}

func (b OSGiBundle) AwaitStopped() error {
	return b.Await("stopped", func() bool {
		state, err := b.State()
		if err != nil {
			log.Warn(err)
			return false
		}
		return state.Stopped()
	}, time.Minute*1)
}

func (b OSGiBundle) Await(state string, condition func() bool, timeout time.Duration) error {
	started := time.Now()
	for {
		if condition() {
			break
		}
		if time.Now().After(started.Add(timeout)) {
			return fmt.Errorf("%s > awaiting bundle '%s' state '%s' reached timeout after %s", b.manager.instance.ID(), b.symbolicName, state, timeout)
		}
		log.Infof("%s > awaiting bundle '%s' state '%s'", b.manager.instance.ID(), b.symbolicName, state)
		time.Sleep(time.Second * 5)
	}
	return nil
}

func (b OSGiBundle) String() string {
	return fmt.Sprintf("bundle '%s'", b.symbolicName)
}

func (b OSGiBundle) MarshalJSON() ([]byte, error) {
	state, err := b.State()
	if err != nil {
		return nil, err
	}
	return json.Marshal(state)
}

func (b OSGiBundle) MarshalYAML() (interface{}, error) {
	return b.State()
}

func (b OSGiBundle) MarshalText() string {
	state, err := b.State()
	if err != nil {
		return fmt.Sprintf("symbolic name '%s' state cannot be read\n", b.symbolicName)
	}
	sb := bytes.NewBufferString("")
	if state.Exists {
		sb.WriteString(fmt.Sprintf("symbolic name '%s'\n", b.symbolicName))
		sb.WriteString(fmtx.TblProps(state.Details))
	} else {
		sb.WriteString(fmt.Sprintf("symbolic name '%s' cannot be found\n", b.symbolicName))
	}
	return sb.String()
}
