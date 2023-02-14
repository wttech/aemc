package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
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
	OutputChanged = "changed"
)

type CLI struct {
	aem    *pkg.Aem
	config *cfg.Config

	cmd   *cobra.Command
	error error

	started time.Time
	ended   time.Time

	outputFormat     string
	outputValue      string
	outputBuffer     *bytes.Buffer
	outputLogFile    string
	outputLogConsole bool
	outputResponse   *OutputResponse
	outputWriter     io.Writer
}

func NewCLI(aem *pkg.Aem, config *cfg.Config) *CLI {
	result := new(CLI)

	result.aem = aem
	result.config = config

	result.outputLogFile = common.LogFile
	result.outputLogConsole = true
	result.outputValue = common.OutputValueAll
	result.outputFormat = fmtx.Text
	result.outputBuffer = bytes.NewBufferString("")
	result.outputResponse = outputResponseDefault()
	result.cmd = result.rootCmd()

	return result
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

func (c *CLI) Exec() {
	c.error = c.cmd.Execute()
}

func (c *CLI) configure() {
	c.config.ConfigureLogger()
	c.configureOutput()
	c.aem.Configure(c.config)
	c.started = time.Now()
}

func (c *CLI) configureOutput() {
	c.outputValue = c.config.Values().Output.Value
	c.outputFormat = strings.ReplaceAll(c.config.Values().Output.Format, "yaml", "yml")
	c.outputLogFile = c.config.Values().Output.Log.File
	c.outputLogConsole = c.config.Values().Output.Log.Console

	if c.outputValue != common.OutputValueNone && c.outputValue != common.OutputValueAll {
		c.outputFormat = fmtx.Text
		c.outputLogConsole = false
	}
	if c.outputFormat != fmtx.Text {
		c.outputLogConsole = false
	}
	if !lo.Contains(cfg.OutputFormats(), c.outputFormat) {
		log.Fatalf("unsupported CLI output format detected '%s'! supported ones are: %s", c.outputFormat, strings.Join(cfg.OutputFormats(), ", "))
	}

	if c.outputFormat == fmtx.Text {
		if !c.outputLogConsole {
			outputWriter := c.openOutputLogFile()
			c.aem.SetOutput(outputWriter)
			log.SetOutput(outputWriter)
		}
	} else {
		outputWriter := io.MultiWriter(c.outputBuffer, c.openOutputLogFile())
		c.aem.SetOutput(outputWriter)
		log.SetOutput(outputWriter)
	}
}

func (c *CLI) openOutputLogFile() *os.File {
	err := pathx.Ensure(path.Dir(c.outputLogFile))
	if err != nil {
		return nil
	}
	file, err := os.OpenFile(c.outputLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf(fmt.Sprintf("cannot open/create AEM output file properly at path '%s': %s", c.outputLogFile, err))
	}
	return file
}

func (c *CLI) elapsed() time.Duration {
	return c.ended.Sub(c.started)
}

// Exit reads response data then prints currently captured output then exits app with proper status code
func (c *CLI) exit() {
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

func (c *CLI) printOutputText() {
	if c.outputValue == common.OutputValueAll {
		c.printDataAll()
		c.printCommandResult()
	} else if c.outputValue == common.OutputValueNone {
		c.printCommandResult()
	} else {
		c.printDataValue()
	}
}

func (c *CLI) printCommandResult() {
	r := c.outputResponse
	msg := fmt.Sprintf("%s", r.Msg)
	if color.NoColor || !c.outputLogConsole {
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
func (c *CLI) printDataValue() {
	value, ok := c.outputResponse.Data[c.outputValue]
	if !ok {
		println("<undefined>")
	} else {
		println(fmtx.MarshalText(value))
	}
}

func (c *CLI) printDataAll() {
	if len(c.outputResponse.Data) > 0 {
		c.printOutputDataIndented(textio.NewPrefixWriter(os.Stdout, ""), c.outputResponse.Data, "")
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
	switch c.outputFormat {
	case fmtx.JSON:
		json, err := fmtx.MarshalJSON(c.outputResponse)
		if err != nil {
			log.Fatalf("cannot serialize CLI output to to target JSON format!")
		}
		fmt.Println(json)
	case fmtx.YML:
		yml, err := fmtx.MarshalYML(c.outputResponse)
		if err != nil {
			log.Fatalf("cannot serialize CLI output to to target YML format!")
		}
		fmt.Println(yml)
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
	format := c.config.Values().Input.Format
	str := c.config.Values().Input.String
	file := c.config.Values().Input.File

	if len(str) > 0 {
		err := fmtx.UnmarshalDataInFormat(format, strings.NewReader(str), out)
		if err != nil {
			return fmt.Errorf("cannot parse string input properly: %w", err)
		}
	} else if file == common.STDIn {
		err := fmtx.UnmarshalDataInFormat(format, bufio.NewReader(os.Stdin), out)
		if err != nil {
			return fmt.Errorf("cannot parse STDIN input properly: %w", err)
		}
	} else {
		if !pathx.Exists(file) {
			return fmt.Errorf("cannot load input file as it does not exist '%s'", file)
		}
		if err := fmtx.UnmarshalFileInFormat(format, file, out); err != nil {
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
		name = stringsx.SnakeCase(name)
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
