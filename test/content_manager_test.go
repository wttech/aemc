//go:build int_test

package test

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/timex"
	"github.com/wttech/aemc/pkg/content"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed resources
var VaultFS embed.FS

func TestPullDir(t *testing.T) {
	testPullContent(t, "/content/mysite", "", nil, "", []string{
		"/content/mysite/.content.xml",
		"/content/mysite/us/.content.xml",
		"/content/mysite/us/en/.content.xml",
	})
}

func TestPullDirWithNamespace(t *testing.T) {
	testPullContent(t, "/conf/mysite/_sling_configs", "", nil, "", []string{
		"/conf/mysite/_sling_configs/.content.xml",
		"/conf/mysite/_sling_configs/com.mysite.pdfviewer.PdfViewerCaConfig/.content.xml",
	})
}

func TestPullOnePage(t *testing.T) {
	testPullContent(t, "", "/content/mysite/us/.content.xml", nil, "", []string{
		"/content/mysite/us/.content.xml",
	})
}

func TestPullOneTemplate(t *testing.T) {
	testPullContent(t, "", "/conf/mysite/settings/wcm/template-types/page/.content.xml", nil, "", []string{
		"/conf/mysite/settings/wcm/template-types/page/.content.xml",
	})
}

func TestPullCqTemplate(t *testing.T) {
	testPullContent(t, "", "/apps/mysite/components/helloworld/_cq_template/.content.xml", nil, "", []string{
		"/apps/mysite/components/helloworld/_cq_template.xml",
	})
}

func TestPullCqDialogFlatten(t *testing.T) {
	testPullContent(t, "", "/apps/mysite/components/helloworld/_cq_dialog.xml", nil, "", []string{
		"/apps/mysite/components/helloworld/_cq_dialog.xml",
	})
}

func TestPullOnlyContentXml(t *testing.T) {
	testPullContent(t, "", "/apps/mysite/components/helloworld/.content.xml", nil, "", []string{
		"/apps/mysite/components/helloworld/.content.xml",
	})
}

func TestPullRepPolicy(t *testing.T) {
	testPullContent(t, "", "/conf/mysite/settings/wcm/policies/_rep_policy.xml", nil, "", []string{
		"/conf/mysite/settings/wcm/policies/_rep_policy.xml",
	})
}

func TestPullXmlFile(t *testing.T) {
	testPullContent(t, "", "/var/workflow/models/mysite/asset_processing.xml", nil, "", []string{
		"/var/workflow/models/mysite/asset_processing.xml",
	})
}

func TestPullTextFile(t *testing.T) {
	testPullContent(t, "", "/apps/mysite/components/helloworld/helloworld.html", nil, "", []string{
		"/apps/mysite/components/helloworld/helloworld.html",
	})
}

func TestPullFilterRoots(t *testing.T) {
	testPullContent(t, "/", "", []string{"/content/mysite"}, "", []string{
		"/content/mysite/.content.xml",
		"/content/mysite/us/.content.xml",
		"/content/mysite/us/en/.content.xml",
	})

}
func TestPullFilterFile(t *testing.T) {
	workDir := pathx.RandomDir(os.TempDir(), "filter")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := copyFile("resources", workDir, "resources/filter.xml"); err != nil {
		t.Fatal(err)
	}
	filterFile := filepath.Join(workDir, pkg.FilterXML)
	testPullContent(t, "/", "", nil, filterFile, []string{
		"/content/mysite/.content.xml",
		"/content/mysite/us/.content.xml",
		"/content/mysite/us/en/.content.xml",
	})
}

func TestPushDir(t *testing.T) {
	testPushContent(t, "/content/mysite", []string{
		"/content/mysite/.content.xml",
		"/content/mysite/us/.content.xml",
		"/content/mysite/us/en/.content.xml",
	})
}

func TestPushDirWithNamespace(t *testing.T) {
	testPushContent(t, "/conf/mysite/_sling_configs", []string{
		"/conf/mysite/_sling_configs/.content.xml",
		"/conf/mysite/_sling_configs/com.mysite.pdfviewer.PdfViewerCaConfig/.content.xml",
	})
}

