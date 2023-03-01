package java

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
)

type Opts struct {
	baseOpts *base.Opts

	HomeDir                 string
	DownloadURL             string
	DownloadURLReplacements map[string]string
	VersionConstraints      version.Constraints
}

func NewOpts(baseOpts *base.Opts) *Opts {
	return &Opts{
		baseOpts: baseOpts,

		HomeDir:     "",
		DownloadURL: "https://github.com/adoptium/temurin11-binaries/releases/download/jdk-11.0.18%2B10/OpenJDK11U-jdk_[[.Arch]]_[[.Os]]_hotspot_11.0.18_10.[[.ArchiveExt]]",
		DownloadURLReplacements: map[string]string{
			// Map GOARCH values to be compatible with Adoptium
			"x86_64": "x64",
			"amd64":  "x64",
			"386":    "x86-32",
			// Enforce non-ARM Java as some AEM features are not working on ARM (e.g Scene7)
			"arm64":   "x64",
			"aarch64": "x64",
		},
		VersionConstraints: version.MustConstraints(version.NewConstraint(">= 11, < 12")),
	}
}

type DownloadLock struct {
	Source string `yaml:"source"`
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
	if o.HomeDir == "" && o.DownloadURL != "" {
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
	url := o.DownloadURL
	for search, replace := range o.DownloadURLReplacements {
		url = strings.ReplaceAll(url, search, replace)
	}
	archiveFile := fmt.Sprintf("%s/%s", o.archiveDir(), httpx.FileNameFromURL(url))
	log.Infof("downloading new JDK from URL '%s' to file '%s'", url, archiveFile)
	if err := httpx.DownloadOnce(url, archiveFile); err != nil {
		return err
	}
	log.Infof("downloaded new JDK from URL '%s' to file '%s'", url, archiveFile)
	if _, err = filex.UnarchiveWithChanged(archiveFile, o.jdkDir()); err != nil {
		return err
	}
	if err := lock.Lock(); err != nil {
		return err
	}
	log.Infof("unarchived new JDK at dir '%s'", o.jdkDir())
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
	homeDir = pathx.Canonical(homeDir)
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
		return pathx.Canonical(homeDir + "/bin/java.exe"), nil
	}
	return pathx.Canonical(homeDir + "/bin/java"), nil
}

func (o *Opts) Env() ([]string, error) {
	homeDir, err := o.FindHomeDir()
	if err != nil {
		return nil, err
	}
	javaPath := pathx.Canonical(homeDir + "/bin")
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
	if len(opts.Download.URL) > 0 {
		o.DownloadURL = opts.Download.URL
	}
	o.DownloadURLReplacements = opts.Download.Replacements
	if len(opts.VersionConstraints) > 0 {
		o.VersionConstraints = version.MustConstraints(version.NewConstraint(opts.VersionConstraints))
	}
}
