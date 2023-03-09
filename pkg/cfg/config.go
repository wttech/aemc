package cfg

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/tplx"
	"os"
	"path/filepath"
	"runtime"
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
	viper  *viper.Viper
	values *ConfigValues
}

func (c *Config) Values() *ConfigValues {
	return c.values
}

func (c *Config) ValuesMap() map[string]any {
	return c.viper.AllSettings()
}

// NewConfig creates a new config
func NewConfig() *Config {
	result := new(Config)
	result.viper = newViper()

	result.readFromFile(FileEffective(), true)
	result.readFromEnv()

	var v ConfigValues
	err := result.viper.Unmarshal(&v)
	if err != nil {
		log.Fatalf(fmt.Sprintf("cannot unmarshal AEM config values properly: %s", err))
	}
	result.values = &v
	return result
}

func newViper() *viper.Viper {
	v := viper.New()
	setDefaults(v)
	return v
}

func (c ConfigValues) String() string {
	yml, err := fmtx.MarshalYML(c)
	if err != nil {
		log.Errorf("Cannot convert config to YML: %s", err)
	}
	return yml
}

func (c *Config) readFromEnv() {
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.viper.SetEnvPrefix(EnvPrefix)
	c.viper.AutomaticEnv()
}

func (c *Config) readFromFile(file string, templating bool) {
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
	var config []byte
	if templating {
		tpl, err := tplx.New(filepath.Base(file)).Delims("[[", "]]").ParseFiles(file)
		if err != nil {
			log.Fatalf("cannot parse AEM config file '%s': %s", file, err)
			return
		}
		var tplOut bytes.Buffer
		if err = tpl.Execute(&tplOut, c.tplData()); err != nil {
			log.Fatalf("cannot render AEM config template properly '%s': %s", file, err)
			return
		}
		config = tplOut.Bytes()
	} else {
		config, err = filex.Read(file)
		if err != nil {
			log.Fatalf("cannot read AEM config file '%s': %s", file, err)
			return
		}
	}
	c.viper.SetConfigType(filepath.Ext(file)[1:])
	if err = c.viper.MergeConfig(bytes.NewReader(config)); err != nil {
		log.Fatalf("cannot load AEM config file properly '%s': %s", file, err)
		return
	}
}

// TODO config defaults are not interpolated, maybe they should with data below
func (c *Config) tplData() map[string]any {
	var ext string
	if osx.IsWindows() {
		ext = "zip"
	} else {
		ext = "tar.gz"
	}
	data := map[string]any{
		"Env":        osx.EnvVarsMap(),
		"Path":       pathx.Normalize("."),
		"Os":         runtime.GOOS,
		"Arch":       runtime.GOARCH,
		"ArchiveExt": ext,
	}
	return data
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

func (c *Config) ExportWithChanged(file string) (bool, error) {
	currentYml, err := fmtx.MarshalYML(c.values)
	if err != nil {
		return false, err
	}

	executable, err := os.Executable()
	if err != nil {
		return false, err
	}
	log.Infof("config file can be used by command '%s=%s %s'", FileEnvVar, pathx.Canonical(file), executable)

	if pathx.Exists(file) {
		oldYml, err := filex.ReadString(file)
		if err != nil {
			return false, err
		}
		if oldYml == currentYml {
			return false, nil
		}
	}
	if err := filex.WriteString(file, currentYml); err != nil {
		return false, err
	}
	return true, nil
}
