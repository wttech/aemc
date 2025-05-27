package keystore

import (
	"bytes"
	"encoding/json"
	"github.com/wttech/aemc/pkg/common/fmtx"
)

type CertificateChain struct {
	Subject      string      `json:"subject"`
	Issuer       string      `json:"issuer"`
	NotBefore    string      `json:"notBefore"`
	NotAfter     string      `json:"notAfter"`
	SerialNumber json.Number `json:"serialNumber"`
}

type PrivateKey struct {
	Alias     string             `json:"alias"`
	EntryType string             `json:"entryType"`
	Algorithm string             `json:"algorithm"`
	Format    string             `json:"format"`
	Chain     []CertificateChain `json:"chain"`
}

func (c *CertificateChain) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblMap("details", "name", "value", map[string]any{
		"subject":      c.Subject,
		"issuer":       c.Issuer,
		"notBefore":    c.NotBefore,
		"notAfter":     c.NotAfter,
		"serialNumber": c.SerialNumber,
	}))
	return bs.String()
}

func (c *PrivateKey) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblMap("details", "name", "value", map[string]any{
		"alias":     c.Alias,
		"entryType": c.EntryType,
		"algorithm": c.Algorithm,
		"format":    c.Format,
		"chain":     c.Chain,
	}))
	return bs.String()
}
