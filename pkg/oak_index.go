package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/oak"
)

type OAKIndex struct {
	manager *OAKIndexManager

	name string
}

func (b OAKIndex) Name() string {
	return b.name
}

type OAKIndexState struct {
	data *oak.IndexListItem

	Name    string         `yaml:"name" json:"name"`
	Exists  bool           `yaml:"exists" json:"exists"`
	Details map[string]any `yaml:"details" json:"details"`
}

func (b OAKIndex) State() (*OAKIndexState, error) {
	data, err := b.manager.Find(b.name)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return &OAKIndexState{
			Name:   b.name,
			Exists: false,
		}, nil
	}

	return &OAKIndexState{
		data: data,

		Name:   b.name,
		Exists: true,
		Details: map[string]any{
			"reindex": data.Reindex,
		},
	}, nil
}

func (b OAKIndex) String() string {
	return fmt.Sprintf("index '%s'", b.name)
}

func (b OAKIndex) MarshalJSON() ([]byte, error) {
	state, err := b.State()
	if err != nil {
		return nil, err
	}
	return json.Marshal(state)
}

func (b OAKIndex) MarshalYAML() (interface{}, error) {
	return b.State()
}

func (b OAKIndex) MarshalText() string {
	state, err := b.State()
	if err != nil {
		return fmt.Sprintf("name '%s' state cannot be read\n", b.name)
	}
	sb := bytes.NewBufferString("")
	if state.Exists {
		sb.WriteString(fmt.Sprintf("name '%s'\n", b.name))
		sb.WriteString(fmtx.TblProps(state.Details))
	} else {
		sb.WriteString(fmt.Sprintf("name '%s' cannot be found\n", b.name))
	}
	return sb.String()
}
