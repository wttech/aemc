package content

import (
	"bufio"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	JcrContentFile      = ".content.xml"
	JcrMixinTypesProp   = "jcr:mixinTypes"
	JcrRootPrefix       = "<jcr:root"
	PropPattern         = "^\\s*([^ =]+)=\"([^\"]+)\"(.*)$"
	NamespacePattern    = "^\\w+:(\\w+)=\"[^\"]+\"$"
	JcrRoot             = "jcr_root"
	ParentsBackupSuffix = ".bak"
)

type Cleaner struct {
	config *Opts
}

func NewCleaner(config *Opts) *Cleaner {
	return &Cleaner{
		config: config,
	}
}

func (c Cleaner) prepare(root string) error {
	if c.config.ParentsBackupEnabled {
		return c.doParentsBackup(root)
	}
	return nil
}

func (c Cleaner) BeforeClean(root string) error {
	if c.config.ParentsBackupEnabled {
		return c.doRootBackup(root)
	}
	return nil
}

func (c Cleaner) Clean(root string) error {
	if err := c.flattenFiles(root); err != nil {
		return err
	}
	if c.config.ParentsBackupEnabled {
		if err := c.undoParentsBackup(root); err != nil {
			return err
		}
	} else {
		if err := c.cleanParents(root); err != nil {
			return err
		}
	}
	if err := c.cleanDotContents(root); err != nil {
		return err
	}
	if err := c.deleteFiles(root); err != nil {
		return err
	}
	if err := c.deleteBackupFiles(root); err != nil {
		return err
	}
	if err := deleteEmptyDirs(root); err != nil {
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

func (c Cleaner) cleanDotContents(root string) error {
	return eachFiles(root, func(path string) error {
		return c.cleanDotContentFile(path)
	})
}

func (c Cleaner) cleanDotContentFile(path string) error {
	if !strings.HasSuffix(path, JcrContentFile) {
		return nil
	}

	log.Printf("Cleaning file %s", path)
	inputLines, err := readLines(path)
	if err != nil {
		return err
	}
	outputLines := c.filterLines(path, inputLines)
	return writeLines(path, outputLines)
}

func (c Cleaner) filterLines(path string, lines []string) []string {
	var result []string
	for _, line := range lines {
		flag, processedLine := c.lineProcess(path, line)
		if flag {
			result[len(result)-1] += processedLine
		} else {
			result = append(result, processedLine)
		}
		if len(result) > 2 && strings.HasSuffix(processedLine, ">") &&
			!strings.HasPrefix(result[len(result)-2], JcrRootPrefix) &&
			strings.HasPrefix(strings.TrimSpace(result[len(result)-2]), "<") &&
			!strings.HasSuffix(result[len(result)-2], ">") &&
			!strings.HasPrefix(strings.TrimSpace(result[len(result)-1]), "<") {
			result[len(result)-2] += " " + strings.TrimSpace(result[len(result)-1])
			result = result[:len(result)-1]
		}
	}
	return c.cleanNamespaces(result)
}

func (c Cleaner) cleanNamespaces(lines []string) []string {
	if !c.config.NamespacesSkipped {
		return lines
	}

	var result []string
	for _, line := range lines {
		if strings.HasPrefix(line, JcrRootPrefix) {
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

func (c Cleaner) lineProcess(path string, line string) (bool, string) {
	groups := stringsx.MatchGroups(line, PropPattern)
	if groups == nil {
		return false, line
	} else if groups[1] == JcrMixinTypesProp {
		return c.normalizeMixins(path, line, groups[2], groups[3])
	} else if matchAnyRule(groups[1], path, c.config.PropertiesSkipped) {
		return true, groups[3]
	}
	return false, line
}

func (c Cleaner) normalizeMixins(path string, line string, propValue string, lineSuffix string) (bool, string) {
	normalizedValue := strings.Trim(propValue, "[]")
	var resultValues []string
	for _, value := range strings.Split(normalizedValue, ",") {
		if !matchAnyRule(value, path, c.config.MixinTypesSkipped) {
			resultValues = append(resultValues, value)
		}
	}
	if len(resultValues) == 0 {
		return true, lineSuffix
	}
	return false, strings.ReplaceAll(line, normalizedValue, strings.Join(resultValues, ","))
}

func (c Cleaner) flattenFiles(root string) error {
	return eachFiles(root, func(path string) error {
		return c.flattenFile(path)
	})
}

func (c Cleaner) flattenFile(path string) error {
	if !matchString(path, c.config.FilesFlattened) {
		return nil
	}

	dest := filepath.Dir(path) + ".xml"
	_, err := os.Stat(dest)
	if os.IsExist(err) {
		log.Printf("Overriding file by flattening %s", path)
	} else {
		log.Printf("Flattening file %s", path)
	}
	return os.Rename(path, dest)
}

func (c Cleaner) deleteFiles(root string) error {
	if err := eachParentFiles(root, func(parent string) error {
		return deleteFile(parent, func() bool {
			return matchAnyRule(parent, parent, c.config.FilesDeleted)
		})
	}); err != nil {
		return err
	}
	if err := eachFiles(root, func(path string) error {
		return deleteFile(path, func() bool {
			return matchAnyRule(path, path, c.config.FilesDeleted)
		})
	}); err != nil {
		return err
	}
	return nil
}

func (c Cleaner) deleteBackupFiles(root string) error {
	patterns := []string{".*" + ParentsBackupSuffix}
	return eachFiles(root, func(path string) error {
		return deleteFile(path, func() bool {
			return matchString(path, patterns)
		})
	})
}

func deleteFile(path string, allowedFunc func() bool) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) || allowedFunc != nil && !allowedFunc() {
		return nil
	}
	log.Printf("Deleting file %s", path)
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
		log.Printf("Deleting empty directory %s", root)

	}
	return nil
}

func (c Cleaner) doParentsBackup(root string) error {
	return eachParentFiles(root, func(parent string) error {
		return eachFilesInDir(parent, func(path string) error {
			if !strings.HasSuffix(path, ParentsBackupSuffix) {
				if err := c.backupFile(path, "Doing backup of parent file: %s"); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

func (c Cleaner) doRootBackup(root string) error {
	entry, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !entry.IsDir() {
		if err = c.backupFile(root, "Doing backup of root file: %s"); err != nil {
			return err
		}
	}
	return eachFiles(root, func(path string) error {
		if matchString(path, c.config.FilesFlattened) {
			if err = c.backupFile(path, "Doing backup of file: %s"); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c Cleaner) undoParentsBackup(root string) error {
	return eachFilesInDir(root, func(path string) error {
		if strings.HasSuffix(path, ParentsBackupSuffix) {
			origin := strings.TrimSuffix(path, ParentsBackupSuffix)
			log.Printf("Undoing backup of parent file: %s", path)
			return os.Rename(path, origin)
		}
		return nil
	})
}

func (c Cleaner) cleanParents(root string) error {
	return eachParentFiles(root, func(parent string) error {
		return eachFilesInDir(parent, func(path string) error {
			if err := deleteFile(path, nil); err != nil {
				return err
			}
			if err := c.cleanDotContentFile(path); err != nil {
				return err
			}
			return nil
		})
	})
}

func eachParentFiles(root string, processFileFunc func(string) error) error {
	parent := root
	for strings.Contains(parent, JcrRoot) && filepath.Base(parent) != JcrRoot {
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

func (c Cleaner) backupFile(path string, format string) error {
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
	log.Printf(format, path)
	return nil
}
