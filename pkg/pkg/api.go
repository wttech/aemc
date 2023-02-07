package pkg

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/intsx"
)

type List struct {
	List  []ListItem `json:"results"`
	Total int        `json:"total"`
}

type ListItem struct {
	PID          string `json:"pid"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	DownloadName string `json:"downloadName" yaml:"download_name"`
	Group        string `json:"group"`
	Version      string `json:"version"`
	Size         int    `json:"size"`

	Created       int  `json:"created"`
	LastWrapped   int  `json:"lastWrapped" yaml:"last_wrapped"`
	LastUnwrapped int  `json:"lastUnwrapped" yaml:"last_unwrapped"`
	LastModified  int  `json:"lastModified" yaml:"last_modified"`
	LastUnpacked  int  `json:"lastUnpacked" yaml:"last_unpacked"`
	Resolved      bool `json:"resolved"`
}

func (pi *ListItem) Built() bool {
	return pi.LastWrapped > 0
}

func (pi *ListItem) Installed() bool {
	return pi.LastUnpacked > 0
}

func (pi *ListItem) LastTouched() int {
	return intsx.MaxOf(0, pi.Created, pi.LastModified, pi.LastWrapped)
}

func (pl List) String() string {
	return fmt.Sprintf("bundle list (total: %d)", pl.Total)
}

func (pl List) MarshalText() string {
	return fmtx.TblRows("list", false, []string{"group", "name", "version", "size", "installed", "built"}, lo.Map(pl.List, func(item ListItem, _ int) map[string]any {
		return map[string]any{
			"group":     item.Group,
			"name":      item.Name,
			"version":   item.Version,
			"size":      humanize.Bytes(uint64(item.Size)),
			"installed": item.Installed(), // TODO date or 'not yet'
			"built":     item.Built(),     // TODO date or 'not yet'
		}
	}))
}

type CommandResult struct {
	Success bool   `json:"success"`
	Message string `json:"msg"`
	Path    string `json:"path"`
}
