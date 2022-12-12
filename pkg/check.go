package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"github.com/wttech/aemc/pkg/osgi"
	"time"
)

type CheckResult struct {
	message string
	ok      bool
	abort   bool
	err     error
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

type TimeoutChecker struct {
	instanceManager *InstanceManager

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

	if len(unstableBundles) > 0 {
		return CheckResult{
			ok:      false,
			message: fmt.Sprintf("%s bundle(s) stable", bundleStablePercent(bundles, unstableBundles)),
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
	var unstableEvents []osgi.Event

	for _, e := range events.List {
		receivedTime := instance.Time(e.Received)
		if !receivedTime.Add(c.ReceivedMaxAge).After(nowTime) {
			continue
		}
		if !stringsx.MatchAnyPattern(e.Topic, c.TopicsUnstable) {
			continue
		}
		/* TODO implement this
		if !checkEventDetails(e.Details, c.DetailsIgnored) {
			continue
		}
		*/
		unstableEvents = append(unstableEvents, e)
	}

	if len(unstableEvents) > 0 {
		return CheckResult{
			ok:      false,
			message: fmt.Sprintf("%d event(s) unstable", len(unstableEvents)),
		}
	}

	return CheckResult{
		ok:      true,
		message: "recent events stable",
	}
}

type StatusStoppedChecker struct {
	instanceManager *InstanceManager
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
			message: fmt.Sprintf("not stopped - %s bundle(s) stable", stablePercent),
		}
	}

	return CheckResult{
		ok:      true,
		message: "stopped (not running)",
	}
}

func bundleStablePercent(bundles *osgi.BundleList, unstableBundles []osgi.BundleListItem) string {
	return stringsx.PercentExplained(bundles.Total()-len(unstableBundles), bundles.Total(), 0)
}
