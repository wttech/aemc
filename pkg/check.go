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
	Check(instance Instance) CheckResult
	Spec() CheckSpec
}

type CheckSpec struct {
	Mandatory bool // indicates if next checks should be skipped if that particular one fails
}

func NewAwaitChecker(opts *CheckOpts, expectedState string) AwaitChecker {
	cv := opts.manager.aem.config.Values()

	return AwaitChecker{
		ExpectedState: expectedState,
		Duration:      cv.GetDuration(fmt.Sprintf("instance.check.await_%s.timeout", expectedState)),
		Started:       time.Now(),
	}
}

func (c AwaitChecker) Check(_ Instance) CheckResult {
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

func (c AwaitChecker) Spec() CheckSpec {
	return CheckSpec{Mandatory: true}
}

type AwaitChecker struct {
	Started       time.Time
	Duration      time.Duration
	ExpectedState string
}

func NewEventStableChecker(opts *CheckOpts) EventStableChecker {
	cv := opts.manager.aem.config.Values()

	return EventStableChecker{
		ReceivedMaxAge: cv.GetDuration("instance.check.event_stable.received_max_age"),
		TopicsUnstable: cv.GetStringSlice("instance.check.event_stable.topics_unstable"),
		DetailsIgnored: cv.GetStringSlice("instance.check.event_stable.details_ignored"),
	}
}

func NewBundleStableChecker(opts *CheckOpts) BundleStableChecker {
	cv := opts.manager.aem.config.Values()

	return BundleStableChecker{
		SymbolicNamesIgnored: cv.GetStringSlice("instance.check.bundle_stable.symbolic_names_ignored"),
	}
}

func (c BundleStableChecker) Spec() CheckSpec {
	return CheckSpec{Mandatory: true}
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
	ReceivedMaxAge time.Duration
	TopicsUnstable []string
	DetailsIgnored []string
}

func (c EventStableChecker) Spec() CheckSpec {
	return CheckSpec{Mandatory: true}
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

func NewInstallerChecker(opts *CheckOpts) InstallerChecker {
	cv := opts.manager.aem.config.Values()

	return InstallerChecker{
		State: cv.GetBool("instance.check.installer.state"),
		Pause: cv.GetBool("instance.check.installer.pause"),
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

func NewStatusStoppedChecker() StatusStoppedChecker {
	return StatusStoppedChecker{}
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

func NewReachableChecker(opts *CheckOpts, reachable bool) ReachableHTTPChecker {
	cv := opts.manager.aem.config.Values()

	return ReachableHTTPChecker{
		Mandatory: reachable,
		Reachable: reachable,
		Timeout:   cv.GetDuration("instance.check.reachable.timeout"),
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
		message: fmt.Sprintf("not reachable (%s)", address),
	}
}

func NewPathReadyChecker(opts *CheckOpts, name string, path string, statusCode int, containedText string) PathHTTPChecker {
	cv := opts.manager.aem.config.Values()

	return PathHTTPChecker{
		Name:           name,
		Path:           path,
		RequestTimeout: cv.GetDuration("instance.check.path_ready.timeout"),
		ResponseCode:   statusCode,
		ResponseText:   containedText,
	}
}

func (c PathHTTPChecker) Spec() CheckSpec {
	return CheckSpec{Mandatory: false}
}

type PathHTTPChecker struct {
	Name           string
	Path           string
	RequestTimeout time.Duration
	ResponseCode   int
	ResponseText   string
}

func (c PathHTTPChecker) Check(instance Instance) CheckResult {
	client := NewResty(instance.HTTP().BaseURL())
	response, err := client.SetTimeout(c.RequestTimeout).NewRequest().Get(c.Path)
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
