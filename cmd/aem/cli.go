package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/iancoleman/strcase"
	"github.com/jmespath-community/go-jmespath"
	"github.com/samber/lo"
	"github.com/segmentio/textio"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	OutputChanged  = "changed"
	OutputInstance = "instance"
)

type CLI struct {
	cmd     *cobra.Command
	started time.Time
	ended   time.Time

	inputFormat string
	inputString string
	inputFile   string

	outputFormat   string
	outputValue    string
	outputBuffer   *bytes.Buffer
	outputLogFile  string
	outputLogMode  string
	outputResponse *OutputResponse
	outputQuery    string
	outputWriter   io.Writer
	outputNoColor  bool

	config *cfg.Config
	aem    *pkg.AEM
}

func NewCLI() *CLI {
	c := new(CLI)
	c.config = cfg.NewConfig()
	c.cmd = c.rootCmd()
	return c
}

// OutputResponse defines a structure of data to be printed
type OutputResponse struct {
	Msg     string         `yaml:"msg" json:"msg"`
	Failed  bool           `yaml:"failed" json:"failed"`
	Changed bool           `yaml:"changed" json:"changed"`
	Log     string         `yaml:"log" json:"log"`
	Data    map[string]any `yaml:"data" json:"data"`
	Ended   time.Time      `yaml:"ended" json:"ended"`
	Elapsed time.Duration  `yaml:"elapsed" json:"elapsed"`
}

func outputResponseDefault() *OutputResponse {
	return &OutputResponse{
		Msg:     "",
		Failed:  false,
		Changed: false,
		Log:     "",
		Data:    map[string]any{},
	}
}

func (c *CLI) Exec() error {
	return c.cmd.Execute()
}

