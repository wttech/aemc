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
	viper *viper.Viper
}

func (c *Config) Values() *viper.Viper {
	return c.viper
}

func NewConfig() *Config {
	result := new(Config)
	result.setDefaults()
	result.readFromFile(FileEffective(), true)
	result.readFromEnv()
	return result
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

func (c *Config) tplString(text string) string {
	result, err := tplx.RenderString(text, c.tplData())
	if err != nil {
		log.Fatalf("cannot render AEM config string\nerror: %s\nvalue:\n%s", err, text)
	}
	return result
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
