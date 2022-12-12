package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/pkg"
)

// Package represents uniquely a single package (locally or remotely)
type Package struct {
	manager *PackageManager

	PID pkg.PID
}

type PackageState struct {
	data *pkg.ListItem

	PID     string         `yaml:"pid" json:"pid"`
	Exists  bool           `yaml:"exists" json:"exists"`
	Details map[string]any `yaml:"details" json:"details"`
}

func (p Package) State() (*PackageState, error) {
	item, err := p.manager.Find(p.PID.String())
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &PackageState{
			PID:    p.PID.String(),
			Exists: false,
		}, nil
	}
	return &PackageState{
		data: item,

		PID:    p.PID.String(),
		Exists: true,
		Details: map[string]any{
			"path":      item.Path,
			"installed": item.Installed(),
			"built":     item.Built(),
			"size":      item.Size,
		},
	}, nil
}

func (p Package) Build() error {
	state, err := p.State()
	if err != nil {
		return err
	}
	if !state.Exists {
		return fmt.Errorf("package '%s' cannot be built as it does not exist on instance '%s'", p.PID.String(), p.manager.instance.ID())
	}
	return p.manager.Build(state.data.Path)
}

func (p *Package) Install() error {
	state, err := p.State()
	if err != nil {
		return err
	}
	if !state.Exists {
		return fmt.Errorf("package '%s' cannot be installed as it does not exist on instance '%s'", p.PID.String(), p.manager.instance.ID())
	}
	return p.manager.Install(state.data.Path)
}

func (p *Package) InstallWithChanged() (bool, error) {
	state, err := p.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		return false, fmt.Errorf("package '%s' cannot be installed as it does not exist on instance '%s'", p.PID.String(), p.manager.instance.ID())
	}
	if !state.data.Installed() { // TODO checksum comparison needed here
		return true, p.manager.Install(state.data.Path)
	}
	return false, nil
}

func (p *Package) Uninstall() error {
	state, err := p.State()
	if err != nil {
		return err
	}
	if !state.Exists {
		return fmt.Errorf("package '%s' cannot be uninstalled as it does not exist on instance '%s'", p.PID.String(), p.manager.instance.ID())
	}
	return p.manager.Uninstall(state.data.Path)
}

func (p *Package) UninstallWithChanged() (bool, error) {
	state, err := p.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		return false, fmt.Errorf("package '%s' cannot be uninstalled as it does not exist on instance '%s'", p.PID.String(), p.manager.instance.ID())
	}
	if state.data.Installed() {
		return true, p.manager.Uninstall(state.data.Path)
	}
	return false, nil
}

func (p Package) Delete() error {
	state, err := p.State()
	if err != nil {
		return err
	}
	if !state.Exists {
		return fmt.Errorf("package '%s' cannot be deleted as it does not exist on instance '%s'", p.PID.String(), p.manager.instance.ID())
	}
	return p.manager.Delete(state.data.Path)
}

func (p Package) DeleteWithChanged() (bool, error) {
	state, err := p.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		return false, nil
	}
	return true, p.manager.Delete(state.data.Path)
}

func (p Package) MarshalJSON() ([]byte, error) {
	state, err := p.State()
	if err != nil {
		return nil, err
	}
	return json.Marshal(state)
}

func (p Package) MarshalYAML() (interface{}, error) {
	return p.State()
}

func (p Package) MarshalText() string {
	state, err := p.State()
	if err != nil {
		return fmt.Sprintf("PID '%s' state cannot be read: %s", state.PID, err)
	}
	sb := bytes.NewBufferString("")
	if state.Exists {
		sb.WriteString(fmt.Sprintf("PID '%s'\n", p.PID.String()))
		sb.WriteString(fmtx.TblProps(state.Details))
	} else {
		sb.WriteString(fmt.Sprintf("PID '%s' cannot be found\n", p.PID.String()))
	}
	return sb.String()
}

func (p Package) String() string {
	return fmt.Sprintf("package '%s'", p.PID.String())
}
