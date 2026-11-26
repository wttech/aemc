package cfg

import (
	"github.com/compose-spec/compose-go/v2/template"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"os"
	"runtime"
)

// Built-in variables available in the config file, next to the environment variables.
// They are reserved: an environment variable of the same name does not override them.
const (
	VarProjectDir = "AEM_PROJECT_DIR"
	VarOs         = "AEM_OS"
	VarArch       = "AEM_ARCH"
	VarArchiveExt = "AEM_ARCHIVE_EXT"
)

func builtinVariables() map[string]string {
	archiveExt := "tar.gz"
	if osx.IsWindows() {
		archiveExt = "zip"
	}
	return map[string]string{
		VarProjectDir: pathx.Canonical("."),
		VarOs:         runtime.GOOS,
		VarArch:       runtime.GOARCH,
		VarArchiveExt: archiveExt,
	}
}

// variables resolves a variable referenced in the config file: a built-in one first, then the environment.
func variables() template.Mapping {
	builtins := builtinVariables()
	return func(name string) (string, bool) {
		if value, ok := builtins[name]; ok {
			return value, true
		}
		return os.LookupEnv(name)
	}
}
