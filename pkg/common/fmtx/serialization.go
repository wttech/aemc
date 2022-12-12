package fmtx

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

const (
	// None skips printing output of the commands
	None string = "none"

	// Text prints output of the commands as human-readable text
	Text string = "text"

	// YML prints output of the commands in YML format
	YML string = "yml"

	// JSON prints output of the commands in JSON format
	JSON string = "json"
)

func UnmarshalDataInFormat(dataFormat string, reader io.Reader, out any) error {
	switch dataFormat {
	case YML:
		return UnmarshalYAML(reader, out)
	case JSON:
		return UnmarshalJSON(reader, out)
	default:
		return fmt.Errorf("cannot decode data to struct; unsupported data format '%s'", dataFormat)
	}
}

func MarshalJSON(i any) (string, error) {
	bytes, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", fmt.Errorf("cannot convert object '%s' to JSON: %w", i, err)
	}
	return string(bytes), nil
}

func UnmarshalJSON(body io.Reader, out any) error {
	err := json.NewDecoder(body).Decode(out)
	if err != nil {
		return fmt.Errorf("cannot decode stream as JSON: %w", err)
	}
	return nil
}

func MarshalYAML(i any) (string, error) {
	bytes, err := yaml.Marshal(i)
	if err != nil {
		return "", fmt.Errorf("cannot convert object '%s' to YML: %w", i, err)
	}
	return string(bytes), nil
}

func UnmarshalYAML(body io.Reader, out any) error {
	err := yaml.NewDecoder(body).Decode(out)
	if err != nil {
		return fmt.Errorf("cannot decode stream as YML: %w", err)
	}
	return nil
}

func UnmarshalFileInFormat(format string, path string, out any) error {
	fileDesc, _ := os.Open(path)
	defer fileDesc.Close()
	err := UnmarshalDataInFormat(format, fileDesc, out)
	if err != nil {
		return fmt.Errorf("cannot parse file '%s' to struct: %w", path, err)
	}
	return nil
}

type TextMarshaler interface {
	MarshalText() string
}

func MarshalText(value any) string {
	marshaller, ok := value.(TextMarshaler)
	if ok {
		return marshaller.MarshalText()
	}
	return fmt.Sprintf("%v", value)
}
