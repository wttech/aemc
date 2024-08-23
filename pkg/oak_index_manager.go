package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"github.com/wttech/aemc/pkg/oak"
	"sort"
	"strings"
	"time"
)

type OAKIndexManager struct {
	instance *Instance

	awaitNotReindexedTimeout time.Duration
}

func NewOAKIndexManager(instance *Instance) *OAKIndexManager {
	cv := instance.manager.aem.config.Values()

	return &OAKIndexManager{
		instance: instance,

		awaitNotReindexedTimeout: cv.GetDuration("instance.oak.index.await_not_reindexed_timeout"),
	}
}

func (im *OAKIndexManager) New(name string) OAKIndex {
	return OAKIndex{
		manager: im,
		name:    name,
	}
}

func (im *OAKIndexManager) Find(name string) (*oak.IndexListItem, error) {
	indexes, err := im.List()
	if err != nil {
		return nil, fmt.Errorf("%s > cannot find index '%s'", im.instance.IDColor(), name)
	}
	item, found := lo.Find(indexes.List, func(i oak.IndexListItem) bool { return name == i.Name })
	if found {
		return &item, nil
	}
	return nil, nil
}

func (im *OAKIndexManager) List() (*oak.IndexList, error) {
	resp, err := im.instance.http.Request().Get(oak.IndexListJson)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot request index list: %w", im.instance.IDColor(), err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot request index list: %s", im.instance.IDColor(), resp.Status())
	}
	var res oak.IndexList
	if err = fmtx.UnmarshalJSON(resp.RawBody(), &res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse index list: %w", im.instance.IDColor(), err)
	}
	res.List = lo.Filter(res.List, func(i oak.IndexListItem, _ int) bool { return i.PrimaryType == oak.IndexPrimaryType })
	return &res, nil
}

func (im *OAKIndexManager) Reindex(name string) error {
	node := im.instance.Repo().Node(fmt.Sprintf("/oak:index/%s", name))
	if err := node.SaveProp("reindex", true); err != nil {
		return fmt.Errorf("%s > cannot reindex '%s': %w", im.instance.IDColor(), name, err)
	}
	return nil
}

func (im *OAKIndexManager) ReindexBatchWithChanged(names []string, reason string) ([]OAKIndex, bool, error) {
	lock := osx.NewLock(fmt.Sprintf("%s/oak/reindex-batch/%s.yml", im.instance.LockDir(), reason), func() (oakReindexAllLock, error) {
		namesSorted := append([]string(nil), names...)
		sort.Strings(namesSorted)
		return oakReindexAllLock{Names: strings.Join(namesSorted, ",")}, nil
	})
	lockState, err := lock.State()
	if err != nil {
		return nil, false, err
	}
	if lockState.UpToDate {
		log.Debugf("%s > reindexing '%s' already done (up-to-date)", im.instance.IDColor(), reason)
		return nil, false, nil
	}
	indexes, err := im.ReindexBatch(names)
	if err != nil {
		return nil, false, err
	}
	if err = lock.Lock(); err != nil {
		return nil, false, err
	}
	return indexes, true, nil
}

type oakReindexAllLock struct {
	Names string `yaml:"names"`
}

func (im *OAKIndexManager) ReindexBatch(names []string) ([]OAKIndex, error) {
	indexes, err := im.FindByName(names)
	if err != nil {
		return nil, err
	}

	total := len(names)
	log.Infof("%s > reindexing batch of indexes (%d)", im.instance.IDColor(), total)

	for i, index := range indexes {
		percent := stringsx.PercentExplained(i+1, total, 0)

		state, err := index.State()
		if err != nil {
			return nil, err
		}

		if state.Reindexed() {
			log.Warnf("%s > reindexing '%s' skipped as already in progress (%s)", im.instance.IDColor(), index.Name(), percent)
			continue
		}
		log.Infof("%s > reindexing '%s' (%s)", im.instance.IDColor(), index.Name(), percent)

		if err = index.Reindex(); err != nil {
			return nil, err
		}
		if err = index.AwaitNotReindexed(); err != nil {
			return nil, err
		}
	}
	log.Infof("%s > reindexed batch of indexes (%d)", im.instance.IDColor(), total)

	return indexes, nil
}

func (im *OAKIndexManager) FindByName(names []string) ([]OAKIndex, error) {
	indexes, err := im.List()
	if err != nil {
		return nil, err
	}
	var res []OAKIndex
	for _, name := range names {
		item, found := lo.Find(indexes.List, func(i oak.IndexListItem) bool { return name == i.Name })
		if !found {
			return nil, fmt.Errorf("%s > index '%s' cannot be found", im.instance.IDColor(), name)
		}
		res = append(res, OAKIndex{
			manager: im,
			name:    item.Name,
		})
	}
	return res, nil
}

func (im *OAKIndexManager) Names() ([]string, error) {
	list, err := im.List()
	if err != nil {
		return nil, err
	}
	return lo.Map(list.List, func(i oak.IndexListItem, _ int) string { return i.Name }), nil
}

func (im *OAKIndexManager) ReindexAll() ([]OAKIndex, error) {
	names, err := im.Names()
	if err != nil {
		return nil, err
	}
	return im.ReindexBatch(names)
}

func (im *OAKIndexManager) ReindexAllWithChanged(reason string) ([]OAKIndex, bool, error) {
	names, err := im.Names()
	if err != nil {
		return nil, false, err
	}
	return im.ReindexBatchWithChanged(names, reason)
}
