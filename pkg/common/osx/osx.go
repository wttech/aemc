package osx

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
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

func EnvVars() map[string]string {
	result := make(map[string]string)
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			result[e[:i]] = e[i+1:]
		}
	}
	return result
}
