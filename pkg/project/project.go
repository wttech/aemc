package project

import (
	"embed"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"io/fs"
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
	return []Kind{KindAuto, KindCloud, KindClassic}
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
		return "", fmt.Errorf("project kind '%s' is not supported", name)
	}
}

//go:embed classic
var classicFiles embed.FS

//go:embed cloud
var cloudFiles embed.FS

func (p Project) InitializeWithChanged(kind Kind) (bool, error) {
	projectChanged, err := p.initializeWithChanged(kind)
	if err != nil {
		return false, err
	}
	configChanged, err := p.config.InitializeWithChanged()
	if err != nil {
		return false, err
	}
	return projectChanged || configChanged, nil
}

func (p Project) initializeWithChanged(kind Kind) (bool, error) {
	if p.config.TemplateFileExists() {
		return false, nil
	}
	if err := p.initialize(kind); err != nil {
		return false, err
	}
	return true, nil
}

func (p Project) initialize(kind Kind) error {
	log.Infof("preparing default files for project of kind '%s'", kind)
	switch kind {
	case KindClassic:
		if err := copyEmbedFiles(&classicFiles, string(kind)+"/"); err != nil {
			return err
		}
	case KindCloud:
		if err := copyEmbedFiles(&cloudFiles, string(kind)+"/"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("project kind '%s' cannot be initialized", kind)
	}
	log.Infof("creating conventional directories")
	if err := pathx.Ensure(common.LibDir); err != nil {
		return err
	}
	if err := pathx.Ensure(common.TmpDir); err != nil {
		return err
	}
	return nil
}

func copyEmbedFiles(efs *embed.FS, dirPrefix string) error {
	return fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		bytes, err := efs.ReadFile(path)
		if err != nil {
			return err
		}
		if err := filex.Write(strings.TrimPrefix(path, dirPrefix), bytes); err != nil {
			return err
		}
		return nil
	})
}

func (p Project) KindDetermine(name string) (Kind, error) {
	var kind Kind = KindAuto
	if name != "" {
		kindCandidate, err := KindOf(name)
		if err != nil {
			return "", err
		}
		kind = kindCandidate
	}
	if kind == KindAuto {
		kindCandidate, err := p.KindInfer()
		if err != nil {
			return "", err
		}
		kind = kindCandidate
	}
	return kind, nil
}

const (
	KindPropFile          = "archetype.properties"
	KindPropName          = "aemVersion"
	KindPropCloudValue    = "cloud"
	KindPropClassicPrefix = "6."
)

func (p Project) KindInfer() (Kind, error) {
	if pathx.Exists(KindPropFile) {
		log.Infof("inferring project kind basing on file '%s' and property '%s'", KindPropFile, KindPropName)
		propLoader := properties.Loader{
			Encoding:         properties.ISO_8859_1,
			DisableExpansion: true,
		}
		props, err := propLoader.LoadFile(KindPropFile)
		if err != nil {
			return "", fmt.Errorf("cannot infer project kind: %w", err)
		}
		propValue := props.GetString(KindPropName, "")

		var kind Kind
		if propValue == KindPropCloudValue {
			kind = KindCloud
		} else if strings.HasPrefix(propValue, KindPropClassicPrefix) {
			kind = KindClassic
		} else {
			return "", fmt.Errorf("cannot infer project kind as value '%s' of property '%s' in file '%s' is not recognized", propValue, KindPropName, KindPropFile)
		}
		log.Infof("inferred project kind basing on file '%s' and property '%s' is '%s'", KindPropFile, KindPropName, kind)
		return kind, nil
	}
	return "", fmt.Errorf("cannot infer project kind as file '%s' does not exist", KindPropFile)
}

func (p Project) GettingStarted() (string, error) {
	text := fmt.Sprintf(strings.Join([]string{
		"The next step is providing AEM files (JAR or SDK ZIP, license, service packs) to directory '" + common.LibDir + "'.",
		"Alternatively, instruct the tool where these files are located by adjusting properties: 'dist_file', 'license_file' in configuration file '" + cfg.FileDefault + "'.",
		"Make sure to exclude the directory '" + common.HomeDir + "'from VCS versioning and IDE indexing.",
		"Finally, use tasks to manage AEM instances:",
		"",

		"sh taskw --list",

		"",
		"It is also possible to run individual AEM Compose CLI commands separately.",
		"Discover available commands by running:",
		"",

		"sh aemw --help",
	}, "\n"))
	return text, nil
}
