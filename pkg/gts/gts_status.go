package gts

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"io"
)

type Status struct {
	// if trust store exist, it doesn't contain property exists
	Created      *bool         `json:"exists"`
	Certificates []Certificate `json:"aliases"`
}

func (s *Status) FindCertificate(certificate x509.Certificate) (*Certificate, error) {
	for i := range s.Certificates {
		isEqualResult, err := s.Certificates[i].IsEqual(certificate)

		if err != nil {
			return nil, err
		}

		if isEqualResult {
			return &s.Certificates[i], nil
		}
	}

	return nil, nil
}

func (s *Status) FindCertificateByAlias(alias string) *Certificate {
	for i := range s.Certificates {
		if s.Certificates[i].Alias == alias {
			return &s.Certificates[i]
		}
	}

	return nil
}

func UnmarshalStatus(readCloser io.ReadCloser) (*Status, error) {
	var status Status

	if err := fmtx.UnmarshalJSON(readCloser, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (s *Status) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmt.Sprintf("Created: %t\n", s.Created == nil || *s.Created))
	bs.WriteString(fmtx.TblRows("List of Certificates", true, []string{"alias"}, lo.Map(s.Certificates, func(c Certificate, _ int) map[string]any {
		return map[string]any{"alias": c.Alias}
	})))
	return bs.String()
}
