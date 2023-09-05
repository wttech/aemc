package pkg

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/httpx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"github.com/wttech/aemc/pkg/common/timex"
	"github.com/wttech/aemc/pkg/common/tplx"
	"github.com/wttech/aemc/pkg/content"
	"github.com/wttech/aemc/pkg/pkg"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PackageManager struct {
	instance *Instance

	SnapshotDeploySkipping bool
	InstallRecursive       bool
	InstallHTMLEnabled     bool
	InstallHTMLConsole     bool
	InstallHTMLStrict      bool
	SnapshotPatterns       []string
	ToggledWorkflows       []string
}

func NewPackageManager(res *Instance) *PackageManager {
	cv := res.manager.aem.config.Values()

	return &PackageManager{
		instance: res,

		SnapshotDeploySkipping: cv.GetBool("instance.package.snapshot_deploy_skipping"),
		InstallHTMLEnabled:     cv.GetBool("instance.package.install_html.enabled"),
		InstallHTMLConsole:     cv.GetBool("instance.package.install_html.console"),
		InstallHTMLStrict:      cv.GetBool("instance.package.install_html.strict"),
		InstallRecursive:       cv.GetBool("instance.package.install_recursive"),
		SnapshotPatterns:       cv.GetStringSlice("instance.package.snapshot_patterns"),
		ToggledWorkflows:       cv.GetStringSlice("instance.package.toggled_workflows"),
	}
}

func (pm *PackageManager) ByPID(pid string) (*Package, error) {
	pidConfig, err := pkg.ParsePID(pid)
	if err != nil {
		return nil, err
	}
	return pm.byPID(*pidConfig)
}

func (pm *PackageManager) ByFile(localPath string) (*Package, error) {
	pidConfig, err := pkg.ReadPIDFromZIP(localPath)
	if err != nil {
		return nil, err
	}
	return pm.byPID(*pidConfig)
}

func (pm *PackageManager) ByPath(remotePath string) (*Package, error) {
	list, err := pm.List()
	if err != nil {
		return nil, err
	}
	item, ok := lo.Find(list.List, func(item pkg.ListItem) bool { return item.Path == remotePath })
	if !ok {
		return nil, fmt.Errorf("%s > package at path '%s' does not exist", pm.instance.ID(), remotePath)
	}
	pid, err := pkg.ParsePID(item.PID)
	if err != nil {
		return nil, err
	}
	return pm.byPID(*pid)
}

func (pm *PackageManager) byPID(pidConfig pkg.PID) (*Package, error) {
	return &Package{manager: pm, PID: pidConfig}, nil
}

func (pm *PackageManager) List() (*pkg.List, error) {
	resp, err := pm.instance.http.Request().Get(ListJson)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot request package list: %w", pm.instance.ID(), err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot request package list: %s", pm.instance.ID(), resp.Status())
	}
	res := new(pkg.List)
	if err = fmtx.UnmarshalJSON(resp.RawBody(), res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse package list response: %w", pm.instance.ID(), err)
	}
	return res, nil
}

func (pm *PackageManager) Find(pid string) (*pkg.ListItem, error) {
	item, err := pm.findInternal(pid)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot find package '%s': %w", pm.instance.ID(), pid, err)
	}
	return item, nil
}

func (pm *PackageManager) findInternal(pid string) (*pkg.ListItem, error) {
	pidConfig, err := pkg.ParsePID(pid)
	if err != nil {
		return nil, err
	}
	resp, err := pm.instance.http.Request().SetQueryParam("name", pidConfig.Name).Get(ListJson)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot request package list: %w", pm.instance.ID(), err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot request package list: %s", pm.instance.ID(), resp.Status())
	}
	res := new(pkg.List)
	if err = fmtx.UnmarshalJSON(resp.RawBody(), res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse package list response: %w", pm.instance.ID(), err)
	}
	item, ok := lo.Find(res.List, func(p pkg.ListItem) bool { return p.PID == pid })
	if ok {
		return &item, nil
	}
	return nil, nil
}

func (pm *PackageManager) IsSnapshot(localPath string) bool {
	return stringsx.MatchSome(pathx.Normalize(localPath), pm.SnapshotPatterns)
}

//go:embed package/default
var packageDefaultFS embed.FS

func copyPackageDefaultFiles(targetTmpDir string, dirPrefix string, data map[string]any) error {
	if err := pathx.DeleteIfExists(targetTmpDir); err != nil {
		return fmt.Errorf("cannot delete temporary dir '%s': %w", targetTmpDir, err)
	}
	return fs.WalkDir(packageDefaultFS, ".", func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}
		bytes, err := packageDefaultFS.ReadFile(path)
		if err != nil {
			return err
		}
		return tplx.RenderFile(targetTmpDir+strings.ReplaceAll(strings.TrimPrefix(path, dirPrefix), "$", ""), string(bytes), data)
	})
}

