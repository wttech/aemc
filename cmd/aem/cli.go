package main

import (
	"bufio"
	"bytes"
	"fmt"
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
	"github.com/wttech/aemc/pkg/common/timex"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	OutputFileDefault = common.LogDir + "/aem.log"
)

type CLI struct {
	aem    *pkg.Aem
	config *cfg.Config

	cmd   *cobra.Command
	error error

	started time.Time
	ended   time.Time

	outputFormat   string
	outputValue    string
	outputBuffer   *bytes.Buffer
	outputFile     string
	outputResponse *OutputResponse
	outputWriter   io.Writer
}

func NewCLI(aem *pkg.Aem, config *cfg.Config) *CLI {
	result := new(CLI)

	result.aem = aem
	result.config = config

	result.outputFormat = fmtx.Text
	result.outputFile = OutputFileDefault
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
	if len(c.outputValue) > 0 {
		c.outputFormat = fmtx.Text
	} else {
		c.outputFile = c.config.Values().Output.File
		c.outputFormat = strings.ReplaceAll(c.config.Values().Output.Format, "yaml", "yml")
	}

	if !lo.Contains(cfg.OutputFormats(), c.outputFormat) {
		log.Fatalf("unsupported CLI output format detected! supported ones are: %s", strings.Join(cfg.OutputFormats(), ", "))
	}

	if c.outputFormat == fmtx.None { // print to file but not to stdout
		outputWriter := c.openOutputFile()
		c.aem.SetOutput(outputWriter)
		log.SetOutput(outputWriter)
	} else if c.outputFormat != fmtx.Text { // print to file but also buffer to later print serialized to stdout
		outputWriter := io.MultiWriter(c.outputBuffer, c.openOutputFile())
		c.aem.SetOutput(outputWriter)
		log.SetOutput(outputWriter)
	}
}

func (c *CLI) openOutputFile() *os.File {
	err := pathx.Ensure(path.Dir(c.outputFile))
	if err != nil {
		return nil
	}
	file, err := os.OpenFile(c.outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf(fmt.Sprintf("cannot open/create AEM output file properly at path '%s': %s", c.outputFile, err))
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

	if c.outputFormat == fmtx.None {
		c.printCommandResult()
	} else if c.outputFormat == fmtx.Text {
		if len(c.outputValue) > 0 {
			c.printOutputValue()
		} else {
			c.printOutputText()
			c.printCommandResult()
		}
	} else {
		c.printOutputMarshaled()
	}

	if c.outputResponse.Failed {
		os.Exit(1)
	}
	os.Exit(0)
}

func (c *CLI) printCommandResult() {
	fmt.Print(fmtx.TblList("command result", [][]any{
		{"message", c.outputResponse.Msg},
		{"changed", c.outputResponse.Changed},
		{"failed", c.outputResponse.Failed},
		{"elapsed", c.outputResponse.Elapsed},
		{"ended", timex.Human(c.outputResponse.Ended)},
	}))
}

// TODO allow to print 'changed', 'failed', 'elapsed', 'ended' as well
func (c *CLI) printOutputValue() {
	value, ok := c.outputResponse.Data[c.outputValue]
	if !ok {
		println("<undefined>")
	} else {
		println(fmtx.MarshalText(value))
	}
}

func (c *CLI) printOutputText() {
	if c.outputResponse.Data != nil {
		c.printOutputTextIndented(textio.NewPrefixWriter(os.Stdout, ""), c.outputResponse.Data)
	}
}

func (c *CLI) printOutputTextIndented(writer *textio.PrefixWriter, value any) {
	rv := reflect.ValueOf(value)
	switch rv.Type().Kind() {
	case reflect.Slice, reflect.Array:
		if rv.Len() == 0 {
			c.printOutputTextIndented(writer, "<empty>")
		} else {
			for i := 0; i < rv.Len(); i++ {
				iv := rv.Index(i).Interface()
				c.printOutputTextIndented(writer, iv)
			}
		}
	case reflect.Map:
		if rv.Len() == 0 {
			c.printOutputTextIndented(writer, "<empty>")
		} else {
			dw := textio.NewPrefixWriter(writer, "  ")
			keys := rv.MapKeys()
			sort.SliceStable(keys, func(k1, k2 int) bool {
				return strings.Compare(fmt.Sprintf("%v", keys[k1].Interface()), fmt.Sprintf("%v", keys[k2].Interface())) < 0
			})
			for _, k := range keys {
				_, _ = writer.WriteString(fmt.Sprintf("%s\n", k)) // TODO camelCase to human
				mv := rv.MapIndex(k).Interface()
				c.printOutputTextIndented(dw, mv)
			}
		}
	default:
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
	} else if file == cfg.InputStdin {
		err := fmtx.UnmarshalDataInFormat(format, bufio.NewReader(os.Stdin), out)
		if err != nil {
			return fmt.Errorf("cannot parse STDIN input properly: %w", err)
		}
	} else {
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

// TODO recursively update map keys (or align)
func (c *CLI) fixOutputName(name string) string {
	if c.outputFormat == fmtx.YML {
		name = stringsx.SnakeCase(name)
	}
	return name
}
