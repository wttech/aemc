package gts

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"strings"
	"time"
)

const (
	AemTimeLayout = "Mon Jan 2 15:04:05 MST 2006"
)

type Certificate struct {
	Alias        string      `json:"alias"`
	Subject      string      `json:"subject"`
	Issuer       string      `json:"issuer"`
	NotBefore    string      `json:"notBefore"`
	NotAfter     string      `json:"notAfter"`
	SerialNumber json.Number `json:"serialNumber"`
}

func (c *Certificate) NotBeforeDate() (time.Time, error) {
	return time.Parse(AemTimeLayout, c.NotBefore)
}

func (c *Certificate) NotAfterDate() (time.Time, error) {
	return time.Parse(AemTimeLayout, c.NotAfter)
}

func standardizeSpaces(a string) string {
	return strings.Join(strings.Fields(a), "")
}

func (c *Certificate) IsEqual(certifiacte x509.Certificate) (bool, error) {
	notBefore, err := c.NotBeforeDate()

	if err != nil {
		return false, err
	}

	notAfter, err := c.NotAfterDate()

	if err != nil {
		return false, err
	}

	return standardizeSpaces(c.Issuer) == standardizeSpaces(certifiacte.Issuer.String()) &&
		standardizeSpaces(c.Subject) == standardizeSpaces(certifiacte.Subject.String()) &&
		notBefore.Equal(certifiacte.NotBefore) &&
		notAfter.Equal(certifiacte.NotAfter) &&
		c.SerialNumber.String() == certifiacte.SerialNumber.String(), nil
}
func (c *Certificate) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblMap("details", "name", "value", map[string]any{
		"alias":        c.Alias,
		"subject":      c.Subject,
		"issuer":       c.Issuer,
		"notBefore":    c.NotBefore,
		"notAfter":     c.NotAfter,
		"serialNumber": c.SerialNumber,
	}))
	return bs.String()
}
