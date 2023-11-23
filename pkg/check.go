package pkg

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/lox"
	"github.com/wttech/aemc/pkg/common/netx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"github.com/wttech/aemc/pkg/osgi"
	"io"
	"strings"
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
	Check(ctx CheckContext, instance Instance) CheckResult
	Spec() CheckSpec
}

type CheckSpec struct {
	Skip      bool // occasionally skip checking e.g. due to performance reasons
	Mandatory bool // indicates if next checks should be skipped if that particular one fails
}

func NewAwaitChecker(opts *CheckOpts, expectedState string) AwaitChecker {
	cv := opts.manager.aem.config.Values()

	return AwaitChecker{
		ExpectedState: expectedState,
		Timeout:       cv.GetDuration(fmt.Sprintf("instance.check.await_%s.timeout", expectedState)),
	}
}

func (c AwaitChecker) Check(ctx CheckContext, _ Instance) CheckResult {
	now := time.Now()

	if now.After(ctx.Started.Add(c.Timeout)) {
		return CheckResult{
			abort:   true,
			message: fmt.Sprintf("timeout after %s, expected state '%s' not reached", c.Timeout, c.ExpectedState),
		}
	}

	return CheckResult{
		ok: true,
	}
}

func (c AwaitChecker) Spec() CheckSpec {
	return CheckSpec{Skip: false, Mandatory: true}
}

type AwaitChecker struct {
	Timeout       time.Duration
	ExpectedState string
}

func NewBundleStableChecker(opts *CheckOpts) BundleStableChecker {
	cv := opts.manager.aem.config.Values()

	return BundleStableChecker{
		Skip:                 cv.GetBool("instance.check.bundle_stable.skip"),
		SymbolicNamesIgnored: cv.GetStringSlice("instance.check.bundle_stable.symbolic_names_ignored"),
	}
}

func (c BundleStableChecker) Spec() CheckSpec {
	return CheckSpec{Skip: c.Skip, Mandatory: true}
}

type BundleStableChecker struct {
	Skip                 bool
	SymbolicNamesIgnored []string
}

