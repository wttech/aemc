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
	//NamespacePattern = "(\\|/)_(mediapro|lr|xmpTPg|prismusagerights|xmpPLUS|dam|s7sitecatalyst|xmpNote|dex|xml|xmp|viewerpreset|rep|stock|psAux|prism|MP|Iptc4xmpExt|photoshop|xmpDM|prl|plus|xmpMM|acdsee|wcmio|exif|xmpRights|bd|tiff|nt|stEvt|s7userdata|jcr|oak|DICOM|mix|oauth|cc|sv|social|crs|album|crx|cm|xmpBJ|cq|sling|adobe_dam|pdfx|pdf|Iptc4xmpCore|rdf|granite|stRef|MicrosoftPhoto|slingevent|dc|vlt)_"
	NamespacePattern = "(\\\\|/)_(mediapro|lr|xmpTPg|prismusagerights|xmpPLUS|dam|s7sitecatalyst|xmpNote|dex|xml|xmp|viewerpreset|rep|stock|psAux|prism|MP|Iptc4xmpExt|photoshop|xmpDM|prl|plus|xmpMM|acdsee|wcmio|exif|xmpRights|bd|tiff|nt|stEvt|s7userdata|jcr|oak|DICOM|mix|oauth|cc|sv|social|crs|album|crx|cm|xmpBJ|cq|sling|adobe_dam|pdfx|pdf|Iptc4xmpCore|rdf|granite|stRef|MicrosoftPhoto|slingevent|dc|vlt)_"
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

func (cm *ContentManager) tmpDir() string {
	if cm.instance.manager.aem.Detached() {
		return os.TempDir()
	}
	return cm.instance.manager.aem.baseOpts.TmpDir
}

func (cm *ContentManager) Download(localFile string, packageOpts PackageCreateOpts) error {
	if packageOpts.PID == "" {
		packageOpts.PID = fmt.Sprintf("aemc:content-download:%s-SNAPSHOT", timex.FileTimestampForNow())
	}
	remotePath, err := cm.pkgMgr().Create(packageOpts)
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

func (cm *ContentManager) PullDir(dir string, clean bool, replace bool, packageOpts PackageCreateOpts) error {
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_pull", ".zip")
	if err := cm.Download(pkgFile, packageOpts); err != nil {
		return err
	}
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	defer func() {
		_ = pathx.DeleteIfExists(pkgFile)
		_ = pathx.DeleteIfExists(workDir)
	}()
	if err := content.Unzip(pkgFile, workDir); err != nil {
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

func (cm *ContentManager) PullFile(file string, clean bool, packageOpts PackageCreateOpts) error {
	pkgFile := pathx.RandomFileName(cm.tmpDir(), "content_pull", ".zip")
	if err := cm.Download(pkgFile, packageOpts); err != nil {
		return err
	}
	workDir := pathx.RandomDir(cm.tmpDir(), "content_pull")
	defer func() {
		_ = pathx.DeleteIfExists(pkgFile)
		_ = pathx.DeleteIfExists(workDir)
	}()
	if err := content.Unzip(pkgFile, workDir); err != nil {
		return err
	}
	dir := filepath.Dir(file)
	_, jcrPath, _ := strings.Cut(dir, content.JCRRoot)
	contentManager := cm.instance.manager.aem.contentManager
	if err := contentManager.BeforePullFile(file); err != nil {
		return err
	}
	if err := filex.CopyDir(filepath.Join(workDir, content.JCRRoot, jcrPath), dir); err != nil {
		return err
	}
	if err := contentManager.AfterPullFile(file); err != nil {
		return err
	}
	if clean {
		cleanFile := determineCleanFile(file)
		if err := contentManager.CleanFile(cleanFile); err != nil {
			return err
		}
		if strings.HasSuffix(file, content.JCRContentFile) {
			root := filepath.Join(dir, content.JCRContentDirName)
			if pathx.Exists(root) {
				if err := contentManager.CleanDir(root); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (cm *ContentManager) Push(clean bool, contentPath string, packageOpts PackageCreateOpts) error {
	if !pathx.Exists(contentPath) {
		return fmt.Errorf("cannot push content as it does not exist '%s'", contentPath)
	}
	if packageOpts.PID == "" {
		packageOpts.PID = fmt.Sprintf("aemc:content-push:%s-SNAPSHOT", timex.FileTimestampForNow())
	}
	if clean {
		workDir := pathx.RandomDir(cm.tmpDir(), "content_push")
		defer func() {
			_ = pathx.DeleteIfExists(workDir)
		}()
		if pathx.IsDir(contentPath) {
			_, jcrPath, _ := strings.Cut(contentPath, content.JCRRoot)
			if err := filex.CopyDir(contentPath, filepath.Join(workDir, content.JCRRoot, jcrPath)); err != nil {
				return err
			}
		} else if pathx.IsFile(contentPath) {
			_, jcrPath, _ := strings.Cut(contentPath, content.JCRRoot)
			jcrDir := filepath.Dir(jcrPath)
			if err := filex.Copy(contentPath, filepath.Join(workDir, content.JCRRoot, jcrPath), true); err != nil {
				return err
			}
			if strings.HasSuffix(contentPath, content.JCRContentFile) {
				contentDir := strings.ReplaceAll(contentPath, content.JCRContentFile, content.JCRContentDirName)
				if pathx.Exists(contentDir) {
					if err := filex.CopyDir(contentDir, filepath.Join(workDir, content.JCRRoot, jcrDir, content.JCRContentDirName)); err != nil {
						return err
					}
				}
			}
		}
		contentManager := cm.instance.manager.aem.contentManager
		if err := contentManager.CleanDir(filepath.Join(workDir, content.JCRRoot)); err != nil {
			return err
		}
		packageOpts.ContentPath = filepath.Join(workDir, content.JCRRoot)
	} else {
		packageOpts.ContentPath = contentPath
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

func determineCleanFile(file string) string {
	if namespacePatternRegex.MatchString(file) && !strings.HasSuffix(file, content.JCRContentFile) {
		return filepath.Join(strings.ReplaceAll(file, content.JCRContentFileSuffix, ""), content.JCRContentFile)
	}
	return file
}

func (cm *ContentManager) Copy(destInstance *Instance, clean bool, packageOpts PackageCreateOpts) error {
	var pkgFile = pathx.RandomFileName(cm.tmpDir(), "content_copy", ".zip")
	defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
	if clean {
		workDir := pathx.RandomDir(cm.tmpDir(), "content_copy")
		defer func() { _ = pathx.DeleteIfExists(workDir) }()
		if err := cm.PullDir(filepath.Join(workDir, content.JCRRoot), clean, false, packageOpts); err != nil {
			return err
		}
		if err := content.Zip(workDir, pkgFile); err != nil {
			return err
		}
	} else {
		if err := cm.Download(pkgFile, packageOpts); err != nil {
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
