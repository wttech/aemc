package tplx

import (
	"html/template"
	"reflect"
	textTemplate "text/template"
)

func New(name string) *template.Template {
	return template.New(name).Funcs(funcMap)
}

// based on <https://github.com/leekchan/gtf/blob/master/gtf.go>
var funcMap = textTemplate.FuncMap{
	"default": func(arg interface{}, value interface{}) interface{} {
		defer recovery()
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
			if v.Len() == 0 {
				return arg
			}
		case reflect.Bool:
			if !v.Bool() {
				return arg
			}
		default:
			return value
		}
		return value
	},
}

func recovery() {
	recover()
}
