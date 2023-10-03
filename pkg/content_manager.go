package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/timex"
	"github.com/wttech/aemc/pkg/content"
	"path/filepath"
	"strings"
)

type ContentManager struct {
	instance *Instance
}

func NewContentManager(instance *Instance) *ContentManager {
	return &ContentManager{instance: instance}
}

func (cm *ContentManager) pkgMgr() *PackageManager {
	return cm.instance.PackageManager()
}

func (cm *ContentManager) tmpDir() string {
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
	defer func() {
		_ = cm.pkgMgr().Delete(remotePath)
	}()
	if err := cm.pkgMgr().Build(remotePath); err != nil {
		return err
	}
	if err := cm.pkgMgr().Download(remotePath, localFile); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) SyncDir(dir string, clean bool, packageOpts PackageCreateOpts) error {
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_sync", ".zip")
	if len(packageOpts.FilterRoots) == 0 && packageOpts.FilterFile == "" {
		packageOpts = PackageCreateOpts{
			FilterRoots: []string{strings.Split(dir, content.JCRRoot)[1]},
			FilterFile:  "",
		}
	}
	if err := cm.Download(pkgFile, packageOpts); err != nil {
		return err
	}
	workDir := pathx.RandomDir(cm.tmpDir(), "content_sync")
	defer func() {
		_ = pathx.DeleteIfExists(pkgFile)
		_ = pathx.DeleteIfExists(workDir)
	}()
	if err := filex.Unarchive(pkgFile, workDir); err != nil {
		return err
	}
	if err := pathx.Ensure(dir); err != nil {
		return err
	}
	before, _, _ := strings.Cut(dir, content.JCRRoot)
	contentManager := cm.instance.manager.aem.contentManager
	if clean {
		if err := contentManager.BeforeClean(dir); err != nil {
			return err
		}
	}
	if err := filex.CopyDir(filepath.Join(workDir, content.JCRRoot), before+content.JCRRoot); err != nil {
		return err
	}
	if clean {
		if err := contentManager.CleanDir(dir); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) SyncFile(file string, clean bool) error {
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_sync", ".zip")
	packageOpts := PackageCreateOpts{
		FilterRoots: []string{strings.ReplaceAll(file, content.JCRContentFile, content.JCRContentNode)},
		FilterFile:  "",
	}
	if err := cm.Download(pkgFile, packageOpts); err != nil {
		return err
	}
	workDir := pathx.RandomDir(cm.tmpDir(), "content_sync")
	defer func() {
		_ = pathx.DeleteIfExists(pkgFile)
		_ = pathx.DeleteIfExists(workDir)
	}()
	if err := filex.Unarchive(pkgFile, workDir); err != nil {
		return err
	}
	dir := filepath.Dir(file)
	if err := pathx.Ensure(dir); err != nil {
		return err
	}
	before, _, _ := strings.Cut(dir, content.JCRRoot)
	contentManager := cm.instance.manager.aem.contentManager
	if clean {
		if err := contentManager.BeforeClean(dir); err != nil {
			return err
		}
	}
	if err := filex.CopyDir(filepath.Join(workDir, content.JCRRoot), before+content.JCRRoot); err != nil {
		return err
	}
	if clean {
		if err := contentManager.CleanFile(file); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) PushDir(dir string, clean bool) error {
	packageOpts := PackageCreateOpts{
		ContentDirs: []string{dir},
	}
	if clean {
		contentManager := cm.instance.manager.aem.contentManager
		if err := contentManager.CleanDir(dir); err != nil {
			return err
		}
	}
	remotePath, err := cm.pkgMgr().Create(packageOpts)
	if err != nil {
		return err
	}
	defer func() {
		_ = cm.pkgMgr().Delete(remotePath)
	}()
	if err = cm.pkgMgr().Install(remotePath); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) PushFile(file string, clean bool) error {
	packageOpts := PackageCreateOpts{
		ContentFiles: []string{file},
	}
	if clean {
		contentManager := cm.instance.manager.aem.contentManager
		if err := contentManager.CleanFile(file); err != nil {
			return err
		}
	}
	remotePath, err := cm.pkgMgr().Create(packageOpts)
	if err != nil {
		return err
	}
	defer func() {
		_ = cm.pkgMgr().Delete(remotePath)
	}()
	if err = cm.pkgMgr().Install(remotePath); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) Copy(destInstance *Instance, clean bool, pkgOpts PackageCreateOpts) error {
	var pkgFile = pathx.RandomFileName(cm.tmpDir(), "content_copy", ".zip")
	defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
	if clean {
		workDir := pathx.RandomDir(cm.tmpDir(), "content_copy")
		defer func() { _ = pathx.DeleteIfExists(workDir) }()
		if err := cm.SyncDir(filepath.Join(workDir, content.JCRRoot), clean, pkgOpts); err != nil {
			return err
		}
		if err := filex.Archive(workDir, pkgFile); err != nil {
			return err
		}
	} else {
		if err := cm.Download(pkgFile, pkgOpts); err != nil {
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
