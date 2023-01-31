package java

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/base"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/httpx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"os"
	"os/exec"
	"strings"
)

type Opts struct {
	baseOpts *base.Opts

	DownloadUrl        string
	HomeDir            string
	VersionConstraints version.Constraints
}

func NewOpts(baseOpts *base.Opts) *Opts {
	return &Opts{
		baseOpts: baseOpts,

		DownloadUrl:        determineDownloadUrl(),
		HomeDir:            "",
		VersionConstraints: version.MustConstraints(version.NewConstraint(">= 11, < 12")),
	}
}

type DownloadLock struct {
	URL string `yaml:"url"`
}

// TODO infer basing on GOARCH and GOOS <https://github.com/adoptium/temurin11-binaries/releases>
func determineDownloadUrl() string {
	return "https://github.com/adoptium/temurin11-binaries/releases/download/jdk-11.0.17%2B8/OpenJDK11U-jdk_x64_mac_hotspot_11.0.17_8.tar.gz"
}

func (o *Opts) downloadLock() osx.Lock[DownloadLock] {
	return osx.NewLock(o.downloadDir()+"/lock/create.yml", func() (DownloadLock, error) {
		return DownloadLock{URL: o.DownloadUrl}, nil
	})
}
func (o *Opts) downloadDir() string {
	return common.ToolDir + "/java" // TODO make it configurable
}

func (o *Opts) downloadJdkDir() string {
	return o.downloadDir() + "/jdk"
}

func (o *Opts) Prepare() error {
	if o.HomeDir != "" {
		log.Debugf("skipping preparing JDK as explicit home dir is defined '%s'", o.HomeDir)
		return nil
	}
	lock := o.downloadLock()
	check, err := lock.State()
	if err != nil {
		return err
	}
	if check.UpToDate {
		log.Debugf("existing JDK '%s' is up-to-date", check.Locked.URL)
		return nil
	}
	log.Infof("preparing new JDK '%s'", o.DownloadUrl)
	archiveFile := o.downloadDir() + "/" + httpx.FileNameFromUrl(o.DownloadUrl)
	if err := httpx.DownloadOnce(o.DownloadUrl, o.downloadDir()+"/"+archiveFile); err != nil {
		return err
	}
	_, err = filex.UnarchiveWithChanged(archiveFile, o.downloadJdkDir())
	if err != nil {
		return err
	}
	if err := lock.Lock(); err != nil {
		return err
	}
	log.Infof("prepared new SDK '%s'", o.DownloadUrl)
	return nil
}

func (o *Opts) Dir() string {
	if o.HomeDir != "" {
		return pathx.Abs(o.HomeDir)
	}
	return o.downloadJdkDir()
}

func (o *Opts) Validate() error {
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
