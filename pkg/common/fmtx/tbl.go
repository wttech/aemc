package fmtx

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
	"reflect"
	"sort"
	"strings"
)

func TblProps(props map[string]any) string {
	return TblMap("properties", "name", "value", props)
}

func TblList(caption string, items [][]any) string {
	sb := bytes.NewBufferString("\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", caption))
	tbl := tablewriter.NewWriter(sb)
	tbl.SetColWidth(TblColWidth)
	tbl.SetHeader([]string{})
	tbl.SetBorder(false)
	tbl.SetAlignment(tablewriter.ALIGN_LEFT)
	for _, item := range items {
		tbl.Append([]string{TblValue(item[0]), TblValue(item[1])})
	}
	tbl.Render()
	sb.WriteString("\n")
	return sb.String()
}

func TblMap(caption, keyLabel, valueLabel string, props map[string]any) string {
	sb := bytes.NewBufferString("\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", caption))
	tbl := tablewriter.NewWriter(sb)
	tbl.SetColWidth(TblColWidth)
	tbl.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tbl.SetHeader([]string{keyLabel, valueLabel})
	tbl.SetBorder(false)
	tbl.SetAlignment(tablewriter.ALIGN_LEFT)
	keys := maps.Keys(props)
	sort.Strings(keys)
	for _, key := range keys {
		tbl.Append([]string{key, TblValue(props[key])})
	}
	tbl.Render()
	sb.WriteString("\n")
	return sb.String()
}

func TblRows(caption string, header []string, rows []map[string]any) string {
	sb := bytes.NewBufferString("\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", caption))
	headerNormalized := []string{"#"}
	headerNormalized = append(headerNormalized, header...)
	rowsNormalized := lo.Map(rows, func(row map[string]any, index int) []string {
		rowVals := []string{TblValue(index + 1)}
		for _, header := range header {
			rowVals = append(rowVals, TblValue(row[header]))
		}
		return rowVals
	})
	tbl := tablewriter.NewWriter(sb)
	tbl.SetColWidth(TblColWidth)
	tbl.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tbl.SetHeader(headerNormalized)
	tbl.SetAlignment(tablewriter.ALIGN_LEFT)
	tbl.SetBorder(false)
	tbl.AppendBulk(rowsNormalized)
	tbl.Render()
	sb.WriteString("\n")
	return sb.String()
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
