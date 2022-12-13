package cfg

import (
	_ "embed"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/osx"
	"os"
	"strings"
)

const (
	FileBaseName = "aem"
	FileType     = "yml"
	FileName     = FileBaseName + "." + FileType
	FilePath     = "./aem/home"
	EnvPrefix    = "AEM"
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
	yml, err := fmtx.MarshalYAML(c)
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
	v.SetConfigName(FileBaseName)
	v.SetConfigType(FileType)
	v.AddConfigPath(filePath())

	if err := v.ReadInConfig(); err != nil {
		log.Tracef("cannot load AEM config file properly: %s", err)
	}
}

func filePath() string {
	path := os.Getenv("AEM_CONFIG_PATH")
	if len(path) == 0 {
		path = FilePath
	}
	return path
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

//go:embed aem.yml
var configYml string

func (c *Config) Init() error {
	file := c.File()
	if osx.PathExists(file) {
		return fmt.Errorf("config file already exists: '%s'", file)
	}
	err := osx.FileWrite(file, configYml)
	if err != nil {
		return fmt.Errorf("cannot create initial config file '%s': '%w'", file, err)
	}
	return nil
}

func (c *Config) File() string {
	return filePath() + "/" + FileName
}

const (
	InputStdin string = "STDIN"
	OutputFile string = "aem/home/aem.log"
)

func InputFormats() []string {
	return []string{fmtx.YML, fmtx.JSON}
}

// OutputFormats returns all available output formats
func OutputFormats() []string {
	return []string{fmtx.Text, fmtx.YML, fmtx.JSON, fmtx.None}
}