func TestPushOnePage(t *testing.T) {
	testPushContent(t, "/content/mysite/us/.content.xml", []string{
		"/content/mysite/us/.content.xml",
	})
}

func TestPushOneTemplate(t *testing.T) {
	testPushContent(t, "/conf/mysite/settings/wcm/template-types/page/.content.xml", []string{
		"/conf/mysite/settings/wcm/template-types/page/.content.xml",
	})
}

func TestPushCqTemplate(t *testing.T) {
	testPushContent(t, "/apps/mysite/components/helloworld/_cq_template/.content.xml", []string{
		"/apps/mysite/components/helloworld/_cq_template.xml",
	})
}

func TestPushCqDialogFlatten(t *testing.T) {
	testPushContent(t, "/apps/mysite/components/helloworld/_cq_dialog.xml", []string{
		"/apps/mysite/components/helloworld/_cq_dialog.xml",
	})
}

func TestPushOnlyContentXml(t *testing.T) {
	testPushContent(t, "/apps/mysite/components/helloworld/.content.xml", []string{
		"/apps/mysite/components/helloworld/.content.xml",
	})
}

func TestPushRepPolicy(t *testing.T) {
	testPushContent(t, "/conf/mysite/settings/wcm/policies/_rep_policy.xml", []string{
		"/conf/mysite/settings/wcm/policies/_rep_policy.xml",
	})
}

func TestPushXmlFile(t *testing.T) {
	testPushContent(t, "/var/workflow/models/mysite/asset_processing.xml", []string{
		"/var/workflow/models/mysite/asset_processing.xml",
	})
}

func TestPushTextFile(t *testing.T) {
	testPushContent(t, "/apps/mysite/components/helloworld/helloworld.html", []string{
		"/apps/mysite/components/helloworld/helloworld.html",
	})
}

func testPullContent(t *testing.T, relDir string, relFile string, filterRoots []string, filterFile string, expectedFiles []string) {
	aem := pkg.DefaultAEM()
	contentManager := pkg.NewContentManager(aem)
	instance := aem.InstanceManager().NewLocalAuthor()
	packageManager := instance.PackageManager()

	remotePath, err := installMainContent(packageManager)
	defer uninstallMainContent(packageManager, instance, remotePath)
	if err != nil {
		t.Fatal(err)
	}

	resultDir, err := pullContent(contentManager, instance, relDir, relFile, filterRoots, filterFile)
	defer func() { _ = pathx.DeleteIfExists(resultDir) }()
	if err != nil {
		t.Fatal(err)
	}

	diffs, err := checkPullResult(resultDir, expectedFiles)
	if err != nil {
		t.Fatal(err)
	}
	if diffs != nil {
		t.Errorf("testPullContent(%s, %s, %v, %s) -> %v", relDir, relFile, filterRoots, filterFile, diffs)
	}
}

func testPushContent(t *testing.T, relPath string, expectedFiles []string) {
	aem := pkg.DefaultAEM()
	contentManager := pkg.NewContentManager(aem)
	instance := aem.InstanceManager().NewLocalAuthor()
	packageManager := instance.PackageManager()

	remotePath, err := installMainContent(packageManager)
	defer uninstallMainContent(packageManager, instance, remotePath)
	if err != nil {
		t.Fatal(err)
	}

	workDir := pathx.RandomDir(os.TempDir(), "content_push")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := copyFiles("resources/new_content", workDir); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(workDir, content.JCRRoot, relPath)
	if err := pushContent(contentManager, instance, path); err != nil {
		t.Fatal(err)
	}

	resultDir, err := pullContent(contentManager, instance, "/", "", []string{
		"/apps/mysite",
		"/conf/mysite",
		"/content/mysite",
		"/var/workflow/models/mysite",
	}, "")
	defer func() { _ = pathx.DeleteIfExists(resultDir) }()
	if err != nil {
		t.Fatal(err)
	}

	diffs, err := checkPushResult(resultDir, expectedFiles)
	if err != nil {
		t.Fatal(err)
	}
	if diffs != nil {
		t.Errorf("testPushContent(%s) -> %v", relPath, diffs)
	}
}

