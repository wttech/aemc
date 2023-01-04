package pathx

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func Current() string {
	path, err := os.Getwd()
	if err != nil {
		log.Fatalf(fmt.Sprintf("cannot determine current working directory: %s", err))
	}
	return path
}

func Abs(path string) string {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf(fmt.Sprintf("cannot determine absolute path for '%s': %s", path, err))
	}
	return path
}

func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func Delete(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("cannot delete path '%s': %w", path, err)
	}
	return nil
}

func Ensure(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("cannot ensure path '%s': %w", path, err)
	}
	return nil
}
