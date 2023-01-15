package execx

import (
	"github.com/wttech/aemc/pkg/common/osx"
	"os/exec"
	"strings"
)

func CommandShell(args []string) *exec.Cmd {
	if osx.IsWindows() {
		args = append([]string{"/C"}, args...)
		return exec.Command("cmd", args...)
	}
	return exec.Command("sh", args...)
}

func CommandString(command string) *exec.Cmd {
	return CommandLine(strings.Split(command, " "))
}

func CommandLine(command []string) *exec.Cmd {
	name := command[0]
	args := command[1:]
	return exec.Command(name, args...)
}
