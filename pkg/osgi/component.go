package osgi

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/fmtx"
)

type ComponentList struct {
	Total int                 `json:"status"`
	List  []ComponentListItem `json:"data"`
}

func (cl ComponentList) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblMap("stats", "stat", "value", map[string]any{
		"total": cl.Total,
	}))
	bs.WriteString("\n")
	bs.WriteString(fmtx.TblRows("list", false, []string{"id", "pid", "state", "bundle ID"}, lo.Map(cl.List, func(c ComponentListItem, _ int) map[string]any {
		return map[string]any{
			"id":        c.ID,
			"pid":       c.PID,
			"state":     c.State,
			"bundle ID": c.BundleID,
		}
	})))
	return bs.String()
}

type ComponentListItem struct {
	ID           string `json:"id"`
	PID          string `json:"pid"`
	BundleID     int    `json:"bundleId"`
	Name         string `json:"name"`
	StateRaw     int    `json:"stateRaw" yaml:"state_raw"`
	State        string `json:"state"`
	Configurable string `json:"configurable"`
}

func (c ComponentListItem) UID() string {
	if c.PID != "" {
		return c.PID
	}
	return c.Name
}

func (c ComponentListItem) Active() bool {
	return c.StateRaw == int(ComponentStateRawActive)
}

func (c ComponentListItem) Satisfied() bool {
	return c.StateRaw == int(ComponentStateRawSatisfied)
}

func (c ComponentListItem) Unsatisfied() bool {
	return c.StateRaw == int(ComponentStateRawUnsatisfied)
}

func (c ComponentListItem) FailedActivation() bool {
	return c.StateRaw == int(ComponentStateRawFailedActivation)
}

func (c ComponentListItem) NoConfig() bool {
	return c.State == ComponentStateNoConfig
}

func (c ComponentListItem) Disabled() bool {
	return c.State == ComponentStateDisabled
}

func (c ComponentListItem) Enabled() bool {
	return !c.Disabled()
}

func (c ComponentListItem) String() string {
	return fmt.Sprintf("component '%s' (state: %s)", c.PID, c.State)
}

type ComponentStateRaw int

const (
	ComponentStateRawUnsatisfied      ComponentStateRaw = 2
	ComponentStateRawSatisfied        ComponentStateRaw = 4
	ComponentStateRawActive           ComponentStateRaw = 8
	ComponentStateRawFailedActivation ComponentStateRaw = 16
	ComponentStateRawUnknown          ComponentStateRaw = -1
)

const (
	ComponentStateActive    = "active"
	ComponentStateSatisfied = "satisfied"
	ComponentStateNoConfig  = "no config"
	ComponentStateDisabled  = "disabled"
)
