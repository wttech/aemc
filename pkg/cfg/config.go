package cfg

import (
	"fmt"
	"github.com/compose-spec/compose-go/v2/template"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"os"
	"strings"
)

const (
	EnvPrefix           = "AEM"
	FileDefault         = common.ConfigDir + "/aem.yml"
	FileEnvVar          = "AEM_CONFIG_FILE"
	TemplateFileDefault = common.DefaultDir + "/" + common.ConfigDirName + "/aem.yml"
	TemplateFileEnvVar  = "AEM_CONFIG_TEMPLATE"
)

// Config defines a place for managing input configuration from various sources (YML file, env vars, etc)
type Config struct {
	viper *viper.Viper
}

func (c *Config) Values() *viper.Viper {
	return c.viper
}

func NewConfig() *Config {
	result := new(Config)
	result.setDefaults()
	result.readFromFile(FileEffective())
	result.readFromEnv()
	return result
}

func (c *Config) readFromEnv() {
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.viper.SetEnvPrefix(EnvPrefix)
	c.viper.AutomaticEnv()
}

func (c *Config) readFromFile(file string) {
	if file == "" {
		log.Debugf("skipping reading AEM config file as it is not provided")
		return
	}
	exists, err := pathx.ExistsStrict(file)
	if err != nil {
		log.Debugf("skipping reading AEM config file '%s': %s", file, err)
		return
	}
	if !exists {
		log.Debugf("skipping reading AEM config file as it does not exist '%s'", file)
		return
	}
	content, err := filex.Read(file)
	if err != nil {
		log.Fatalf("cannot read AEM config file '%s': %s", file, err)
		return
	}
	values, err := interpolateYAML(content, variables())
	if err != nil {
		log.Fatalf("cannot load AEM config file '%s': %s", file, err)
		return
	}
	if err = c.viper.MergeConfigMap(values); err != nil {
		log.Fatalf("cannot merge AEM config file '%s': %s", file, err)
		return
	}
}

// interpolateString replaces variables in a config value the same way as in the config file.
func (c *Config) interpolateString(text string) string {
	result, err := template.Substitute(text, variables())
	if err != nil {
		log.Fatalf("cannot interpolate AEM config value\nerror: %s\nvalue:\n%s", err, text)
	}
	return result
}

func FileEffective() string {
	var file string
	for _, file := range []string{File(), TemplateFile()} {
		if pathx.Exists(file) {
			return file
		}
	}
	return file
}

func File() string {
	path := os.Getenv(FileEnvVar)
	if path == "" {
		path = FileDefault
	}
	return path
}

func (c *Config) FileExists() bool {
	return pathx.Exists(File())
}

func TemplateFile() string {
	path := os.Getenv(TemplateFileEnvVar)
	if path == "" {
		path = TemplateFileDefault
	}
	return path
}

func (c *Config) TemplateFileExists() bool {
	return pathx.Exists(TemplateFile())
}

func (c *Config) InitializeWithChanged() (bool, error) {
	file := File()
	if pathx.Exists(file) {
		return false, nil
	}
	templateFile := TemplateFile()
	if !pathx.Exists(templateFile) {
		return false, fmt.Errorf("config file template does not exist: '%s'", templateFile)
	}
	if err := filex.Copy(templateFile, file, false); err != nil {
		return false, fmt.Errorf("cannot copy config file template: '%w'", err)
	}
	return true, nil
}

func InputFormats() []string {
	return []string{fmtx.YML, fmtx.JSON}
}

func OutputFormats() []string {
	return []string{fmtx.Text, fmtx.YML, fmtx.JSON}
}

const (
	OutputLogNone    = "none"
	OutputLogFile    = "file"
	OutputLogConsole = "console"
	OutputLogBoth    = "both"
)

func OutputLogModes() []string {
	return []string{OutputLogConsole, OutputLogFile, OutputLogBoth, OutputLogNone}
}

func (c *Config) Export(file string) error {
	if err := c.Values().SafeWriteConfigAs(file); err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	log.Infof("exported config file can be used by command '%s=%s %s'", FileEnvVar, pathx.Canonical(file), executable)
	return nil
}
