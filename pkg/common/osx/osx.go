package osx

import (
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"os"
	"runtime"
	"strings"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

func isLinux() bool {
	return !IsWindows() && !IsDarwin()
}

func EnvVarsMap() map[string]string {
	result := make(map[string]string)
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			result[e[:i]] = e[i+1:]
		}
	}
	return result
}

func EnvVarsWithout(names ...string) []string {
	var result []string
	for _, pair := range os.Environ() {
		key := stringsx.Before(pair, "=")
		if !lo.Contains(names, key) {
			result = append(result, pair)
		}
	}
	return result
}
