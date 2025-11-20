package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/sdk"
	"path/filepath"
	"strings"
)

func NewSDK(vendorManager *VendorManager) *SDK {
	cv := vendorManager.aem.Config().Values()

	return &SDK{
		vendorManager: vendorManager,

		OS: cv.GetString("vendor.sdk.os"),
	}
}

type SDK struct {
	vendorManager *VendorManager

	OS string
}

func (s SDK) Dir() string {
	return fmt.Sprintf("%s/%s", s.vendorManager.aem.baseOpts.ToolDir, "sdk")
}

type SDKLock struct {
	Version string `yaml:"version"`
}

func (s SDK) lock(zipFile string) osx.Lock[SDKLock] {
	return osx.NewLock(s.Dir()+"/lock/create.yml", func() (SDKLock, error) {
		return SDKLock{Version: pathx.NameWithoutExt(zipFile)}, nil
	})
}

func (s SDK) PrepareWithChanged() (bool, error) {
	zipFile, err := s.vendorManager.quickstart.FindDistFile()
	if err != nil {
		return false, err
	}
	lock := s.lock(zipFile)
	check, err := lock.State()
	if err != nil {
		return false, err
	}
	if check.UpToDate {
		log.Debugf("existing SDK '%s' is up-to-date", zipFile)
		return false, nil
	}
	log.Infof("preparing new SDK '%s'", zipFile)
	err = s.prepare(zipFile)
	if err != nil {
		return false, err
	}
	err = lock.Lock()
	if err != nil {
		return false, err
	}
	log.Infof("prepared new SDK '%s'", zipFile)

	jar, err := s.QuickstartJar()
	if err != nil {
		return false, err
	}
	log.Debugf("found JAR '%s' in unpacked SDK '%s'", jar, zipFile)

	return true, nil
}

func (s SDK) prepare(zipFile string) error {
	if err := pathx.DeleteIfExists(s.Dir()); err != nil {
		return err
	}
	if err := s.unpackSdk(zipFile); err != nil {
		return err
	}
	if err := s.unpackDispatcher(); err != nil {
		return err
	}
	return nil
}

func (s SDK) unpackSdk(zipFile string) error {
	log.Infof("unpacking SDK ZIP '%s' to dir '%s'", zipFile, s.Dir())
	err := filex.Unarchive(zipFile, s.Dir())
	if err != nil {
		return fmt.Errorf("cannot unpack SDK ZIP '%s' to dir '%s': %w", zipFile, s.Dir(), err)
	}
	log.Infof("unpacked SDK ZIP '%s' to dir '%s'", zipFile, s.Dir())
	return nil
}

func (s SDK) QuickstartJar() (string, error) {
	return s.findFile("*-quickstart-*.jar")
}

func (s SDK) DispatcherDir() string {
	return s.Dir() + "/dispatcher"
}

func (s SDK) dispatcherToolsUnixScript() (string, error) {
	return s.findFile("*-dispatcher-tools-*-unix.sh")
}

func (s SDK) dispatcherToolsWindowsZip() (string, error) {
	return s.findFile("*-dispatcher-tools-*-windows.zip")
}

func (s SDK) findFile(pattern string) (string, error) {
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

func (s SDK) unpackDispatcher() error {
	os, err := s.determineOs()
	if err != nil {
		return err
	}

	if os == sdk.OSWindows {
		zip, err := s.dispatcherToolsWindowsZip()
		if err != nil {
			return err
		}
		log.Infof("unpacking SDK dispatcher tools ZIP '%s' to dir '%s'", zip, s.DispatcherDir())
		err = filex.Unarchive(zip, s.DispatcherDir())
		if err != nil {
			return err
		}
		log.Infof("unpacked SDK dispatcher tools ZIP '%s' to dir '%s'", zip, s.DispatcherDir())
	} else {
		script, err := s.dispatcherToolsUnixScript()
		if err != nil {
			return err
		}

		log.Infof("extracting SDK dispatcher tools from Makeself script '%s' to dir '%s'", script, s.DispatcherDir())
		if err := filex.UnarchiveMakeself(script, s.DispatcherDir()); err != nil {
			return fmt.Errorf("cannot extract SDK dispatcher tools from self-extracting script '%s': %w", script, err)
		}
		log.Infof("extracted SDK dispatcher tools from Makeself script '%s' to dir '%s'", script, s.DispatcherDir())

	}
	return nil
}

func (s SDK) determineOs() (string, error) {
	os := s.OS
	if !lo.Contains(sdk.OsTypes(), os) {
		return "", fmt.Errorf("unsupported SDK OS type '%s', supported types are: %s", os, strings.Join(sdk.OsTypes(), ", "))
	}
	if os != sdk.OSAuto {
		return os, nil
	}
	if osx.IsWindows() {
		return sdk.OSWindows, nil
	}
	return sdk.OSUnix, nil
}

func (s SDK) Destroy() error {
	return pathx.Delete(s.Dir())
}
