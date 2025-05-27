package keystore

import (
	"bytes"
	"io"

	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/fmtx"
)

type Status struct {
	Created     bool         `json:"exists"`
	PrivateKeys []PrivateKey `json:"aliases"`
}

func UnmarshalStatus(readCloser io.ReadCloser) (*Status, error) {
	// if key store exist, it doesn't contain property exists
	var status = Status{Created: true, PrivateKeys: []PrivateKey{}}

	if err := fmtx.UnmarshalJSON(readCloser, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (s *Status) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblMap("details", "name", "value", map[string]any{
		"created": s.Created,
	}))
	bs.WriteString("\n")
	bs.WriteString(fmtx.TblRows("private keys", true, []string{"alias"}, lo.Map(s.PrivateKeys, func(c PrivateKey, _ int) map[string]any {
		return map[string]any{"alias": c.Alias}
	})))
	return bs.String()
}

func (s *Status) HasAlias(privateKeyAlias string) bool {
	return lo.ContainsBy(s.PrivateKeys, func(c PrivateKey) bool {
		return c.Alias == privateKeyAlias
	})
}
