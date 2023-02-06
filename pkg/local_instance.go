package pkg

import (
	"fmt"
	"github.com/magiconair/properties"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/cryptox"
	"github.com/wttech/aemc/pkg/common/execx"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/netx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"github.com/wttech/aemc/pkg/common/timex"
	"github.com/wttech/aemc/pkg/instance"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/osx"
)

type LocalInstance struct {
	instance *Instance

	Version    string
	JvmOpts    []string
	StartOpts  []string
	RunModes   []string
	EnvVars    []string
	SecretVars []string
	SlingProps []string
}

type LocalInstanceState struct {
	ID           string   `yaml:"id" json:"id"`
	URL          string   `json:"url" json:"url"`
	AemVersion   string   `yaml:"aem_version" json:"aemVersion"`
	Attributes   []string `yaml:"attributes" json:"attributes"`
	RunModes     []string `yaml:"run_modes" json:"runModes"`
	HealthChecks []string `yaml:"health_checks" json:"healthChecks"`
	Dir          string   `yaml:"dir" json:"dir"`
}

const (
	LocalInstanceScriptStart     = "start"
	LocalInstanceScriptStop      = "stop"
	LocalInstanceScriptStatus    = "status"
	LocalInstanceBackupExtension = "aemb.tar.zst"
	LocalInstanceUser            = "admin"
	LocalInstanceWorkDirName     = common.AppId
	LocalInstanceNameCommon      = "common"
	LocalInstanceSecretsDir      = "conf/secret"
)

func (li LocalInstance) Instance() *Instance {
	return li.instance
}

func NewLocal(i *Instance) *LocalInstance {
	li := &LocalInstance{instance: i}
	li.Version = "1"
	li.StartOpts = []string{}
	li.RunModes = []string{"local"}
	li.JvmOpts = []string{
		"-Djava.io.tmpdir=" + pathx.Abs(common.TmpDir),
		"-Duser.language=en", "-Duser.country=US", "-Duser.timezone=UTC", "-Duser.name=" + common.AppId,
	}
	li.EnvVars = []string{}
	li.SecretVars = []string{}
	li.SlingProps = []string{}
	return li
}

func (li LocalInstance) State() LocalInstanceState {
	return LocalInstanceState{
		ID:           li.instance.id,
		URL:          li.instance.http.BaseURL(),
		Attributes:   li.instance.Attributes(),
		AemVersion:   li.instance.AemVersion(),
		RunModes:     li.instance.RunModes(),
		HealthChecks: li.instance.HealthChecks(),
		Dir:          li.Dir(),
	}
}

func (li LocalInstance) Opts() *LocalOpts {
	return li.instance.manager.LocalOpts
}

func (li LocalInstance) Name() string {
	id := li.instance.IDInfo()
	if id.Classifier != "" {
		return string(id.Role) + instance.IDDelimiter + id.Classifier
	}
	return string(id.Role)
}

func (li LocalInstance) Dir() string {
	return pathx.Normalize(pathx.Abs(fmt.Sprintf("%s/%s", li.Opts().UnpackDir, li.Name())))
}

func (li LocalInstance) WorkDir() string {
	return fmt.Sprintf("%s/%s", li.Dir(), LocalInstanceWorkDirName)
}

func (li LocalInstance) OverrideDirs() []string {
	return lo.Map([]string{
		fmt.Sprintf("%s/%s", li.Opts().OverrideDir, LocalInstanceNameCommon),
		fmt.Sprintf("%s/%s", li.Opts().OverrideDir, li.Name()),
	}, func(p string, _ int) string {
		return pathx.Normalize(pathx.Abs(p))
	})
}

func (li LocalInstance) overrideDirsChecksum() (string, error) {
	return filex.ChecksumPaths(lo.Filter(li.OverrideDirs(), func(d string, _ int) bool { return pathx.Exists(d) }), []string{})
}

func (li LocalInstance) LockDir() string {
	return fmt.Sprintf("%s/%s/lock", li.WorkDir(), common.VarDirName)
}

