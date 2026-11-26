package pkg

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/project"
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

func (p Project) ScaffoldWithChanged(kind project.Kind, runner project.Runner) (bool, error) {
	if p.IsScaffolded() {
		return false, nil
	}
	if err := p.Scaffold(kind, runner); err != nil {
		return false, err
	}
	return true, nil
}

func (p Project) Scaffold(kind project.Kind, runner project.Runner) error {
	if err := p.scaffoldDefaultFiles(kind, runner); err != nil {
		return err
	}
	if err := p.scaffoldGitIgnore(kind, runner); err != nil {
		return err
	}
	if err := p.scaffoldLocalEnvFile(kind); err != nil {
		return err
	}
	return nil
}

func (p Project) scaffoldDefaultFiles(kind project.Kind, runner project.Runner) error {
	log.Infof("preparing default files for project of kind '%s' run by '%s'", kind, runner)
	var kindFiles embed.FS
	switch kind {
	case project.KindInstance:
		kindFiles = project.InstanceFiles
	case project.KindAppClassic:
		kindFiles = project.AppClassicFiles
	case project.KindAppCloud:
		kindFiles = project.AppCloudFiles
	default:
		return fmt.Errorf("project kind '%s' cannot be initialized", kind)
	}
	for _, dir := range []struct {
		files embed.FS
		path  string
	}{
		{kindFiles, string(kind)},
		{project.RunnerFiles, "runner/" + string(runner) + "/common"},
		{project.RunnerFiles, "runner/" + string(runner) + "/" + string(kind)},
	} {
		if err := copyEmbedFiles(dir.files, dir.path); err != nil {
			return err
		}
	}
	return nil
}

// copyEmbedFiles copies the files under dir to the project; '$' in a path is dropped, so '$.mvn' becomes '.mvn' (embed skips dot files).
func copyEmbedFiles(efs embed.FS, dir string) error {
	sub, err := fs.Sub(efs, dir)
	if err != nil {
		return err
	}
	return fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && path == "." {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		bytes, err := fs.ReadFile(sub, path)
		if err != nil {
			return err
		}
		return filex.Write(strings.ReplaceAll(path, "$", ""), bytes)
	})
}

// need to be in sync with osx.EnvVarsLoad()
func (p Project) scaffoldGitIgnore(kind project.Kind, runner project.Runner) error {
	lines := []string{"", "# " + common.AppName, common.HomeDir + "/"}
	if p.IsAppKind(kind) {
		lines = append(lines, common.DispatcherHomeDir+"/")
	}
	lines = append(lines, runner.GitIgnored()...)
	lines = append(lines, "."+osx.EnvFileExt, "."+osx.EnvFileExt+".*", "")
	return filex.AppendString(project.GitIgnoreFile, osx.LineSep()+strings.Join(lines, osx.LineSep()))
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
				"# AEM Application",
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

// RunnerDetected is the runner of a scaffolded project, recognised by its task file.
func (p Project) RunnerDetected() project.Runner {
	if pathx.Exists(project.TaskFile) {
		return project.RunnerTask
	}
	return project.RunnerMise
}

func (p Project) ScaffoldGettingStarted(runner project.Runner) string {
	lines := []string{
		"AEM Compose project now contains required files.",
		"",
		"Consider saving the project to VCS repository.",
		"The next step is to initialize the project by running:",
		"",
		runner.TaskCommand("init"),
	}
	if runner == project.RunnerMise {
		lines = append(lines, "", "Without mise installed, run './bin/mise' instead of 'mise', e.g. on CI.")
	}
	return strings.Join(lines, "\n")
}

func (p Project) InitGettingStartedError() string {
	return strings.Join([]string{
		"AEM Compose project is not yet ready to use!",
		"",
		"Be sure to provide AEM files (SDK ZIP or Quickstart JAR + License + Service Packs) to directory '" + p.aem.BaseOpts().LibDir + "'.",
	}, "\n")
}

func (p Project) InitGettingStartedSuccess() string {
	runner := p.RunnerDetected()
	return strings.Join([]string{
		"AEM Compose project is ready to use!",
		"",
		fmt.Sprintf("Consider excluding the directories from VCS versioning and IDE indexing: %s", strings.Join(p.DirsIgnored(), ", ")),
		"Discover available commands by running:",
		"",
		runner.AemCommand("--help"),
		"",
		"Multiple commands are usually organized into tasks. Discover them by running:",
		"",
		runner.TaskListCommand(),
		"",
		"Now it is a time to setup AEM environment then build and deploy the AEM application. Typically it is done by running:",
		"",
		runner.TaskCommand("setup"),
	}, "\n")
}

func (p Project) InitWithChanged() (bool, error) {
	changed := false

	baseChanged, err := p.aem.BaseOpts().PrepareWithChanged()
	changed = changed || baseChanged
	if err != nil {
		return changed, err
	}

	vendorPrepared, err := p.aem.VendorManager().PrepareWithChanged(true)
	changed = changed || vendorPrepared
	if err != nil {
		return changed, err
	}

	return changed, nil
}
