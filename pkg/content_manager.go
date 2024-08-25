package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/timex"
	"github.com/wttech/aemc/pkg/content"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	NamespacePattern = "(\\\\|/)_([a-zA-Z0-9]+)_"
)

var (
	namespacePatternRegex *regexp.Regexp
)

func init() {
	namespacePatternRegex = regexp.MustCompile(NamespacePattern)
}

type ContentManager struct {
	instance *Instance
}

func NewContentManager(instance *Instance) *ContentManager {
	return &ContentManager{instance: instance}
}

func (cm *ContentManager) pkgMgr() *PackageManager {
	return cm.instance.PackageManager()
}

func (cm *ContentManager) contentManager() *content.Manager {
	return cm.instance.manager.aem.ContentManager()
}

func (cm *ContentManager) tmpDir() string {
	if cm.instance.manager.aem.Detached() {
		return os.TempDir()
	}
	return cm.instance.manager.aem.baseOpts.TmpDir
}

func (cm *ContentManager) Download(localFile string, opts PackageCreateOpts) error {
	if opts.PID == "" {
		opts.PID = fmt.Sprintf("aemc:content-download:%s-SNAPSHOT", timex.FileTimestampForNow())
	}
	remotePath, err := cm.pkgMgr().Create(opts)
	if err != nil {
		return err
	}
	defer func() { _ = cm.pkgMgr().Delete(remotePath) }()
	if err := cm.pkgMgr().Build(remotePath); err != nil {
		return err
	}
	if err := cm.pkgMgr().Download(remotePath, localFile); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) PullDir(dir string, clean bool, replace bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	pkgFile := pathx.RandomFileName(workDir, "content_pull", ".zip")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := cm.Download(pkgFile, opts); err != nil {
		return err
	}
	if err := content.Unzip(pkgFile, workDir); err != nil {
		return err
	}
	if err := pathx.Ensure(dir); err != nil {
		return err
	}
	mainDir, _, _ := strings.Cut(dir, content.JCRRoot)
	if replace {
		if err := cm.contentManager().Prepare(dir); err != nil {
			return err
		}
	}
	if err := cm.contentManager().BeforePullDir(dir); err != nil {
		return err
	}
	if err := filex.CopyDir(filepath.Join(workDir, content.JCRRoot), filepath.Join(mainDir, content.JCRRoot)); err != nil {
		return err
	}
	if err := cm.contentManager().AfterPullDir(dir); err != nil {
		return err
	}
	if clean {
		if err := cm.contentManager().CleanDir(dir); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) PullFile(file string, clean bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	pkgFile := pathx.RandomFileName(workDir, "content_pull", ".zip")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := cm.Download(pkgFile, opts); err != nil {
		return err
	}
	if err := content.Unzip(pkgFile, workDir); err != nil {
		return err
	}
	cleanFile := determineCleanFile(file)
	if pathx.Exists(file) && file != cleanFile {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	_, jcrPath, _ := strings.Cut(cleanFile, content.JCRRoot)
	if err := filex.Copy(filepath.Join(workDir, content.JCRRoot, jcrPath), cleanFile, true); err != nil {
		return err
	}
	if clean {
		if err := cm.contentManager().CleanFile(cleanFile); err != nil {
			return err
		}
	}
	return nil
}

func determineCleanFile(file string) string {
	if namespacePatternRegex.MatchString(file) && !strings.HasSuffix(file, content.JCRContentFile) {
		return filepath.Join(strings.ReplaceAll(file, content.XmlFileSuffix, ""), content.JCRContentFile)
	}
	return file
}

func (cm *ContentManager) Copy(destInstance *Instance, clean bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_copy")
	pkgFile := pathx.RandomFileName(workDir, "content_copy", ".zip")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if clean {
		if err := cm.PullDir(filepath.Join(workDir, content.JCRRoot), true, false, opts); err != nil {
			return err
		}
		if err := content.Zip(workDir, pkgFile); err != nil {
			return err
		}
	} else {
		if err := cm.Download(pkgFile, opts); err != nil {
			return err
		}
	}
	remotePath, err := destInstance.PackageManager().Upload(pkgFile)
	if err != nil {
		return err
	}
	defer func() { _ = destInstance.PackageManager().Delete(remotePath) }()
	if err = destInstance.PackageManager().Install(remotePath); err != nil {
		return err
	}
	return nil
}
