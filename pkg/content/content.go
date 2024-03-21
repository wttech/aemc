package content

import (
	"bufio"
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/wttech/aemc/pkg/base"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	JCRRoot                   = "jcr_root"
	JCRContentFile            = ".content.xml"
	JCRContentFileSuffix      = ".xml"
	JCRMixinTypesProp         = "jcr:mixinTypes"
	JCRRootPrefix             = "<jcr:root"
	PropPattern               = "^\\s*([^ =]+)=\"([^\"]+)\"(.*)$"
	NamespacePattern          = "^\\w+:(\\w+)=\"[^\"]+\"$"
	ParentsBackupSuffix       = ".bak"
	ParentsBackupDirIndicator = ".bakdir"
	JCRContentNode            = "jcr:content"
	JCRContentPrefix          = "<jcr:content"
	JCRContentDirName         = "_jcr_content"
)

type Manager struct {
	baseOpts *base.Opts

	FilesDeleted         []PathRule
	FilesFlattened       []string
	PropertiesSkipped    []PathRule
	MixinTypesSkipped    []PathRule
	NamespacesSkipped    bool
	ParentsBackupEnabled bool
}

func NewManager(baseOpts *base.Opts) *Manager {
	cv := baseOpts.Config().Values()

	return &Manager{
		baseOpts: baseOpts,

		FilesDeleted:         determinePathRules(cv.Get("content.clean.files_deleted")),
		FilesFlattened:       cv.GetStringSlice("content.clean.files_flattened"),
		PropertiesSkipped:    determinePathRules(cv.Get("content.clean.properties_skipped")),
		MixinTypesSkipped:    determinePathRules(cv.Get("content.clean.mixin_types_skipped")),
		NamespacesSkipped:    cv.GetBool("content.clean.namespaces_skipped"),
		ParentsBackupEnabled: cv.GetBool("content.clean.parents_backup_enabled"),
	}
}

func (c Manager) Prepare(root string) error {
	return deleteDir(root)
}

func (c Manager) BeforeClean(root string) error {
	if c.ParentsBackupEnabled {
		return c.doParentsBackup(root)
	}
	return nil
}

func (c Manager) CleanDir(root string) error {
	if err := c.cleanDotContents(root); err != nil {
		return err
	}
	if c.ParentsBackupEnabled {
		if err := c.undoParentsBackup(root); err != nil {
			return err
		}
	} else {
		if err := c.cleanParents(root); err != nil {
			return err
		}
	}
	if err := c.flattenFiles(root); err != nil {
		return err
	}
	if err := c.deleteFiles(root); err != nil {
		return err
	}
	if err := deleteEmptyDirs(root); err != nil {
		return err
	}
	return nil
}

func (c Manager) CleanFile(path string) error {
	if !pathx.Exists(path) {
		return fmt.Errorf("file does not exist: %s", path)
	}
	if err := c.cleanDotContentFile(path); err != nil {
		return err
	}
	if err := c.flattenFile(path); err != nil {
		return err
	}
	if err := deleteEmptyDirs(filepath.Dir(path)); err != nil {
		return err
	}
	return nil
}

func eachFilesInDir(root string, processFileFunc func(path string) error) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			if err = processFileFunc(filepath.Join(root, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func eachFiles(root string, processFileFunc func(string) error) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}
		return processFileFunc(path)
	})
}

func (c Manager) cleanDotContents(root string) error {
	return eachFiles(root, func(path string) error {
		return c.cleanDotContentFile(path)
	})
}

func (c Manager) cleanDotContentFile(path string) error {
	if !strings.HasSuffix(path, JCRContentFileSuffix) {
		return nil
	}

	log.Infof("cleaning file %s", path)
	inputLines, err := readLines(path)
	if err != nil {
		return err
	}
	outputLines := c.filterLines(path, inputLines)
	return writeLines(path, outputLines)
}

