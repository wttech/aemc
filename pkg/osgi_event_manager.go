package pkg

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/osgi"
)

const (
	EventsPath     = "/system/console/events"
	EventsPathJson = EventsPath + ".json"
)

type OSGiEventManager struct {
	instance *Instance
}

func (em *OSGiEventManager) List() (*osgi.EventList, error) {
	resp, err := em.instance.http.Request().Get(EventsPathJson)
	if err != nil {
		return nil, fmt.Errorf("%s > cannot request event list: %w", em.instance.ID(), err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("%s > cannot request event list: %s", em.instance.ID(), resp.Status())
	}
	var res = new(osgi.EventList)
	if err = fmtx.UnmarshalJSON(resp.RawBody(), res); err != nil {
		return nil, fmt.Errorf("%s > cannot parse event list response: %w", em.instance.ID(), err)
	}
	return res, nil
}
