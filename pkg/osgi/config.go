package osgi

import (
	"bytes"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"strings"
)

type ConfigPIDs struct {
	PIDs []ConfigPID
}

type ConfigPID struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	HasConfig bool   `json:"has_config"`
	FPID      string `json:"fpid"`
	NameHint  string `json:"nameHint"`
}

type ConfigListItem struct {
	PID                  string                    `json:"pid"`
	Title                string                    `json:"title"`
	Description          string                    `json:"description"`
	Properties           map[string]map[string]any `json:"properties"`
	AdditionalProperties string                    `json:"additionalProperties"`
	FactoryPID           string                    `json:"factoryPid"`
	BundleLocation       string                    `json:"bundle_location"`
	ServiceLocation      string                    `json:"service_location"`
}

func (c ConfigListItem) PropertyValues() map[string]any {
	var result = map[string]any{}
	for k, def := range c.Properties {
		value, ok := def["value"]
		if ok {
			result[k] = value
			continue
		}
		values, ok := def["values"]
		if ok {
			result[k] = values
		}
	}
	return result
}

// CID Extracts the constant ID, which is used to find config. See FPIDDummy for explanation.
func (c ConfigListItem) CID() string {
	for _, prop := range strings.Split(c.AdditionalProperties, ",") {
		if strings.HasPrefix(prop, CidPrefix) {
			return prop[len(CidPrefix):]
		}
	}
	return ""
}

type ConfigList struct {
	List []ConfigListItem
}

func (cl ConfigList) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblRows("list", true, []string{"pid"}, lo.Map(cl.List, func(c ConfigListItem, _ int) map[string]any {
		return map[string]any{"pid": c.PID}
	})))
	return bs.String()
}

const (
	// FPIDDummy holds a special endpoint name in Apache Felix, which is used to create new factory config.
	// It is replaced by real PID upon save, so it is not possible to use it to find config later.
	// That's why we need to use CID instead.
	FPIDDummy = "[Temporary PID replaced by real PID upon save]"
	CidPrefix = "aemComposeId~"
	CidValue  = common.AppId
)