func (c *CLI) MustExec() {
	if err := c.Exec(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

// onStart initializes CLI settings (not the NewCLI method) because since that moment PFlags are available (they are bound to Viper config)
// note that using 'c.aem' before that moment may lead to unexpected behavior
func (c *CLI) onStart() {
	c.aem = pkg.NewAEM(c.config)
	cv := c.config.Values()

	c.inputFormat = cv.GetString("input.format")
	c.inputString = cv.GetString("input.string")
	c.inputFile = cv.GetString("input.file")

	c.outputBuffer = bytes.NewBufferString("")
	c.outputResponse = outputResponseDefault()
	c.outputValue = cv.GetString("output.value")
	c.outputFormat = strings.ReplaceAll(cv.GetString("output.format"), "yaml", "yml")
	c.outputQuery = cv.GetString("output.query")
	c.outputLogFile = cv.GetString("output.log.file")
	c.outputLogMode = cv.GetString("output.log.mode")

	if c.outputValue != common.OutputValueNone && c.outputValue != common.OutputValueAll {
		c.outputFormat = fmtx.Text
	}

	noColor := cv.GetBool("output.no_color")
	if c.outputFormat != fmtx.Text || c.outputLogMode != cfg.OutputLogConsole {
		noColor = true
	}
	c.outputNoColor = noColor
	color.NoColor = noColor

	log.SetFormatter(&log.TextFormatter{
		ForceColors:     !c.outputNoColor,
		DisableColors:   c.outputNoColor,
		TimestampFormat: cv.GetString("log.timestamp_format"),
		FullTimestamp:   cv.GetBool("log.full_timestamp"),
	})
	levelName := cv.GetString("log.level")
	level, err := log.ParseLevel(levelName)
	if err != nil {
		log.Fatalf("unsupported CLI log level specified: '%s'", levelName)
	}
	log.SetLevel(level)

	if !lo.Contains(cfg.OutputFormats(), c.outputFormat) {
		log.Fatalf("unsupported CLI output format detected '%s'! supported ones are: %s", c.outputFormat, strings.Join(cfg.OutputFormats(), ", "))
	}

	if c.outputFormat == fmtx.Text {
		outputWriterLogDefault := log.StandardLogger().Out
		switch c.outputLogMode {
		case cfg.OutputLogNone:
			outputWriter := io.Discard
			log.SetOutput(outputWriter)
			c.aem.SetOutput(outputWriter)
			break
		case cfg.OutputLogFile:
			outputWriter := c.openOutputLogFile()
			log.SetOutput(outputWriter)
			c.aem.SetOutput(outputWriter)
			break
		case cfg.OutputLogBoth:
			log.SetOutput(io.MultiWriter(outputWriterLogDefault, c.openOutputLogFile()))
			c.aem.SetOutput(io.MultiWriter(os.Stdout, c.openOutputLogFile()))
			break
		case cfg.OutputLogConsole:
			log.SetOutput(outputWriterLogDefault)
			c.aem.SetOutput(os.Stdout)
			break
		default:
			log.Fatalf("unsupported CLI output log mode specified: '%s'", c.outputLogMode)
		}
	} else {
		outputWriter := io.MultiWriter(c.outputBuffer, c.openOutputLogFile())
		c.aem.SetOutput(outputWriter)
		log.SetOutput(outputWriter)
	}

	c.started = time.Now()
}

// onEnd reads response data then prints currently captured output then exits app with proper status code
func (c *CLI) onEnd() {
	c.ended = time.Now()
	c.outputResponse.Ended = c.ended
	c.outputResponse.Elapsed = c.elapsed()
	c.outputResponse.Log = c.outputBuffer.String()

	if c.outputFormat == fmtx.Text {
		c.printOutputText()
	} else {
		c.printOutputMarshaled()
	}

	if c.outputResponse.Failed {
		os.Exit(1)
	}
	os.Exit(0)
}

func (c *CLI) elapsed() time.Duration {
	return c.ended.Sub(c.started)
}

func (c *CLI) openOutputLogFile() *os.File {
	err := pathx.Ensure(path.Dir(c.outputLogFile))
	if err != nil {
		return nil
	}
	file, err := os.OpenFile(c.outputLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf(fmt.Sprintf("cannot open/create CLI output file '%s': %s", c.outputLogFile, err))
	}
	return file
}

func (c *CLI) printOutputText() {
	if c.outputValue == common.OutputValueAll {
		c.printCommandResult()
		c.printDataAll()
	} else if c.outputValue == common.OutputValueNone {
		c.printCommandResult()
	} else if c.outputValue == common.OutputValueOnly {
		c.printDataAll()
	} else {
		c.printDataValue(c.outputValue)
	}
}

func (c *CLI) printCommandResult() {
	r := c.outputResponse
	msg := fmt.Sprintf("%s", r.Msg)

	if c.outputNoColor {
		entry := log.WithField("changed", r.Changed).WithField("elapsed", r.Elapsed)
		if r.Failed {
			entry.Errorf(msg)
		} else {
			entry.Infof(msg)
		}
	} else {
		if r.Failed {
			log.Errorf(color.RedString(msg))
		} else {
			if r.Changed {
				log.Info(color.YellowString(msg))
			} else {
				log.Info(color.GreenString(msg))
			}
		}
	}
}

// TODO allow to print 'changed', 'failed', 'elapsed', 'ended' as well
func (c *CLI) printDataValue(key string) {
	value, ok := c.outputResponse.Data[key]
	if !ok {
		fmt.Println("<undefined>")
	} else {
		fmt.Println(fmtx.MarshalText(value))
	}
}

func (c *CLI) printDataAll() {
	if len(c.outputResponse.Data) > 0 {
		c.printOutputDataIndented(textio.NewPrefixWriter(c.aem.Output(), ""), c.outputResponse.Data, "")
	}
}

func (c *CLI) printOutputDataIndented(writer *textio.PrefixWriter, value any, key string) {
	rv := reflect.ValueOf(value)
	switch rv.Type().Kind() {
	case reflect.Slice, reflect.Array:
		if rv.Len() == 0 {
			c.printOutputDataIndented(writer, "<empty>", "")
		} else {
			for i := 0; i < rv.Len(); i++ {
				iv := rv.Index(i).Interface()
				c.printOutputDataIndented(writer, iv, "")
			}
		}
	case reflect.Map:
		if rv.Len() == 0 {
			c.printOutputDataIndented(writer, "<empty>", "")
		} else {
			dw := textio.NewPrefixWriter(writer, "  ")
			keys := rv.MapKeys()
			sort.SliceStable(keys, func(k1, k2 int) bool {
				return strings.Compare(fmt.Sprintf("%v", keys[k1].Interface()), fmt.Sprintf("%v", keys[k2].Interface())) < 0
			})
			for _, k := range keys {
				mapKey := fmt.Sprintf("%s", k)
				_, _ = writer.WriteString(color.BlueString(stringsx.HumanCase(mapKey)) + "\n")
				mv := rv.MapIndex(k).Interface()
				c.printOutputDataIndented(dw, mv, mapKey)
			}
		}
	default:
		boolValue, boolOk := value.(bool)
		if boolOk && key == OutputChanged {
			value = formatValueChanged(boolValue)
		}
		_, _ = writer.WriteString(strings.TrimSuffix(fmtx.MarshalText(value), "\n") + "\n")
	}
}

func (c *CLI) printOutputMarshaled() {
	outputTransformed, err := jmespath.Search(c.outputQuery, c.outputResponse)
	if err != nil {
		log.Fatalf("cannot transform CLI output data: %s", err)
	}
	switch c.outputFormat {
	case fmtx.JSON:
		json, err := fmtx.MarshalJSON(outputTransformed)
		if err != nil {
			log.Fatalf("cannot serialize CLI output to target JSON format!")
		}
		fmt.Println(json)
		break
	case fmtx.YML:
		yml, err := fmtx.MarshalYML(outputTransformed)
		if err != nil {
			log.Fatalf("cannot serialize CLI output to target YML format!")
		}
		fmt.Println(yml)
		break
	default:
		log.Fatalf("cannot serialize CLI output to unsupported format '%s'", c.outputFormat)
	}
}

func (c *CLI) Ok(message string) {
	c.Success(message, false)
}

func (c *CLI) Changed(message string) {
	c.Success(message, true)
}

func (c *CLI) Success(message string, changed bool) {
	c.outputResponse.Failed = false
	c.outputResponse.Changed = changed
	c.outputResponse.Msg = message
}

func (c *CLI) Fail(msg string) {
	c.outputResponse.Failed = true
	c.outputResponse.Msg = msg
}

func (c *CLI) Error(err error) {
	c.Fail(fmt.Sprintf("%s", err))
}

func (c *CLI) ReadInput(out any) error {
	if c.inputString != "" {
		err := fmtx.UnmarshalDataInFormat(c.inputFormat, io.NopCloser(strings.NewReader(c.inputString)), out)
		if err != nil {
			return fmt.Errorf("cannot parse string input properly: %w", err)
		}
	} else if c.inputFile == common.STDIn {
		err := fmtx.UnmarshalDataInFormat(c.inputFormat, io.NopCloser(bufio.NewReader(os.Stdin)), out)
		if err != nil {
			return fmt.Errorf("cannot parse STDIN input properly: %w", err)
		}
	} else {
		if !pathx.Exists(c.inputFile) {
			return fmt.Errorf("cannot load input file as it does not exist '%s'", c.inputFile)
		}
		if err := fmtx.UnmarshalFileInFormat(c.inputFormat, c.inputFile, out); err != nil {
			return err
		}
	}
	return nil
}

func (c *CLI) SetOutput(name string, data any) {
	c.setOutput(name, data)
}

func (c *CLI) AddOutput(name string, data any) {
	c.addOutput(name, data)
}

func (c *CLI) setOutput(name string, data any) {
	c.outputResponse.Data[c.fixOutputName(name)] = data
}

func (c *CLI) addOutput(name string, data any) {
	var result []any
	existing, ok := c.outputResponse.Data[name]
	if ok {
		existingTyped, ok := existing.([]any)
		if ok {
			result = append(result, existingTyped...)
		}
	}
	result = append(result, data)

	c.setOutput(name, result)
}

func (c *CLI) fixOutputName(name string) string {
	if c.outputFormat == fmtx.YML {
		name = strcase.ToSnake(name)
	}
	return name
}

func formatValueFailed(failed bool) string {
	text := "false"
	if failed {
		text = color.RedString("true")
	}
	return text
}

func formatValueChanged(changed bool) string {
	text := "false"
	if changed {
		text = color.YellowString("true")
	}
	return text
}

func InstancesChanged(instanceData []map[string]any) []pkg.Instance {
	var result []pkg.Instance
	for _, data := range instanceData {
		if data[OutputChanged] == true {
			result = append(result, data[OutputInstance].(pkg.Instance))
		}
	}
	return result
}
