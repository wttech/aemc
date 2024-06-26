package pkg

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"time"
)

type CheckOpts struct {
	manager *InstanceManager

	Warmup        time.Duration
	Interval      time.Duration
	DoneThreshold int
	DoneNever     bool
	AwaitStrict   bool
	Skip          bool

	Reachable       ReachableHTTPChecker
	BundleStable    BundleStableChecker
	EventStable     EventStableChecker
	ComponentStable ComponentStableChecker
	Installer       InstallerChecker
	AwaitStarted    AwaitChecker
	Unreachable     ReachableHTTPChecker
	StatusStopped   StatusStoppedChecker
	AwaitStopped    AwaitChecker
	LoginPage       PathHTTPChecker
}

func NewCheckOpts(manager *InstanceManager) *CheckOpts {
	cv := manager.aem.config.Values()

	result := &CheckOpts{manager: manager}

	result.Warmup = cv.GetDuration("instance.check.warmup")
	result.Interval = cv.GetDuration("instance.check.interval")
	result.DoneThreshold = cv.GetInt("instance.check.done_threshold")
	result.AwaitStrict = cv.GetBool("instance.local.await_strict")
	result.Skip = cv.GetBool("instance.check.skip")

	result.Reachable = NewReachableChecker(result, true)
	result.BundleStable = NewBundleStableChecker(result)
	result.EventStable = NewEventStableChecker(result)
	result.ComponentStable = NewComponentStableChecker(result)
	result.AwaitStarted = NewAwaitChecker(result, "started")
	result.Installer = NewInstallerChecker(result)
	result.StatusStopped = NewStatusStoppedChecker()
	result.AwaitStopped = NewAwaitChecker(result, "stopped")
	result.Unreachable = NewReachableChecker(result, false)
	result.LoginPage = NewLoginPageChecker(result)

	return result
}

type CheckContext struct {
	Started time.Time
}

type checkContextKey struct{}

func (im *InstanceManager) CheckContext() context.Context {
	return context.WithValue(context.Background(), checkContextKey{}, CheckContext{
		Started: time.Now(),
	})
}

func (im *InstanceManager) CheckUntilDone(instances []Instance, opts *CheckOpts, checks []Checker) error {
	return im.checkUntilDone(im.CheckContext(), instances, opts, checks)
}

func (im *InstanceManager) checkUntilDone(ctx context.Context, instances []Instance, opts *CheckOpts, checks []Checker) error {
	if len(instances) == 0 {
		log.Debug("no instances to check")
		return nil
	}
	time.Sleep(opts.Warmup)
	doneTimes := 0
	for {
		done, err := im.checkIfDone(ctx, instances, checks)
		if err != nil {
			return err
		}
		if done {
			if !opts.DoneNever {
				doneTimes++
				if doneTimes <= opts.DoneThreshold {
					log.Info(InstancesMsg(instances, fmt.Sprintf("checked (%d/%d)", doneTimes, opts.DoneThreshold)))
				}
				if doneTimes == opts.DoneThreshold {
					break
				}
			}
		} else {
			doneTimes = 0
		}
		time.Sleep(opts.Interval)
	}
	return nil
}

func (im *InstanceManager) CheckIfDone(instances []Instance, checks []Checker) (bool, error) {
	return im.checkIfDone(im.CheckContext(), instances, checks)
}

func (im *InstanceManager) checkIfDone(ctx context.Context, instances []Instance, checks []Checker) (bool, error) {
	instanceResults, err := im.check(ctx, instances, checks)
	if err != nil {
		return false, nil
	}
	ok := lo.EveryBy(instanceResults, func(results []CheckResult) bool {
		return lo.EveryBy(results, func(result CheckResult) bool { return result.ok })
	})
	return ok, nil
}

func (im *InstanceManager) Check(instances []Instance, checks []Checker) ([][]CheckResult, error) {
	return im.check(im.CheckContext(), instances, checks)
}

func (im *InstanceManager) check(ctx context.Context, instances []Instance, checks []Checker) ([][]CheckResult, error) {
	return InstanceProcess(im.aem, instances, func(i Instance) ([]CheckResult, error) { return im.checkOne(ctx, i, checks) })
}

func (im *InstanceManager) CheckOne(i Instance, checks []Checker) ([]CheckResult, error) {
	return im.checkOne(im.CheckContext(), i, checks)
}

func (im *InstanceManager) checkOne(ctx context.Context, i Instance, checks []Checker) ([]CheckResult, error) {
	var results []CheckResult
	for _, check := range checks {
		if check.Spec().Skip {
			continue
		}
		result := check.Check(ctx.Value(checkContextKey{}).(CheckContext), i)
		results = append(results, result)
		resultText := result.Text()
		if result.abort {
			log.Fatal(InstanceMsg(i, resultText))
		}
		if resultText != "" {
			if result.ok {
				log.Info(InstanceMsg(i, resultText))
			} else {
				log.Warn(InstanceMsg(i, resultText))
			}
		}
		if !result.ok && check.Spec().Mandatory {
			break
		}
	}
	return results, nil
}

func (im *InstanceManager) AwaitStartedOne(instance Instance) error {
	return im.AwaitStarted([]Instance{instance})
}

func (im *InstanceManager) AwaitStartedAll() error {
	return im.AwaitStarted(im.All())
}

func (im *InstanceManager) AwaitStarted(instances []Instance) error {
	if len(instances) == 0 || im.CheckOpts.Skip {
		return nil
	}
	log.Info(InstancesMsg(instances, "awaiting started"))
	var checkers []Checker
	if im.LocalOpts.ServiceMode {
		checkers = []Checker{
			im.CheckOpts.AwaitStarted,
			im.CheckOpts.Reachable,
			im.CheckOpts.LoginPage,
		}
	} else {
		checkers = []Checker{
			im.CheckOpts.AwaitStarted,
			im.CheckOpts.Reachable,
			im.CheckOpts.BundleStable,
			im.CheckOpts.EventStable,
			im.CheckOpts.Installer,
			im.CheckOpts.LoginPage,
			im.CheckOpts.ComponentStable,
		}
	}
	return im.CheckUntilDone(instances, im.CheckOpts, checkers)
}

func (im *InstanceManager) AwaitStoppedOne(instance Instance) error {
	return im.AwaitStopped([]Instance{instance})
}

func (im *InstanceManager) AwaitStoppedAll() error {
	return im.AwaitStopped(im.Locals())
}

func (im *InstanceManager) AwaitStopped(instances []Instance) error {
	if len(instances) == 0 || im.CheckOpts.Skip {
		return nil
	}
	log.Info(InstancesMsg(instances, "awaiting stopped"))
	return im.CheckUntilDone(instances, im.CheckOpts, []Checker{
		im.CheckOpts.AwaitStopped,
		im.CheckOpts.StatusStopped,
		im.CheckOpts.Unreachable,
	})
}
