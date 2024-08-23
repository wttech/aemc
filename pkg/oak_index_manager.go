package pkg

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/oak"
)

type OAKIndexManager struct {
	instance *Instance
}

func NewOAKIndexManager(instance *Instance) *OAKIndexManager {
	return &OAKIndexManager{
		instance: instance,
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

func (im *OAKIndexManager) ReindexAll() (*oak.IndexList, error) {
	indexes, err := im.List()
	if err != nil {
		return nil, err
	}

	for _, i := range indexes.List {
		if i.Reindex {
			log.Warnf("%s > index '%s' is currently being reindexed, skipping", im.instance.IDColor(), i.Name)
			continue
		}
		index := im.New(i.Name)
		if err = im.Reindex(i.Name); err != nil {
			return nil, err
		}
		if err = index.AwaitNotReindexed(); err != nil {
			return nil, err
		}
	}

	return indexes, nil
}
