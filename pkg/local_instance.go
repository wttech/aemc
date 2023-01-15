package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common/execx"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
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

	Version  string
	JvmOpts  []string
	RunModes []string
}

type LocalInstanceState struct {
	ID         string   `yaml:"id" json:"id"`
	URL        string   `json:"url" json:"url"`
	Attributes []string `yaml:"attributes" json:"attributes"`
	Dir        string   `yaml:"dir" json:"dir"`
}

func (li LocalInstance) Instance() *Instance {
	return li.instance
}

func NewLocal(i *Instance) *LocalInstance {
	return &LocalInstance{
		instance: i,
		Version:  "1",
		JvmOpts:  []string{"-server", "-Xmx1024m", "-Djava.awt.headless=true"},
		RunModes: []string{},
	}
}

func (li LocalInstance) State() LocalInstanceState {
	return LocalInstanceState{
		ID:         li.instance.id,
		URL:        li.instance.http.BaseURL(),
		Attributes: li.instance.Attributes(),
		Dir:        li.Dir(),
	}
}

func (li LocalInstance) Opts() *LocalOpts {
	return li.instance.manager.LocalOpts
}

func (li LocalInstance) Dir() string {
	return pathx.Abs(fmt.Sprintf("%s/%s", li.Opts().UnpackDir, li.instance.ID()))
}

func (li LocalInstance) binScript(name string) string {
	return fmt.Sprintf("%s/crx-quickstart/bin/%s", li.Dir(), name)
}

func (li LocalInstance) DebugPort() string {
	return "1" + li.instance.http.Port()
}

func (li LocalInstance) LicenseFile() string {
	return li.Dir() + "/" + LicenseFilename
}

func (li LocalInstance) Create() error {
	if err := pathx.Ensure(li.Dir()); err != nil {
		return fmt.Errorf("cannot create dir for instance '%s': %w", li.instance.ID(), err)
	}
	if err := li.unpackJarFile(); err != nil {
		return err
	}
	if err := li.copyLicenseFile(); err != nil {
		return err
	}
	if err := li.createLock().Lock(); err != nil {
		return err
	}
	return nil
}

func (li LocalInstance) createLock() osx.Lock[localInstanceCreateLock] {
	return osx.NewLock(fmt.Sprintf("%s/lock/create.yml", li.Dir()), localInstanceCreateLock{
		Created: time.Now(),
	})
}

type localInstanceCreateLock struct {
	Created time.Time
}

func (li LocalInstance) unpackJarFile() error {
	log.Infof("unpacking files for instance '%s'", li.instance.ID())
	jar, err := li.Opts().Jar()
	if err != nil {
		return err
	}
	cmd := li.verboseCommand(
		pathx.Abs(li.Opts().JavaOpts.Executable()), "-jar",
		pathx.Abs(jar), "-unpack",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot unpack files for instance '%s': %w", li.instance.ID(), err)
	}
	return nil
}

func (li LocalInstance) copyLicenseFile() error {
	source := pathx.Abs(li.Opts().Quickstart.LicenseFile)
	dest := pathx.Abs(li.LicenseFile())
	log.Infof("copying license file from '%s' to '%s' for instance '%s'", source, dest, li.instance.ID())
	err := filex.Copy(source, dest)
	if err != nil {
		return fmt.Errorf("cannot copy license file for instance '%s': %s", li.instance.ID(), err)
	}
	return nil
}

func (li LocalInstance) IsCreated() bool {
	return li.createLock().IsLocked()
}

func (li LocalInstance) Start() error {
	if !li.IsCreated() {
		return fmt.Errorf("cannot start instance as it is not created")
	}
	// TODO enforce 'java' to be always from JAVA_PATH (update $PATH var accordingly)
	cmd := li.verboseCommand(li.binScript("start"))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot execute start script for instance '%s': %w", li.instance.ID(), err)
	}
	if err := li.startLock().Lock(); err != nil {
		return err
	}
	return nil
}

