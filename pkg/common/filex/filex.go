package filex

import (
	"fmt"
	"github.com/wttech/aemc/pkg/common/pathx"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Write(path string, text string) error {
	err := pathx.Ensure(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("cannot ensure path '%s': %w", path, err)
	}
	err = os.WriteFile(path, []byte(text), 0755)
	if err != nil {
		return fmt.Errorf("cannot write to file '%s': %w", path, err)
	}
	return nil
}

func Read(path string) ([]byte, error) {
	if !pathx.Exists(path) {
		return nil, fmt.Errorf("cannot read file as it does not exist at path '%s'", path)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file '%s': %w", path, err)
	}
	return bytes, nil
}

var fileCopyBufferSize = 4 * 1024 // 4 kB <https://stackoverflow.com/a/3034155>

func Copy(sourcePath, destinationPath string) error {
	sourceStat, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	if !sourceStat.Mode().IsRegular() {
		return fmt.Errorf("cannot copy file from '%s' to '%s' as source does not exist (or is not a regular file)", sourcePath, destinationPath)
	}
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	_, err = os.Stat(destinationPath)
	if err == nil {
		return fmt.Errorf("cannot copy file from '%s' to '%s' as destination already exists", sourcePath, destinationPath)
	}
	destination, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destination.Close()
	buf := make([]byte, fileCopyBufferSize)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return err
}

func Ext(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

func NameWithoutExt(path string) string {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)]
}
