package tplx

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"strings"
	"text/template"
)

var (
	DelimLeft  = "[["
	DelimRight = "]]"
)

func New(name string) *template.Template {
	return template.New(name).Funcs(funcMap)
}

var funcMap = sprig.TxtFuncMap()

func init() {
	funcMap["canonicalPath"] = func(pathSegments ...string) string {
		defer recovery()
		return pathx.Canonical(strings.Join(pathSegments, "/"))
	}
}

func recovery() {
	recover()
}

func RenderKey(key string, data any) (string, error) {
	return RenderString(DelimLeft+"."+key+DelimRight, data)
}

func RenderString(tplContent string, data any) (string, error) {
	tplParsed, err := New("string-template").Delims(DelimLeft, DelimRight).Parse(tplContent)
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
