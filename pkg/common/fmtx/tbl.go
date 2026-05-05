package fmtx

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
)

func TblProps(props map[string]any) string {
	return TblMap("properties", "name", "value", props)
}

func TblList(caption string, items [][]any) string {
	sb := bytes.NewBufferString("\n")
	sb.WriteString(fmt.Sprintf("%s\n", caption))
	tbl := newTable(sb, false)
	tbl.Configure(func(cfg *tablewriter.Config) {
		cfg.Behavior.Header.Hide = tw.On
	})
	for _, item := range items {
		tbl.Append(TblValue(item[0]), TblValue(item[1]))
	}
	tbl.Render()
	sb.WriteString("\n")
	return sb.String()
}

func TblMap(caption, keyLabel, valueLabel string, props map[string]any) string {
	sb := bytes.NewBufferString("\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", caption))
	tbl := newTable(sb, true)
	tbl.Header(keyLabel, valueLabel)
	keys := maps.Keys(props)
	sort.Strings(keys)
	for _, key := range keys {
		tbl.Append(key, TblValue(props[key]))
	}
	tbl.Render()
	sb.WriteString("\n")
	return sb.String()
}

func TblRows(caption string, enum bool, header []string, rows []map[string]any) string {
	sb := bytes.NewBufferString("\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", caption))
	headerNormalized := []string{}
	if enum {
		headerNormalized = append(headerNormalized, "#")
	}
	headerNormalized = append(headerNormalized, header...)
	rowsNormalized := lo.Map(rows, func(row map[string]any, index int) []any {
		rowVals := []any{}
		if enum {
			rowVals = append(rowVals, TblValue(index+1))
		}
		for _, header := range header {
			rowVals = append(rowVals, TblValue(row[header]))
		}
		return rowVals
	})
	tbl := newTable(sb, true)
	tbl.Header(lo.ToAnySlice(headerNormalized)...)
	tbl.Bulk(rowsNormalized)
	tbl.Render()
	sb.WriteString("\n")
	return sb.String()
}

func newTable(w io.Writer, showHeaderLine bool) *tablewriter.Table {
	lines := tw.LinesNone
	if showHeaderLine {
		lines = tw.Lines{ShowHeaderLine: tw.On}
	}
	return tablewriter.NewTable(w,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders:  tw.BorderNone,
			Settings: tw.Settings{Separators: tw.SeparatorsNone, Lines: lines},
			Symbols:  tw.NewSymbols(tw.StyleASCII),
		})),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
				Formatting: tw.CellFormatting{AutoFormat: tw.On},
			},
			Row: tw.CellConfig{
				Alignment:    tw.CellAlignment{Global: tw.AlignLeft},
				ColMaxWidths: tw.CellWidth{Global: TblColWidth},
			},
		}),
	)
}

func TblValue(value any) string {
	result := ""
	if value != nil {
		rv := reflect.ValueOf(value)
		kind := rv.Type().Kind()
		if kind == reflect.Map {
			mapValues := map[string]string{}
			for _, key := range rv.MapKeys() {
				mapValue := rv.MapIndex(key)
				mapValues[key.String()] = tblValue(mapValue)
			}
			keys := maps.Keys(mapValues)
			sort.Strings(keys)
			result = strings.Join(lo.Map(keys, func(k string, index int) string {
				return fmt.Sprintf("%s = %v", k, mapValues[k])
			}), ", ")
		} else if kind == reflect.Array || kind == reflect.Slice {
			var listValue []string
			for i := 0; i < rv.Len(); i++ {
				iv := rv.Index(i).Interface()
				listValue = append(listValue, tblValue(iv))
			}
			result = strings.Join(listValue, ", ")
		} else {
			result = tblValue(value)
		}
	}
	if len(result) == 0 {
		return "<empty>"
	}
	return result
}

func tblValue(value any) string {
	return fmt.Sprintf("%v", value)
}

const (
	TblColWidth = 120
)
