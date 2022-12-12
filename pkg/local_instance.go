package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/osx"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
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

func (li *LocalInstance) Instance() *Instance {
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

func (li *LocalInstance) Opts() *LocalOpts {
	return li.instance.manager.LocalOpts
}

func (li *LocalInstance) Dir() string {
	return fmt.Sprintf("%s/%s", li.Opts().RootPath, li.instance.ID())
}

func (li *LocalInstance) binScript(name string) string {
	return fmt.Sprintf("%s/crx-quickstart/bin/%s", li.Dir(), name)
}

func (li *LocalInstance) DebugPort() string {
	return "1" + li.instance.http.Port()
}

func (li *LocalInstance) Create() error {
	if err := osx.PathEnsure(li.Dir()); err != nil {
		return fmt.Errorf("cannot create dir for instance '%s': %w", li.instance.ID(), err)
	}

	cmd := li.verboseCommand(
		li.Opts().JavaOpts.Executable(), "-jar",
		li.Opts().QuickstartOpts.DistPath, "-unpack",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot unpack files for instance '%s': %w", li.instance.ID(), err)
	}

	err := li.createLockSave()
	if err != nil {
		return err
	}

	return nil
}

func (li *LocalInstance) createLockSave() error {
	err := osx.FileWrite(li.createLockPath(), time.Now().String())
	if err != nil {
		return fmt.Errorf("cannot save lock 'create' for instance '%s': %w", li.instance.ID(), err)
	}
	return nil
}

func (li *LocalInstance) createLockPath() string {
	return fmt.Sprintf("%s/lock/create.txt", li.Dir())
}

func (li *LocalInstance) IsCreated() bool {
	return osx.PathExists(li.createLockPath())
}

func (li *LocalInstance) Start() error {
	if !li.IsCreated() {
		return fmt.Errorf("cannot start instance as it is not created")
	}

	// TODO enforce 'java' to be always from JAVA_PATH (update $PATH var accordingly)
	cmd := li.verboseCommand("/bin/sh", li.binScript("start"))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot start instance '%s': %w", li.instance.ID(), err)
	}

	err := li.upLockSave()
	if err != nil {
		return err
	}

	return nil
}

func (li *LocalInstance) Stop() error {
	if !li.IsCreated() {
		return fmt.Errorf("cannot stop instance as it is not created")
	}

	// TODO enforce 'java' to be always from JAVA_PATH (update $PATH var accordingly)
	cmd := li.verboseCommand("/bin/sh", li.binScript("stop"))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot stop instance '%s': %w", li.instance.ID(), err)
	}

	err := li.upLockDelete()
	if err != nil {
		return err
	}

	return nil
}

func (li *LocalInstance) Restart() error {
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

func (li *LocalInstance) Await(state string, condition func() bool, timeout time.Duration) error {
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

func (li *LocalInstance) AwaitRunning() error {
	return li.Await("running", func() bool { return li.IsRunning() }, time.Minute*30)
}

func (li *LocalInstance) AwaitNotRunning() error {
	return li.Await("not running", func() bool { return !li.IsRunning() }, time.Minute*10)
}

type LocalStatus int

const (
	LocalStatusRunning     LocalStatus = 0
	LocalStatusDead        LocalStatus = 1
	LocalStatusNotRunning  LocalStatus = 3
	LocalStatusUnreachable LocalStatus = 4
	LocalStatusUnknown     LocalStatus = -1
)

func (s LocalStatus) String() string {
	return localStatuses()[s]
}

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
	}
}

func (li *LocalInstance) Status() (LocalStatus, error) {
	if !li.IsCreated() {
		return LocalStatusUnknown, fmt.Errorf("cannot check status of instance as it is not created")
	}

	cmd := li.quietCommand("/bin/sh", li.binScript("status"))

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return LocalStatus(exitCode), nil
}

func (li *LocalInstance) IsRunning() bool {
	if !li.IsCreated() {
		return false
	}

	status, err := li.Status()
	if err != nil {
		return false
	}

	return status == LocalStatusRunning
}

func (li *LocalInstance) Delete() error {
	if li.IsRunning() {
		return fmt.Errorf("cannot delete instance as it is running")
	}
	err := osx.PathDelete(li.Dir())
	if err != nil {
		return fmt.Errorf("cannot delete instance properly: %w", err)

	}
	return nil
}

func (li *LocalInstance) verboseCommand(name string, arg ...string) *exec.Cmd {
	cmd := li.quietCommand(name, arg...)

	out := li.instance.manager.aem.output
	cmd.Stdout = out
	cmd.Stderr = out

	return cmd
}

func (li *LocalInstance) quietCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Dir = li.Dir()
	cmd.Env = append(os.Environ(),
		"JAVA_HOME="+li.Opts().JavaOpts.HomePath,
		"CQ_PORT="+li.instance.http.Port(),
		"CQ_RUNMODE="+li.instance.local.RunModesString(),
		"CQ_JVM_OPTS="+li.instance.local.JVMOptsString(),
	)

	return cmd
}

func (li *LocalInstance) RunModesString() string {
	result := []string{string(li.instance.IDInfo().Role)}
	result = append(result, li.RunModes...)
	sort.Strings(result)
	return strings.Join(lo.Uniq[string](result), ",")
}

func (li *LocalInstance) JVMOptsString() string {
	result := li.JvmOpts
	sort.Strings(result)
	return strings.Join(result, " ")
}

func (li *LocalInstance) OutOfDate() bool {
	return !li.UpToDate()
}

func (li *LocalInstance) UpToDate() bool {
	return li.upLockCurrent() == li.upLockSaved()
}

func (li *LocalInstance) upLockCurrent() upToDateLock {
	return upToDateLock{
		Version:  li.Version,
		HTTPPort: li.instance.HTTP().Port(),
		RunModes: li.RunModesString(),
		JVMOpts:  li.JVMOptsString(),
	}
}

func (li *LocalInstance) upLockSave() error {
	yml, err := fmtx.MarshalYAML(li.upLockCurrent())
	if err != nil {
		return fmt.Errorf("cannot serialize up lock of instance '%s' to YML: %w", li.instance.ID(), err)
	}
	err = osx.FileWrite(li.upLockPath(), yml)
	if err != nil {
		return fmt.Errorf("cannot save instance up lock file '%s': %w", li.upLockPath(), err)
	}
	return nil
}

func (li *LocalInstance) upLockDelete() error {
	err := osx.PathDelete(li.upLockPath())
	if err != nil {
		return fmt.Errorf("cannot delete instance up lock file '%s': %w", li.upLockPath(), err)
	}
	return nil
}

func (li *LocalInstance) upLockSaved() upToDateLock {
	var result = upToDateLock{}
	if osx.PathExists(li.upLockPath()) {
		err := fmtx.UnmarshalFileInFormat(fmtx.YML, li.upLockPath(), &result)
		if err != nil {
			log.Warn(fmt.Sprintf("cannot read instance up lock file '%s': %s", li.upLockPath(), err))
		}
	}
	return result
}

// up lock helps with restarting AEM when it is out-of-date (when changed: run modes, http port, jvm opts)
// but should not block turning on/off on demand as out-of-date-ness is too complex to diagnose
func (li *LocalInstance) upLockPath() string {
	return fmt.Sprintf("%s/lock/up.yml", li.Dir())
}

type upToDateLock struct {
	Version string

	JVMOpts  string
	RunModes string
	HTTPPort string
}
