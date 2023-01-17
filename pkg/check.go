package pkg

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/netx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"github.com/wttech/aemc/pkg/osgi"
	"time"
)

type CheckResult struct {
	message string
	ok      bool
	err     error
	abort   bool
}

func (c *CheckResult) Message() string {
	return c.message
}

func (c *CheckResult) Ok() bool {
	return c.ok
}

func (c *CheckResult) Err() error {
	return c.err
}

type Checker interface {
	Check(instance Instance) CheckResult
	Spec() CheckSpec
}

type CheckSpec struct {
	Mandatory bool // indicates if next checks should be skipped if that particular one fails
}

func NewTimeoutChecker(expectedState string, duration time.Duration) TimeoutChecker {
	return TimeoutChecker{
		ExpectedState: expectedState,
		Duration:      duration,
		Started:       time.Now(),
	}
}

func (c TimeoutChecker) Check(_ Instance) CheckResult {
	now := time.Now()
	if now.After(c.Started.Add(c.Duration)) {
		return CheckResult{
			abort:   true,
			message: fmt.Sprintf("timeout after %s, expected state '%s' not reached", c.Duration, c.ExpectedState),
		}
	}

	return CheckResult{
		ok: true,
	}
}

func (c TimeoutChecker) Spec() CheckSpec {
	return CheckSpec{Mandatory: true}
}

type TimeoutChecker struct {
	Started       time.Time
	Duration      time.Duration
	ExpectedState string
}

func NewEventStableChecker() EventStableChecker {
	return EventStableChecker{
		ReceivedMaxAge: time.Second * 5,
		TopicsUnstable: []string{
			"org/osgi/framework/ServiceEvent/*",
			"org/osgi/framework/FrameworkEvent/*",
			"org/osgi/framework/BundleEvent/*",
		},
		DetailsIgnored: []string{
			"*.*MBean",
			"org.osgi.service.component.runtime.ServiceComponentRuntime",
			"java.util.ResourceBundle",
		},
	}
}

func NewBundleStableChecker() BundleStableChecker {
	return BundleStableChecker{
		SymbolicNamesIgnored: []string{},
	}
}

func (c BundleStableChecker) Spec() CheckSpec {
	return CheckSpec{Mandatory: true}
}

func NewInstallerChecker() InstallerChecker {
	return InstallerChecker{
		State: true,
		Pause: true,
	}
}

func NewStatusStoppedChecker() StatusStoppedChecker {
	return StatusStoppedChecker{}
}

type BundleStableChecker struct {
	SymbolicNamesIgnored []string
}

func (c BundleStableChecker) Check(instance Instance) CheckResult {
	bundles, err := instance.osgi.bundleManager.List()
	if err != nil {
		return CheckResult{
			ok:      false,
			message: "bundles unknown",
			err:     err,
		}
	}
	unstableBundles := bundles.FindUnstable()
	unstableBundleCount := len(unstableBundles)

	if unstableBundleCount > 0 {
		var message string
		if unstableBundleCount <= 10 {
			message = fmt.Sprintf("some bundles unstable (%d): %s", unstableBundleCount, unstableBundles[0].SymbolicName)
		} else {
			message = fmt.Sprintf("many bundles unstable (%s): %s", bundleStablePercent(bundles, unstableBundles), unstableBundles[0].SymbolicName)
		}
		return CheckResult{
			ok:      false,
			message: message,
		}
	}
	return CheckResult{
		ok:      true,
		message: "all bundles stable",
	}
}

type EventStableChecker struct {
	ReceivedMaxAge time.Duration
	TopicsUnstable []string
	DetailsIgnored []string
}

func (c EventStableChecker) Spec() CheckSpec {
	return CheckSpec{Mandatory: false}
}