func (li LocalInstance) QuickstartDir() string {
	return fmt.Sprintf("%s/%s", li.Dir(), "crx-quickstart")
}

func (li LocalInstance) binScriptWindows(name string) string {
	return fmt.Sprintf("%sbin/%s.bat", li.QuickstartDir(), name)
}

func (li LocalInstance) binScriptUnix(name string) string {
	return pathx.Normalize(fmt.Sprintf("%s/bin/%s", li.QuickstartDir(), name))
}

func (li LocalInstance) binCbpExecutable() string {
	return pathx.Normalize(fmt.Sprintf("%s/bin/cbp.exe", li.WorkDir()))
}

func (li LocalInstance) LicenseFile() string {
	return li.Dir() + "/" + LicenseFilename
}

func (li LocalInstance) Create() error {
	log.Infof("creating instance '%s'", li.instance.ID())
	if err := pathx.DeleteIfExists(li.Dir()); err != nil {
		return fmt.Errorf("cannot clean up dir for instance '%s': %w", li.instance.ID(), err)
	}
	if err := pathx.Ensure(li.Dir()); err != nil {
		return fmt.Errorf("cannot create dir for instance '%s': %w", li.instance.ID(), err)
	}
	if err := li.unpackJarFile(); err != nil {
		return err
	}
	if err := li.copyLicenseFile(); err != nil {
		return err
	}
	if err := li.copyCbpExecutable(); err != nil {
		return err
	}
	if err := li.correctFiles(); err != nil {
		return err
	}
	if err := li.createLock().Lock(); err != nil {
		return err
	}
	log.Infof("created instance '%s'", li.instance.ID())
	return nil
}

func (li LocalInstance) createLock() osx.Lock[localInstanceCreateLock] {
	return osx.NewLock(fmt.Sprintf("%s/create.yml", li.LockDir()), func() (localInstanceCreateLock, error) {
		return localInstanceCreateLock{Created: time.Now()}, nil
	})
}

type localInstanceCreateLock struct {
	Created time.Time `yaml:"created"`
}

func (li LocalInstance) unpackJarFile() error {
	log.Infof("unpacking files for instance '%s'", li.instance.ID())
	jar, err := li.Opts().Jar()
	if err != nil {
		return err
	}
	cmd, err := li.Opts().JavaOpts.Command("-jar", pathx.Abs(jar), "-unpack")
	if err != nil {
		return err
	}
	cmd.Dir = li.Dir()
	li.instance.manager.aem.CommandOutput(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot unpack files for instance '%s': %w", li.instance.ID(), err)
	}
	return nil
}

func (li LocalInstance) copyLicenseFile() error {
	source := pathx.Abs(li.Opts().Quickstart.LicenseFile)
	dest := pathx.Abs(li.LicenseFile())
	log.Infof("copying license file from '%s' to '%s'", source, dest)
	err := filex.Copy(source, dest)
	if err != nil {
		return fmt.Errorf("cannot copy license file from '%s' to '%s': %s", source, dest, err)
	}
	return nil
}

func (li LocalInstance) copyCbpExecutable() error {
	dest := li.binCbpExecutable()
	log.Infof("copying CBP executable to '%s'", dest)
	if err := filex.Write(dest, instance.CbpExecutable); err != nil {
		return fmt.Errorf("cannot copy CBP executable to '%s': %s", dest, err)
	}
	return nil
}

