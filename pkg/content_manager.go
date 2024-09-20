package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/timex"
	"github.com/wttech/aemc/pkg/content"
	"github.com/wttech/aemc/pkg/pkg"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	NamespacePattern = "\\\\|/_[a-zA-Z0-9]+_"
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

func (cm *ContentManager) vaultCli() *VaultCli {
	return NewVaultCli(cm.instance.manager.aem)
}

func (cm *ContentManager) tmpDir() string {
	if cm.instance.manager.aem.Detached() {
		return os.TempDir()
	}
	return cm.instance.manager.aem.baseOpts.TmpDir
}

func (cm *ContentManager) Download(localFile string, vault bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_download")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if opts.PID == "" {
		opts.PID = fmt.Sprintf("aemc:content-download:%s-SNAPSHOT", timex.FileTimestampForNow())
	}
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
		if err := content.Zip(workDir, localFile); err != nil {
			return err
		}
	} else {
		remotePath, err := cm.pkgMgr().Create(opts)
		if err != nil {
			return err
		}
		defer func() { _ = cm.pkgMgr().Delete(remotePath) }()
		if err = cm.pkgMgr().Build(remotePath); err != nil {
			return err
		}
		if err = cm.pkgMgr().Download(remotePath, localFile); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) PullDir(dir string, clean bool, vault bool, replace bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_pull", ".zip")
	defer func() {
		_ = pathx.DeleteIfExists(workDir)
		_ = pathx.DeleteIfExists(pkgFile)
	}()
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
	} else {
		if err := cm.Download(pkgFile, false, opts); err != nil {
			return err
		}
		if err := content.Unzip(pkgFile, workDir); err != nil {
			return err
		}
	}
	if replace {
		if err := cm.contentManager().PrepareDir(dir); err != nil {
			return err
		}
	}
	_, jcrPath, _ := strings.Cut(dir, content.JCRRoot)
	if err := filex.CopyDir(filepath.Join(workDir, content.JCRRoot, jcrPath), dir); err != nil {
		return err
	}
	if clean {
		if err := cm.contentManager().CleanDir(dir); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) PullFile(file string, clean bool, vault bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_pull", ".zip")
	defer func() {
		_ = pathx.DeleteIfExists(workDir)
		_ = pathx.DeleteIfExists(pkgFile)
	}()
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
	} else {
		if err := cm.Download(pkgFile, false, opts); err != nil {
			return err
		}
		if err := content.Unzip(pkgFile, workDir); err != nil {
			return err
		}
	}
	cleanFile := DetermineCleanFile(file)
	if file != cleanFile {
		if err := cm.contentManager().PrepareFile(file); err != nil {
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

func (cm *ContentManager) Push(path string, clean bool, vault bool, opts PackageCreateOpts) error {
	if !pathx.Exists(path) {
		return fmt.Errorf("cannot push content as it does not exist '%s'", path)
	}
	workDir := pathx.RandomDir(cm.tmpDir(), "content_push")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if opts.PID == "" {
		opts.PID = fmt.Sprintf("aemc:content-push:%s-SNAPSHOT", timex.FileTimestampForNow())
	}
	if clean {
		if err := copyContentFiles(path, workDir); err != nil {
			return err
		}
		if err := cm.contentManager().CleanDir(filepath.Join(workDir, content.JCRRoot)); err != nil {
			return err
		}
		opts.ContentPath = filepath.Join(workDir, content.JCRRoot)
	} else {
		opts.ContentPath = path
	}
	if vault {
		mainDir, jcrPath, _ := strings.Cut(path, content.JCRRoot)
		vaultCliArgs := []string{
			"vlt",
			"--credentials", fmt.Sprintf("%s:%s", cm.instance.user, cm.instance.password),
			"import",
			fmt.Sprintf("%s/crx/-/jcr:root%s", cm.instance.http.baseURL, jcrPath),
			mainDir,
		}
		if err := cm.vaultCli().CommandShell(vaultCliArgs); err != nil {
			return err
		}
	} else {
		remotePath, err := cm.pkgMgr().Create(opts)
		if err != nil {
			return err
		}
		defer func() { _ = cm.pkgMgr().Delete(remotePath) }()
		if err = cm.pkgMgr().Install(remotePath); err != nil {
			return err
		}
	}
	return nil
}

func DetermineCleanFile(file string) string {
	if namespacePatternRegex.MatchString(file) && !strings.HasSuffix(file, content.JCRContentFile) {
		return filepath.Join(strings.ReplaceAll(file, content.XmlFileSuffix, ""), content.JCRContentFile)
	}
	return file
}

func (cm *ContentManager) Copy(destInstance *Instance, clean bool, vault bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_copy")
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_copy", ".zip")
	defer func() {
		_ = pathx.DeleteIfExists(workDir)
		_ = pathx.DeleteIfExists(pkgFile)
	}()
	if vault {
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
	} else {
		if clean {
			if err := cm.PullDir(filepath.Join(workDir, content.JCRRoot), true, false, false, opts); err != nil {
				return err
			}
			if err := content.Zip(workDir, pkgFile); err != nil {
				return err
			}
		} else {
			if err := cm.Download(pkgFile, false, opts); err != nil {
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
	}
	return nil
}
