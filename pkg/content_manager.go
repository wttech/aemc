package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/content"
	"github.com/wttech/aemc/pkg/pkg"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	FlattenFilePattern = "[\\\\/]_[a-zA-Z0-9]+_[^\\\\/]+\\.xml$"
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

func (cm *ContentManager) contentManager() *content.Manager {
	return cm.instance.manager.aem.ContentManager()
}

func (cm *ContentManager) vaultCli() *VaultCli {
	return NewVaultCli(cm.instance.manager.aem)
}
func (cm *ContentManager) tmpDir() string {
	if cm.instance.manager.aem.Detached() {
		return os.TempDir()
	}
	return cm.instance.manager.aem.baseOpts.TmpDir
}

func (cm *ContentManager) pullContent(workDir string, vault bool, opts PackageCreateOpts) error {
	if vault {
		if err := copyPackageAllFiles(workDir, opts); err != nil {
			return err
		}
		filterFile := filepath.Join(workDir, pkg.MetaPath, pkg.VltDir, FilterXML)
		vaultCliArgs := []string{
			"vlt",
			"--credentials", fmt.Sprintf("%s:%s", cm.instance.user, cm.instance.password),
			"checkout",
			"--force",
			"--filter", filterFile,
			fmt.Sprintf("%s/crx/server/crx.default", cm.instance.http.baseURL),
			workDir,
		}
		if err := cm.vaultCli().CommandShell(vaultCliArgs); err != nil {
			return err
		}
		if err := cm.contentManager().DeleteFiles(filepath.Join(workDir, content.JCRRoot)); err != nil {
			return err
		}
	} else {
		pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_pull", ".zip")
		defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
		if err := cm.downloadByPkgMgr(pkgFile, opts); err != nil {
			return err
		}
		if err := content.Unzip(pkgFile, workDir); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) downloadByPkgMgr(localFile string, opts PackageCreateOpts) error {
	remotePath, err := cm.pkgMgr().Create(opts)
	defer func() { _ = cm.pkgMgr().Delete(remotePath) }()
	if err != nil {
		return err
	}
	if err = cm.pkgMgr().Build(remotePath); err != nil {
		return err
	}
	if err = cm.pkgMgr().Download(remotePath, localFile); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) Download(localFile string, clean bool, vault bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_download")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if !clean && !vault {
		return cm.downloadByPkgMgr(localFile, opts)
	}
	if err := cm.pullContent(workDir, vault, opts); err != nil {
		return err
	}
	if clean {
		if err := cm.contentManager().Clean(filepath.Join(workDir, content.JCRRoot)); err != nil {
			return err
		}
	}
	if err := content.Zip(workDir, localFile); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) PullDir(dir string, clean bool, vault bool, replace bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := cm.pullContent(workDir, vault, opts); err != nil {
		return err
	}
	if replace {
		if err := cm.contentManager().DeleteFiles(dir); err != nil {
			return err
		}
	}
	_, jcrPath, _ := strings.Cut(dir, content.JCRRoot)
	if err := filex.CopyDir(filepath.Join(workDir, content.JCRRoot, jcrPath), dir); err != nil {
		return err
	}
	if clean {
		if err := cm.contentManager().Clean(dir); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) PullFile(file string, clean bool, vault bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := cm.pullContent(workDir, vault, opts); err != nil {
		return err
	}
	syncFile := DetermineSyncFile(file)
	if file != syncFile {
		if err := cm.contentManager().DeleteFile(file, nil); err != nil {
			return err
		}
	}
	_, jcrPath, _ := strings.Cut(syncFile, content.JCRRoot)
	if err := filex.Copy(filepath.Join(workDir, content.JCRRoot, jcrPath), syncFile, true); err != nil {
		return err
	}
	if clean {
		if err := cm.contentManager().Clean(syncFile); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) pushContent(destInstance *Instance, vault bool, opts PackageCreateOpts) error {
	if vault {
		mainDir, _, _ := strings.Cut(opts.ContentPath, content.JCRRoot)
		jcrPath := DetermineFilterRoot(opts.ContentPath)
		vaultCliArgs := []string{
			"vlt",
			"--credentials", fmt.Sprintf("%s:%s", destInstance.user, destInstance.password),
			"import",
			fmt.Sprintf("%s/crx/-/jcr:root%s", destInstance.http.baseURL, jcrPath),
			mainDir,
		}
		if err := cm.vaultCli().CommandShell(vaultCliArgs); err != nil {
			return err
		}
	} else {
		remotePath, err := destInstance.PackageManager().Create(opts)
		defer func() { _ = destInstance.PackageManager().Delete(remotePath) }()
		if err != nil {
			return err
		}
		if err = destInstance.PackageManager().Install(remotePath); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) Push(path string, clean bool, vault bool, opts PackageCreateOpts) error {
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
			if err := cm.contentManager().Clean(filepath.Join(workDir, content.JCRRoot)); err != nil {
				return err
			}
		}
		opts.ContentPath = filepath.Join(workDir, content.JCRRoot)
	}
	if err := cm.pushContent(cm.instance, vault, opts); err != nil {
		return err
	}
	return nil
}

func (cm *ContentManager) copyByPkgMgr(destInstance *Instance, clean bool, opts PackageCreateOpts) error {
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_copy", ".zip")
	defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
	if clean {
		workDir := pathx.RandomDir(cm.tmpDir(), "content_copy")
		defer func() { _ = pathx.DeleteIfExists(workDir) }()
		if err := cm.pullContent(workDir, false, opts); err != nil {
			return err
		}
		if clean {
			if err := cm.contentManager().Clean(filepath.Join(workDir, content.JCRRoot)); err != nil {
				return err
			}
		}
		if err := content.Zip(workDir, pkgFile); err != nil {
			return err
		}
	} else {
		if err := cm.downloadByPkgMgr(pkgFile, opts); err != nil {
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

func (cm *ContentManager) copyByVaultCli(destInstance *Instance, clean bool, opts PackageCreateOpts) error {
	if clean || opts.FilterFile != "" {
		workDir := pathx.RandomDir(cm.tmpDir(), "content_copy")
		defer func() { _ = pathx.DeleteIfExists(workDir) }()
		if err := cm.pullContent(workDir, true, opts); err != nil {
			return err
		}
		if clean {
			if err := cm.contentManager().Clean(filepath.Join(workDir, content.JCRRoot)); err != nil {
				return err
			}
		}
		opts.ContentPath = filepath.Join(workDir, content.JCRRoot)
		if err := cm.pushContent(destInstance, true, opts); err != nil {
			return err
		}
	} else {
		parsedURLSrc, err := url.Parse(cm.instance.http.baseURL)
		if err != nil {
			return err
		}
		parsedURLDest, err := url.Parse(destInstance.http.baseURL)
		if err != nil {
			return err
		}
		for _, filterRoot := range opts.FilterRoots {
			vaultCliArgs := []string{
				"vlt",
				"rcp",
				"-b", "100",
				"-r",
				"-u",
				fmt.Sprintf("%s://%s:%s@%s/crx/-/jcr:root%s",
					parsedURLSrc.Scheme, cm.instance.user, cm.instance.password,
					parsedURLSrc.Host, filterRoot),
				fmt.Sprintf("%s://%s:%s@%s/crx/-/jcr:root%s",
					parsedURLDest.Scheme, destInstance.user, destInstance.password,
					parsedURLDest.Host, filterRoot),
			}
			if err = cm.vaultCli().CommandShell(vaultCliArgs); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cm *ContentManager) Copy(destInstance *Instance, clean bool, vault bool, opts PackageCreateOpts) error {
	if vault {
		return cm.copyByVaultCli(destInstance, clean, opts)
	}
	return cm.copyByPkgMgr(destInstance, clean, opts)
}

func DetermineSyncFile(file string) string {
	if regexp.MustCompile(FlattenFilePattern).MatchString(file) {
		return filepath.Join(strings.ReplaceAll(file, content.XmlFileSuffix, ""), content.JCRContentFile)
	}
	return file
}