func pullContent(contentManager *pkg.ContentManager, instance pkg.Instance, relDir string, relFile string, filterRoots []string, filterFile string) (string, error) {
	resultDir := pathx.RandomDir(os.TempDir(), "content_pull")
	var dir string
	if relDir != "" {
		dir = filepath.Join(resultDir, content.JCRRoot, relDir)
	}
	var file string
	if relFile != "" {
		file = filepath.Join(resultDir, content.JCRRoot, relFile)
	}
	if dir != "" {
		if err := contentManager.PullDir(&instance, dir, true, false, pkg.PackageCreateOpts{
			PID:         fmt.Sprintf("aemc:content-pull:%s-SNAPSHOT", timex.FileTimestampForNow()),
			FilterRoots: determineFilterRoots(dir, file, filterRoots, filterFile),
			FilterFile:  filterFile,
		}); err != nil {
			return resultDir, err
		}
	} else if file != "" {
		if err := contentManager.PullFile(&instance, file, true, false, pkg.PackageCreateOpts{
			PID:         fmt.Sprintf("aemc:content-pull:%s-SNAPSHOT", timex.FileTimestampForNow()),
			FilterRoots: determineFilterRoots(dir, file, filterRoots, filterFile),
			FilterFile:  filterFile,
		}); err != nil {
			return resultDir, err
		}
	}
	return resultDir, nil
}

func pushContent(contentManager *pkg.ContentManager, instance pkg.Instance, path string) error {
	if err := contentManager.Push([]pkg.Instance{instance}, true, pkg.PackageCreateOpts{
		PID:             fmt.Sprintf("aemc:content-push:%s-SNAPSHOT", timex.FileTimestampForNow()),
		FilterRoots:     determineFilterRoots(path, path, nil, ""),
		ExcludePatterns: determineExcludePatterns(path),
		ContentPath:     path,
	}); err != nil {
		return nil
	}
	return nil
}

func installMainContent(packageManager *pkg.PackageManager) (string, error) {
	workDir := pathx.RandomDir(os.TempDir(), "content_pull")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := copyFiles("resources/main_content", workDir); err != nil {
		return "", err
	}
	pkgFile := pathx.RandomFileName(os.TempDir(), "content_pull", ".zip")
	defer func() { _ = pathx.DeleteIfExists(pkgFile) }()
	if err := content.Zip(workDir, pkgFile); err != nil {
		return "", err
	}
	remotePath, err := packageManager.Upload(pkgFile)
	if err != nil {
		return "", err
	}
	if err = packageManager.Install(remotePath); err != nil {
		return remotePath, err
	}
	return remotePath, nil
}

func uninstallMainContent(packageManager *pkg.PackageManager, instance pkg.Instance, remotePath string) {
	_ = packageManager.Delete(remotePath)
	_ = instance.Repo().Delete("/apps/mysite")
	_ = instance.Repo().Delete("/conf/mysite")
	_ = instance.Repo().Delete("/content/mysite")
	_ = instance.Repo().Delete("/var/workflow/models/mysite")
}