func (c BundleStableChecker) Check(_ CheckContext, instance Instance) CheckResult {
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
		randomBundleSymbolicName := lox.Random(unstableBundles).SymbolicName
		if unstableBundleCount <= 10 {
			message = fmt.Sprintf("some bundles unstable (%d): '%s'", unstableBundleCount, randomBundleSymbolicName)
		} else {
			message = fmt.Sprintf("many bundles unstable (%s): '%s'", bundleStablePercent(bundles, unstableBundles), randomBundleSymbolicName)
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
	Skip           bool
	ReceivedMaxAge time.Duration
	TopicsUnstable []string
	DetailsIgnored []string
}

func NewEventStableChecker(opts *CheckOpts) EventStableChecker {
	cv := opts.manager.aem.config.Values()

	return EventStableChecker{
		Skip:           cv.GetBool("instance.check.event_stable.skip"),
		ReceivedMaxAge: cv.GetDuration("instance.check.event_stable.received_max_age"),
		TopicsUnstable: cv.GetStringSlice("instance.check.event_stable.topics_unstable"),
		DetailsIgnored: cv.GetStringSlice("instance.check.event_stable.details_ignored"),
	}
}

func (c EventStableChecker) Spec() CheckSpec {
	return CheckSpec{Skip: c.Skip, Mandatory: true}
}

func (c EventStableChecker) Check(_ CheckContext, instance Instance) CheckResult {
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
		message := fmt.Sprintf("some events unstable (%d): '%s'", unstableEventCount, unstableEvents[0].Details())
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

type ComponentStableChecker struct {
	Skip                     bool
	IgnoredPIDs              []string
	FailedActivationPIDs     []string
	UnsatisfiedReferencePIDs []string
}

func NewComponentStableChecker(opts *CheckOpts) ComponentStableChecker {
	cv := opts.manager.aem.config.Values()

	return ComponentStableChecker{
		Skip:                     cv.GetBool("instance.check.component_stable.skip"),
		IgnoredPIDs:              cv.GetStringSlice("instance.check.component_stable.ignored_pids"),
		FailedActivationPIDs:     cv.GetStringSlice("instance.check.component_stable.failed_activation_pids"),
		UnsatisfiedReferencePIDs: cv.GetStringSlice("instance.check.component_stable.unsatisfied_reference_pids"),
	}
}

func (c ComponentStableChecker) Spec() CheckSpec {
	return CheckSpec{Skip: c.Skip, Mandatory: true}
}

func (c ComponentStableChecker) Check(_ CheckContext, instance Instance) CheckResult {
	components, err := instance.osgi.componentManager.List()
	if err != nil {
		return CheckResult{
			ok:      false,
			message: "components unknown",
			err:     err,
		}
	}

	failedComponents := lo.Filter(components.List, func(component osgi.ComponentListItem, _ int) bool {
		return !stringsx.MatchSome(component.PID, c.IgnoredPIDs) && stringsx.MatchSome(component.PID, c.FailedActivationPIDs) && component.State == osgi.ComponentStateFailedActivation
	})
	failedComponentCount := len(failedComponents)
	if failedComponentCount > 0 {
		message := fmt.Sprintf("some components failed activation (%d): '%s'", failedComponentCount, failedComponents[0].PID)
		return CheckResult{
			ok:      false,
			message: message,
		}
	}

	unsatisfiedComponents := lo.Filter(components.List, func(component osgi.ComponentListItem, _ int) bool {
		return !stringsx.MatchSome(component.PID, c.IgnoredPIDs) && stringsx.MatchSome(component.PID, c.UnsatisfiedReferencePIDs) && component.State == osgi.ComponentStateUnsatisfiedReference
	})
	unsatisfiedComponentCount := len(unsatisfiedComponents)
	if unsatisfiedComponentCount > 0 {
		message := fmt.Sprintf("some components unsatisfied (%d): '%s'", unsatisfiedComponentCount, unsatisfiedComponents[0].PID)
		return CheckResult{
			ok:      false,
			message: message,
		}
	}

	return CheckResult{
		ok:      true,
		message: "all components stable",
	}
}

type InstallerChecker struct {
	Skip  bool
	State bool
	Pause bool
}

func NewInstallerChecker(opts *CheckOpts) InstallerChecker {
	cv := opts.manager.aem.config.Values()

	return InstallerChecker{
		Skip:  cv.GetBool("instance.check.installer.skip"),
		State: cv.GetBool("instance.check.installer.state"),
		Pause: cv.GetBool("instance.check.installer.pause"),
	}
}

func (c InstallerChecker) Spec() CheckSpec {
	return CheckSpec{Skip: c.Skip, Mandatory: false}
}

func (c InstallerChecker) Check(_ CheckContext, instance Instance) CheckResult {
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
		if state.IsActive() {
			return CheckResult{
				ok:      false,
				message: fmt.Sprintf("installer active (%d)", state.ActiveResources()),
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

func NewStatusStoppedChecker() StatusStoppedChecker {
	return StatusStoppedChecker{}
}

type StatusStoppedChecker struct{}

func (c StatusStoppedChecker) Spec() CheckSpec {
	return CheckSpec{Skip: false, Mandatory: true}
}

func (c StatusStoppedChecker) Check(_ CheckContext, instance Instance) CheckResult {
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

func NewReachableChecker(opts *CheckOpts, reachable bool) ReachableHTTPChecker {
	cv := opts.manager.aem.config.Values()

	return ReachableHTTPChecker{
		Skip:      cv.GetBool("instance.check.reachable.skip"),
		Mandatory: reachable,
		Reachable: reachable,
		Timeout:   cv.GetDuration("instance.check.reachable.timeout"),
	}
}

type ReachableHTTPChecker struct {
	Skip      bool
	Mandatory bool
	Reachable bool
	Timeout   time.Duration
}

func (c ReachableHTTPChecker) Spec() CheckSpec {
	return CheckSpec{Skip: c.Skip, Mandatory: c.Mandatory}
}

func (c ReachableHTTPChecker) Check(_ CheckContext, instance Instance) CheckResult {
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
		message: fmt.Sprintf("not reachable (%s)", address),
	}
}

func NewLoginPageChecker(opts *CheckOpts) PathHTTPChecker {
	cv := opts.manager.aem.config.Values()
	return NewPathReadyChecker(opts, cv.GetBool("instance.check.login_page.skip"),
		"login page",
		cv.GetString("instance.check.login_page.path"),
		cv.GetInt("instance.check.login_page.status_code"),
		cv.GetString("instance.check.login_page.contained_text"),
	)
}

func NewPathReadyChecker(opts *CheckOpts, skip bool, name string, path string, statusCode int, containedText string) PathHTTPChecker {
	cv := opts.manager.aem.config.Values()

	return PathHTTPChecker{
		Skip:           skip,
		Name:           name,
		Path:           path,
		RequestTimeout: cv.GetDuration("instance.check.path_ready.timeout"),
		ResponseCode:   statusCode,
		ResponseText:   containedText,
	}
}

func (c PathHTTPChecker) Spec() CheckSpec {
	return CheckSpec{Skip: c.Skip, Mandatory: false}
}

type PathHTTPChecker struct {
	Skip           bool
	Name           string
	Path           string
	RequestTimeout time.Duration
	ResponseCode   int
	ResponseText   string
}

func (c PathHTTPChecker) Check(_ CheckContext, instance Instance) CheckResult {
	response, err := instance.http.RequestWithTimeout(c.RequestTimeout).Get(c.Path)
	if err != nil {
		return CheckResult{
			ok:      false,
			message: fmt.Sprintf("%s request error", c.Name),
			err:     err,
		}
	}
	if c.ResponseCode > 0 && response.StatusCode() != c.ResponseCode {
		return CheckResult{
			ok:      false,
			message: fmt.Sprintf("%s responds with unexpected code (%d)", c.Name, response.StatusCode()),
		}
	}
	if c.ResponseText != "" {
		textBytes, err := io.ReadAll(response.RawBody())
		if err != nil {
			return CheckResult{
				ok:      false,
				message: fmt.Sprintf("%s response read error", c.Name),
				err:     err,
			}
		}
		text := string(textBytes)
		if !strings.Contains(text, c.ResponseText) {
			return CheckResult{
				ok:      false,
				message: fmt.Sprintf("%s responds without text: %s", c.Name, c.ResponseText),
			}
		}
	}
	return CheckResult{
		ok:      true,
		message: fmt.Sprintf("%s ready", c.Name),
	}
}

func bundleStablePercent(bundles *osgi.BundleList, unstableBundles []osgi.BundleListItem) string {
	return stringsx.PercentExplained(bundles.Total()-len(unstableBundles), bundles.Total(), 0)
}