func (li LocalInstance) startLock() osx.Lock[localInstanceStartLock] {
	return osx.NewLock(fmt.Sprintf("%s/lock/start.yml", li.Dir()), localInstanceStartLock{
		Version:  li.Version,
		HTTPPort: li.instance.HTTP().Port(),
		RunModes: li.RunModesString(),
		JVMOpts:  li.JVMOptsString(),
	})
}

type localInstanceStartLock struct {
	Version  string
	JVMOpts  string
	RunModes string
	HTTPPort string
}

func (li LocalInstance) Stop() error {
	if !li.IsCreated() {
		return fmt.Errorf("cannot stop instance as it is not created")
	}
	// TODO enforce 'java' to be always from JAVA_PATH (update $PATH var accordingly)
	cmd := li.verboseCommand(li.binScript("stop"))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot execute stop script for instance '%s': %w", li.instance.ID(), err)
	}
	if err := li.startLock().Unlock(); err != nil {
		return err
	}
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
	pid, err := li.PID()
	if err != nil {
		return err
	}
	var cmd *exec.Cmd
	if osx.IsWindows() {
		cmd = li.verboseCommand("taskkill", "/F", "/PID", fmt.Sprintf("%d", pid))
	} else {
		cmd = li.verboseCommand("kill", "-9", fmt.Sprintf("%d", pid))
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot execute kill command for instance '%s': %w", li.instance.ID(), err)
	}
	return nil
}

func (li LocalInstance) Await(state string, condition func() bool, timeout time.Duration) error {
	started := time.Now()
	for {
		if condition() {
			break
		}
		if time.Now().After(started.Add(timeout)) {
			return fmt.Errorf("instance '%s' awaiting state '%s' reached timeout after %s", state, li.Instance().ID(), timeout)
		}
		log.Infof("instance '%s' is awaiting state '%s'", li.instance.ID(), state)
		time.Sleep(time.Second * 5)
	}
	return nil
}

func (li LocalInstance) AwaitRunning() error {
	return li.Await("running", func() bool { return li.IsRunning() }, time.Minute*30)
}

func (li LocalInstance) AwaitNotRunning() error {
	return li.Await("not running", func() bool { return !li.IsRunning() }, time.Minute*10)
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
	cmd := li.quietCommand(li.binScript("status"))
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
	if err := pathx.Delete(li.Dir()); err != nil {
		return fmt.Errorf("cannot delete instance properly: %w", err)

	}
	return nil
}

func (li LocalInstance) verboseCommand(args ...string) *exec.Cmd {
	cmd := li.quietCommand(args...)

	out := li.instance.manager.aem.output
	cmd.Stdout = out
	cmd.Stderr = out

	return cmd
}

func (li LocalInstance) quietCommand(args ...string) *exec.Cmd {
	cmd := execx.CommandShell(args)
	cmd.Dir = li.Dir()
	cmd.Env = append(os.Environ(),
		"JAVA_HOME="+pathx.Abs(li.Opts().JavaOpts.HomeDir),
		"CQ_PORT="+li.instance.http.Port(),
		"CQ_RUNMODE="+li.instance.local.RunModesString(),
		"CQ_JVM_OPTS="+li.instance.local.JVMOptsString(),
	)

	return cmd
}

func (li LocalInstance) RunModesString() string {
	result := []string{string(li.instance.IDInfo().Role)}
	result = append(result, li.RunModes...)
	sort.Strings(result)
	return strings.Join(lo.Uniq[string](result), ",")
}

func (li LocalInstance) JVMOptsString() string {
	result := li.JvmOpts
	sort.Strings(result)
	return strings.Join(result, " ")
}

func (li LocalInstance) OutOfDate() bool {
	return !li.UpToDate()
}

func (li LocalInstance) UpToDate() bool {
	upToDate, err := li.startLock().IsUpToDate()
	if err != nil {
		log.Debugf("cannot check if instance '%s' is up-to-date: %s", li.instance.ID(), err)
		return false
	}
	return upToDate
}

func (li LocalInstance) PID() (int, error) {
	file := fmt.Sprintf("%s/crx-quickstart/conf/cq.pid", li.Dir())
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
