package oak

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"sort"
	"strings"
)

type IndexList struct {
	List []IndexListItem `json:"__children__"`
}

type IndexListItem struct {
	Name                     string   `json:"__name__"`
	Type                     string   `json:"type"`
	Async                    any      `json:"async"`         /* string or []string */
	IncludedPaths            any      `json:"includedPaths"` /* string or []string */
	QueryPaths               any      `json:"queryPaths"`    /* string or []string */
	Reindex                  bool     `json:"reindex"`
	ReindexCount             int      `json:"reindexCount"`
	EvaluatePathRestrictions bool     `json:"evaluatePathRestrictions"`
	DeclaringNodeTypes       []string `json:"declaringNodeTypes"`
	PropertyNames            []string `json:"propertyNames"`
	Tags                     []string `json:"tags"`
}

func (il *IndexList) Total() int {
	return len(il.List)
}

func (il IndexList) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblMap("stats", "stat", "value", map[string]any{
		"total": il.Total(),
	}))
	bs.WriteString("\n")

	var indexesSorted []IndexListItem
	indexesSorted = append(indexesSorted, il.List...)
	sort.SliceStable(indexesSorted, func(i, j int) bool {
		return strings.Compare(indexesSorted[i].Name, indexesSorted[j].Name) < 0
	})

	bs.WriteString(fmtx.TblRows("list", false, []string{"name", "type", "async", "reindex", "reindex count", "tags"}, lo.Map(indexesSorted, func(i IndexListItem, _ int) map[string]any {
		return map[string]any{
			"name":          i.Name,
			"type":          i.Type,
			"async":         i.Async,
			"reindex":       i.Reindex,
			"reindex count": i.ReindexCount,
			"tags":          i.Tags,
		}
	})))
	return bs.String()
}

func (i IndexListItem) String() string {
	return fmt.Sprintf("index '%s' (reindex: %v)", i.Name, i.Reindex)
}
