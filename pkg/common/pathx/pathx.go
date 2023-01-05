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
		log.Fatalf("cannot determine current working directory: %s", err)
	}
	return path
}

func Abs(path string) string {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("cannot determine absolute path for '%s': %s", path, err)
	}
	return path
}

func Exists(path string) bool {
	exists, err := ExistsStrict(path)
	return err == nil && exists
}

func ExistsStrict(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("cannot check path existence '%s': %w", path, err)
}

func Delete(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("cannot delete path '%s': %w", path, err)
	}
	return nil
}

func DeleteIfExists(path string) error {
	exists, err := ExistsStrict(path)
	if err != nil {
		return err
	}
	if exists {
		return Delete(path)
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
