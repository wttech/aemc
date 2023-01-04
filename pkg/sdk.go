package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/osx"
	"path/filepath"
	"time"
)

type Sdk struct {
	localOpts *LocalOpts
}

func (s Sdk) Dir() string {
	return s.localOpts.UnpackDir + "/sdk"
}

func (s Sdk) lockFile() string {
	return s.Dir() + "/lock/create.yml"
}

type Lock struct {
	Version  string
	Unpacked time.Time
}

func (s Sdk) Prepare() error {
	zipFile := s.localOpts.Quickstart.DistFile
	versionNew := osx.FileNameWithoutExt(zipFile)
	upToDate := false
	if osx.PathExists(s.lockFile()) {
		var lock Lock
		err := fmtx.UnmarshalFile(s.lockFile(), &lock)
		if err != nil {
			return fmt.Errorf("cannot read instance SDK lock file '%s': %w", s.lockFile(), err)
		}
		upToDate = versionNew == lock.Version
	}
	if upToDate {
		log.Debugf("existing instance SDK '%s' is up-to-date", versionNew)
		return nil
	}
	log.Infof("preparing new instance SDK '%s'", versionNew)
	err := s.prepare(zipFile)
	if err != nil {
		return err
	}
	lock := Lock{Version: versionNew, Unpacked: time.Now()}
	err = fmtx.MarshalToFile(s.lockFile(), lock)
	if err != nil {
		return fmt.Errorf("cannot save instance SDK lock file '%s': %w", s.lockFile(), err)
	}
	log.Infof("prepared new instance SDK '%s'", versionNew)

	jar, err := s.Jar()
	if err != nil {
		return err
	}
	log.Debugf("found JAR '%s' in unpacked SDK '%s'", jar, versionNew)

	return nil
}

func (s Sdk) prepare(zipFile string) error {
	err := osx.PathDelete(s.Dir())
	if err != nil {
		return err
	}
	err = s.unpackSdk(zipFile)
	if err != nil {
		return err
	}
	return nil
}

func (s Sdk) unpackSdk(zipFile string) error {
	log.Debugf("unpacking SDK ZIP '%s' to dir '%s'", zipFile, s.Dir())
	err := osx.ArchiveExtract(zipFile, s.Dir())
	if err != nil {
		return fmt.Errorf("cannot unpack SDK ZIP '%s' to dir '%s': %w", zipFile, s.Dir(), err)
	}
	return nil
}

func (s Sdk) Jar() (string, error) {
	paths, err := filepath.Glob(s.Dir() + "/*.jar")
	if err != nil {
		return "", fmt.Errorf("cannot find JAR in unpacked ZIP dir '%s': %w", s.Dir(), err)
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("cannot find JAR in unpacked ZIP dir '%s'", s.Dir())
	}
	jar := paths[0]
	return jar, nil
}

func (s Sdk) Destroy() error {
	return osx.PathDelete(s.Dir())
}
