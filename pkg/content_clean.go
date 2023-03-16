package pkg

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wttech/aemc/pkg/cfg"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	JcrMixinTypesProp  = "jcr:mixinTypes"
	JcrRootPrefix      = "<jcr:root"
	ContentPropPattern = "^([^ =]+)=\"([^\"]+)\"$"
	NamespacePattern   = "^\\w+:(\\w+)=\"[^\"]+\"$"
	JcrRoot            = "jcr_root"
)

func filesDotContent() {} // TODO
func filesDeleted()    {} // TODO
func filesFlattened()  {} // TODO

func lineProcess(path string, line string, config *cfg.ConfigValues) string {
	return normalizeLine(path, line, config)
}

func contentProcess(lines []string, config *cfg.ConfigValues) []string {
	return normalizeContent(lines, config)
}

func prepare(root string, config *cfg.ConfigValues) {
	if config.Content.ParentsBackupEnabled {
		doParentsBackup(root)
	}
}

func beforeClean(root string, config *cfg.ConfigValues) {
	if config.Content.ParentsBackupEnabled {
		doRootBackup(root)
	}
}

func Clean(root string, config *cfg.ConfigValues) error {
	err := flattenFiles(root, config)
	if err == nil {
		if config.Content.ParentsBackupEnabled {
			err = undoParentsBackup(root)
		} else {
			err = cleanParents(root, config)
		}
	}
	if err == nil {
		err = cleanDotContents(root, config)
	}
	if err == nil {
		err = deleteFiles(root, config)
	}
	if err == nil {
		err = deleteBackupFiles(root)
	}
	if err == nil {
		err = deleteEmptyDirs(root)
	}
	return err
}

func eachFilesInDir(root string, processFile func(string, bool) error) error {
	files, err := ioutil.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range files {
		err = processFile(filepath.Join(root, entry.Name()), entry.IsDir())
		if err != nil {
			return err
		}
	}
	return nil
}

func eachFiles(root string, filter func(string) bool, action func(string) error) error {
	return eachFilesInDir(root, func(path string, isDir bool) error {
		var err error
		if isDir {
			err = eachFiles(path, filter, action)
		} else if filter(path) {
			err = action(path)
		}
		return err
	})
}

func cleanDotContents(root string, config *cfg.ConfigValues) error {
	return eachFiles(root, func(path string) bool {
		return matchString(config.Content.FilesDotContent, path)
	}, func(path string) error {
		return cleanDotContentFile(path, config)
	})
}

func cleanDotContentFile(path string, config *cfg.ConfigValues) error {
	inputLines, err := readLines(path)
	var outputLines []string
	if err == nil {
		outputLines = filterLines(path, inputLines, config)
	}
	if err == nil {
		err = writeLines(path, outputLines)
	}
	return err
}

func filterLines(path string, lines []string, config *cfg.ConfigValues) []string {
	var result []string
	for _, line := range lines {
		processedLine := lineProcess(path, line, config)
		if len(processedLine) == 0 {
			if strings.HasSuffix(result[len(result)-1], ">") {
				// skip line
			} else if strings.HasSuffix(strings.TrimSpace(line), "/>") {
				result = append(result, result[len(result)-1]+"/>")
				result = result[:len(result)-2]
			} else if strings.HasSuffix(strings.TrimSpace(line), ">") {
				result = append(result, result[len(result)-1]+">")
				result = result[:len(result)-2]
			} else {
				// skip line
			}
		} else {
			result = append(result, processedLine)
		}
	}
	return contentProcess(result, config)
}

func normalizeContent(lines []string, config *cfg.ConfigValues) []string {
	return mergeSinglePropertyLines(cleanNamespaces(lines, config))
}

func mergeSinglePropertyLines(lines []string) []string {
	var result []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		lineTrimmed := strings.TrimSpace(line)
		if strings.HasPrefix(lineTrimmed, JcrRootPrefix) || i == len(lines)-1 {
			result = append(result, line)
		} else if strings.HasPrefix(lineTrimmed, "<") && !strings.HasSuffix(lineTrimmed, ">") {
			i++
			nextLine := lines[i]
			nextLineTrimmed := strings.TrimSpace(nextLine)
			if !strings.HasPrefix(nextLineTrimmed, "<") && strings.HasSuffix(nextLineTrimmed, ">") {
				result = append(result, fmt.Sprintf("%s %s", line, nextLineTrimmed))
			} else {
				result = append(result, line)
				result = append(result, nextLine)
			}
		} else {
			result = append(result, line)
		}
	}
	return result
}

