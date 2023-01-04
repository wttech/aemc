package osx

import (
	"os"
	"runtime"
	"strings"
)

const (
	ShellPath = "/bin/sh"
)

func EnvVars() map[string]string {
	result := make(map[string]string)
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			result[e[:i]] = e[i+1:]
		}
	}
	return result
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

func isLinux() bool {
	return !IsWindows() && !IsDarwin()
}