func (pm *PackageManager) Create(pid string, rootPaths []string, filterFile string) (string, error) {
	log.Infof("%s > creating package '%s'", pm.instance.ID(), pid)
	pidConfig, err := pkg.ParsePID(pid)
	if err != nil {
		return "", err
	}
	var response *resty.Response
	if len(rootPaths) == 0 && filterFile == "" {
		response, err = pm.instance.http.Request().
			SetFormData(map[string]string{
				"packageName":    pidConfig.Name,
				"packageVersion": pidConfig.Version,
				"groupName":      pidConfig.Group,
			}).
			Post(ExecPath + "?cmd=create")
	} else {
		tmpDir := pathx.RandomTemporaryPathName(pm.tmpDir(), "tmppck")
		tmpFile := pathx.RandomTemporaryFileName(pm.tmpDir(), "tmppck", ".zip")
		defer func() {
			_ = pathx.DeleteIfExists(tmpDir)
			_ = pathx.DeleteIfExists(tmpFile)
		}()
		data := map[string]any{
			"Pid":     pid,
			"Group":   pidConfig.Group,
			"Name":    pidConfig.Name,
			"Version": pidConfig.Version,
			"Roots":   rootPaths,
		}
		if err = copyPackageDefaultFiles(tmpDir, "package/default", data); err != nil {
			return "", err
		}
		if filterFile != "" {
			if err = filex.Copy(filterFile, filepath.Join(tmpDir, "META-INF", "vault", "filter.xml"), true); err != nil {
				return "", err
			}
		}
		if err = filex.Archive(tmpDir, tmpFile); err != nil {
			return "", err
		}
		response, err = pm.instance.http.Request().
			SetFile("package", tmpFile).
			SetMultipartFormData(map[string]string{"force": "true"}).
			Post(ServiceJsonPath + "/?cmd=upload")
	}
	if err != nil {
		return "", fmt.Errorf("%s > cannot create package '%s': %w", pm.instance.ID(), pid, err)
	} else if response.IsError() {
		return "", fmt.Errorf("%s > cannot create package '%s': %s", pm.instance.ID(), pid, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return "", fmt.Errorf("%s > cannot create package '%s'; cannot parse response: %w", pm.instance.ID(), pid, err)
	}
	if !status.Success {
		return "", fmt.Errorf("%s > cannot create package '%s'; unexpected status: %s", pm.instance.ID(), pid, status.Message)
	}
	log.Infof("%s > created package '%s'", pm.instance.ID(), pid)
	return status.Path, nil
}

type Filter struct {
	Root  string `json:"root"`
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Modifier string `json:"modifier"`
	Pattern  string `json:"pattern"`
}

func (pm *PackageManager) UpdateFilters(remotePath string, pid string, filters []Filter) error {
	log.Infof("%s > updating package '%s'", pm.instance.ID(), pid)
	pidConfig, err := pkg.ParsePID(pid)
	if err != nil {
		return err
	}
	filtersJson, err := json.Marshal(filters)
	if err != nil {
		return err
	}
	response, err := pm.instance.http.Request().
		SetMultipartFormData(map[string]string{
			"path":        remotePath,
			"packageName": pidConfig.Name,
			"groupName":   pidConfig.Group,
			"version":     pidConfig.Version,
			"filter":      string(filtersJson),
		}).
		Post(UpdatePath)
	if err != nil {
		return fmt.Errorf("%s > cannot update filters for package '%s': %w", pm.instance.ID(), pid, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot update filters for package '%s': %s", pm.instance.ID(), pid, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot update filters for package '%s'; cannot parse response: %w", pm.instance.ID(), pid, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot update filters for package '%s'; unexpected status: %s", pm.instance.ID(), pid, status.Message)
	}
	log.Infof("%s > update filters for package '%s'", pm.instance.ID(), pid)
	return nil
}

func (pm *PackageManager) Download(remotePath string, localFile string) error {
	log.Infof("%s > downloading package '%s'", pm.instance.ID(), remotePath)
	opts := httpx.DownloadOpts{}
	opts.URL = pm.instance.HTTP().BaseURL() + remotePath
	opts.File = localFile
	opts.Override = true
	opts.AuthBasicUser = pm.instance.User()
	opts.AuthBasicPassword = pm.instance.Password()
	_, err := httpx.DownloadWithChanged(opts)
	if err != nil {
		return fmt.Errorf("%s > cannot download package '%s': %w", pm.instance.ID(), remotePath, err)
	}
	log.Infof("%s > downloaded package '%s'", pm.instance.ID(), remotePath)
	return nil
}

func (pm *PackageManager) Build(remotePath string) error {
	log.Infof("%s > building package '%s'", pm.instance.ID(), remotePath)
	response, err := pm.instance.http.Request().Post(ServiceJsonPath + remotePath + "?cmd=build")
	if err != nil {
		return fmt.Errorf("%s > cannot build package '%s': %w", pm.instance.ID(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot build package '%s': %s", pm.instance.ID(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot build package '%s'; cannot parse response: %w", pm.instance.ID(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot build package '%s'; unexpected status: %s", pm.instance.ID(), remotePath, status.Message)
	}
	log.Infof("%s > built package '%s'", pm.instance.ID(), remotePath)
	return nil
}

func (pm *PackageManager) UploadWithChanged(localPath string) (bool, error) {
	if pm.IsSnapshot(localPath) {
		_, err := pm.Upload(localPath)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	p, err := pm.ByFile(localPath)
	if err != nil {
		return false, err
	}
	state, err := p.State()
	if err != nil {
		return false, err
	}
	if !state.Exists {
		_, err = pm.Upload(localPath)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (pm *PackageManager) Upload(localPath string) (string, error) {
	log.Infof("%s > uploading package '%s'", pm.instance.ID(), localPath)
	response, err := pm.instance.http.Request().
		SetFile("package", localPath).
		SetMultipartFormData(map[string]string{"force": "true"}).
		Post(ServiceJsonPath + "/?cmd=upload")
	if err != nil {
		return "", fmt.Errorf("%s > cannot upload package '%s': %w", pm.instance.ID(), localPath, err)
	} else if response.IsError() {
		return "", fmt.Errorf("%s > cannot upload package '%s': %s", pm.instance.ID(), localPath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return "", fmt.Errorf("%s > cannot upload package '%s'; cannot parse response: %w", pm.instance.ID(), localPath, err)
	}
	if !status.Success {
		return "", fmt.Errorf("%s > cannot upload package '%s'; unexpected status: %s", pm.instance.ID(), localPath, status.Message)
	}
	log.Infof("%s > uploaded package '%s'", pm.instance.ID(), localPath)
	return status.Path, nil
}

func (pm *PackageManager) Install(remotePath string) error {
	if pm.InstallHTMLEnabled {
		return pm.installHTML(remotePath)
	}
	return pm.installJSON(remotePath)
}

func (pm *PackageManager) installJSON(remotePath string) error {
	log.Infof("%s > installing package '%s'", pm.instance.ID(), remotePath)
	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "install", "recursive": fmt.Sprintf("%v", pm.InstallRecursive)}).
		Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("%s > cannot install package '%s': %w", pm.instance.ID(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot install package '%s': '%s'", pm.instance.ID(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot install package '%s'; cannot parse JSON response: %w", pm.instance.ID(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot install package '%s'; unexpected status: %s", pm.instance.ID(), remotePath, status.Message)
	}
	log.Infof("%s > installed package '%s'", pm.instance.ID(), remotePath)
	return nil
}

func (pm *PackageManager) installHTML(remotePath string) error {
	log.Infof("%s > installing package '%s'", pm.instance.ID(), remotePath)

	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "install", "recursive": fmt.Sprintf("%v", pm.InstallRecursive)}).
		Post(ServiceHtmlPath + remotePath)
	if err != nil {
		return fmt.Errorf("%s > cannot install package '%s': %w", pm.instance.ID(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot install package '%s': '%s'", pm.instance.ID(), remotePath, response.Status())
	}

	success := false
	successWithErrors := false

	htmlFilePath := fmt.Sprintf("%s/package/install/%s-%s.html", pm.instance.CacheDir(), filepath.Base(remotePath), timex.FileTimestampForNow())
	var htmlWriter *bufio.Writer

	if !pm.InstallHTMLConsole {
		if err := pathx.Ensure(filepath.Dir(htmlFilePath)); err != nil {
			return err
		}
		htmlFile, err := os.OpenFile(htmlFilePath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return fmt.Errorf("%s > cannot install package '%s': cannot open HTML report file '%s'", pm.instance.ID(), remotePath, htmlFilePath)
		}
		defer htmlFile.Close()
		htmlWriter = bufio.NewWriter(htmlFile)
		defer htmlWriter.Flush()
	}

	scanner := bufio.NewScanner(response.RawBody())
	for scanner.Scan() {
		htmlLine := scanner.Text()
		if !success && strings.Contains(htmlLine, pkg.InstallSuccess) {
			success = true
		}
		if !successWithErrors && strings.Contains(htmlLine, pkg.InstallSuccessWithErrors) {
			successWithErrors = true
		}
		if !pm.InstallHTMLConsole {
			_, err := htmlWriter.WriteString(htmlLine + osx.LineSep())
			if err != nil {
				return fmt.Errorf("%s > cannot install package '%s': cannot write to HTML report file '%s'", pm.instance.ID(), remotePath, htmlFilePath)
			}
		} else {
			fmt.Println(htmlLine)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%s > cannot install package '%s': cannot parse HTML response: %w", pm.instance.ID(), remotePath, err)
	}

	failure := !success && !successWithErrors
	if failure || (successWithErrors && pm.InstallHTMLStrict) {
		if pm.InstallHTMLConsole {
			return fmt.Errorf("%s > cannot install package '%s': HTML output contains errors", pm.instance.ID(), remotePath)
		}
		return fmt.Errorf("%s > cannot install package '%s': HTML report contains errors '%s'", pm.instance.ID(), remotePath, htmlFilePath)
	}
	if successWithErrors {
		log.Warnf("%s > installed package '%s': HTML response contains errors: %s", pm.instance.ID(), remotePath, err)
		return nil
	}
	log.Infof("%s > installed package '%s'", pm.instance.ID(), remotePath)
	return nil
}

func (pm *PackageManager) DeployWithChanged(localPath string) (bool, error) {
	if pm.IsSnapshot(localPath) {
		return pm.deploySnapshot(localPath)
	}
	return pm.deployRegular(localPath)
}

func (pm *PackageManager) deployRegular(localPath string) (bool, error) {
	deployed, err := pm.IsDeployed(localPath)
	if err != nil {
		return false, err
	}
	if !deployed {
		return true, pm.Deploy(localPath)
	}
	return false, nil
}

func (pm *PackageManager) deploySnapshot(localPath string) (bool, error) {
	checksum, err := filex.ChecksumFile(localPath)
	if err != nil {
		return false, err
	}
	deployed, err := pm.IsDeployed(localPath)
	if err != nil {
		return false, err
	}
	var lock = pm.deployLock(localPath, checksum)
	if deployed && pm.SnapshotDeploySkipping && lock.IsLocked() {
		lockData, err := lock.Locked()
		if err != nil {
			return false, err
		}
		if checksum == lockData.Checksum {
			log.Infof("%s > skipped deploying package '%s'", pm.instance.ID(), localPath)
			return false, nil
		}
	}
	if err := pm.Deploy(localPath); err != nil {
		return false, err
	}
	if err := lock.Lock(); err != nil {
		return false, err
	}
	return true, nil
}

func (pm *PackageManager) IsDeployed(localPath string) (bool, error) {
	p, err := pm.ByFile(localPath)
	if err != nil {
		return false, err
	}
	state, err := p.State()
	if err != nil {
		return false, err
	}
	return state.Exists && state.Data.Installed(), nil
}

func (pm *PackageManager) Deploy(localPath string) error {
	remotePath, err := pm.Upload(localPath)
	if err != nil {
		return err
	}
	return pm.instance.workflowManager.ToggleLaunchers(pm.ToggledWorkflows, func() error {
		return pm.Install(remotePath)
	})
}

func (pm *PackageManager) deployLock(file string, checksum string) osx.Lock[packageDeployLock] {
	name := filepath.Base(file)
	return osx.NewLock(fmt.Sprintf("%s/package/deploy/%s.yml", pm.instance.LockDir(), name), func() (packageDeployLock, error) {
		return packageDeployLock{Deployed: time.Now(), Checksum: checksum}, nil
	})
}

type packageDeployLock struct {
	Deployed time.Time `yaml:"deployed"`
	Checksum string    `yaml:"checksum"`
}

func (pm *PackageManager) Uninstall(remotePath string) error {
	log.Infof("%s > uninstalling package '%s'", pm.instance.ID(), remotePath)
	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "uninstall"}).
		Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("%s > cannot uninstall package '%s': %w", pm.instance.ID(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot uninstall package '%s': %s", pm.instance.ID(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot uninstall package '%s'; cannot parse response: %w", pm.instance.ID(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot uninstall package '%s'; unexpected status: %s", pm.instance.ID(), remotePath, status.Message)
	}
	log.Infof("%s > uninstalled package '%s'", pm.instance.ID(), remotePath)
	return nil
}

func (pm *PackageManager) Delete(remotePath string) error {
	log.Infof("%s > deleting package '%s'", pm.instance.ID(), remotePath)
	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "delete"}).
		Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("%s > cannot delete package '%s': %w", pm.instance.ID(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot delete package '%s': %s", pm.instance.ID(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot delete package '%s'; cannot parse response: %w", pm.instance.ID(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot delete package '%s'; unexpected status: %s", pm.instance.ID(), remotePath, status.Message)
	}
	log.Infof("%s > deleted package '%s'", pm.instance.ID(), remotePath)
	return nil
}

func (pm *PackageManager) tmpDir() string {
	return pm.instance.manager.aem.baseOpts.TmpDir
}

func (pm *PackageManager) DownloadPackage(pid string, roots []string, filter string) (string, error) {
	if pid == "" {
		pid = "my_packages:aemc_content:" + time.Now().Format("2006.102.304") + "-SNAPSHOT"
	}
	remotePath, err := pm.Create(pid, roots, filter)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = pm.Delete(remotePath)
	}()
	if err := pm.Build(remotePath); err != nil {
		return "", err
	}
	tmpResultFile := filepath.Join(pm.instance.manager.aem.baseOpts.TmpDir, filepath.Base(remotePath))
	if err := pm.Download(remotePath, tmpResultFile); err != nil {
		return "", err
	}
	return tmpResultFile, nil
}

func (pm *PackageManager) DownloadContent(pid string, root string, roots []string, filter string, clean bool, unpack bool) error {
	tmpResultFile, err := pm.DownloadPackage(pid, roots, filter)
	if err != nil {
		return err
	}
	tmpResultDir := pathx.RandomTemporaryPathName(pm.tmpDir(), "pkg_download_content")
	defer func() {
		_ = pathx.DeleteIfExists(tmpResultDir)
		_ = pathx.DeleteIfExists(tmpResultFile)
	}()
	if unpack {
		if err = filex.Unarchive(tmpResultFile, tmpResultDir); err != nil {
			return err
		}
		if err := pathx.Ensure(root); err != nil {
			return err
		}
		before, _, _ := strings.Cut(root, content.JCRRoot)
		contentManager := pm.instance.manager.aem.contentManager
		if clean {
			if err = contentManager.BeforeClean(root); err != nil {
				return err
			}
		}
		if err = filex.CopyDir(filepath.Join(tmpResultDir, content.JCRRoot), before+content.JCRRoot); err != nil {
			return err
		}
		if clean {
			if err = contentManager.Clean(root); err != nil {
				return err
			}
		}
	} else {
		if err = filex.Copy(tmpResultFile, filepath.Join(root, filepath.Base(tmpResultFile)), true); err != nil {
			return err
		}
	}
	return nil
}

func (pm *PackageManager) CopyContent(destInstance *Instance, pid string, roots []string, filter string, clean bool) error {

	var tmpResultFile string
	if clean {
		tmpResultFile = pathx.RandomTemporaryFileName(pm.tmpDir(), "pkg_copy_content", ".zip")
		tmpResultDir := pathx.RandomTemporaryPathName(pm.tmpDir(), "pkg_copy_content")
		defer func() {
			_ = pathx.DeleteIfExists(tmpResultDir)
			_ = pathx.DeleteIfExists(tmpResultFile)
		}()
		if err := pm.DownloadContent(filepath.Join(tmpResultDir, content.JCRRoot), "", roots, filter, true, true); err != nil {
			return err
		}
		if err := filex.Archive(tmpResultDir, tmpResultFile); err != nil {
			return err
		}
	} else {
		var err error
		tmpResultFile, err = pm.DownloadPackage(pid, roots, filter)
		if err != nil {
			return err
		}
	}
	defer func() { _ = pathx.DeleteIfExists(tmpResultFile) }()
	remotePath, err := destInstance.PackageManager().Upload(tmpResultFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = destInstance.PackageManager().Delete(remotePath)
	}()
	if err = destInstance.PackageManager().Install(remotePath); err != nil {
		return err
	}
	return nil
}

const (
	MgrPath         = "/crx/packmgr"
	ServicePath     = MgrPath + "/service"
	ServiceJsonPath = ServicePath + "/.json"
	ServiceHtmlPath = ServicePath + "/.html"
	ListJson        = MgrPath + "/list.jsp"
	IndexPath       = MgrPath + "/index.jsp"
	ExecPath        = ServicePath + "/exec.json"
	UpdatePath      = MgrPath + "/update.jsp"

	FilterXML = "filter.xml"
)
