package tplx

import (
	"bytes"
	"fmt"
	"github.com/wttech/aemc/pkg/common/filex"
	"text/template"
)

// Delimiters of the templates embedded in AEM Compose (package and script resources).
var (
	DelimLeft  = "[["
	DelimRight = "]]"
)

func New(name string) *template.Template {
	return template.New(name)
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
