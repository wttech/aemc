package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"time"
)

type CheckOpts struct {
	Warmup        time.Duration
	Interval      time.Duration
	DoneThreshold int
	DoneNever     bool
	AwaitStrict   bool

	Reachable           ReachableHTTPChecker
	BundleStable        BundleStableChecker
	EventStable         EventStableChecker
	Installer           InstallerChecker
	AwaitStartedTimeout TimeoutChecker
	Unreachable         ReachableHTTPChecker
	StatusStopped       StatusStoppedChecker
	AwaitStoppedTimeout TimeoutChecker
}

func (im *InstanceManager) NewCheckOpts() *CheckOpts {
	return &CheckOpts{
		Warmup:        time.Second * 1,
		Interval:      time.Second * 5,
		DoneThreshold: 3,

		Reachable:           NewReachableChecker(true),
		BundleStable:        NewBundleStableChecker(),
		EventStable:         NewEventStableChecker(),
		AwaitStartedTimeout: NewTimeoutChecker("started", time.Minute*30),
		Installer:           NewInstallerChecker(),
		StatusStopped:       NewStatusStoppedChecker(),
		AwaitStoppedTimeout: NewTimeoutChecker("stopped", time.Minute*10),
		Unreachable:         NewReachableChecker(false),
	}
}

func (im *InstanceManager) CheckUntilDone(instances []Instance, opts *CheckOpts, checks []Checker) error {
	if len(instances) == 0 {
		log.Debugf("no instances to check")
		return nil
	}
	time.Sleep(opts.Warmup)
	doneTimes := 0
	for {
		done, err := im.CheckIfDone(instances, checks)
		if err != nil {
			return err
		}
		if done {
			if !opts.DoneNever {
				doneTimes++
				if doneTimes <= opts.DoneThreshold {
					log.Infof(InstanceMsg(instances, fmt.Sprintf("checked (%d/%d)", doneTimes, opts.DoneThreshold)))
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
	instanceResults, err := im.Check(instances, checks)
	if err != nil {
		return false, nil
	}
	ok := lo.EveryBy(instanceResults, func(results []CheckResult) bool {
		return lo.EveryBy(results, func(result CheckResult) bool { return result.ok })
	})
	return ok, nil
}

func (im *InstanceManager) Check(instances []Instance, checks []Checker) ([][]CheckResult, error) {
	return InstanceProcess(im.aem, instances, func(i Instance) ([]CheckResult, error) { return im.CheckOne(i, checks) })
}

func (im *InstanceManager) CheckOne(i Instance, checks []Checker) ([]CheckResult, error) {
	var results []CheckResult
	for _, check := range checks {
		result := check.Check(i)
		results = append(results, result)
		if result.abort {
			log.Fatalf("instance '%s': %s", i.ID(), result.message)
		}
		if result.err != nil {
			log.Infof("instance '%s': %s", i.ID(), result.err)
		} else if len(result.message) > 0 {
			log.Infof("instance '%s': %s", i.ID(), result.message)
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
	if len(instances) == 0 {
		return nil
	}
	log.Infof(InstanceMsg(instances, "awaiting up"))
	return im.CheckUntilDone(instances, im.CheckOpts, []Checker{
		im.CheckOpts.AwaitStartedTimeout,
		im.CheckOpts.Reachable,
		im.CheckOpts.BundleStable,
		im.CheckOpts.EventStable,
		im.CheckOpts.Installer,
	})
}

func (im *InstanceManager) AwaitStoppedOne(instance Instance) error {
	return im.AwaitStopped([]Instance{instance})
}

func (im *InstanceManager) AwaitStoppedAll() error {
	return im.AwaitStopped(im.Locals())
}

func (im *InstanceManager) AwaitStopped(instances []Instance) error {
	if len(instances) == 0 {
		return nil
	}
	log.Infof(InstanceMsg(instances, "awaiting down"))
	return im.CheckUntilDone(instances, im.CheckOpts, []Checker{
		im.CheckOpts.AwaitStoppedTimeout,
		im.CheckOpts.StatusStopped,
		im.CheckOpts.Unreachable,
	})
}
