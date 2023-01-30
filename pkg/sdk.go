package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/execx"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"path/filepath"
)

const (
	SdkToolDirName = "sdk"
)

func NewSdk(localOpts *LocalOpts) *Sdk {
	return &Sdk{localOpts: localOpts}
}

type Sdk struct {
	localOpts *LocalOpts
}

func (s Sdk) Dir() string {
	return s.localOpts.ToolDir + "/" + SdkToolDirName
}

type SdkLock struct {
	Version string `yaml:"version"`
}

func (s Sdk) lock(zipFile string) osx.Lock[SdkLock] {
	return osx.NewLock(s.Dir()+"/lock/create.yml", func() (SdkLock, error) {
		return SdkLock{Version: pathx.NameWithoutExt(zipFile)}, nil
	})
}

func (s Sdk) Prepare(zipFile string) error {
	lock := s.lock(zipFile)

	check, err := lock.State()
	if err != nil {
		return err
	}
	if check.UpToDate {
		log.Debugf("existing SDK '%s' is up-to-date", check.Current.Version)
		return nil
	}
	log.Infof("preparing new SDK '%s'", check.Current.Version)
	err = s.prepare(zipFile)
	if err != nil {
		return err
	}
	err = lock.Lock()
	if err != nil {
		return err
	}
	log.Infof("prepared new SDK '%s'", check.Current.Version)

	jar, err := s.QuickstartJar()
	if err != nil {
		return err
	}
	log.Debugf("found JAR '%s' in unpacked SDK '%s'", jar, check.Current.Version)

	return nil
}

func (s Sdk) prepare(zipFile string) error {
	err := pathx.Delete(s.Dir())
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
	err := filex.Unarchive(zipFile, s.Dir())
	if err != nil {
		return fmt.Errorf("cannot unpack SDK ZIP '%s' to dir '%s': %w", zipFile, s.Dir(), err)
	}
	log.Infof("unpacked SDK ZIP '%s' to dir '%s'", zipFile, s.Dir())
	return nil
}

func (s Sdk) QuickstartJar() (string, error) {
	return s.findFile("*-quickstart-*.jar")
}

func (s Sdk) DispatcherDir() string {
	return s.Dir() + "/dispatcher"
}

func (s Sdk) dispatcherToolsUnixScript() (string, error) {
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
		err = filex.Unarchive(zip, s.DispatcherDir())
		log.Infof("unpacked SDK dispatcher tools ZIP '%s' to dir '%s'", zip, s.DispatcherDir())
		if err != nil {
			return err
		}
	} else {
		script, err := s.dispatcherToolsUnixScript()
		if err != nil {
			return err
		}
		log.Infof("unpacking SDK dispatcher tools using script '%s' to dir '%s'", script, s.DispatcherDir())
		cmd := execx.CommandShell([]string{script, "--target", s.DispatcherDir()})
		s.localOpts.manager.aem.CommandOutput(cmd)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("cannot run SDK dispatcher tools unpacking script '%s': %w", script, err)
		}
		log.Infof("unpacked SDK dispatcher tools using script '%s' to dir '%s'", script, s.DispatcherDir())
	}
	return nil
}

func (s Sdk) Destroy() error {
	return pathx.Delete(s.Dir())
}
