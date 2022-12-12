package osgi

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/stringsx"
)

type BundleList struct {
	Status  string           `json:"status"`
	Numbers []int            `json:"s"`
	List    []BundleListItem `json:"data"`
}

func (bl *BundleList) Total() int {
	return bl.Numbers[0]
}

func (bl *BundleList) Active() int {
	return bl.Numbers[1]
}

func (bl *BundleList) ActiveFragments() int {
	return bl.Numbers[2]
}

func (bl *BundleList) Resolved() int {
	return bl.Numbers[3]
}

func (bl *BundleList) Installed() int {
	return bl.Numbers[4]
}

func (bl *BundleList) StatusUnknown() bool {
	return len(bl.List) == 0
}

func (bl *BundleList) StatusLabelled() string {
	return fmt.Sprintf("%dt|%dba|%dfa|%dbr", bl.Total(), bl.Active(), bl.ActiveFragments(), bl.Resolved())
}

func (bl *BundleList) StablePercent() string {
	return stringsx.PercentExplained(bl.Total()-(bl.Resolved()+bl.Installed()), bl.Total(), 0)
}

func (bl *BundleList) FindStable() []BundleListItem {
	return lo.Filter(bl.List, func(b BundleListItem, _ int) bool { return b.Stable() })
}

func (bl *BundleList) FindUnstable() []BundleListItem {
	return lo.Filter(bl.List, func(b BundleListItem, _ int) bool { return !b.Stable() })
}

func (bl BundleList) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblMap("stats", "stat", "value", map[string]any{
		"total":     bl.Total(),
		"active":    bl.Active(),
		"fragments": bl.ActiveFragments(),
		"resolved":  bl.Resolved(),
	}))
	bs.WriteString("\n")
	bs.WriteString(fmtx.TblRows("list", []string{"symbolic name", "state", "category", "version"}, lo.Map(bl.List, func(b BundleListItem, _ int) map[string]any {
		return map[string]any{
			"symbolic name": b.SymbolicName,
			"state":         b.State,
			"category":      b.Category,
			"version":       b.Version,
		}
	})))
	return bs.String()
}

type BundleListItem struct {
	ID           int    `json:"id"`
	Fragment     bool   `json:"fragment"`
	StateRaw     int    `json:"stateRaw" yaml:"state_raw"`
	State        string `json:"state"`
	Version      string `json:"version"`
	SymbolicName string `json:"symbolicName" yaml:"symbolic_name"`
	Category     string `json:"category"`
}

func (b *BundleListItem) Stable() bool {
	if b.Fragment {
		return b.StateRaw == int(StateResolved)
	}
	return b.StateRaw == int(StateActive)
}

func (b BundleListItem) String() string {
	return fmt.Sprintf("bundle '%s' (state: %s)", b.SymbolicName, b.State)
}

type StateRaw int

const (
	StateUninstalled StateRaw = 0x00000001
	StateInstalled   StateRaw = 0x00000002
	StateResolved    StateRaw = 0x00000004
	StateStarting    StateRaw = 0x00000008
	StateStopping    StateRaw = 0x00000010
	StateActive      StateRaw = 0x00000020
	StateUnknown     StateRaw = -1
)
