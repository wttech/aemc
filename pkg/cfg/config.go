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
	"strings"
)

const (
	EnvPrefix           = "AEM"
	InputStdin          = "STDIN"
	OutputFileDefault   = common.LogDir + "/aem.log"
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

// NewConfig creates a new config
func NewConfig() *Config {
	result := new(Config)
	result.viper = newViper()

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
	readFromFile(v)
	readFromEnv(v)

	return v
}

func (c ConfigValues) String() string {
	yml, err := fmtx.MarshalYML(c)
	if err != nil {
		log.Errorf("Cannot convert config to YML: %s", err)
	}
	return yml
}

func readFromEnv(v *viper.Viper) {
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()
}

func readFromFile(v *viper.Viper) {
	file := File()
	exists, err := pathx.ExistsStrict(file)
	if err != nil {
		log.Debugf("skipping reading AEM config file '%s': %s", file, err)
		return
	}
	if !exists {
		log.Debugf("skipping reading AEM config file as it does not exist '%s'", file)
		return
	}
	tpl, err := tplx.New(filepath.Base(file)).ParseFiles(file)
	if err != nil {
		log.Fatalf("cannot parse AEM config file '%s': %s", file, err)
		return
	}
	data := map[string]any{
		"Env":  osx.EnvVarsMap(),
		"Path": pathx.Abs("."),
	}
	var tplOut bytes.Buffer
	if err = tpl.Execute(&tplOut, data); err != nil {
		log.Fatalf("cannot render AEM config template properly '%s': %s", file, err)
	}
	v.SetConfigType(filepath.Ext(file)[1:])
	if err = v.ReadConfig(bytes.NewReader(tplOut.Bytes())); err != nil {
		log.Fatalf("cannot load AEM config file properly '%s': %s", file, err)
	}
}

func File() string {
	path := os.Getenv(FileEnvVar)
	if path == "" {
		path = FileDefault
	}
	return path
}

func TemplateFile() string {
	path := os.Getenv(TemplateFileEnvVar)
	if path == "" {
		path = TemplateFileDefault
	}
	return path
}

func (c *Config) IsInitialized() bool {
	return pathx.Exists(File())
}

func (c *Config) Initialize() error {
	file := File()
	if pathx.Exists(file) {
		return fmt.Errorf("config file already exists: '%s'", file)
	}
	templateFile := TemplateFile()
	if !pathx.Exists(templateFile) {
		return fmt.Errorf("config file template does not exist: '%s'", templateFile)
	}
	ymlTplStr, err := filex.ReadString(templateFile)
	if err != nil {
		return err
	}
	ymlTplParsed, err := tplx.New("config-tpl").Delims("[[", "]]").Parse(ymlTplStr)
	if err != nil {
		return fmt.Errorf("cannot parse config file template '%s': '%w'", templateFile, err)
	}
	var yml bytes.Buffer
	if err := ymlTplParsed.Execute(&yml, map[string]any{ /* TODO future hook (not sure if needed or not */ }); err != nil {
		return fmt.Errorf("cannot render config file template '%s': '%w'", templateFile, err)
	}
	if err = filex.WriteString(file, yml.String()); err != nil {
		return fmt.Errorf("cannot save config file '%s': '%w'", file, err)
	}
	return nil
}

func InputFormats() []string {
	return []string{fmtx.YML, fmtx.JSON}
}

func OutputFormats() []string {
	return []string{fmtx.Text, fmtx.YML, fmtx.JSON, fmtx.None}
}

func (c *Config) ConfigureLogger() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: c.values.Log.TimestampFormat,
		FullTimestamp:   c.values.Log.FullTimestamp,
	})
	level, err := log.ParseLevel(c.values.Log.Level)
	if err != nil {
		log.Fatalf("unsupported log level specified: '%s'", c.values.Log.Level)
	}
	log.SetLevel(level)
}
