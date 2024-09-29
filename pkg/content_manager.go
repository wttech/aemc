package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/content"
	"github.com/wttech/aemc/pkg/pkg"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	FlattenFilePattern = "[\\\\/]_[a-zA-Z0-9]+_[^\\\\/]+\\.xml$"
)

type ContentManager struct {
	aem            *AEM
	contentManager *content.Manager
	vaultCli       *VaultCli
}

func NewContentManager(aem *AEM) *ContentManager {
	return &ContentManager{
		aem:            aem,
		contentManager: content.NewManager(aem.baseOpts),
		vaultCli:       NewVaultCli(aem),
	}
}

func (cm *ContentManager) tmpDir() string {
	if cm.aem.Detached() {
		return os.TempDir()
	}
	return cm.aem.baseOpts.TmpDir
}

func (cm *ContentManager) Clean(path string) error {
	return cm.contentManager.Clean(path)
}

func (cm *ContentManager) pullContent(instance *Instance, workDir string, vault bool, opts PackageCreateOpts) error {
	if vault {
		if err := copyPackageAllFiles(workDir, opts); err != nil {
			return err
		}
		filterFile := filepath.Join(workDir, pkg.MetaPath, pkg.VltDir, FilterXML)
		if err := cm.vaultCli.PullContent(instance, workDir, filterFile); err != nil {
			return err
		}
		if err := cm.contentManager.DeleteFiles(filepath.Join(workDir, content.JCRRoot)); err != nil {
			return err
		}
	} else {
		pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_pull", ".zip")
		defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
		if err := cm.downloadByPkgMgr(instance, pkgFile, opts); err != nil {
			return err
		}
		if err := content.Unzip(pkgFile, workDir); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) downloadByPkgMgr(instance *Instance, localFile string, opts PackageCreateOpts) error {
	remotePath, err := instance.PackageManager().Create(opts)
	defer func() { _ = instance.PackageManager().Delete(remotePath) }()
	if err != nil {
		return err
	}
	if err = instance.PackageManager().Build(remotePath); err != nil {
		return err
	}
	if err = instance.PackageManager().Download(remotePath, localFile); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) Download(instance *Instance, localFile string, clean bool, vault bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_download")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if !clean && !vault {
		return cm.downloadByPkgMgr(instance, localFile, opts)
	}
	if err := cm.pullContent(instance, workDir, vault, opts); err != nil {
		return err
	}
	if clean {
		if err := cm.contentManager.Clean(filepath.Join(workDir, content.JCRRoot)); err != nil {
			return err
		}
	}
	if err := content.Zip(workDir, localFile); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) PullDir(instance *Instance, dir string, clean bool, vault bool, replace bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := cm.pullContent(instance, workDir, vault, opts); err != nil {
		return err
	}
	if replace {
		if err := cm.contentManager.DeleteFiles(dir); err != nil {
			return err
		}
	}
	_, jcrPath, _ := strings.Cut(dir, content.JCRRoot)
	if err := filex.CopyDir(filepath.Join(workDir, content.JCRRoot, jcrPath), dir); err != nil {
		return err
	}
	if clean {
		if err := cm.contentManager.Clean(dir); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) PullFile(instance *Instance, file string, clean bool, vault bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := cm.pullContent(instance, workDir, vault, opts); err != nil {
		return err
	}
	syncFile := DetermineSyncFile(file)
	if file != syncFile {
		if err := cm.contentManager.DeleteFile(file, nil); err != nil {
			return err
		}
	}
	_, jcrPath, _ := strings.Cut(syncFile, content.JCRRoot)
	if err := filex.Copy(filepath.Join(workDir, content.JCRRoot, jcrPath), syncFile, true); err != nil {
		return err
	}
	if clean {
		if err := cm.contentManager.Clean(syncFile); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) pushContent(instance *Instance, vault bool, opts PackageCreateOpts) error {
	if vault {
		mainDir, _, _ := strings.Cut(opts.ContentPath, content.JCRRoot)
		jcrPath := DetermineFilterRoot(opts.ContentPath)
		if err := cm.vaultCli.PushContent(instance, mainDir, jcrPath); err != nil {
			return err
		}
	} else {
		remotePath, err := instance.PackageManager().Create(opts)
		defer func() { _ = instance.PackageManager().Delete(remotePath) }()
		if err != nil {
			return err
		}
		if err = instance.PackageManager().Install(remotePath); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) Push(instance *Instance, path string, clean bool, vault bool, opts PackageCreateOpts) error {
	if !pathx.Exists(path) {
		return fmt.Errorf("cannot push content as it does not exist '%s'", path)
	}
	workDir := pathx.RandomDir(cm.tmpDir(), "content_push")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if clean || vault && pathx.IsFile(path) {
		if err := copyPackageAllFiles(workDir, opts); err != nil {
			return err
		}
		if clean {
			if err := cm.contentManager.Clean(filepath.Join(workDir, content.JCRRoot)); err != nil {
				return err
			}
		}
		opts.ContentPath = filepath.Join(workDir, content.JCRRoot)
	}
	if err := cm.pushContent(instance, vault, opts); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) copyByPkgMgr(srcInstance *Instance, destInstance *Instance, clean bool, opts PackageCreateOpts) error {
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_copy", ".zip")
	defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
	if clean {
		workDir := pathx.RandomDir(cm.tmpDir(), "content_copy")
		defer func() { _ = pathx.DeleteIfExists(workDir) }()
		if err := cm.pullContent(srcInstance, workDir, false, opts); err != nil {
			return err
		}
		if clean {
			if err := cm.contentManager.Clean(filepath.Join(workDir, content.JCRRoot)); err != nil {
				return err
			}
		}
		if err := content.Zip(workDir, pkgFile); err != nil {
			return err
		}
	} else {
		if err := cm.downloadByPkgMgr(srcInstance, pkgFile, opts); err != nil {
			return err
		}
	}
	remotePath, err := destInstance.PackageManager().Upload(pkgFile)
	defer func() { _ = destInstance.PackageManager().Delete(remotePath) }()
	if err != nil {
		return err
	}
	if err = destInstance.PackageManager().Install(remotePath); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) copyByVaultCli(srcInstance *Instance, destInstance *Instance, clean bool, rcpArgs string, opts PackageCreateOpts) error {
	if clean || opts.FilterFile != "" {
		workDir := pathx.RandomDir(cm.tmpDir(), "content_copy")
		defer func() { _ = pathx.DeleteIfExists(workDir) }()
		if err := cm.pullContent(srcInstance, workDir, true, opts); err != nil {
			return err
		}
		if clean {
			if err := cm.contentManager.Clean(filepath.Join(workDir, content.JCRRoot)); err != nil {
				return err
			}
		}
		opts.ContentPath = filepath.Join(workDir, content.JCRRoot)
		if err := cm.pushContent(destInstance, true, opts); err != nil {
			return err
		}
	} else {
		if rcpArgs == "" {
			rcpArgs = "-b 100 -r -u"
		}
		for _, filterRoot := range opts.FilterRoots {
			if err := cm.vaultCli.CopyContent(srcInstance, destInstance, strings.Fields(rcpArgs), filterRoot); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cm *ContentManager) Copy(srcInstance *Instance, destInstance *Instance, clean bool, vault bool, rcpArgs string, opts PackageCreateOpts) error {
	if vault {
		return cm.copyByVaultCli(srcInstance, destInstance, clean, rcpArgs, opts)
	}
	return cm.copyByPkgMgr(srcInstance, destInstance, clean, opts)
}

func DetermineSyncFile(file string) string {
	if regexp.MustCompile(FlattenFilePattern).MatchString(file) {
		return filepath.Join(strings.ReplaceAll(file, content.XmlFileSuffix, ""), content.JCRContentFile)
	}
	return file
}
