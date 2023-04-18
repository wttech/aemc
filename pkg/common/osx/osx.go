package osx

import (
	"github.com/joho/godotenv"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"os"
	"runtime"
	"strings"
)

var (
	EnvFileExt = "env"
	EnvVar     = "AEM_ENV"
	EnvLocal   = "local"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

func IsLinux() bool {
	return !IsWindows() && !IsDarwin()
}

func EnvVarsLoad() {
	name := os.Getenv(EnvVar)
	if name == "" {
		name = EnvLocal
	}
	for _, file := range []string{
		name + "." + EnvFileExt,
		"." + EnvFileExt + "." + name,
		"." + EnvFileExt,
	} {
		if pathx.Exists(file) {
			if err := godotenv.Overload(file); err != nil {
				log.Fatalf("cannot load env file '%s': %s", file, err)
			}
		}
	}
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

func LineSep() string {
	if pathx.Sep() == "\\" {
		return "\r\n"
	}
	return "\n"
}

func PathVarSep() string {
	if pathx.Sep() == "\\" {
		return ";"
	}
	return ":"
}
