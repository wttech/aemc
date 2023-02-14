package project

import (
	"embed"
	"fmt"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/filex"
	"io/fs"
	"os"
	"strings"
)

type Project struct {
	config *cfg.Config
}

func New(config *cfg.Config) *Project {
	return &Project{config: config}
}

type Kind string

const (
	KindAuto    = "auto"
	KindClassic = "classic"
	KindCloud   = "cloud"
)

func Kinds() []Kind {
	return []Kind{KindCloud, KindClassic}
}

func KindStrings() []string {
	return lo.Map(Kinds(), func(k Kind, _ int) string { return string(k) })
}

func KindOf(name string) (Kind, error) {
	if name == KindAuto {
		return KindAuto, nil
	} else if name == KindCloud {
		return KindCloud, nil
	} else if name == KindClassic {
		return KindClassic, nil
	} else {
		return "", fmt.Errorf("unsupport project kind '%s'", name)
	}
}

//go:embed classic
var classicFiles embed.FS

//go:embed cloud
var cloudFiles embed.FS

func (p Project) IsInitialized() bool {
	return p.config.IsInitialized()
}

func (p Project) Initialize(kind Kind) error {
	switch kind {
	case KindClassic:
		return copyEmbedFiles(&classicFiles)
	case KindCloud:
		return copyEmbedFiles(&cloudFiles)
	default:
		return fmt.Errorf("unsupport project type '%s'", kind)
	}
	return p.config.Initialize()
}

func copyEmbedFiles(efs *embed.FS) error {
	return fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		bytes, err := efs.ReadFile(path)
		if err != nil {
			return err
		}
		if err := filex.Write(path, bytes); err != nil {
			return err
		}
		return nil
	})
}

func (p Project) DetermineKind(name string) (Kind, error) {
	var kind Kind
	if name != "" {
		kindCandidate, err := KindOf(name)
		if err != nil {
			return "", err
		}
		kind = kindCandidate
	} else {
		kind = p.InferKind()
	}
	if kind == KindAuto {
		kind = p.InferKind()
	}
	return kind, nil
}

func (p Project) InferKind() Kind {
	return KindClassic // TODO read archetype.properties ; check aemVersion=cloud
}

func (p Project) FindScriptNames() ([]string, error) {
	scriptFiles, err := os.ReadDir(common.ScriptDir)
	if err != nil {
		return nil, fmt.Errorf("cannot list scripts in dir '%s'", common.ScriptDir)
	}
	var scripts []string
	for _, file := range scriptFiles {
		if strings.HasSuffix(file.Name(), ".sh") {
			scripts = append(scripts, strings.TrimSuffix(file.Name(), ".sh"))
		}
	}
	return scripts, nil
}

func (p Project) InitializeWithChanged(kind Kind) (bool, error) {
	if p.IsInitialized() {
		return false, nil
	}
	if err := p.Initialize(kind); err != nil {
		return false, err
	}
	return true, nil
}
