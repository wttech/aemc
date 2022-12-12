package osgi

import (
	"fmt"
	"github.com/essentialkaos/go-jar"
)

func ReadBundleManifest(localPath string) (*BundleManifest, error) {
	manifest, err := jar.ReadFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read OSGi bundle manifest from file '%s'", localPath)
	}
	return &BundleManifest{SymbolicName: manifest[AttributeSymbolicName], Version: manifest[AttributeVersion]}, nil
}

type BundleManifest struct {
	SymbolicName string
	Version      string
}

const (
	AttributeSymbolicName = "Bundle-SymbolicName"
	AttributeVersion      = "Bundle-Version"
)
