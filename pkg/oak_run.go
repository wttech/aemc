package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/httpx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/tplx"
	"github.com/wttech/aemc/pkg/instance"
	"os/exec"
	"path/filepath"
)

const (
	OakRunToolDirName = "oak-run"
)

func NewOakRun(localOpts *LocalOpts) *OakRun {
	return &OakRun{
		localOpts: localOpts,

		DownloadURL: "https://repo1.maven.org/maven2/org/apache/jackrabbit/oak-run/1.44.0/oak-run-1.44.0.jar",
		StorePath:   "crx-quickstart/repository/segmentstore",
	}
}

type OakRun struct {
	localOpts *LocalOpts

	DownloadURL string
	StorePath   string
}

type OakRunLock struct {
	DownloadURL string `yaml:"download_url"`
}

func (or OakRun) Dir() string {
	return or.localOpts.ToolDir + "/" + OakRunToolDirName
}

func (or OakRun) lock() osx.Lock[OakRunLock] {
	return osx.NewLock(or.Dir()+"/lock/create.yml", func() OakRunLock { return OakRunLock{DownloadURL: or.DownloadURL} })
}

func (or OakRun) Prepare() error {
	lock := or.lock()

	upToDate, err := lock.IsUpToDate()
	if err != nil {
		return err
	}
	if upToDate {
		log.Debugf("existing instance Oak Run '%s' is up-to-date", lock.DataCurrent().DownloadURL)
		return nil
	}
	log.Infof("preparing new instance Oak Run '%s'", lock.DataCurrent().DownloadURL)
	err = or.prepare()
	if err != nil {
		return err
	}
	err = lock.Lock()
	if err != nil {
		return err
	}
	log.Infof("prepared new instance OakRun '%s'", lock.DataCurrent().DownloadURL)

	return nil
}

func (or OakRun) JarFile() string {
	return pathx.Abs(fmt.Sprintf("%s/%s", or.Dir(), filepath.Base(or.DownloadURL)))
}

func (or OakRun) prepare() error {
	jarFile := or.JarFile()
	log.Infof("downloading Oak Run JAR from URL '%s' to file '%s'", or.DownloadURL, jarFile)
	if err := httpx.DownloadOnce(or.DownloadURL, jarFile); err != nil {
		return err
	}
	log.Infof("downloaded Oak Run JAR from URL '%s' to file '%s'", or.DownloadURL, jarFile)
	return nil
}

func (or OakRun) SetPassword(instanceDir string, user string, password string) error {
	log.Infof("password setting for user '%s' on instance at dir '%s'", user, instanceDir)

	scriptFile := fmt.Sprintf("%s/%s/tmp/oak-run/set-password.groovy", instanceDir, LocalInstanceWorkDirName)
	if err := tplx.RenderFile(scriptFile, instance.OakRunSetPassword, map[string]any{"User": user, "Password": password}); err != nil {
		return err
	}
	defer func() {
		pathx.DeleteIfExists(scriptFile)
	}()
	if err := or.RunScript(instanceDir, scriptFile); err != nil {
		return err
	}
	log.Infof("password set for user '%s' on instance at dir '%s'", user, instanceDir)
	return nil
}

func (or OakRun) RunScript(instanceDir string, scriptFile string) error {
	storeDir := fmt.Sprintf("%s/%s", instanceDir, or.StorePath)
	// TODO https://issues.apache.org/jira/browse/OAK-5961 (handle JAnsi problem)
	cmd := exec.Command("java",
		"-Djava.io.tmpdir=aem/home/tmp",
		"-jar", or.JarFile(),
		"console", storeDir, "--read-write", fmt.Sprintf(":load %s", scriptFile),
	)
	or.localOpts.manager.aem.CommandOutput(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot run Oak Run script '%s': %w", scriptFile, err)
	}
	return nil
}