func (c Manager) filterLines(path string, lines []string) []string {
	var result []string
	for _, line := range lines {
		flag, processedLine := c.lineProcess(path, line)
		if flag {
			result[len(result)-1] += processedLine
		} else {
			result = append(result, processedLine)
		}
		if len(result) > 2 && strings.HasSuffix(processedLine, ">") &&
			!strings.HasPrefix(result[len(result)-2], JCRRootPrefix) &&
			strings.HasPrefix(strings.TrimSpace(result[len(result)-2]), "<") &&
			!strings.HasSuffix(result[len(result)-2], ">") &&
			!strings.HasPrefix(strings.TrimSpace(result[len(result)-1]), "<") {
			result[len(result)-2] += " " + strings.TrimSpace(result[len(result)-1])
			result = result[:len(result)-1]
		}
	}
	return c.cleanNamespaces(result)
}

func (c Manager) cleanNamespaces(lines []string) []string {
	if !c.NamespacesSkipped {
		return lines
	}

	var result []string
	for _, line := range lines {
		if strings.HasPrefix(line, JCRRootPrefix) {
			var rootResult []string
			for _, part := range strings.Split(line, " ") {
				groups := stringsx.MatchGroups(part, NamespacePattern)
				if groups == nil {
					rootResult = append(rootResult, part)
				} else {
					flag := lo.SomeBy(lines, func(line string) bool {
						return strings.Contains(line, groups[1]+":")
					})
					if flag {
						rootResult = append(rootResult, part)
					}
				}
			}
			result = append(result, strings.Join(rootResult, " "))
		} else {
			result = append(result, line)
		}
	}
	return result
}

func (c Manager) lineProcess(path string, line string) (bool, string) {
	groups := stringsx.MatchGroups(line, PropPattern)
	if groups == nil {
		return false, line
	} else if groups[1] == JCRMixinTypesProp {
		return c.normalizeMixins(path, line, groups[2], groups[3])
	} else if matchAnyRule(groups[1], path, c.PropertiesSkipped) {
		return true, groups[3]
	}
	return false, line
}

func (c Manager) normalizeMixins(path string, line string, propValue string, lineSuffix string) (bool, string) {
	normalizedValue := strings.Trim(propValue, "[]")
	var resultValues []string
	for _, value := range strings.Split(normalizedValue, ",") {
		if !matchAnyRule(value, path, c.MixinTypesSkipped) {
			resultValues = append(resultValues, value)
		}
	}
	if len(resultValues) == 0 {
		return true, lineSuffix
	}
	return false, strings.ReplaceAll(line, normalizedValue, strings.Join(resultValues, ","))
}

func (c Manager) flattenFiles(root string) error {
	return eachFiles(root, func(path string) error {
		return c.flattenFile(path)
	})
}

func (c Manager) flattenFile(path string) error {
	if !matchString(path, c.FilesFlattened) {
		return nil
	}

	dest := filepath.Dir(path) + ".xml"
	if pathx.Exists(dest) {
		log.Infof("flattening file (override): %s", path)
	} else {
		log.Infof("flattening file: %s", path)
	}
	return os.Rename(path, dest)
}

func (c Manager) deleteFiles(root string) error {
	if err := eachParentFiles(root, func(parent string) error {
		return eachFilesInDir(parent, func(path string) error {
			return deleteFile(path, func() bool {
				return matchAnyRule(path, path, c.FilesDeleted)
			})
		})
	}); err != nil {
		return err
	}
	if err := eachFiles(root, func(path string) error {
		return deleteFile(path, func() bool {
			return matchAnyRule(path, path, c.FilesDeleted)
		})
	}); err != nil {
		return err
	}
	return nil
}

func deleteDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil
	}
	log.Infof("deleting dir %s", dir)
	return os.Remove(dir)
}

func deleteFile(path string, allowedFunc func() bool) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) || allowedFunc != nil && !allowedFunc() {
		return nil
	}
	log.Infof("deleting file %s", path)
	return os.Remove(path)
}

func deleteEmptyDirs(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if err = deleteEmptyDirs(filepath.Join(root, entry.Name())); err != nil {
				return err
			}
		}
	}
	entries, err = os.ReadDir(root)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		if err = os.Remove(root); err != nil {
			return err
		}
		log.Infof("deleting empty directory %s", root)

	}
	return nil
}

