package project

import (
	"fmt"
	"github.com/samber/lo"
)

type Kind string

const (
	KindAuto       = "auto"
	KindInstance   = "instance"
	KindAppClassic = "app_classic"
	KindAppCloud   = "app_cloud"
	KindUnknown    = "unknown"

	KindPropName          = "aemVersion"
	KindPropCloudValue    = "cloud"
	KindPropClassicPrefix = "6."

	GitIgnoreFile   = ".gitignore"
	PropFile        = "archetype.properties"
	PackagePropName = "package"
)

func Kinds() []Kind {
	return []Kind{KindInstance, KindAppCloud, KindAppClassic}
}

func KindStrings() []string {
	return lo.Map(Kinds(), func(k Kind, _ int) string { return string(k) })
}

func KindOf(name string) (Kind, error) {
	if name == KindAuto {
		return KindAuto, nil
	} else if name == KindInstance {
		return KindInstance, nil
	} else if name == KindAppCloud {
		return KindAppCloud, nil
	} else if name == KindAppClassic {
		return KindAppClassic, nil
	}
	return "", fmt.Errorf("project kind '%s' is not supported", name)
}
