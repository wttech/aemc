package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/oak"
	"time"
)

type OAKIndex struct {
	manager *OAKIndexManager

	name string
}

func (i OAKIndex) Name() string {
	return i.name
}

type OAKIndexState struct {
	data *oak.IndexListItem

	Name    string         `yaml:"name" json:"name"`
	Exists  bool           `yaml:"exists" json:"exists"`
	Details map[string]any `yaml:"details" json:"details"`
}

func (i OAKIndex) assumeExists() (*OAKIndexState, error) {
	state, err := i.State()
	if err != nil {
		return state, err
	}
	if !state.Exists {
		return state, fmt.Errorf("%s > index '%s' does not exist", i.manager.instance.IDColor(), i.name)
	}
	return state, nil
}

func (i OAKIndex) ReindexWithChanged() (bool, error) {
	state, err := i.assumeExists()
	if err != nil {
		return false, err
	}
	if state.data.Reindex {
		return false, nil
	}
	return true, i.manager.Reindex(state.data.Name)
}

func (i OAKIndex) Reindex() error {
	state, err := i.assumeExists()
	if err != nil {
		return err
	}
	err = i.manager.Reindex(state.data.Name)
	if err != nil {
		return fmt.Errorf("%s > cannot reindex index '%s': %w", i.manager.instance.IDColor(), i.name, err)
	}
	return nil
}

func (i OAKIndex) State() (*OAKIndexState, error) {
	data, err := i.manager.Find(i.name)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return &OAKIndexState{
			Name:   i.name,
			Exists: false,
		}, nil
	}

	return &OAKIndexState{
		data: data,

		Name:   i.name,
		Exists: true,
		Details: map[string]any{
			"type":                     data.Type,
			"async":                    data.Async,
			"includePaths":             data.IncludedPaths,
			"queryPaths":               data.QueryPaths,
			"reindex":                  data.Reindex,
			"reindexCount":             data.ReindexCount,
			"evaluatePathRestrictions": data.EvaluatePathRestrictions,
			"declaringNodeTypes":       data.DeclaringNodeTypes,
			"propertyNames":            data.PropertyNames,
			"tags":                     data.Tags,
		},
	}, nil
}

func (i OAKIndex) String() string {
	return fmt.Sprintf("index '%s'", i.name)
}

func (i OAKIndex) MarshalJSON() ([]byte, error) {
	state, err := i.State()
	if err != nil {
		return nil, err
	}
	return json.Marshal(state)
}

func (i OAKIndex) MarshalYAML() (interface{}, error) {
	return i.State()
}

func (i OAKIndex) MarshalText() string {
	state, err := i.State()
	if err != nil {
		return fmt.Sprintf("name '%s' state cannot be read\n", i.name)
	}
	sb := bytes.NewBufferString("")
	if state.Exists {
		sb.WriteString(fmt.Sprintf("name '%s'\n", i.name))
		sb.WriteString(fmtx.TblProps(state.Details))
	} else {
		sb.WriteString(fmt.Sprintf("name '%s' cannot be found\n", i.name))
	}
	return sb.String()
}

func (i OAKIndex) AwaitNotReindexed() error {
	return i.Await("not reindexed", func() bool {
		state, err := i.State()
		if err != nil {
			log.Warn(err)
			return false
		}
		return state.Exists && !state.data.Reindex
	}, time.Minute*1)
}

func (i OAKIndex) Await(state string, condition func() bool, timeout time.Duration) error {
	started := time.Now()
	for {
		if condition() {
			break
		}
		if time.Now().After(started.Add(timeout)) {
			return fmt.Errorf("%s > awaiting index '%s' state '%s' reached timeout after %s", i.manager.instance.IDColor(), i.name, state, timeout)
		}
		log.Infof("%s > awaiting index '%s' state '%s'", i.manager.instance.IDColor(), i.name, state)
		time.Sleep(time.Second * 5)
	}
	return nil
}
