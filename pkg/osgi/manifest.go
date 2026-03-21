package osgi

import (
	"fmt"
	"github.com/essentialkaos/go-jar"
)

func ReadBundleManifest(localPath string) (*BundleManifest, error) {
	manifest, err := jar.ReadFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read OSGi bundle manifest from file '%s': %w", localPath, err)
	}
	return &BundleManifest{SymbolicName: manifest[AttributeSymbolicName], Version: manifest[AttributeVersion]}, nil
}

type BundleManifest struct {
	SymbolicName string `yaml:"symbolic_name" json:"symbolicName"`
	Version      string `yaml:"version" json:"version"`
}

func (m BundleManifest) MarshalText() string {
	return fmt.Sprintf("symbolic name '%s'\nversion '%s'\n", m.SymbolicName, m.Version)
}

const (
	AttributeSymbolicName = "Bundle-SymbolicName"
	AttributeVersion      = "Bundle-Version"
)
