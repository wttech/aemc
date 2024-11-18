package pkg

import (
	"embed"
	"fmt"
	"github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/project"
	"io/fs"
	"strings"
)

type Project struct {
	aem *AEM
}

func NewProject(aem *AEM) *Project {
	return &Project{aem}
}

func (p Project) IsAppKind(kind project.Kind) bool {
	return kind == project.KindAppClassic || kind == project.KindAppCloud
}

func (p Project) IsScaffolded() bool {
	return p.aem.config.TemplateFileExists()
}

func (p Project) ScaffoldWithChanged(kind project.Kind) (bool, error) {
	if p.IsScaffolded() {
		return false, nil
	}
	if err := p.Scaffold(kind); err != nil {
		return false, err
	}
	return true, nil
}

func (p Project) Scaffold(kind project.Kind) error {
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

func (p Project) scaffoldDefaultFiles(kind project.Kind) error {
	log.Infof("preparing default files for project of kind '%s'", kind)
	switch kind {
	case project.KindInstance:
		if err := copyEmbedFiles(&project.CommonFiles, "common/"); err != nil {
			return err
		}
		if err := copyEmbedFiles(&project.InstanceFiles, "instance/"); err != nil {
			return err
		}
	case project.KindAppClassic:
		if err := copyEmbedFiles(&project.CommonFiles, "common/"); err != nil {
			return err
		}
		if err := copyEmbedFiles(&project.AppClassicFiles, "app_classic/"); err != nil {
			return err
		}
	case project.KindAppCloud:
		if err := copyEmbedFiles(&project.CommonFiles, "common/"); err != nil {
			return err
		}
		if err := copyEmbedFiles(&project.AppCloudFiles, "app_cloud/"); err != nil {
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
func (p Project) scaffoldGitIgnore(kind project.Kind) error {
	switch kind {
	case project.KindAppClassic, project.KindAppCloud:
		return filex.AppendString(project.GitIgnoreFile, osx.LineSep()+strings.Join([]string{
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
		return filex.AppendString(project.GitIgnoreFile, osx.LineSep()+strings.Join([]string{
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

func (p Project) scaffoldLocalEnvFile(kind project.Kind) error {
	if p.IsAppKind(kind) && p.HasProps() {
		prop, err := p.Prop(project.PackagePropName)
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

func (p Project) KindDetermine(name string) (project.Kind, error) {
	var kind project.Kind = project.KindAuto
	if name != "" {
		kindCandidate, err := project.KindOf(name)
		if err != nil {
			return "", err
		}
		kind = kindCandidate
	}
	if kind == project.KindAuto {
		kindCandidate, err := p.KindInfer()
		if err != nil {
			return "", err
		}
		kind = kindCandidate
	}
	return kind, nil
}

func (p Project) HasProps() bool {
	return pathx.Exists(project.PropFile)
}

func (p Project) Prop(name string) (string, error) {
	propLoader := properties.Loader{
		Encoding:         properties.ISO_8859_1,
		DisableExpansion: true,
	}
	props, err := propLoader.LoadFile(project.PropFile)
	if err != nil {
		return "", fmt.Errorf("cannot read project property '%s' from file '%s': %w", name, project.PropFile, err)
	}
	propValue := props.GetString(name, "")
	return propValue, nil
}

func (p Project) KindInfer() (project.Kind, error) {
	if p.HasProps() {
		log.Infof("inferring project kind basing on file '%s' and property '%s'", project.PropFile, project.KindPropName)
		propValue, err := p.Prop(project.KindPropName)
		if err != nil {
			return "", err
		}
		var kind project.Kind
		if propValue == project.KindPropCloudValue {
			kind = project.KindAppCloud
		} else if strings.HasPrefix(propValue, project.KindPropClassicPrefix) {
			kind = project.KindAppClassic
		} else {
			return "", fmt.Errorf("cannot infer project kind as value '%s' of property '%s' in file '%s' is not recognized", propValue, project.KindPropName, project.PropFile)
		}
		log.Infof("inferred project kind basing on file '%s' and property '%s' is '%s'", project.PropFile, project.KindPropName, kind)
		return kind, nil
	}
	return project.KindUnknown, nil
}

func (p Project) DirsIgnored() []string {
	return []string{common.HomeDir, common.DispatcherHomeDir}
}

func (p Project) ScaffoldGettingStarted() string {
	libDir := p.aem.BaseOpts().LibDir

	text := fmt.Sprintf(strings.Join([]string{
		"As a next step provide AEM files (JAR or sdk ZIP, license, service packs) to directory '" + libDir + "'.",
		"Alternatively, instruct the tool where these files are located by adjusting properties: 'dist_file', 'license_file' in configuration file '" + cfg.FileDefault + "'.",
		"",
		"Use tasks to manage AEM instances and more:",
		"",
		"sh taskw --list",
		"",
		"It is also possible to run individual AEM Compose CLI commands separately.",
		"Discover available commands by running:",
		"",
		"sh aemw --help",
	}, "\n"))
	return text
}

func (p Project) InitGettingStartedError() string {
	libDir := p.aem.BaseOpts().LibDir
	text := fmt.Sprintf(strings.Join([]string{
		"Be sure to provide AEM files (JAR or sdk ZIP, license, service packs) to directory '" + libDir + "'.",
	}, "\n"))
	return text
}

func (p Project) InitGettingStartedSuccess() string {
	text := fmt.Sprintf(strings.Join([]string{
		"AEM Compose project is ready to use.",
		"",
		fmt.Sprintf("Make sure to exclude the directories from VCS versioning and IDE indexing: %s", strings.Join(p.DirsIgnored(), ", ")),
		"",
		"Use tasks to manage AEM instances and more:",
		"",
		"sh taskw --list",
		"",
		"It is also possible to run individual AEM Compose CLI commands separately.",
		"Discover available commands by running:",
		"",
		"sh aemw --help",
	}, "\n"))
	return text
}