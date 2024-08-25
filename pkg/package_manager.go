package pkg

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
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
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PackageManager struct {
	instance *Instance

	UploadOptimized           bool
	InstallRecursive          bool
	InstallSaveThreshold      int
	InstallACHandling         string
	InstallDependencyHandling string
	InstallHTMLEnabled        bool
	InstallHTMLConsole        bool
	InstallHTMLStrict         bool
	SnapshotDeploySkipping    bool
	SnapshotIgnored           bool
	SnapshotPatterns          []string
	ToggledWorkflows          []string
}

func NewPackageManager(res *Instance) *PackageManager {
	cv := res.manager.aem.config.Values()

	return &PackageManager{
		instance: res,

		UploadOptimized:           cv.GetBool("instance.package.upload_optimized"),
		InstallRecursive:          cv.GetBool("instance.package.install_recursive"),
		InstallSaveThreshold:      cv.GetInt("instance.package.install_save_threshold"),
		InstallACHandling:         cv.GetString("instance.package.install_ac_handling"),
		InstallDependencyHandling: cv.GetString("instance.package.install_dependency_handling"),
		InstallHTMLEnabled:        cv.GetBool("instance.package.install_html.enabled"),
		InstallHTMLConsole:        cv.GetBool("instance.package.install_html.console"),
		InstallHTMLStrict:         cv.GetBool("instance.package.install_html.strict"),
		SnapshotDeploySkipping:    cv.GetBool("instance.package.snapshot_deploy_skipping"),
		SnapshotIgnored:           cv.GetBool("instance.package.snapshot_ignored"),
		SnapshotPatterns:          cv.GetStringSlice("instance.package.snapshot_patterns"),
		ToggledWorkflows:          cv.GetStringSlice("instance.package.toggled_workflows"),
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
		return nil, fmt.Errorf("%s > package at path '%s' does not exist", pm.instance.IDColor(), remotePath)
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
		return nil, fmt.Errorf("%s > cannot request package list: %w", pm.instance.IDColor(), err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot request package list: %s", pm.instance.IDColor(), resp.Status())
	}
	res := new(pkg.List)
	if err = fmtx.UnmarshalJSON(resp.RawBody(), res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse package list response: %w", pm.instance.IDColor(), err)
	}
	return res, nil
}

func (pm *PackageManager) Find(pid string) (*pkg.ListItem, error) {
	item, err := pm.findInternal(pid)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot find package '%s': %w", pm.instance.IDColor(), pid, err)
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
		return nil, fmt.Errorf("%s > cannot request package list: %w", pm.instance.IDColor(), err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot request package list: %s", pm.instance.IDColor(), resp.Status())
	}
	res := new(pkg.List)
	if err = fmtx.UnmarshalJSON(resp.RawBody(), res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse package list response: %w", pm.instance.IDColor(), err)
	}
	item, ok := lo.Find(res.List, func(p pkg.ListItem) bool { return p.PID == pid })
	if ok {
		return &item, nil
	}
	return nil, nil
}

func (pm *PackageManager) IsSnapshot(localPath string) bool {
	return !pm.SnapshotIgnored && stringsx.MatchSome(pathx.Normalize(localPath), pm.SnapshotPatterns)
}

func copyPackageDefaultFiles(targetTmpDir string, data map[string]any) error {
	if err := pathx.DeleteIfExists(targetTmpDir); err != nil {
		return fmt.Errorf("cannot delete temporary dir '%s': %w", targetTmpDir, err)
	}
	return fs.WalkDir(pkg.VaultFS, ".", func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}
		bytes, err := pkg.VaultFS.ReadFile(path)
		if err != nil {
			return err
		}
		return tplx.RenderFile(targetTmpDir+strings.ReplaceAll(strings.TrimPrefix(path, "vault"), "$", ""), string(bytes), data)
	})
}

type PackageCreateOpts struct {
	PID         string
	FilterRoots []string
	FilterFile  string
	ContentDir  string
	ContentFile string
}

