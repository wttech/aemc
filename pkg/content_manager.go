package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/timex"
	"github.com/wttech/aemc/pkg/common/tplx"
	"github.com/wttech/aemc/pkg/content"
	"github.com/wttech/aemc/pkg/pkg"
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

func (cm *ContentManager) vaultCli() *VaultCli {
	return NewVaultCli(cm.instance.manager.aem)
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

func (cm *ContentManager) PullDir(dir string, clean bool, vault bool, replace bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if vault {
		if err := cm.executeVaultCliCommand("checkout", workDir, opts); err != nil {
			return err
		}
	} else {
		pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_pull", ".zip")
		defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
		if err := cm.Download(pkgFile, opts); err != nil {
			return err
		}
		if err := content.Unzip(pkgFile, workDir); err != nil {
			return err
		}
	}
	if err := pathx.Ensure(dir); err != nil {
		return err
	}
	mainDir, _, _ := strings.Cut(dir, content.JCRRoot)
	contentManager := cm.instance.manager.aem.contentManager
	if replace {
		if err := contentManager.Prepare(dir); err != nil {
			return err
		}
	}
	if err := contentManager.BeforePullDir(dir); err != nil {
		return err
	}
	if err := filex.CopyDir(filepath.Join(workDir, content.JCRRoot), filepath.Join(mainDir, content.JCRRoot)); err != nil {
		return err
	}
	if err := contentManager.AfterPullDir(dir); err != nil {
		return err
	}
	if clean {
		if err := contentManager.CleanDir(dir); err != nil {
			return err
		}
	}
	return nil
}

func (cm *ContentManager) PullFile(file string, clean bool, vault bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if vault {
		if err := cm.executeVaultCliCommand("checkout", workDir, opts); err != nil {
			return err
		}
	} else {
		pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_pull", ".zip")
		defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
		if err := cm.Download(pkgFile, opts); err != nil {
			return err
		}
		if err := content.Unzip(pkgFile, workDir); err != nil {
			return err
		}
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
		contentManager := cm.instance.manager.aem.contentManager
		if err := contentManager.CleanFile(cleanFile); err != nil {
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
		contentManager := cm.instance.manager.aem.contentManager
		if err := contentManager.CleanDir(filepath.Join(workDir, content.JCRRoot)); err != nil {
			return err
		}
		opts.ContentPath = filepath.Join(workDir, content.JCRRoot)
	} else {
		opts.ContentPath = path
	}
	if vault {
		// TODO implement Vault-Cli command
		if err := cm.executeVaultCliCommand("commit", workDir, opts); err != nil {
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

func determineCleanFile(file string) string {
	if namespacePatternRegex.MatchString(file) && !strings.HasSuffix(file, content.JCRContentFile) {
		return filepath.Join(strings.ReplaceAll(file, content.XmlFileSuffix, ""), content.JCRContentFile)
	}
	return file
}

func (cm *ContentManager) Copy(destInstance *Instance, clean bool, vault bool, opts PackageCreateOpts) error {
	workDir := pathx.RandomDir(cm.tmpDir(), "content_copy")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if vault {
		// TODO implement Vault-Cli command
		if err := cm.executeVaultCliCommand("", workDir, opts); err != nil {
			return err
		}
	} else {
		pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_copy", ".zip")
		defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
		if clean {
			if err := cm.PullDir(filepath.Join(workDir, content.JCRRoot), true, false, false, opts); err != nil {
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
	}
	return nil
}

func (cm *ContentManager) determineTmpFilterFile(workDir string, opts PackageCreateOpts) (string, error) {
	tmpFilterFile := filepath.Join(workDir, FilterXML)
	if opts.FilterFile != "" {
		if err := filex.Copy(opts.FilterFile, tmpFilterFile, true); err != nil {
			return tmpFilterFile, err
		}
	} else {
		bytes, err := pkg.VaultFS.ReadFile("vault/META-INF/vault/filter.xml")
		if err != nil {
			return tmpFilterFile, err
		}
		data := map[string]any{
			"FilterRoots":     opts.FilterRoots,
			"ExcludePatterns": opts.ExcludePatterns,
		}
		if err := tplx.RenderFile(tmpFilterFile, string(bytes), data); err != nil {
			return tmpFilterFile, err
		}
	}
	return tmpFilterFile, nil
}
func (cm *ContentManager) executeVaultCliCommand(command string, workDir string, opts PackageCreateOpts) error {
	tmpFilterFile, err := cm.determineTmpFilterFile(workDir, opts)
	defer func() { _ = pathx.DeleteIfExists(tmpFilterFile) }()
	if err != nil {
		return err
	}
	vaultCliArgs := []string{
		"vlt",
		"--credentials", fmt.Sprintf("%s:%s", cm.instance.user, cm.instance.password),
		command,
		"--force",
		"--filter", tmpFilterFile,
		fmt.Sprintf("%s/crx/server/crx.default", cm.instance.http.baseURL),
		workDir,
	}
	if err := cm.vaultCli().CommandShell(vaultCliArgs); err != nil {
		return err
	}
	return nil
}
