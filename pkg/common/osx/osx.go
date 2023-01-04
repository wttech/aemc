package osx

import (
	"fmt"
	"github.com/mholt/archiver/v3"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	ShellPath = "/bin/sh"
)

func PathCurrent() string {
	path, err := os.Getwd()
	if err != nil {
		log.Fatalf(fmt.Sprintf("cannot determine current working directory: %s", err))
	}
	return path
}

func PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func PathDelete(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("cannot delete path '%s': %w", path, err)
	}
	return nil
}

func PathEnsure(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("cannot ensure path '%s': %w", path, err)
	}
	return nil
}

func FileWrite(path string, text string) error {
	err := PathEnsure(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("cannot ensure path '%s': %w", path, err)
	}
	err = os.WriteFile(path, []byte(text), 0755)
	if err != nil {
		return fmt.Errorf("cannot write to file '%s': %w", path, err)
	}
	return nil
}

func FileRead(path string) ([]byte, error) {
	if !PathExists(path) {
		return nil, fmt.Errorf("cannot read file as it does not exist at path '%s'", path)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file '%s': %w", path, err)
	}
	return bytes, nil
}

var fileCopyBufferSize = 4 * 1024 // 4 kB <https://stackoverflow.com/a/3034155>

func FileCopy(sourcePath, destinationPath string) error {
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

func FileExt(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

func FileNameWithoutExt(path string) string {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)]
}

func EnvVars() map[string]string {
	result := make(map[string]string)
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			result[e[:i]] = e[i+1:]
		}
	}
	return result
}

func PathAbs(path string) string {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf(fmt.Sprintf("cannot determine absolute path for '%s': %s", path, err))
	}
	return path
}

func ArchiveExtract(sourceFile string, targetDir string) error {
	err := PathEnsure(targetDir)
	if err != nil {
		return err
	}
	err = archiver.Unarchive(sourceFile, targetDir)
	if err != nil {
		return err
	}
	return nil
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

func isLinux() bool {
	return !IsWindows() && !IsDarwin()
}
