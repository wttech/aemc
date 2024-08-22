package pkg

import (
	"fmt"
	"github.com/samber/lo"
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
	return &res, nil
}