func (pm *PackageManager) Create(opts PackageCreateOpts) (string, error) {
	log.Infof("%s > creating package '%s'", pm.instance.IDColor(), opts.PID)
	pidConfig, err := pkg.ParsePID(opts.PID)
	if err != nil {
		return "", err
	}

	tmpDir := pathx.RandomDir(pm.tmpDir(), "pkg_create")
	tmpFile := pathx.RandomFileName(pm.tmpDir(), "pkg_create", ".zip")
	defer func() {
		_ = pathx.DeleteIfExists(tmpDir)
		_ = pathx.DeleteIfExists(tmpFile)
	}()
	if len(opts.FilterRoots) == 0 && opts.FilterFile == "" {
		opts.FilterRoots = []string{determineFilterRoot(opts)}
	}
	data := map[string]any{
		"Pid":         opts.PID,
		"Group":       pidConfig.Group,
		"Name":        pidConfig.Name,
		"Version":     pidConfig.Version,
		"FilterRoots": opts.FilterRoots,
	}
	if err = copyPackageDefaultFiles(tmpDir, data); err != nil {
		return "", err
	}
	if opts.FilterFile != "" {
		if err = filex.Copy(opts.FilterFile, filepath.Join(tmpDir, "META-INF", "vault", FilterXML), true); err != nil {
			return "", err
		}
	}
	if err = content.Zip(tmpDir, tmpFile); err != nil {
		return "", err
	}
	response, err := pm.instance.http.Request().
		SetFile("package", tmpFile).
		SetMultipartFormData(map[string]string{"force": "true"}).
		Post(ServiceJsonPath + "/?cmd=upload")
	if err != nil {
		return "", fmt.Errorf("%s > cannot create package '%s': %w", pm.instance.IDColor(), opts.PID, err)
	} else if response.IsError() {
		return "", fmt.Errorf("%s > cannot create package '%s': %s", pm.instance.IDColor(), opts.PID, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return "", fmt.Errorf("%s > cannot create package '%s'; cannot parse response: %w", pm.instance.IDColor(), opts.PID, err)
	}
	if !status.Success {
		return "", fmt.Errorf("%s > cannot create package '%s'; unexpected status: %s", pm.instance.IDColor(), opts.PID, status.Message)
	}
	log.Infof("%s > created package '%s'", pm.instance.IDColor(), opts.PID)
	return status.Path, nil
}

func determineFilterRoot(opts PackageCreateOpts) string {
	if opts.ContentDir != "" {
		return strings.Split(opts.ContentDir, content.JCRRoot)[1]
	}

	if opts.ContentFile != "" {
		contentFile := strings.Split(opts.ContentFile, content.JCRRoot)[1]
		if content.IsContentFile(opts.ContentFile) {
			return strings.ReplaceAll(contentFile, content.JCRContentFile, content.JCRContentNode)
		} else if strings.HasSuffix(contentFile, content.JCRContentFile) {
			contentFile = namespacePatternRegex.ReplaceAllString(contentFile, "$1:")
			return filepath.Dir(contentFile)
		} else if strings.HasSuffix(contentFile, content.JCRContentFileSuffix) {
			contentFile = namespacePatternRegex.ReplaceAllString(contentFile, "$1:")
			return strings.ReplaceAll(contentFile, content.JCRContentFileSuffix, "")
		}
		return contentFile
	}

	return ""
}

func (pm *PackageManager) Copy(remotePath string, destInstance *Instance) error {
	localPath := pathx.RandomFileName(pm.tmpDir(), "pkg_copy", ".zip")
	defer func() { _ = pathx.DeleteIfExists(localPath) }()
	if err := pm.Download(remotePath, localPath); err != nil {
		return err
	}
	destRemotePath, err := destInstance.PackageManager().Upload(localPath)
	if err != nil {
		return err
	}
	if err := destInstance.PackageManager().Install(destRemotePath); err != nil {
		return err
	}
	return nil
}

func (pm *PackageManager) tmpDir() string {
	if pm.instance.manager.aem.Detached() {
		return os.TempDir()
	}
	return pm.instance.manager.aem.baseOpts.TmpDir
}

type PackageFilter struct {
	Root  string              `json:"root"`
	Rules []PackageFilterRule `json:"rules"`
}

func NewPackageFilters(rootPaths []string) []PackageFilter {
	var filters []PackageFilter
	for _, root := range rootPaths {
		filters = append(filters, PackageFilter{Root: root, Rules: []PackageFilterRule{}})
	}
	return filters
}

type PackageFilterRule struct {
	Modifier string `json:"modifier"`
	Pattern  string `json:"pattern"`
}

func (pm *PackageManager) UpdateFilters(remotePath string, pid string, filters []PackageFilter) error {
	log.Infof("%s > updating filters of package '%s'", pm.instance.IDColor(), pid)
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
		return fmt.Errorf("%s > cannot update filters of package '%s': %w", pm.instance.IDColor(), pid, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot update filters of package '%s': %s", pm.instance.IDColor(), pid, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot update filters of package '%s'; cannot parse response: %w", pm.instance.IDColor(), pid, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot update filters of package '%s'; unexpected status: %s", pm.instance.IDColor(), pid, status.Message)
	}
	log.Infof("%s > updated filters of package '%s'", pm.instance.IDColor(), pid)
	return nil
}

func (pm *PackageManager) Download(remotePath string, localFile string) error {
	log.Infof("%s > downloading package '%s'", pm.instance.IDColor(), remotePath)
	if err := httpx.DownloadWithOpts(httpx.DownloadOpts{
		Client:   pm.instance.http.Client(),
		URL:      remotePath,
		File:     localFile,
		Override: true,
	}); err != nil {
		return fmt.Errorf("%s > cannot download package '%s': %w", pm.instance.IDColor(), remotePath, err)
	}
	log.Infof("%s > downloaded package '%s'", pm.instance.IDColor(), remotePath)
	return nil
}

func (pm *PackageManager) Build(remotePath string) error {
	log.Infof("%s > building package '%s'", pm.instance.IDColor(), remotePath)
	response, err := pm.instance.http.Request().Post(ServiceJsonPath + remotePath + "?cmd=build")
	if err != nil {
		return fmt.Errorf("%s > cannot build package '%s': %w", pm.instance.IDColor(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot build package '%s': %s", pm.instance.IDColor(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot build package '%s'; cannot parse response: %w", pm.instance.IDColor(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot build package '%s'; unexpected status: %s", pm.instance.IDColor(), remotePath, status.Message)
	}
	log.Infof("%s > built package '%s'", pm.instance.IDColor(), remotePath)
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
	if pm.UploadOptimized {
		return pm.uploadOptimized(localPath)
	}
	return pm.uploadBuffered(localPath)
}

// https://medium.com/@owlwalks/sending-big-file-with-minimal-memory-in-golang-8f3fc280d2c
// https://github.com/go-resty/resty/issues/309#issuecomment-1750659170
func (pm *PackageManager) uploadOptimized(localPath string) (string, error) {
	log.Infof("%s > uploading package '%s'", pm.instance.IDColor(), localPath)
	r, w := io.Pipe()
	m := multipart.NewWriter(w)
	go func() {
		defer func(w *io.PipeWriter) { _ = w.Close() }(w)
		defer func(m *multipart.Writer) { _ = m.Close() }(m)
		part, err := m.CreateFormFile("package", filepath.Base(localPath))
		if err != nil {
			return
		}
		file, err := os.Open(localPath)
		if err != nil {
			return
		}
		defer func(file *os.File) { _ = file.Close() }(file)
		if _, err = io.Copy(part, file); err != nil {
			return
		}
	}()
	request, err := http.NewRequest("POST", pm.instance.HTTP().BaseURL()+ServiceJsonPath+"/?cmd=upload&force=true", r)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", m.FormDataContentType())
	request.SetBasicAuth(pm.instance.user, pm.instance.password)
	cv := pm.instance.manager.aem.config.Values()
	transport := &http.Transport{}
	if cv.GetBool("instance.http.ignore_ssl_errors") {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{Transport: transport}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("%s > cannot upload package '%s': %w", pm.instance.IDColor(), localPath, err)
	} else if response.StatusCode > 399 {
		return "", fmt.Errorf("%s > cannot upload package '%s': %s", pm.instance.IDColor(), localPath, response.Status)
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.Body, &status); err != nil {
		return "", fmt.Errorf("%s > cannot upload package '%s'; cannot parse response: %w", pm.instance.IDColor(), localPath, err)
	}
	if !status.Success {
		return "", fmt.Errorf("%s > cannot upload package '%s'; %s", pm.instance.IDColor(), localPath, pm.interpretFail(status.Message))
	}
	log.Infof("%s > uploaded package '%s'", pm.instance.IDColor(), localPath)
	return status.Path, nil
}

func (pm *PackageManager) uploadBuffered(localPath string) (string, error) {
	log.Infof("%s > uploading package '%s'", pm.instance.IDColor(), localPath)
	response, err := pm.instance.http.Request().
		SetFile("package", localPath).
		SetMultipartFormData(map[string]string{"force": "true"}).
		Post(ServiceJsonPath + "/?cmd=upload")
	if err != nil {
		return "", fmt.Errorf("%s > cannot upload package '%s': %w", pm.instance.IDColor(), localPath, err)
	} else if response.IsError() {
		return "", fmt.Errorf("%s > cannot upload package '%s': %s", pm.instance.IDColor(), localPath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return "", fmt.Errorf("%s > cannot upload package '%s'; cannot parse response: %w", pm.instance.IDColor(), localPath, err)
	}
	if !status.Success {
		return "", fmt.Errorf("%s > cannot upload package '%s'; %s", pm.instance.IDColor(), localPath, pm.interpretFail(status.Message))
	}
	log.Infof("%s > uploaded package '%s'", pm.instance.IDColor(), localPath)
	return status.Path, nil
}

func (pm *PackageManager) interpretFail(message string) string {
	if strings.Contains(strings.ToLower(message), "inaccessible value") {
		return fmt.Sprintf("probably no disk space left (server respond with '%s')", message) // https://forums.adobe.com/thread/2338290
	}
	if strings.Contains(strings.ToLower(message), "package file parameter missing") {
		return fmt.Sprintf("probably no disk space left (server respond with '%s')", message)
	}
	return fmt.Sprintf("unexpected status: %s", message)
}

func (pm *PackageManager) Install(remotePath string) error {
	if pm.InstallHTMLEnabled {
		return pm.installHTML(remotePath)
	}
	return pm.installJSON(remotePath)
}

func (pm *PackageManager) installJSON(remotePath string) error {
	log.Infof("%s > installing package '%s'", pm.instance.IDColor(), remotePath)
	response, err := pm.instance.http.Request().SetFormData(pm.installParams()).Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("%s > cannot install package '%s': %w", pm.instance.IDColor(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot install package '%s': '%s'", pm.instance.IDColor(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot install package '%s'; cannot parse JSON response: %w", pm.instance.IDColor(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot install package '%s'; unexpected status: %s", pm.instance.IDColor(), remotePath, status.Message)
	}
	log.Infof("%s > installed package '%s'", pm.instance.IDColor(), remotePath)
	return nil
}

func (pm *PackageManager) installHTML(remotePath string) error {
	log.Infof("%s > installing package '%s'", pm.instance.IDColor(), remotePath)

	response, err := pm.instance.http.Request().SetFormData(pm.installParams()).Post(ServiceHtmlPath + remotePath)
	if err != nil {
		return fmt.Errorf("%s > cannot install package '%s': %w", pm.instance.IDColor(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot install package '%s': '%s'", pm.instance.IDColor(), remotePath, response.Status())
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
			return fmt.Errorf("%s > cannot install package '%s': cannot open HTML report file '%s'", pm.instance.IDColor(), remotePath, htmlFilePath)
		}
		defer func() { _ = htmlFile.Close() }()
		htmlWriter = bufio.NewWriter(htmlFile)
		defer func() { _ = htmlWriter.Flush() }()
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
				return fmt.Errorf("%s > cannot install package '%s': cannot write to HTML report file '%s'", pm.instance.IDColor(), remotePath, htmlFilePath)
			}
		} else {
			fmt.Println(htmlLine)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%s > cannot install package '%s': cannot parse HTML response: %w", pm.instance.IDColor(), remotePath, err)
	}

	failure := !success && !successWithErrors
	if failure || (successWithErrors && pm.InstallHTMLStrict) {
		if pm.InstallHTMLConsole {
			return fmt.Errorf("%s > cannot install package '%s': HTML output contains errors", pm.instance.IDColor(), remotePath)
		}
		return fmt.Errorf("%s > cannot install package '%s': HTML report contains errors '%s'", pm.instance.IDColor(), remotePath, htmlFilePath)
	}
	if successWithErrors {
		log.Warnf("%s > installed package '%s': HTML response contains errors: %s", pm.instance.IDColor(), remotePath, err)
		return nil
	}
	log.Infof("%s > installed package '%s'", pm.instance.IDColor(), remotePath)
	return nil
}

func (pm *PackageManager) installParams() map[string]string {
	return map[string]string{
		"cmd":                "install",
		"recursive":          fmt.Sprintf("%v", pm.InstallRecursive),
		"autosave":           fmt.Sprintf("%d", pm.InstallSaveThreshold),
		"acHandling":         pm.InstallACHandling,
		"dependencyHandling": pm.InstallDependencyHandling,
	}
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
			log.Infof("%s > skipped deploying package '%s'", pm.instance.IDColor(), localPath)
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
	log.Infof("%s > uninstalling package '%s'", pm.instance.IDColor(), remotePath)
	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "uninstall"}).
		Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("%s > cannot uninstall package '%s': %w", pm.instance.IDColor(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot uninstall package '%s': %s", pm.instance.IDColor(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot uninstall package '%s'; cannot parse response: %w", pm.instance.IDColor(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot uninstall package '%s'; unexpected status: %s", pm.instance.IDColor(), remotePath, status.Message)
	}
	log.Infof("%s > uninstalled package '%s'", pm.instance.IDColor(), remotePath)
	return nil
}

func (pm *PackageManager) Delete(remotePath string) error {
	log.Infof("%s > deleting package '%s'", pm.instance.IDColor(), remotePath)
	response, err := pm.instance.http.Request().
		SetFormData(map[string]string{"cmd": "delete"}).
		Post(ServiceJsonPath + remotePath)
	if err != nil {
		return fmt.Errorf("%s > cannot delete package '%s': %w", pm.instance.IDColor(), remotePath, err)
	} else if response.IsError() {
		return fmt.Errorf("%s > cannot delete package '%s': %s", pm.instance.IDColor(), remotePath, response.Status())
	}
	var status pkg.CommandResult
	if err = fmtx.UnmarshalJSON(response.RawBody(), &status); err != nil {
		return fmt.Errorf("%s > cannot delete package '%s'; cannot parse response: %w", pm.instance.IDColor(), remotePath, err)
	}
	if !status.Success {
		return fmt.Errorf("%s > cannot delete package '%s'; unexpected status: %s", pm.instance.IDColor(), remotePath, status.Message)
	}
	log.Infof("%s > deleted package '%s'", pm.instance.IDColor(), remotePath)
	return nil
}

const (
	MgrPath         = "/crx/packmgr"
	ServicePath     = MgrPath + "/service"
	ServiceJsonPath = ServicePath + "/.json"
	ServiceHtmlPath = ServicePath + "/.html"
	ListJson        = MgrPath + "/list.jsp"
	UpdatePath      = MgrPath + "/update.jsp"

	FilterXML = "filter.xml"
)