func cleanNamespaces(lines []string, config *cfg.ConfigValues) []string {
	if !config.Content.NamespacesSkipped {
		return lines
	}

	var result []string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), JcrRootPrefix) {
			var rootResult []string
			parts := strings.Split(line, " ")
			for _, part := range parts {
				groups := regexp.MustCompile(NamespacePattern).FindStringSubmatch(part)
				if groups == nil {
					rootResult = append(rootResult, part)
				} else {
					flag := false
					for i := 0; i < len(lines) && !flag; i++ {
						flag = strings.Contains(lines[i], groups[1]+":")
					}
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

func normalizeLine(path string, line string, config *cfg.ConfigValues) string {
	return normalizeMixins(path, skipProperties(path, line, config), config)
}

func skipProperties(path string, line string, config *cfg.ConfigValues) string {
	return eachProp(line, func(propOccurrence string, _ string) string {
		if matchAnyRule(propOccurrence, path, config.Content.PropertiesSkipped) {
			return ""
		} else {
			return line
		}
	})
}

func normalizeMixins(path string, line string, config *cfg.ConfigValues) string {
	return eachProp(line, func(propName string, propValue string) string {
		if propName == JcrMixinTypesProp {
			normalizedValue, _ := strings.CutPrefix(propValue, "[")
			normalizedValue, _ = strings.CutSuffix(normalizedValue, "]")
			var resultValues []string
			for _, value := range strings.Split(normalizedValue, ",") {
				if !matchAnyRule(value, path, config.Content.MixinTypesSkipped) {
					resultValues = append(resultValues, value)
				}
			}
			if len(resultValues) == 0 || len(normalizedValue) == 0 {
				return ""
			} else {
				resultValue := strings.Join(resultValues, ",")
				return strings.ReplaceAll(line, normalizedValue, resultValue)
			}
		} else {
			return line
		}
	})
}

func eachProp(line string, processProp func(string, string) string) string {
	normalizedLine, _ := strings.CutSuffix(strings.TrimSpace(line), "/>")
	normalizedLine, _ = strings.CutSuffix(normalizedLine, ">")
	groups := regexp.MustCompile(ContentPropPattern).FindStringSubmatch(normalizedLine)
	if groups == nil {
		return line
	} else {
		return processProp(groups[1], groups[2])
	}
}

func flattenFiles(root string, config *cfg.ConfigValues) error {
	return eachFiles(root, func(path string) bool {
		return matchString(config.Content.FilesFlattened, path)
	}, flattenFile)
}

func flattenFile(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) || info.IsDir() {
		return err
	}
	newpath := filepath.Dir(path) + ".xml"
	_, err = os.Stat(newpath)
	if os.IsExist(err) {
		log.Infof("Overriding file by flattening %s", path)
	} else {
		log.Infof("Flattening file %s", path)
	}
	return os.Rename(path, newpath)
}

func deleteFiles(root string, config *cfg.ConfigValues) error {
	err := eachParentFiles(root, func(parent string) error {
		if matchAnyRule(parent, parent, config.Content.FilesDeleted) {
			return deleteFile(parent)
		}
		return nil
	})
	if err == nil {
		err = eachFiles(root, func(path string) bool {
			return matchAnyRule(path, path, config.Content.FilesDeleted)
		}, deleteFile)
	}
	return err
}

func deleteBackupFiles(root string) error { return nil } // TODO

func deleteFile(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return err
	}
	log.Printf("Deleting file %s", path)
	return os.Remove(path)
}

func deleteEmptyDirs(root string) error   { return nil } // TODO
func doParentsBackup(root string)         {}             // TODO
func doRootBackup(root string)            {}             // TODO
func undoParentsBackup(root string) error { return nil } // TODO

func cleanParents(root string, config *cfg.ConfigValues) error {
	return eachParentFiles(root, func(parent string) error {
		return eachFilesInDir(parent, func(path string, isDir bool) error {
			if !isDir {
				err := deleteFile(path)
				if err == nil {
					err = cleanDotContentFile(path, config)
				}
				return err
			}
			return nil
		})
	})
}

func eachParentFiles(root string, processFile func(string) error) error {
	parent := filepath.Dir(root)
	for parent != "" {
		err := processFile(parent)
		if err == nil {
			return err
		}
		if filepath.Base(parent) == JcrRoot {
			break
		}
		parent = filepath.Dir(parent)
	}
	return nil
}

func matchAnyRule(s string, path string, rules []cfg.PathRule) bool {
	for _, rule := range rules {
		if matchRule(s, path, rule) {
			return true
		}
	}
	return false
}

func matchRule(s string, path string, rule cfg.PathRule) bool {
	return matchString(rule.Patterns, s) && !matchString(rule.ExcludedPaths, path) && (len(rule.IncludedPaths) == 0 || matchString(rule.IncludedPaths, path))
}

func matchString(patterns []string, s string) bool {
	if len(patterns) == 0 {
		return false
	}
	for _, pattern := range patterns {
		result, _ := regexp.MatchString(pattern, s)
		if result {
			return true
		}
	}
	return false
}

func readLines(path string) ([]string, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	err = sc.Err()
	return lines, err
}

func writeLines(path string, lines []string) error {
	return nil
}