func (c Manager) doParentsBackup(root string) error {
	return eachParentFiles(root, func(parent string) error {
		return eachFilesInDir(parent, func(path string) error {
			if !strings.HasSuffix(path, ParentsBackupSuffix) {
				if err := c.backupFile(path, "doing backup of parent file: %s"); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

func (c Manager) undoParentsBackup(root string) error {
	return eachParentFiles(root, func(parent string) error {
		indicator := false
		if err := eachFilesInDir(parent, func(path string) error {
			indicator = indicator || strings.HasSuffix(path, ParentsBackupDirIndicator)
			return nil
		}); err != nil {
			return err
		}
		if !indicator {
			return nil
		}

		if err := eachFilesInDir(parent, func(path string) error {
			return deleteFile(path, func() bool {
				return !strings.HasSuffix(path, ParentsBackupSuffix) || !strings.HasSuffix(path, ParentsBackupDirIndicator)
			})
		}); err != nil {
			return err
		}

		return eachFilesInDir(parent, func(path string) error {
			if strings.HasSuffix(path, ParentsBackupSuffix) {
				origin := strings.TrimSuffix(path, ParentsBackupSuffix)
				log.Infof("undoing backup of parent file: %s", path)
				return os.Rename(path, origin)
			}
			return nil
		})
	})
}

func (c Manager) cleanParents(root string) error {
	return eachParentFiles(root, func(parent string) error {
		return eachFilesInDir(parent, func(path string) error {
			return c.cleanDotContentFile(path)
		})
	})
}

func eachParentFiles(root string, processFileFunc func(string) error) error {
	parent := root
	for strings.Contains(parent, JCRRoot) && filepath.Base(parent) != JCRRoot {
		parent = filepath.Dir(parent)
		if err := processFileFunc(parent); err != nil {
			return err
		}
	}
	return nil
}

func matchAnyRule(value string, path string, rules []PathRule) bool {
	return lo.SomeBy(rules, func(rule PathRule) bool {
		return matchRule(value, path, rule)
	})
}

func matchRule(value string, path string, rule PathRule) bool {
	return matchString(value, rule.Patterns) && !matchString(path, rule.ExcludedPaths) && (len(rule.IncludedPaths) == 0 || matchString(path, rule.IncludedPaths))
}

func matchString(value string, patterns []string) bool {
	return stringsx.MatchSome(value, patterns)
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	content := strings.Join(lines, "\n") + "\n"
	_, err = file.WriteString(content)
	return err
}

func (c Manager) backupFile(path string, format string) error {
	dir := filepath.Dir(path)
	indicator, err := os.Create(filepath.Join(dir, ParentsBackupDirIndicator))
	if err != nil {
		return err
	}
	defer func() { _ = indicator.Close() }()

	source, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()

	destination, err := os.Create(path + ParentsBackupSuffix)
	if err != nil {
		return err
	}
	defer func() { _ = destination.Close() }()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}
	log.Infof(format, path)
	return nil
}

type PathRule struct {
	Patterns      []string
	ExcludedPaths []string
	IncludedPaths []string
}

func determinePathRules(values any) []PathRule {
	var result []PathRule
	for _, value := range cast.ToSlice(values) {
		result = append(result, PathRule{
			Patterns:      determineStringSlice(value, "patterns"),
			ExcludedPaths: determineStringSlice(value, "excluded_paths"),
			IncludedPaths: determineStringSlice(value, "included_paths"),
		})
	}
	return result
}

func determineStringSlice(values any, key string) []string {
	var result []string
	value := cast.ToStringMap(values)[key]
	if value != nil {
		result = cast.ToStringSlice(value)
	}
	return result
}

func IsContentFile(path string) bool {
	if !strings.HasSuffix(path, JCRContentFile) {
		return false
	}

	inputLines, err := readLines(path)
	if err != nil {
		return false
	}

	for _, inputLine := range inputLines {
		if strings.Contains(inputLine, JCRContentPrefix) {
			return true
		}
	}
	return false
}
