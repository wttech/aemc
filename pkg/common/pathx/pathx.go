package pathx

import (
	"fmt"
	"github.com/gobwas/glob"
	ignore "github.com/sabhiram/go-gitignore"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

func IsDir(path string) bool {
	exists, err := IsDirStrict(path)
	return err == nil && exists
}

func IsDirStrict(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("cannot check if path is a dir '%s': %w", path, err)
	}
	return stat.IsDir(), nil
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

func Ext(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

func NameWithoutExt(path string) string {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)]
}

func GlobOne(pathPattern string) (string, error) {
	dir := stringsx.BeforeLast(pathPattern, "/")
	pattern := stringsx.AfterLast(pathPattern, "/")
	paths, err := GlobDir(dir, pattern)
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return "", fmt.Errorf("cannot find any file matching pattern '%s'", pathPattern)
	}
	return paths[len(paths)-1], nil
}

// GlobDir is a modified version of 'go/1.19.2/libexec/src/path/filepath/match.go'
func GlobDir(dir string, pattern string) ([]string, error) {
	m := []string{}
	patternCompiled, err := glob.Compile(pattern)
	if err != nil {
		return m, err
	}
	if !Exists(dir) {
		return m, fmt.Errorf("cannot glob non-existing dir '%s' with pattern '%s'", dir, pattern)
	}
	fi, err := os.Stat(dir)
	if err != nil {
		return m, fmt.Errorf("cannot glob dir '%s' with pattern '%s' : %w", dir, pattern, err)
	}
	if !fi.IsDir() {
		return m, err
	}
	d, err := os.Open(dir)
	if err != nil {
		return m, err
	}
	defer d.Close()
	names, _ := d.Readdirnames(-1)
	sort.Strings(names)
	for _, n := range names {
		matched := patternCompiled.Match(n)
		if matched {
			m = append(m, filepath.Join(dir, n))
		}
	}
	return m, nil
}

type IgnoreMatcher struct {
	matcher *ignore.GitIgnore
}

func NewIgnoreMatcher(patterns []string) IgnoreMatcher {
	return IgnoreMatcher{matcher: ignore.CompileIgnoreLines(patterns...)}
}
func (m *IgnoreMatcher) Match(path string) bool {
	return m.matcher.MatchesPath(path)
}

func Sep() string {
	return string(os.PathSeparator)
}

func Normalize(path string) string {
	return strings.ReplaceAll(strings.ReplaceAll(path, "\\", Sep()), "/", Sep())
}
