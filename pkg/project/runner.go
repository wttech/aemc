package project

import (
	"fmt"
	"github.com/samber/lo"
)

// Runner is the tool that installs the project tools and runs its tasks.
type Runner string

const (
	// RunnerMise uses mise (https://mise.jdx.dev) for tools (AEM Compose, Java, Maven) and tasks; no wrapper scripts.
	RunnerMise Runner = "mise"
	// RunnerTask uses Task (https://taskfile.dev) for tasks and the wrapper scripts 'aemw', 'taskw' and 'mvnw' for tools.
	RunnerTask Runner = "task"

	RunnerDefault = RunnerMise
)

func Runners() []Runner {
	return []Runner{RunnerMise, RunnerTask}
}

func RunnerStrings() []string {
	return lo.Map(Runners(), func(r Runner, _ int) string { return string(r) })
}

func RunnerOf(name string) (Runner, error) {
	for _, r := range Runners() {
		if string(r) == name {
			return r, nil
		}
	}
	return "", fmt.Errorf("project runner '%s' is not supported", name)
}

// TaskCommand is the command line running a project task.
func (r Runner) TaskCommand(task string) string {
	if r == RunnerTask {
		return "sh taskw " + task
	}
	return "mise run " + task
}

// TaskListCommand is the command line listing the project tasks.
func (r Runner) TaskListCommand() string {
	if r == RunnerTask {
		return "sh taskw --list"
	}
	return "mise tasks"
}

// AemCommand is the command line running AEM Compose in the project.
func (r Runner) AemCommand(args string) string {
	if r == RunnerTask {
		return "sh aemw " + args
	}
	return "aem " + args
}

// GitIgnored lists the paths of the runner state to keep out of VCS.
func (r Runner) GitIgnored() []string {
	if r == RunnerTask {
		return []string{".task/"}
	}
	return []string{}
}
