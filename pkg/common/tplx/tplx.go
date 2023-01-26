package tplx

import (
	"bytes"
	"fmt"
	"github.com/wttech/aemc/pkg/common/filex"
	"reflect"
	"text/template"
)

func New(name string) *template.Template {
	return template.New(name).Funcs(funcMap)
}

// based on <https://github.com/leekchan/gtf/blob/master/gtf.go>
var funcMap = template.FuncMap{
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

func RenderString(tplContent string, data any) (string, error) {
	tplParsed, err := New("string-template").Parse(tplContent)
	if err != nil {
		return "", err
	}
	var tplOutput bytes.Buffer
	if err := tplParsed.Execute(&tplOutput, data); err != nil {
		return "", err
	}
	return tplOutput.String(), nil
}

func RenderFile(file string, content string, data map[string]any) error {
	scriptContent, err := RenderString(content, data)
	if err != nil {
		return err
	}
	if err := filex.WriteString(file, scriptContent); err != nil {
		return fmt.Errorf("cannot render template file '%s': %w", file, err)
	}
	return nil
}