func copyFiles(fsSrcDir string, osDestDir string) error {
	if err := fs.WalkDir(VaultFS, fsSrcDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if err = copyFile(fsSrcDir, osDestDir, path); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func copyFile(fsSrcDir string, osDestDir string, path string) error {
	relPath, err := filepath.Rel(fsSrcDir, path)
	if err != nil {
		return err
	}
	relPath = strings.ReplaceAll(relPath, "$", "")
	newPath := filepath.Join(osDestDir, relPath)
	data, err := VaultFS.ReadFile(path)
	if err != nil {
		return err
	}
	if err = filex.Write(newPath, data); err != nil {
		return err
	}
	return nil
}

func determineFilterRoots(dir string, file string, filterRoots []string, filterFile string) []string {
	if len(filterRoots) > 0 {
		return filterRoots
	}
	if filterFile != "" {
		return nil
	}
	if dir != "" {
		return []string{pkg.DetermineFilterRoot(dir)}
	}
	if file != "" {
		return []string{pkg.DetermineFilterRoot(file)}
	}
	return nil
}

func determineExcludePatterns(file string) []string {
	if !pathx.IsFile(file) || !strings.HasSuffix(file, content.JCRContentFile) || content.IsPageContentFile(file) {
		return nil
	}
	dir := filepath.Dir(file)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var excludePatterns []string
	for _, entry := range entries {
		if entry.Name() != content.JCRContentFile {
			jcrPath := pkg.DetermineFilterRoot(filepath.Join(dir, entry.Name()))
			excludePattern := fmt.Sprintf("%s(/.*)?", jcrPath)
			excludePatterns = append(excludePatterns, excludePattern)
		}
	}
	return excludePatterns
}

func determineResultFiles(workDir string) (map[string]string, error) {
	files := map[string]string{}
	if err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.Contains(path, "/mysite/") {
			return nil
		}
		_, relPath, _ := strings.Cut(path, content.JCRRoot)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[relPath] = calculateHashCode(data)
		return nil
	}); err != nil {
		return nil, err
	}
	return files, nil
}

func determineTestFiles(workDir string) (map[string]string, error) {
	files := map[string]string{}
	if err := fs.WalkDir(VaultFS, workDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.Contains(path, "/mysite/") {
			return nil
		}
		_, relPath, _ := strings.Cut(path, content.JCRRoot)
		relPath = strings.ReplaceAll(relPath, "$", "")
		data, err := VaultFS.ReadFile(path)
		if err != nil {
			return err
		}
		files[relPath] = calculateHashCode(data)
		return nil
	}); err != nil {
		return nil, err
	}
	return files, nil
}

func determineExpectedFiles(workDir string, expectedFiles []string) (map[string]string, error) {
	files := map[string]string{}
	for _, expectedFile := range expectedFiles {
		path := filepath.Join(workDir, expectedFile)
		path = strings.ReplaceAll(path, "/.", "/$.")
		path = strings.ReplaceAll(path, "/_", "/$_")
		data, err := VaultFS.ReadFile(path)
		if err != nil {
			return nil, err
		}
		files[expectedFile] = calculateHashCode(data)
	}
	return files, nil
}

func determineDifferenceFiles(actualFiles map[string]string, expectedFiles map[string]string) map[string]string {
	diffFiles := map[string]string{}
	for relPath, hashCode := range actualFiles {
		hashCode2, exists := expectedFiles[relPath]
		if !exists {
			diffFiles[relPath] = ""
		}
		if hashCode != hashCode2 {
			diffFiles[relPath] = hashCode
		}
	}
	for relPath := range expectedFiles {
		if _, exists := actualFiles[relPath]; !exists {
			diffFiles[relPath] = ""
		}
	}
	return diffFiles
}

func determineKeySet(filesMap map[string]string) []string {
	var filesList []string
	for relPath := range filesMap {
		filesList = append(filesList, relPath)
	}
	return filesList
}

func checkPullResult(resultDir string, expectedFiles []string) ([]string, error) {
	files1, err := determineResultFiles(resultDir)
	if err != nil {
		return nil, err
	}
	files2, err := determineExpectedFiles("resources/main_content/jcr_root", expectedFiles)
	if err != nil {
		return nil, err
	}
	diffFiles := determineDifferenceFiles(files1, files2)
	return determineKeySet(diffFiles), nil
}

func checkPushResult(resultDir string, expectedFiles []string) ([]string, error) {
	files1, err := determineResultFiles(resultDir)
	if err != nil {
		return nil, err
	}
	files2, err := determineTestFiles("resources/main_content/jcr_root")
	if err != nil {
		return nil, err
	}
	diffFiles := determineDifferenceFiles(files1, files2)
	files3, err := determineExpectedFiles("resources/new_content/jcr_root", expectedFiles)
	if err != nil {
		return nil, err
	}
	diffFiles2 := determineDifferenceFiles(diffFiles, files3)
	return determineKeySet(diffFiles2), nil
}

func calculateHashCode(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
