// Command wrapper generates the wrapper scripts of the Task runner ('aemw', 'taskw') from one template,
// so their shared logic (platform detection, verified download, unarchiving) is written once.
// Run it with 'go generate ./pkg/project' after changing the versions in 'pkg/project/versions.go'.
package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/wttech/aemc/pkg/project"
)

// OutputDir is where the generated wrappers are placed, relative to 'pkg/project'.
const OutputDir = "runner/task/common"

//go:embed wrapper.sh.tmpl
var wrapperTemplate string

type Wrapper struct {
	File             string
	Title            string
	VersionEnvVar    string
	Version          string
	InstalledVersion string
	ReleaseURL       string
	ArchiveName      string
	ChecksumsFile    string
	Dir              string
	Bin              string
	HomeDir          string
	Exports          []string
}

func Wrappers() []Wrapper {
	return []Wrapper{
		{
			File:             "aemw",
			Title:            "AEM Compose CLI",
			VersionEnvVar:    "AEM_CLI_VERSION",
			Version:          project.AemcVersion,
			InstalledVersion: "installed",
			ReleaseURL:       "https://github.com/wttech/aemc/releases/download",
			ArchiveName:      "aemc-cli",
			ChecksumsFile:    "checksums.txt",
			Dir:              "aemc-cli",
			Bin:              "aem",
			HomeDir:          "aem/home",
		},
		{
			File:          "taskw",
			Title:         "Task CLI",
			VersionEnvVar: "TASK_VERSION",
			Version:       project.TaskVersion,
			ReleaseURL:    "https://github.com/go-task/task/releases/download",
			ArchiveName:   "task",
			ChecksumsFile: "task_checksums.txt",
			Dir:           "task",
			Bin:           "task",
			HomeDir:       "aem/home",
			Exports: []string{
				"",
				"# https://taskfile.dev/api/#env",
				"export TASK_COLOR_GREEN=35",
				"",
				"# https://taskfile.dev/experiments/env-precedence",
				"export TASK_X_ENV_PRECEDENCE=1",
			},
		},
	}
}

func Render(w Wrapper) (string, error) {
	tpl, err := template.New("wrapper").Parse(wrapperTemplate)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	if err := tpl.Execute(&out, w); err != nil {
		return "", err
	}
	return out.String(), nil
}

func main() {
	for _, w := range Wrappers() {
		content, err := Render(w)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot render wrapper '%s': %s\n", w.File, err)
			os.Exit(1)
		}
		file := filepath.Join(OutputDir, w.File)
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "cannot create dir for wrapper '%s': %s\n", file, err)
			os.Exit(1)
		}
		if err := os.WriteFile(file, []byte(content), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "cannot write wrapper '%s': %s\n", file, err)
			os.Exit(1)
		}
	}
}