func (c EventStableChecker) Check(instance Instance) CheckResult {
	events, err := instance.osgi.eventManager.List()
	if err != nil {
		return CheckResult{
			ok:      false,
			message: "events unknown",
			err:     err,
		}
	}

	nowTime := instance.Now()
	unstableEvents := lo.Filter(events.List, func(e osgi.Event, _ int) bool {
		receivedTime := instance.Time(e.Received)
		if !receivedTime.Add(c.ReceivedMaxAge).After(nowTime) {
			return false
		}
		if !stringsx.MatchSome(e.Topic, c.TopicsUnstable) {
			return false
		}
		if stringsx.MatchSome(e.Details(), c.DetailsIgnored) {
			return false
		}
		return true
	})
	unstableEventCount := len(unstableEvents)

	if unstableEventCount > 0 {
		message := fmt.Sprintf("some events unstable (%d): %s", unstableEventCount, unstableEvents[0].Details())
		return CheckResult{
			ok:      false,
			message: message,
		}
	}

	return CheckResult{
		ok:      true,
		message: "recent events stable",
	}
}

type InstallerChecker struct {
	State bool
	Pause bool
}

func (c InstallerChecker) Spec() CheckSpec {
	return CheckSpec{Mandatory: false}
}

func (c InstallerChecker) Check(instance Instance) CheckResult {
	installer := instance.Sling().Installer()
	if c.State {
		state, err := installer.State()
		if err != nil {
			return CheckResult{
				ok:      false,
				message: "installer state unknown",
				err:     err,
			}
		}
		if state.IsBusy() {
			return CheckResult{
				ok:      false,
				message: fmt.Sprintf("installer busy (%d)", state.ActiveResources()),
				err:     err,
			}
		}
	}
	if c.Pause {
		pauseCount, err := installer.CountPauses()
		if err != nil {
			return CheckResult{
				ok:      false,
				message: "installer pause unknown",
				err:     err,
			}
		}
		if pauseCount > 0 {
			return CheckResult{
				ok:      false,
				message: fmt.Sprintf("installer paused (%d)", pauseCount),
				err:     err,
			}
		}
	}
	return CheckResult{
		ok:      true,
		message: "installer idle",
	}
}

type StatusStoppedChecker struct{}

func (c StatusStoppedChecker) Spec() CheckSpec {
	return CheckSpec{Mandatory: true}
}

func (c StatusStoppedChecker) Check(instance Instance) CheckResult {
	if !instance.IsLocal() {
		return CheckResult{
			ok:      true,
			message: "stopped unknown",
		}
	}

	running := instance.local.IsRunning()
	if running {
		bundles, err := instance.osgi.bundleManager.List()
		if err != nil {
			return CheckResult{
				ok:      false,
				message: "not stopped (bundles unknown)",
			}
		}
		unstableBundles := bundles.FindUnstable()
		stablePercent := bundleStablePercent(bundles, unstableBundles)
		return CheckResult{
			ok:      false,
			message: fmt.Sprintf("not stopped (%s bundles stable)", stablePercent),
		}
	}

	return CheckResult{
		ok:      true,
		message: "stopped (not running)",
	}
}

func NewReachableChecker(reachable bool) ReachableHTTPChecker {
	return ReachableHTTPChecker{
		Mandatory: reachable,
		Reachable: reachable,
		Timeout:   time.Second * 3,
	}
}

type ReachableHTTPChecker struct {
	Mandatory bool
	Reachable bool
	Timeout   time.Duration
}

func (c ReachableHTTPChecker) Spec() CheckSpec {
	return CheckSpec{Mandatory: c.Mandatory}
}

func (c ReachableHTTPChecker) Check(instance Instance) CheckResult {
	address := fmt.Sprintf("%s:%s", instance.http.Hostname(), instance.http.Port())
	reachable, _ := netx.IsReachable(instance.http.Hostname(), instance.http.Port(), c.Timeout)
	if c.Reachable == reachable {
		return CheckResult{ok: true}
	}
	if reachable {
		return CheckResult{
			ok:      false,
			message: fmt.Sprintf("still reachable: %s", address),
		}
	}
	return CheckResult{
		ok:      false,
		message: fmt.Sprintf("not reachable: %s", address),
	}
}

func bundleStablePercent(bundles *osgi.BundleList, unstableBundles []osgi.BundleListItem) string {
	return stringsx.PercentExplained(bundles.Total()-len(unstableBundles), bundles.Total(), 0)
}
