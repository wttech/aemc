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
	Data *pkg.ListItem

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
		Data: item,

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
		return fmt.Errorf("%s > package '%s' cannot be built as it does not exist", p.manager.instance.ID(), p.PID.String())
	}
	return p.manager.Build(state.Data.Path)
}

func (p Package) Install() error {
	state, err := p.State()
	if err != nil {
		return err
	}
	if !state.Exists {
		return fmt.Errorf("%s > package '%s' cannot be installed as it does not exist", p.manager.instance.ID(), p.PID.String())
	}
	return p.manager.Install(state.Data.Path)
}

func (p Package) InstallWithChanged() (bool, error) {
	state, err := p.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		return false, fmt.Errorf("%s > package '%s' cannot be installed as it does not exist", p.manager.instance.ID(), p.PID.String())
	}
	if !state.Data.Installed() {
		return true, p.manager.Install(state.Data.Path)
	}
	return false, nil
}

func (p Package) Uninstall() error {
	state, err := p.State()
	if err != nil {
		return err
	}
	if !state.Exists {
		return fmt.Errorf("%s > package '%s' cannot be uninstalled as it does not exist", p.manager.instance.ID(), p.PID.String())
	}
	return p.manager.Uninstall(state.Data.Path)
}

func (p Package) UninstallWithChanged() (bool, error) {
	state, err := p.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		return false, fmt.Errorf("%s > package '%s' cannot be uninstalled as it does not exist", p.manager.instance.ID(), p.PID.String())
	}
	if state.Data.Installed() {
		return true, p.manager.Uninstall(state.Data.Path)
	}
	return false, nil
}

func (p Package) Delete() error {
	state, err := p.State()
	if err != nil {
		return err
	}
	if !state.Exists {
		return fmt.Errorf("%s > package '%s' cannot be deleted as it does not exist", p.manager.instance.ID(), p.PID.String())
	}
	return p.manager.Delete(state.Data.Path)
}

func (p Package) DeleteWithChanged() (bool, error) {
	state, err := p.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		return false, nil
	}
	return true, p.manager.Delete(state.Data.Path)
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

func (p Package) Create(rootPaths []string, filterFile string) error {
	state, err := p.State()
	if err != nil {
		return err
	}
	if state.Exists {
		return fmt.Errorf("%s > package '%s' cannot be created as it already exists", p.manager.instance.ID(), p.PID.String())
	}
	return p.manager.Create(state.PID, rootPaths, filterFile)
}

func (p Package) UpdateFilters(filters []Filter) error {
	state, err := p.State()
	if err != nil {
		return err
	}
	if !state.Exists {
		return fmt.Errorf("%s > filters for package '%s' cannot be updated as it does not exist", p.manager.instance.ID(), p.PID.String())
	}
	return p.manager.UpdateFilters(state.Data.Path, state.PID, filters)
}

func (p Package) Download(localFile string) error {
	state, err := p.State()
	if err != nil {
		return err
	}
	if !state.Exists {
		return fmt.Errorf("%s > package '%s' cannot be downloaded as it does not exist", p.manager.instance.ID(), p.PID.String())
	}
	return p.manager.Download(state.Data.Path, localFile)
}