func (li LocalInstance) correctFiles() error {
	filex.AmendString(li.binScriptUnix(LocalInstanceScriptStart), func(content string) string {
		content = strings.ReplaceAll(content, // introduce CQ_START_OPTS (not available by default)
			"set START_OPTS=start -c %CurrDirName% -i launchpad",
			"set START_OPTS=start -c %CurrDirName% -i launchpad %CQ_START_OPTS%",
		)
		return content
	})
	filex.AmendString(li.binScriptWindows(LocalInstanceScriptStart), func(content string) string {
		content = strings.ReplaceAll(content, // update 'timeout' to 'ping' as of it does not work when called from process without GUI
			"timeout /T 1 /NOBREAK >nul",
			"ping 127.0.0.1 -n 3 > nul",
		)
		content = strings.ReplaceAll(content, // force instance to be launched in background (it is windowed by default)
			"start \"CQ\" cmd.exe /C java %CQ_JVM_OPTS% -jar %CurrDirName%\\%CQ_JARFILE% %START_OPTS%",
			LocalInstanceWorkDirName+"\\bin\\cbp.exe cmd.exe /C \"java %CQ_JVM_OPTS% -jar %CurrDirName%\\%CQ_JARFILE% %START_OPTS% 1> %CurrDirName%\\logs\\stdout.log 2>&1\"",
		)
		content = strings.ReplaceAll(content, // introduce CQ_START_OPTS (not available by default)
			"set START_OPTS=start -c %CurrDirName% -i launchpad",
			"set START_OPTS=start -c %CurrDirName% -i launchpad %CQ_START_OPTS%",
		)
		return content
	})
	return nil
}

func (li LocalInstance) IsCreated() bool {
	return li.createLock().IsLocked()
}

func (li LocalInstance) IsInitialized() bool {
	return li.startLock().IsLocked()
}

func (li LocalInstance) Start() error {
	if !li.IsCreated() {
		return fmt.Errorf("cannot start instance '%s' as it is not created", li.instance.ID())
	}
	if err := li.update(); err != nil {
		return err
	}
	log.Infof("starting instance '%s'", li.instance.ID())
	if err := li.checkPortsOpen(); err != nil {
		return err
	}
	cmd, err := li.binScriptCommand(LocalInstanceScriptStart, true)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot execute start script for instance '%s': %w", li.instance.ID(), err)
	}
	if err := li.awaitAuth(); err != nil {
		return err
	}
	if err := li.startLock().Lock(); err != nil {
		return err
	}
	log.Infof("started instance '%s'", li.instance.ID())
	return nil
}

func (li LocalInstance) StartAndAwait() error {
	if err := li.Start(); err != nil {
		return err
	}
	if err := li.instance.manager.AwaitStartedOne(*li.instance); err != nil {
		return err
	}
	return nil
}

