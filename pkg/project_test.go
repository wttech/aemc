package pkg

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wttech/aemc/pkg/project"
	"gopkg.in/yaml.v3"
)

// TestProjectScaffold scaffolds every project kind with every runner and checks the files and that tasks refer to existing tasks.
func TestProjectScaffold(t *testing.T) {
	expectedFiles := map[project.Runner][]string{
		project.RunnerMise: {".gitignore", ".mvn/jvm.config", "aem/default/etc/aem.yml", "bin/mise", "bin/mise.cmd", "local.env", "mise.toml"},
		project.RunnerTask: {".gitignore", ".mvn/wrapper/maven-wrapper.properties", "Taskfile.yml", "aem/default/etc/aem.yml", "aemw", "local.env", "mvnw", "taskw"},
	}
	for _, runner := range project.Runners() {
		for _, kind := range project.Kinds() {
			t.Run(string(runner)+"/"+string(kind), func(t *testing.T) {
				t.Chdir(t.TempDir())
				require.NoError(t, NewProject(DefaultAEM()).Scaffold(kind, runner))

				files := scaffoldedFiles(t)
				for _, file := range expectedFiles[runner] {
					assert.Contains(t, files, file)
				}
				assert.Equal(t, runner, NewProject(DefaultAEM()).RunnerDetected())

				switch runner {
				case project.RunnerMise:
					assertMiseTaskRefs(t)
				case project.RunnerTask:
					assertTaskTaskRefs(t)
				}
			})
		}
	}
}

func scaffoldedFiles(t *testing.T) []string {
	var files []string
	require.NoError(t, filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			files = append(files, filepath.ToSlash(path))
		}
		return err
	}))
	sort.Strings(files)
	return files
}

func assertMiseTaskRefs(t *testing.T) {
	content, err := os.ReadFile("mise.toml")
	require.NoError(t, err)
	var config struct {
		Tasks map[string]struct {
			Alias   any   `toml:"alias"`
			Depends []any `toml:"depends"`
			Run     any   `toml:"run"`
		} `toml:"tasks"`
	}
	require.NoError(t, toml.Unmarshal(content, &config))
	require.NotEmpty(t, config.Tasks)
	for name, task := range config.Tasks {
		refs := append([]any{}, task.Depends...)
		if runs, ok := task.Run.([]any); ok {
			for _, run := range runs {
				if ref, ok := run.(map[string]any); ok {
					refs = append(refs, ref["task"])
				}
			}
		}
		for _, ref := range refs {
			_, exists := config.Tasks[ref.(string)]
			assert.True(t, exists, "task '%s' refers to missing task '%v'", name, ref)
		}
	}
}

var taskRefPattern = regexp.MustCompile(`task: ([\w:-]+)`)

func assertTaskTaskRefs(t *testing.T) {
	content, err := os.ReadFile(project.TaskFile)
	require.NoError(t, err)
	var config struct {
		Tasks map[string]struct {
			Aliases []string `yaml:"aliases"`
			Deps    []string `yaml:"deps"`
		} `yaml:"tasks"`
	}
	require.NoError(t, yaml.Unmarshal(content, &config))
	require.NotEmpty(t, config.Tasks)
	names := map[string]bool{}
	for name, task := range config.Tasks {
		names[name] = true
		for _, alias := range task.Aliases {
			names[alias] = true
		}
	}
	refs := []string{}
	for _, task := range config.Tasks {
		refs = append(refs, task.Deps...)
	}
	for _, match := range taskRefPattern.FindAllStringSubmatch(string(content), -1) {
		refs = append(refs, match[1])
	}
	for _, ref := range refs {
		assert.True(t, names[ref], "missing task '%s'", ref)
	}
}
