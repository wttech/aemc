package pkg

import (
	"archive/zip"
	"fmt"
	"github.com/antchfx/xmlquery"
	"strings"
)

type PID struct {
	Group   string `json:"group" yaml:"group"`
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
}

func (d PID) String() string {
	return fmt.Sprintf("%s:%s:%s", d.Group, d.Name, d.Version)
}

func ParsePID(str string) (*PID, error) {
	parts := strings.Split(str, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("package dependency '%s' has different format than expected 'group:name:version'", str)
	}
	return &PID{parts[0], parts[1], parts[2]}, nil
}

func ReadPIDFromZIP(path string) (*PID, error) {
	zf, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("package '%s' cannot be read: %w", path, err)
	}
	defer zf.Close()
	for _, entry := range zf.File {
		if entry.Name == VltProperties {
			return readPIDFromZipEntry(path, entry)
		}
	}
	return nil, fmt.Errorf("package '%s' has no properties file '%s' required to determine PID", path, VltProperties)
}

func readPIDFromZipEntry(path string, file *zip.File) (*PID, error) {
	fh, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("package '%s' has properties file '%s' that cannot be read: %w", path, VltProperties, err)
	}
	defer fh.Close()
	doc, err := xmlquery.Parse(fh)
	if err != nil {
		return nil, fmt.Errorf("package '%s' has properties file '%s' that cannot be parsed: %w", path, VltProperties, err)
	}
	return &PID{
		doc.SelectElement("//entry[@key=\"group\"]").InnerText(),
		doc.SelectElement("//entry[@key=\"name\"]").InnerText(),
		doc.SelectElement("//entry[@key=\"version\"]").InnerText(),
	}, nil
}
