package pkg

import (
	_ "embed"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/httpx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/instance"
	"path/filepath"
)

const (
	OakRunSourceEmbedded = "embedded/oak-run-1.46.0.jar"
)

type OakRun struct {
	localOpts *LocalOpts

	Source string
}

type OakRunLock struct {
	Source string
}

func (or OakRun) Dir() string {
	return or.localOpts.UnpackDir + "/oak-run"
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
	return fmt.Sprintf("%s/%s", or.Dir(), filepath.Base(or.Source))
}

func (or OakRun) prepare() error {
	jarFile := or.JarFile()
	if or.Source == OakRunSourceEmbedded {
		log.Infof("copying embedded Oak Run JAR to file '%s'", jarFile)
		if err := filex.Write(jarFile, instance.CbpExecutable); err != nil {
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

func (or OakRun) SetPassword(quickstartDir string, password string) error {
	log.Infof("setting password for instance quickstart dir '%s'", quickstartDir)
	return nil
}
