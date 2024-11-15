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
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"io/fs"
	"strings"
)

type Project struct {
	config *cfg.Config
}

func New(config *cfg.Config) *Project {
	return &Project{config}
}

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

func (p Project) IsAppKind(kind Kind) bool {
	return kind == KindAppClassic || kind == KindAppCloud
}

//go:embed common
var commonFiles embed.FS

//go:embed instance
var instanceFiles embed.FS

//go:embed app_classic
var appClassicFiles embed.FS

//go:embed app_cloud
var appCloudFiles embed.FS

func (p Project) IsScaffolded() bool {
	return p.config.TemplateFileExists()
}

func (p Project) ScaffoldWithChanged(kind Kind) (bool, error) {
	if p.IsScaffolded() {
		return false, nil
	}
	if err := p.Scaffold(kind); err != nil {
		return false, err
	}
	return true, nil
}

func (p Project) Scaffold(kind Kind) error {
	if err := p.scaffoldConventionalDirs(); err != nil {
		return err
	}
	if err := p.scaffoldDefaultFiles(kind); err != nil {
		return err
	}
	if err := p.scaffoldGitIgnore(kind); err != nil {
		return err
	}
	if err := p.scaffoldLocalEnvFile(kind); err != nil {
		return err
	}
	return nil
}

func (p Project) scaffoldConventionalDirs() error {
	log.Infof("ensuring conventional project directories")
	if err := pathx.Ensure(common.LibDir); err != nil {
		return err
	}
	if err := pathx.Ensure(common.TmpDir); err != nil {
		return err
	}
	return nil
}

func (p Project) scaffoldDefaultFiles(kind Kind) error {
	log.Infof("preparing default files for project of kind '%s'", kind)
	switch kind {
	case KindInstance:
		if err := copyEmbedFiles(&commonFiles, "common/"); err != nil {
			return err
		}
		if err := copyEmbedFiles(&instanceFiles, "instance/"); err != nil {
			return err
		}
	case KindAppClassic:
		if err := copyEmbedFiles(&commonFiles, "common/"); err != nil {
			return err
		}
		if err := copyEmbedFiles(&appClassicFiles, "app_classic/"); err != nil {
			return err
		}
	case KindAppCloud:
		if err := copyEmbedFiles(&commonFiles, "common/"); err != nil {
			return err
		}
		if err := copyEmbedFiles(&appCloudFiles, "app_cloud/"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("project kind '%s' cannot be initialized", kind)
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

// need to be in sync with osx.EnvVarsLoad()
func (p Project) scaffoldGitIgnore(kind Kind) error {
	switch kind {
	case KindAppClassic, KindAppCloud:
		return filex.AppendString(GitIgnoreFile, osx.LineSep()+strings.Join([]string{
			"",
			"# " + common.AppName,
			common.HomeDir + "/",
			common.DispatcherHomeDir + "/",
			".task/",
			"." + osx.EnvFileExt,
			"." + osx.EnvFileExt + ".*",
			"",
		}, osx.LineSep()))
	default:
		return filex.AppendString(GitIgnoreFile, osx.LineSep()+strings.Join([]string{
			"",
			"# " + common.AppName,
			common.HomeDir + "/",
			".task/",
			"." + osx.EnvFileExt,
			"." + osx.EnvFileExt + ".*",
			"",
		}, osx.LineSep()))
	}
}

func (p Project) scaffoldLocalEnvFile(kind Kind) error {
	if p.IsAppKind(kind) && p.HasProps() {
		prop, err := p.Prop(PackagePropName)
		if err != nil {
			return err
		}
		propTrimmed := strings.TrimSpace(prop)
		if propTrimmed != "" {
			if err := filex.AppendString(osx.EnvLocalFile, osx.LineSep()+strings.Join([]string{
				"",
				"AEM_PACKAGE=" + prop,
				"",
			}, osx.LineSep())); err != nil {
				return err
			}
		}
	}
	return nil
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

func (p Project) HasProps() bool {
	return pathx.Exists(PropFile)
}

func (p Project) Prop(name string) (string, error) {
	propLoader := properties.Loader{
		Encoding:         properties.ISO_8859_1,
		DisableExpansion: true,
	}
	props, err := propLoader.LoadFile(PropFile)
	if err != nil {
		return "", fmt.Errorf("cannot read project property '%s' from file '%s': %w", name, PropFile, err)
	}
	propValue := props.GetString(name, "")
	return propValue, nil
}

func (p Project) KindInfer() (Kind, error) {
	if p.HasProps() {
		log.Infof("inferring project kind basing on file '%s' and property '%s'", PropFile, KindPropName)
		propValue, err := p.Prop(KindPropName)
		if err != nil {
			return "", err
		}
		var kind Kind
		if propValue == KindPropCloudValue {
			kind = KindAppCloud
		} else if strings.HasPrefix(propValue, KindPropClassicPrefix) {
			kind = KindAppClassic
		} else {
			return "", fmt.Errorf("cannot infer project kind as value '%s' of property '%s' in file '%s' is not recognized", propValue, KindPropName, PropFile)
		}
		log.Infof("inferred project kind basing on file '%s' and property '%s' is '%s'", PropFile, KindPropName, kind)
		return kind, nil
	}
	return KindUnknown, nil
}

func (p Project) DirsIgnored() []string {
	return []string{common.HomeDir, common.DispatcherHomeDir}
}

func (p Project) GettingStarted() (string, error) {
	text := fmt.Sprintf(strings.Join([]string{
		"As a next step provide AEM files (JAR or sdk ZIP, license, service packs) to directory '" + common.LibDir + "'.",
		"Alternatively, instruct the tool where these files are located by adjusting properties: 'dist_file', 'license_file' in configuration file '" + cfg.FileDefault + "'.",
		"",
		fmt.Sprintf("Make sure to exclude the directories from VCS versioning and IDE indexing: %s", strings.Join(p.DirsIgnored(), ", ")),
		"",
		"Finally, use tasks to manage AEM instances and more:",
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
