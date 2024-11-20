package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/httpx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
)

type JavaManager struct {
	vendorManager *VendorManager

	HomeDir                 string
	DownloadURL             string
	DownloadURLReplacements map[string]string
	VersionConstraints      string
}

func NewJavaManager(manager *VendorManager) *JavaManager {
	cv := manager.aem.Config().Values()

	return &JavaManager{
		vendorManager: manager,

		HomeDir:                 cv.GetString("vendor.java.home_dir"),
		DownloadURL:             cv.GetString("vendor.java.download.url"),
		DownloadURLReplacements: cv.GetStringMapString("vendor.java.download.replacements"),
		VersionConstraints:      cv.GetString("vendor.java.version_constraints"),
	}
}

type DownloadLock struct {
	Source string `yaml:"source"`
}

func (jm *JavaManager) toolDir() string {
	return fmt.Sprintf("%s/%s", jm.vendorManager.aem.baseOpts.ToolDir, "java")
}

func (jm *JavaManager) archiveDir() string {
	return fmt.Sprintf("%s/%s", jm.toolDir(), "archive")
}

func (jm *JavaManager) downloadLock() osx.Lock[DownloadLock] {
	return osx.NewLock(fmt.Sprintf("%s/lock/create.yml", jm.toolDir()), func() (DownloadLock, error) { return DownloadLock{Source: jm.DownloadURL}, nil })
}

func (jm *JavaManager) jdkDir() string {
	return fmt.Sprintf("%s/%s", jm.toolDir(), "jdk")
}

func (jm *JavaManager) PrepareWithChanged() (bool, error) {
	changed := false
	if jm.HomeDir == "" && jm.DownloadURL != "" {
		downloaded, err := jm.download()
		changed = downloaded
		if err != nil {
			return downloaded, err
		}
	}
	if err := jm.checkVersion(); err != nil {
		return changed, err
	}
	return changed, nil
}

func (jm *JavaManager) download() (bool, error) {
	lock := jm.downloadLock()
	check, err := lock.State()
	if err != nil {
		return false, err
	}
	if check.UpToDate {
		log.Debugf("existing JDK '%s' is up-to-date", check.Locked.Source)
		return false, nil
	}
	log.Infof("preparing new JDK at dir '%s'", jm.jdkDir())
	if err = jm.prepare(err); err != nil {
		return false, err
	}
	if err := lock.Lock(); err != nil {
		return false, err
	}
	log.Infof("prepared new JDK at dir '%s'", jm.jdkDir())
	return true, nil
}

func (jm *JavaManager) prepare(err error) error {
	if err := pathx.DeleteIfExists(jm.jdkDir()); err != nil {
		return err
	}
	url := jm.DownloadURL
	for search, replace := range jm.DownloadURLReplacements {
		url = strings.ReplaceAll(url, search, replace)
	}
	archiveFile := fmt.Sprintf("%s/%s", jm.archiveDir(), httpx.FileNameFromURL(url))
	log.Infof("downloading new JDK from URL '%s' to file '%s'", url, archiveFile)
	if err := httpx.DownloadOnce(url, archiveFile); err != nil {
		return err
	}
	log.Infof("downloaded new JDK from URL '%s' to file '%s'", url, archiveFile)
	if _, err = filex.UnarchiveWithChanged(archiveFile, jm.jdkDir()); err != nil {
		return err
	}
	return nil
}

func (jm *JavaManager) checkVersion() error {
	currentVersion, err := jm.CurrentVersion()
	if err != nil {
		return err
	}
	if jm.VersionConstraints != "" {
		versionConstraints, err := version.NewConstraint(jm.VersionConstraints)
		if err != nil {
			return fmt.Errorf("java version constraint '%s' is invalid: %w", jm.VersionConstraints, err)
		}
		if !versionConstraints.Check(currentVersion) {
			return fmt.Errorf("java current version '%s' does not meet contraints '%s'", currentVersion, jm.VersionConstraints)
		}
	}
	return nil
}

func (jm *JavaManager) FindHomeDir() (string, error) {
	var homeDir string
	if jm.HomeDir == "" {
		files, err := os.ReadDir(jm.jdkDir())
		if err != nil {
			return "", err
		}
		var dir string
		for _, file := range files {
			if file.IsDir() && strings.HasPrefix(file.Name(), "jdk") {
				dir = fmt.Sprintf("%s/%s", jm.jdkDir(), file.Name())
				break
			}
		}
		if dir == "" {
			return "", fmt.Errorf("java home dir cannot be found in unarchived JDK contents under path '%s'", jm.archiveDir())
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
		homeDir = jm.HomeDir
	}
	homeDir = pathx.Canonical(homeDir)
	if !pathx.Exists(homeDir) {
		return "", fmt.Errorf("java home dir does not exist at path '%s'", homeDir)
	}
	return homeDir, nil
}

func (jm *JavaManager) Executable() (string, error) {
	homeDir, err := jm.FindHomeDir()
	if err != nil {
		return "", err
	}
	if osx.IsWindows() {
		return pathx.Canonical(homeDir + "/bin/java.exe"), nil
	}
	return pathx.Canonical(homeDir + "/bin/java"), nil
}

func (jm *JavaManager) Env() ([]string, error) {
	homeDir, err := jm.FindHomeDir()
	if err != nil {
		return nil, err
	}
	javaPath := pathx.Canonical(homeDir + "/bin")
	envOthers := osx.EnvVarsWithout("PATH", "JAVA_HOME", "TERM")
	envFinal := append([]string{
		"PATH=" + javaPath + osx.PathVarSep() + os.Getenv("PATH"),
		"JAVA_HOME=" + homeDir,
		"TERM=xterm", // unpacking AEM jar requires this on some OSes
	}, envOthers...)
	return envFinal, nil
}

func (jm *JavaManager) Command(args ...string) (*exec.Cmd, error) {
	executable, err := jm.Executable()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(executable, args...)
	env, err := jm.Env()
	if err != nil {
		return nil, err
	}
	cmd.Env = env
	return cmd, nil
}

func (jm *JavaManager) CurrentVersion() (*version.Version, error) {
	currentText, err := jm.readCurrentVersion()
	if err != nil {
		return nil, err
	}
	current, err := version.NewVersion(currentText)
	if err != nil {
		return nil, fmt.Errorf("current java version '%s' cannot be parsed: %w", currentText, err)
	}
	return current, nil
}

func (jm *JavaManager) readCurrentVersion() (string, error) {
	cmd, err := jm.Command("-version")
	if err != nil {
		return "", err
	}
	bytes, err := cmd.CombinedOutput()
	output := string(bytes)
	if err != nil {
		log.Error(output)
		return "", fmt.Errorf("cannot check java version properly: '%s': %w", output, err)
	}
	lines := strings.Split(output, "\n")
	line, ok := lo.Find(lines, func(line string) bool { return strings.Contains(line, " version \"") })
	if !ok {
		return "", fmt.Errorf("cannot extract java version from output")
	}
	result := stringsx.Between(line, "\"", "\"")
	result = strings.Split(result, "_")[0]
	return strings.TrimSpace(result), nil
}
