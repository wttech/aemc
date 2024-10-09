package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/httpx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/tplx"
	"github.com/wttech/aemc/pkg/instance"
	"path/filepath"
)

func NewOakRun(localOpts *LocalOpts) *OakRun {
	cv := localOpts.manager.aem.Config().Values()

	return &OakRun{
		localOpts: localOpts,

		DownloadURL: cv.GetString("instance.local.oak_run.download_url"),
		StorePath:   cv.GetString("instance.local.oak_run.store_path"),
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
	return or.localOpts.manager.aem.baseOpts.ToolDir + "/oak-run"
}

func (or OakRun) lock() osx.Lock[OakRunLock] {
	return osx.NewLock(or.Dir()+"/lock/create.yml", func() (OakRunLock, error) { return OakRunLock{DownloadURL: or.DownloadURL}, nil })
}

func (or OakRun) Prepare() error {
	lock := or.lock()
	check, err := lock.State()
	if err != nil {
		return err
	}
	if check.UpToDate {
		log.Debugf("existing OakRun '%s' is up-to-date", or.DownloadURL)
		return nil
	}
	log.Infof("preparing new OakRun '%s'", or.DownloadURL)
	err = or.prepare()
	if err != nil {
		return err
	}
	err = lock.Lock()
	if err != nil {
		return err
	}
	log.Infof("prepared new OakRun '%s'", or.DownloadURL)

	return nil
}

func (or OakRun) JarFile() string {
	return pathx.Canonical(fmt.Sprintf("%s/%s", or.Dir(), filepath.Base(or.DownloadURL)))
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
	defer func() { pathx.DeleteIfExists(scriptFile) }()
	if err := or.RunScript(instanceDir, scriptFile); err != nil {
		return err
	}
	log.Infof("password set for user '%s' on instance at dir '%s'", user, instanceDir)
	return nil
}

func (or OakRun) RunScript(instanceDir string, scriptFile string) error {
	storeDir := fmt.Sprintf("%s/%s", instanceDir, or.StorePath)
	cmd, err := or.localOpts.manager.aem.javaOpts.Command(
		"-Djava.io.tmpdir="+pathx.Canonical(or.localOpts.manager.aem.baseOpts.TmpDir),
		"-jar", or.JarFile(),
		"console", storeDir, "--read-write", fmt.Sprintf(":load %s", scriptFile),
	)
	if err != nil {
		return err
	}
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(bytes))
		return fmt.Errorf("cannot run Oak Run script '%s': %w", scriptFile, err)
	}
	return nil
}
