package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/osx"
	"os/exec"
	"path/filepath"
	"time"
)

type Sdk struct {
	localOpts *LocalOpts
}

func (s Sdk) Dir() string {
	return s.localOpts.UnpackDir + "/sdk"
}

func (s Sdk) DispatcherDir() string {
	return s.Dir() + "/dispatcher"
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

	jar, err := s.QuickstartJar()
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
	err = s.unpackDispatcher()
	if err != nil {
		return err
	}
	return nil
}

func (s Sdk) unpackSdk(zipFile string) error {
	log.Infof("unpacking SDK ZIP '%s' to dir '%s'", zipFile, s.Dir())
	err := osx.ArchiveExtract(zipFile, s.Dir())
	if err != nil {
		return fmt.Errorf("cannot unpack SDK ZIP '%s' to dir '%s': %w", zipFile, s.Dir(), err)
	}
	log.Infof("unpacked SDK ZIP '%s' to dir '%s'", zipFile, s.Dir())
	return nil
}

func (s Sdk) QuickstartJar() (string, error) {
	return s.findFile("*-quickstart-*.jar")
}

func (s Sdk) dispatcherToolsUnixSh() (string, error) {
	return s.findFile("*-dispatcher-tools-*-unix.sh")
}

func (s Sdk) dispatcherToolsWindowsZip() (string, error) {
	return s.findFile("*-dispatcher-tools-*-windows.zip")
}

func (s Sdk) findFile(pattern string) (string, error) {
	paths, err := filepath.Glob(s.Dir() + "/" + pattern)
	if err != nil {
		return "", fmt.Errorf("cannot find file matching pattern '%s' in unpacked ZIP dir '%s': %w", pattern, s.Dir(), err)
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("cannot find file matching pattern '%s' in unpacked ZIP dir '%s'", pattern, s.Dir())
	}
	jar := paths[0]
	return jar, nil
}

func (s Sdk) unpackDispatcher() error {
	if osx.IsWindows() {
		zip, err := s.dispatcherToolsWindowsZip()
		if err != nil {
			return err
		}
		log.Infof("unpacking SDK dispatcher tools ZIP '%s' to dir '%s'", zip, s.DispatcherDir())
		err = osx.ArchiveExtract(zip, s.DispatcherDir())
		log.Infof("unpacked SDK dispatcher tools ZIP '%s' to dir '%s'", zip, s.DispatcherDir())
		if err != nil {
			return err
		}
	} else {
		sh, err := s.dispatcherToolsUnixSh()
		if err != nil {
			return err
		}
		log.Infof("unpacking SDK dispatcher tools using script '%s' to dir '%s'", sh, s.DispatcherDir())
		out := s.localOpts.manager.aem.output
		cmd := exec.Command(osx.ShellPath, sh, "--target", s.DispatcherDir())
		cmd.Stdout = out
		cmd.Stderr = out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("cannot run SDK dispatcher tools unpacking script '%s': %w", sh, err)
		}
		log.Infof("unpacked SDK dispatcher tools using script '%s' to dir '%s'", sh, s.DispatcherDir())
	}
	return nil
}

func (s Sdk) Destroy() error {
	return osx.PathDelete(s.Dir())
}
