package java

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"os"
	"os/exec"
	"strings"
)

type Opts struct {
	HomeDir            string
	VersionConstraints version.Constraints
}

func NewOpts() *Opts {
	return &Opts{
		HomeDir:            os.Getenv("JAVA_HOME"),
		VersionConstraints: version.MustConstraints(version.NewConstraint(">= 11, < 12")),
	}
}

func (o *Opts) Validate() error {
	if len(o.HomeDir) == 0 {
		return fmt.Errorf("java home dir is not set; fix it by setting config property 'java.home_dir' or environment variable 'JAVA_HOME'")
	}
	if !pathx.Exists(o.Executable()) {
		return fmt.Errorf("java executable '%s' does not exist", o.Executable())
	}
	currentVersion, err := o.CurrentVersion()
	if err != nil {
		return err
	}
	if !o.VersionConstraints.Check(currentVersion) {
		return fmt.Errorf("java current version '%s' does not meet contraints '%s'", currentVersion, o.VersionConstraints)
	}
	return nil
}

func (o *Opts) Executable() string {
	if osx.IsWindows() {
		return pathx.Normalize(o.HomeDir) + "/bin/java.exe"
	}
	return pathx.Normalize(o.HomeDir) + "/bin/java"
}

func (o *Opts) Env() []string {
	javaDir := pathx.Abs(o.HomeDir)
	javaPath := javaDir + "/bin"
	others := osx.EnvVarsWithout("PATH", "JAVA_HOME")
	return append([]string{
		"PATH=" + javaPath + ":" + os.Getenv("PATH"),
		"JAVA_HOME=" + javaDir,
	}, others...)
}

func (o *Opts) Command(args ...string) *exec.Cmd {
	cmd := exec.Command(o.Executable(), args...)
	cmd.Env = o.Env()
	return cmd
}

func (o *Opts) CurrentVersion() (*version.Version, error) {
	currentText, err := o.readCurrentVersion()
	if err != nil {
		return nil, err
	}
	current, err := version.NewVersion(currentText)
	if err != nil {
		return nil, fmt.Errorf("current java version '%s' cannot be parsed: %w", currentText, err)
	}
	return current, nil
}

func (o *Opts) readCurrentVersion() (string, error) {
	bytes, err := o.Command("-version").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cannot check java version properly: %w", err)
	}
	output := string(bytes)
	lines := strings.Split(output, "\n")
	line, ok := lo.Find(lines, func(line string) bool { return strings.Contains(line, " version \"") })
	if !ok {
		return "", fmt.Errorf("cannot extract java version from output")
	}
	result := stringsx.Between(line, "\"", "\"")
	return strings.TrimSpace(result), nil
}

func (o *Opts) Configure(config *cfg.Config) {
	opts := config.Values().Java

	if len(opts.HomeDir) > 0 {
		o.HomeDir = pathx.Normalize(opts.HomeDir)
	}
	if len(opts.VersionConstraints) > 0 {
		o.VersionConstraints = version.MustConstraints(version.NewConstraint(opts.VersionConstraints))
	}
}
