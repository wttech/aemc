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
	FPID                 string                    `json:"factoryPid"`
	Title                string                    `json:"title"`
	Description          string                    `json:"description"`
	Properties           map[string]map[string]any `json:"properties"`
	AdditionalProperties string                    `json:"additionalProperties"`
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

func (c ConfigListItem) Alias() string {
	for _, prop := range strings.Split(c.AdditionalProperties, ",") {
		if strings.HasPrefix(prop, ConfigAliasPropPrefix) {
			return prop[len(ConfigAliasPropPrefix):]
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
	ConfigPIDPlaceholder  = "[Temporary PID replaced by real PID upon save]" // https://github.com/apache/felix-dev/blob/master/webconsole/src/main/java/org/apache/felix/webconsole/internal/configuration/ConfigurationUtil.java#L36
	ConfigAliasSeparator  = "~"
	ConfigAliasPropPrefix = "alias" + ConfigAliasSeparator
	ConfigAliasPropValue  = common.AppId
)