func (li LocalInstance) update() error {
	if !li.IsInitialized() { // first boot
		if err := li.copyOverrideDirs(); err != nil {
			return err
		}
		if err := li.recreateSlingPropsFile(); err != nil {
			return err
		}
		if err := li.recreateSecretsDir(); err != nil {
			return err
		}
	} else { // next boot
		state, err := li.startLock().State()
		if err != nil {
			return err
		}
		if state.Locked.Password != state.Current.Password {
			if err := li.setPassword(); err != nil {
				return err
			}
		}
		if state.Locked.Overrides != state.Current.Overrides {
			if err := li.copyOverrideDirs(); err != nil {
				return err
			}
		}
		if state.Locked.SlingProps != state.Current.SlingProps {
			if err := li.recreateSlingPropsFile(); err != nil {
				return err
			}
		}
		if state.Locked.SecretVars != state.Current.SecretVars {
			if err := li.recreateSecretsDir(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (li LocalInstance) setPassword() error {
	return li.Opts().OakRun.SetPassword(li.Dir(), LocalInstanceUser, li.instance.password)
}

func (li LocalInstance) copyOverrideDirs() error {
	for _, src := range lo.Filter(li.OverrideDirs(), func(s string, _ int) bool { return pathx.Exists(s) }) {
		log.Infof("copying instance override files from dir '%s' to '%s'", src, li.Dir())
		if err := filex.CopyDir(src, li.Dir()); err != nil {
			return err
		}
		log.Infof("copied instance override files from dir '%s' to '%s'", src, li.Dir())
	}
	return nil
}

func (li LocalInstance) secretsDir() string {
	return fmt.Sprintf("%s/%s", li.QuickstartDir(), LocalInstanceSecretsDir)
}

func (li LocalInstance) recreateSlingPropsFile() error {
	filePath := fmt.Sprintf("%s/conf/sling.properties", li.QuickstartDir())
	log.Infof("configuring instance Sling properties in file '%s'", filePath)
	propsCombined := append(li.SlingProps, "org.apache.felix.configadmin.plugin.interpolation.secretsdir=${sling.home}/"+LocalInstanceSecretsDir)
	propsLoaded, err := properties.LoadString(strings.Join(propsCombined, "\n"))
	if err != nil {
		return fmt.Errorf("cannot parse Sling properties of instance '%s'", li.instance.ID())
	}
	file, err := os.Create(filePath)
	defer file.Close()
	_, err = propsLoaded.Write(file, properties.ISO_8859_1)
	if err != nil {
		return fmt.Errorf("cannot save Sling properties file '%s'", filePath)
	}
	log.Infof("configured instance Sling properties in file '%s'", filePath)
	return nil
}

func (li LocalInstance) recreateSecretsDir() error {
	dir := li.secretsDir()
	if err := pathx.DeleteIfExists(dir); err != nil {
		return err
	}
	if len(li.SecretVars) > 0 {
		log.Infof("configuring instance secret vars in dir '%s'", dir)
		for _, secretVar := range li.SecretVars {
			k := stringsx.Before(secretVar, "=")
			v := stringsx.After(secretVar, "=")
			if err := filex.WriteString(fmt.Sprintf("%s/%s", dir, k), v); err != nil {
				return err
			}
		}
		log.Infof("configured instance secret vars in dir '%s'", dir)
	}
	return nil
}

func (li LocalInstance) checkPortsOpen() error {
	host := li.instance.http.Hostname()
	ports := []string{li.instance.http.Port()}
	for _, port := range ports {
		reachable, _ := netx.IsReachable(host, port, time.Second*3)
		if reachable {
			return fmt.Errorf("some process is already running on address '%s:%s'", host, port)
		}
	}
	return nil
}

func (li LocalInstance) startLock() osx.Lock[localInstanceStartLock] {
	return osx.NewLock(fmt.Sprintf("%s/start.yml", li.LockDir()), func() (localInstanceStartLock, error) {
		var zero localInstanceStartLock
		overrides, err := li.overrideDirsChecksum()
		if err != nil {
			return zero, err
		}
		return localInstanceStartLock{
			Version:    li.Version,
			HTTPPort:   li.instance.HTTP().Port(),
			RunModes:   strings.Join(li.RunModes, ","),
			JVMOpts:    strings.Join(li.JvmOpts, " "),
			Password:   cryptox.HashString(li.instance.password),
			EnvVars:    strings.Join(li.EnvVars, ","),
			SecretVars: cryptox.HashString(strings.Join(li.SecretVars, ",")),
			SlingProps: strings.Join(li.SlingProps, ","),
			Overrides:  overrides,
		}, nil
	})
}

type localInstanceStartLock struct {
	Version    string `yaml:"version"`
	JVMOpts    string `yaml:"jvm_opts"`
	RunModes   string `yaml:"run_modes"`
	HTTPPort   string `yaml:"http_port"`
	Password   string `yaml:"password"`
	Overrides  string `yaml:"overrides"`
	EnvVars    string `yaml:"env_vars"`
	SecretVars string `yaml:"secret_vars"`
	SlingProps string `yaml:"sling_props"`
}

func (li LocalInstance) Stop() error {
	if !li.IsCreated() {
		return fmt.Errorf("cannot stop instance as it is not created")
	}
	log.Infof("stopping instance '%s'", li.instance.ID())
	cmd, err := li.binScriptCommand(LocalInstanceScriptStop, true)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot execute stop script for instance '%s': %w", li.instance.ID(), err)
	}
	log.Infof("stopped instance '%s'", li.instance.ID())
	return nil
}

func (li LocalInstance) StopAndAwait() error {
	if err := li.Stop(); err != nil {
		return err
	}
	if err := li.instance.manager.AwaitStoppedOne(*li.instance); err != nil {
		return err
	}
	return nil
}

func (li LocalInstance) Clean() error {
	log.Infof("cleaning instance '%s'", li.instance.ID())
	err := pathx.DeleteIfExists(li.pidFile())
	if err != nil {
		return err
	}
	log.Infof("cleaned instance '%s'", li.instance.ID())
	return nil
}

func (li LocalInstance) Restart() error {
	downErr := li.Stop()
	if downErr != nil {
		return downErr
	}
	err := li.AwaitNotRunning()
	if err != nil {
		return err
	}
	upErr := li.Start()
	if upErr != nil {
		return upErr
	}
	return nil
}

func (li LocalInstance) IsKillable() bool {
	if !li.IsCreated() {
		return false
	}
	pid, err := li.PID()
	if err != nil {
		return false
	}
	return pid > 0
}

func (li LocalInstance) Kill() error {
	log.Infof("killing instance '%s'", li.instance.ID())
	pid, err := li.PID()
	if err != nil {
		return err
	}
	var cmd *exec.Cmd
	if osx.IsWindows() {
		cmd = execx.CommandShell([]string{"taskkill", "/F", "/PID", fmt.Sprintf("%d", pid)})
	} else {
		cmd = exec.Command("kill", "-9", fmt.Sprintf("%d", pid))
	}
	cmd.Dir = li.Dir()
	li.instance.manager.aem.CommandOutput(cmd)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot execute kill command for instance '%s' with PID '%d': %w", li.instance.ID(), pid, err)
	}
	file := li.pidFile()
	if err := pathx.DeleteIfExists(file); err != nil {
		return fmt.Errorf("cannot delete PID file '%s' for instance '%s': %w", file, li.instance.ID(), err)
	}
	log.Infof("killed instance '%s'", li.instance.ID())
	return nil
}

func (li LocalInstance) Await(stateChecker func() error, timeout time.Duration) error {
	started := time.Now()
	for {
		err := stateChecker()
		if err == nil {
			break
		}
		if time.Now().After(started.Add(timeout)) {
			return fmt.Errorf("instance '%s' awaiting reached timeout after %s", li.Instance().ID(), timeout)
		}
		log.Infof("%s | %s", li.instance.ID(), err)
		time.Sleep(time.Second * 5)
	}
	return nil
}

func (li LocalInstance) AwaitRunning() error {
	return li.Await(func() error {
		if !li.IsRunning() {
			return fmt.Errorf("not yet running")
		}
		return nil
	}, time.Minute*30)
}

func (li LocalInstance) AwaitNotRunning() error {
	return li.Await(func() error {
		if li.IsRunning() {
			return fmt.Errorf("still running")
		}
		return nil
	}, time.Minute*10)
}

// awaitAuth waits for a custom password to be in use (initially the default one is used instead)
func (li LocalInstance) awaitAuth() error {
	if li.IsInitialized() || li.instance.password == instance.PasswordDefault {
		return nil
	}

	log.Infof("awaiting auth on instance '%s'", li.instance.ID())
	err := li.Await(func() error {
		address := fmt.Sprintf("%s:%s", li.instance.http.Hostname(), li.instance.http.Port())
		reachable, _ := netx.IsReachable(li.instance.http.Hostname(), li.instance.http.Port(), time.Second*3)
		if !reachable {
			return fmt.Errorf("not reachable (%s)", address)
		}
		_, err := li.instance.osgi.bundleManager.List()
		if err != nil {
			defaultInstance, err := li.instance.manager.NewByURL(li.instance.http.BaseURL())
			if err != nil {
				return err
			}
			bundles, err := defaultInstance.osgi.bundleManager.List()
			if err != nil {
				return err
			}
			return fmt.Errorf("starting bundles (%s)", bundles.StablePercent())
		}
		return nil
	}, time.Minute*10)
	if err != nil {
		return err
	}
	log.Infof("awaited auth on instance '%s'", li.instance.ID())
	return nil
}

type LocalStatus int

const (
	LocalStatusUnknown     LocalStatus = -1
	LocalStatusRunning     LocalStatus = 0
	LocalStatusDead        LocalStatus = 1
	LocalStatusNotRunning  LocalStatus = 3
	LocalStatusUnreachable LocalStatus = 4
	LocalStatusError       LocalStatus = 127
)

func localStatusOf(name string) LocalStatus {
	for sk, sn := range localStatuses() {
		if sn == name {
			return sk
		}
	}
	return LocalStatusUnknown
}

func localStatuses() map[LocalStatus]string {
	return map[LocalStatus]string{
		LocalStatusRunning:     "running",
		LocalStatusDead:        "dead",
		LocalStatusNotRunning:  "not running",
		LocalStatusUnreachable: "unreachable",
		LocalStatusUnknown:     "unknown",
		LocalStatusError:       "status error",
	}
}

func (s LocalStatus) String() string {
	text, ok := localStatuses()[s]
	if !ok {
		return fmt.Sprintf("unknown (%d)", s)
	}
	return text
}

func (li LocalInstance) Status() (LocalStatus, error) {
	if !li.IsCreated() {
		return LocalStatusUnknown, fmt.Errorf("cannot check status of instance as it is not created")
	}
	cmd, err := li.binScriptCommand(LocalInstanceScriptStatus, false)
	if err != nil {
		return LocalStatusUnknown, err
	}
	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}
	return LocalStatus(exitCode), nil
}

func (li LocalInstance) IsRunning() bool {
	if !li.IsCreated() {
		return false
	}
	status, err := li.Status()
	if err != nil {
		return false
	}
	return status == LocalStatusRunning
}

func (li LocalInstance) Delete() error {
	if li.IsRunning() {
		return fmt.Errorf("cannot delete instance as it is running")
	}
	log.Infof("deleting instance '%s'", li.instance.ID())
	if err := pathx.Delete(li.Dir()); err != nil {
		return fmt.Errorf("cannot delete instance properly: %w", err)

	}
	log.Infof("deleted instance '%s'", li.instance.ID())
	return nil
}

func (li LocalInstance) binScriptCommand(name string, verbose bool) (*exec.Cmd, error) {
	var args []string
	if osx.IsWindows() { // note 'call' usage here; without it on Windows exit code is always 0
		args = []string{"call", li.binScriptWindows(name)}
	} else {
		args = []string{li.binScriptUnix(name)}
	}
	cmd := execx.CommandShell(args)
	cmd.Dir = li.Dir()
	env, err := li.Opts().JavaOpts.Env()
	if err != nil {
		return nil, err
	}
	env = append(env,
		"CQ_PORT="+li.instance.http.Port(),
		"CQ_RUNMODE="+li.instance.local.RunModesString(),
		"CQ_JVM_OPTS="+li.instance.local.JVMOptsString(),
		"CQ_START_OPTS="+li.instance.local.StartOptsString(),
	)
	env = append(env, li.EnvVars...)
	cmd.Env = env
	if verbose {
		li.instance.manager.aem.CommandOutput(cmd)
	}
	return cmd, nil
}

func (li LocalInstance) RunModesString() string {
	result := []string{string(li.instance.IDInfo().Role)}
	result = append(result, li.RunModes...)
	sort.Strings(result)
	return strings.Join(lo.Uniq[string](result), ",")
}

func (li LocalInstance) JVMOptsString() string {
	result := append([]string{}, li.JvmOpts...)

	// at the first boot admin password could be customized via property, at the next boot only via Oak Run
	if !li.IsInitialized() && li.instance.password != instance.PasswordDefault {
		result = append(result, fmt.Sprintf("-Dadmin.password=%s", li.instance.password))
	}
	sort.Strings(result)
	return strings.Join(result, " ")
}

func (li LocalInstance) StartOptsString() string {
	return strings.Join(li.StartOpts, " ")
}

func (li LocalInstance) OutOfDate() bool {
	return !li.UpToDate()
}

func (li LocalInstance) UpToDate() bool {
	check, err := li.startLock().State()
	if err != nil {
		log.Debugf("cannot check if instance '%s' is up-to-date: %s", li.instance.ID(), err)
		return false
	}
	return check.UpToDate
}

func (li LocalInstance) pidFile() string {
	return fmt.Sprintf("%s/crx-quickstart/conf/cq.pid", li.Dir())
}

func (li LocalInstance) PID() (int, error) {
	file := li.pidFile()
	str, err := filex.ReadString(file)
	if err != nil {
		return 0, fmt.Errorf("cannot read instance PID file '%s'", file)
	}
	strTrimmed := strings.TrimSpace(str)
	num, err := strconv.Atoi(strTrimmed)
	if err != nil {
		return 0, fmt.Errorf("cannot convert value '%s' to integer read from instance PID file '%s'", strTrimmed, file)
	}
	return num, nil
}

func (li LocalInstance) ProposeBackupFileToMake() string {
	nameParts := []string{li.Name()}
	if li.IsRunning() {
		nameParts = append(nameParts, li.instance.AemVersion())
	}
	nameParts = append(nameParts, timex.FileTimestampForNow())
	return fmt.Sprintf("%s/%s.%s", li.Opts().BackupDir, strings.Join(nameParts, "-"), LocalInstanceBackupExtension)
}

func (li LocalInstance) MakeBackup(file string) error {
	if !li.IsCreated() {
		return fmt.Errorf("cannot make backup of instance '%s' to file '%s' - instance not created", li.instance.ID(), file)
	}
	if li.IsRunning() {
		return fmt.Errorf("cannot make a backup of instance '%s' to file '%s' - instance cannot be running", li.instance.ID(), file)
	}
	log.Infof("making backup of instance '%s' to file '%s'", li.instance.ID(), file)
	_, err := filex.ArchiveWithChanged(li.Dir(), file)
	if err != nil {
		return fmt.Errorf("cannot make a backup of instance '%s' to file '%s': %w", li.instance.ID(), file, err)
	}
	log.Infof("made backup of instance '%s' to file '%s'", li.instance.ID(), file)
	return nil
}

func (li LocalInstance) ProposeBackupFileToUse() (string, error) {
	var pathPattern string
	if li.IsRunning() {
		aemVersion, err := li.instance.status.AemVersion()
		if err != nil {
			return "", err
		}
		pathPattern = fmt.Sprintf("%s/%s-%s-*.%s", li.Opts().BackupDir, li.Name(), aemVersion, LocalInstanceBackupExtension)
	} else {
		pathPattern = fmt.Sprintf("%s/%s-*.%s", li.Opts().BackupDir, li.Name(), LocalInstanceBackupExtension)
	}
	file, err := pathx.GlobSome(pathPattern)
	if err != nil {
		return "", fmt.Errorf("no backup file found to use for instance '%s': %w", li.instance.ID(), err)
	}
	return file, nil
}

func (li LocalInstance) UseBackup(file string, deleteCreated bool) error {
	if li.IsRunning() {
		return fmt.Errorf("cannot use backup of instance '%s' from file '%s' - instance cannot be running", li.instance.ID(), file)
	}
	if li.IsCreated() {
		if !deleteCreated {
			return fmt.Errorf("cannot use backup of instance '%s' from file '%s' - instance already created", li.instance.ID(), file)
		}
		if err := li.Delete(); err != nil {
			return err
		}
	}
	log.Infof("using backup of instance '%s' from file '%s'", li.instance.ID(), file)
	_, err := filex.UnarchiveWithChanged(file, li.Dir())
	if err != nil {
		return fmt.Errorf("cannot use a backup of instance '%s' from file '%s': %w", li.instance.ID(), file, err)
	}
	log.Infof("used backup of instance '%s' from file '%s'", li.instance.ID(), file)
	return nil
}
