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
	"runtime"
	"strings"
)

const (
	DownloadVersion = "11.0.18+10"
)

type Opts struct {
	baseOpts *base.Opts

	HomeDir            string
	DownloadURL        string
	VersionConstraints version.Constraints
}

func NewOpts(baseOpts *base.Opts) *Opts {
	return &Opts{
		baseOpts: baseOpts,

		HomeDir:            "",
		DownloadURL:        proposeDownloadURL(),
		VersionConstraints: version.MustConstraints(version.NewConstraint(">= 11, < 12")),
	}
}

type DownloadLock struct {
	Source string `yaml:"source"`
}

func proposeDownloadURL() string {
	var ext string
	if osx.IsWindows() {
		ext = "zip"
	} else {
		ext = "tar.gz"
	}
	versionFolder := strings.ReplaceAll(DownloadVersion, "+", "%2B")
	versionSuffix := strings.ReplaceAll(DownloadVersion, "+", "_")
	os := strings.NewReplacer("darwin", "mac").Replace(runtime.GOOS)
	arch := strings.NewReplacer(
		"x86_64", "x64",
		"386", "x86-32",
		// enforce non-ARM Java as some AEM features are not working on ARM (e.g Scene 7)
		"arm64", "x64",
		"aarch64", "x64",
	).Replace(runtime.GOARCH)
	return fmt.Sprintf("https://github.com/adoptium/temurin11-binaries/releases/download/jdk-%s/OpenJDK11U-jdk_%s_%s_hotspot_%s.%s", versionFolder, arch, os, versionSuffix, ext)
}

func (o *Opts) workDir() string {
	return common.ToolDir + "/java"
}

func (o *Opts) archiveDir() string {
	return fmt.Sprintf("%s/%s", o.workDir(), "archive")
}

func (o *Opts) downloadLock() osx.Lock[DownloadLock] {
	return osx.NewLock(fmt.Sprintf("%s/lock/create.yml", o.workDir()), func() (DownloadLock, error) { return DownloadLock{Source: o.DownloadURL}, nil })
}

func (o *Opts) jdkDir() string {
	return fmt.Sprintf("%s/%s", o.workDir(), "jdk")
}

func (o *Opts) Prepare() error {
	if o.DownloadURL != "" {
		if err := o.download(); err != nil {
			return err
		}
	}
	if err := o.checkVersion(); err != nil {
		return err
	}
	return nil
}

func (o *Opts) download() error {
	lock := o.downloadLock()
	check, err := lock.State()
	if err != nil {
		return err
	}
	if check.UpToDate {
		log.Debugf("existing JDK '%s' is up-to-date", check.Locked.Source)
		return nil
	}
	log.Infof("downloading new JDK '%s'", o.DownloadURL)
	archiveFile := fmt.Sprintf("%s/%s", o.archiveDir(), httpx.FileNameFromURL(o.DownloadURL))
	if err := httpx.DownloadOnce(o.DownloadURL, archiveFile); err != nil {
		return err
	}
	_, err = filex.UnarchiveWithChanged(archiveFile, o.jdkDir())
	if err != nil {
		return err
	}
	if err := lock.Lock(); err != nil {
		return err
	}
	log.Infof("downloaded new JDK '%s'", o.DownloadURL)
	return nil
}

func (o *Opts) checkVersion() error {
	currentVersion, err := o.CurrentVersion()
	if err != nil {
		return err
	}
	if !o.VersionConstraints.Check(currentVersion) {
		return fmt.Errorf("java current version '%s' does not meet contraints '%s'", currentVersion, o.VersionConstraints)
	}
	return nil
}

func (o *Opts) FindHomeDir() (string, error) {
	var homeDir string
	if o.HomeDir == "" {
		files, err := os.ReadDir(o.jdkDir())
		if err != nil {
			return "", err
		}
		var dir string
		for _, file := range files {
			if file.IsDir() && strings.HasPrefix(file.Name(), "jdk") {
				dir = fmt.Sprintf("%s/%s", o.jdkDir(), file.Name())
				break
			}
		}
		if dir == "" {
			return "", fmt.Errorf("java home dir cannot be found in unarchived JDK contents under path '%s'", o.archiveDir())
		}
		if err != nil {
			return "", err
		}
		if osx.IsDarwin() {
			homeDir = fmt.Sprintf("%s/Contents/Home", dir)
		} else {
			homeDir = dir
		}
	} else {
		homeDir = o.HomeDir
	}
	homeDir = pathx.Abs(homeDir)
	if !pathx.Exists(homeDir) {
		return "", fmt.Errorf("java home dir does not exist at path '%s'", homeDir)
	}
	return homeDir, nil
}

func (o *Opts) Executable() (string, error) {
	homeDir, err := o.FindHomeDir()
	if err != nil {
		return "", err
	}
	if osx.IsWindows() {
		return homeDir + "/bin/java.exe", nil
	}
	return homeDir + "/bin/java", nil
}

func (o *Opts) Env() ([]string, error) {
	homeDir, err := o.FindHomeDir()
	if err != nil {
		return nil, err
	}
	javaPath := homeDir + "/bin"
	envOthers := osx.EnvVarsWithout("PATH", "JAVA_HOME")
	envFinal := append([]string{
		"PATH=" + javaPath + ":" + os.Getenv("PATH"),
		"JAVA_HOME=" + homeDir,
	}, envOthers...)
	return envFinal, nil
}

func (o *Opts) Command(args ...string) (*exec.Cmd, error) {
	executable, err := o.Executable()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(executable, args...)
	env, err := o.Env()
	if err != nil {
		return nil, err
	}
	cmd.Env = env
	return cmd, nil
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
	cmd, err := o.Command("-version")
	if err != nil {
		return "", err
	}
	bytes, err := cmd.CombinedOutput()
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
		o.HomeDir = opts.HomeDir
	}
	if len(opts.DownloadURL) > 0 {
		o.DownloadURL = opts.DownloadURL
	}
	if len(opts.VersionConstraints) > 0 {
		o.VersionConstraints = version.MustConstraints(version.NewConstraint(opts.VersionConstraints))
	}
}
