package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/httpx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/tplx"
	"github.com/wttech/aemc/pkg/instance"
	"os/exec"
	"path/filepath"
)

const (
	OakRunSourceEmbedded = "embedded/oak-run-1.42.0.jar"
	OakRunToolDirName    = "oak-run"
)

func NewOakRun(localOpts *LocalOpts) *OakRun {
	return &OakRun{
		localOpts: localOpts,

		Source:    OakRunSourceEmbedded,
		StorePath: "crx-quickstart/repository/segmentstore",
	}
}

type OakRun struct {
	localOpts *LocalOpts

	Source    string
	StorePath string
}

type OakRunLock struct {
	Source string
}

func (or OakRun) Dir() string {
	return or.localOpts.ToolDir + "/" + OakRunToolDirName
}

func (or OakRun) lock() osx.Lock[OakRunLock] {
	return osx.NewLock(or.Dir()+"/lock/create.yml", func() OakRunLock { return OakRunLock{Source: or.Source} })
}

func (or OakRun) Prepare() error {
	lock := or.lock()

	upToDate, err := lock.IsUpToDate()
	if err != nil {
		return err
	}
	if upToDate {
		log.Debugf("existing instance Oak Run '%s' is up-to-date", lock.DataCurrent().Source)
		return nil
	}
	log.Infof("preparing new instance Oak Run '%s'", lock.DataCurrent().Source)
	err = or.prepare()
	if err != nil {
		return err
	}
	err = lock.Lock()
	if err != nil {
		return err
	}
	log.Infof("prepared new instance OakRun '%s'", lock.DataCurrent().Source)

	return nil
}

func (or OakRun) JarFile() string {
	return pathx.Abs(fmt.Sprintf("%s/%s", or.Dir(), filepath.Base(or.Source)))
}

func (or OakRun) prepare() error {
	jarFile := or.JarFile()
	if or.Source == OakRunSourceEmbedded {
		log.Infof("copying embedded Oak Run JAR to file '%s'", jarFile)
		if err := filex.Write(jarFile, instance.OakRunJar); err != nil {
			return fmt.Errorf("cannot embedded Oak Run JAR to file '%s': %s", jarFile, err)
		}
		log.Infof("copied embedded Oak Run JAR to file '%s'", jarFile)
	} else {
		log.Infof("downloading Oak Run JAR from URL '%s' to file '%s'", or.Source, jarFile)
		if err := httpx.DownloadOnce(or.Source, jarFile); err != nil {
			return err
		}
		log.Infof("downloaded Oak Run JAR from URL '%s' to file '%s'", or.Source, jarFile)
	}
	return nil
}

func (or OakRun) SetPassword(instanceDir string, user string, password string) error {
	log.Infof("password setting for user '%s' on instance at dir '%s'", user, instanceDir)
	if err := or.RunScript(instanceDir, "set-password", instance.OakRunSetPassword, map[string]any{"User": user, "Password": password}); err != nil {
		return err
	}
	log.Infof("password set for user '%s' on instance at dir '%s'", user, instanceDir)
	return nil
}

func (or OakRun) RunScript(instanceDir string, scriptName, scriptTpl string, scriptData map[string]any) error {
	scriptContent, err := tplx.RenderString(scriptTpl, scriptData)
	if err != nil {
		return err
	}
	scriptFile := fmt.Sprintf("%s/%s/tmp/oak-run/%s.groovy", instanceDir, LocalInstanceWorkDirName, scriptName)
	if err := filex.WriteString(scriptFile, scriptContent); err != nil {
		return fmt.Errorf("cannot save Oak Run script '%s': %w", scriptFile, err)
	}
	defer func() {
		pathx.DeleteIfExists(scriptFile)
	}()
	storeDir := fmt.Sprintf("%s/%s", instanceDir, or.StorePath)
	// TODO https://issues.apache.org/jira/browse/OAK-5961 (handle JAnsi problem)
	cmd := exec.Command("java", "-jar", or.JarFile(), "console", storeDir, "--read-write", fmt.Sprintf(":load %s", scriptFile))
	or.localOpts.manager.aem.CommandOutput(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot run Oak Run script '%s': %w", scriptFile, err)
	}
	return nil
}
