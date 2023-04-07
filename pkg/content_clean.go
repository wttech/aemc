package pkg

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	JcrContentFile     = ".content.xml"
	JcrMixinTypesProp  = "jcr:mixinTypes"
	JcrRootPrefix      = "<jcr:root"
	ContentPropPattern = "^\\s*([^ =]+)=\"([^\"]+)\"(.*)$"
	NamespacePattern   = "^\\w+:(\\w+)=\"[^\"]+\"$"
	JcrRoot            = "jcr_root"
)

func prepare(root string, config *Content) error {
	if config.ParentsBackupEnabled {
		return doParentsBackup(root, config)
	}
	return nil
}

func BeforeClean(root string, config *Content) error {
	if config.ParentsBackupEnabled {
		return doRootBackup(root, config)
	}
	return nil
}

func Clean(root string, config *Content) error {
	err := flattenFiles(root, config)
	if err == nil {
		if config.ParentsBackupEnabled {
			err = undoParentsBackup(root, config)
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
		err = deleteBackupFiles(root, config)
	}
	if err == nil {
		err = deleteEmptyDirs(root)
	}
	return err
}

func eachFilesInDir(root string, processFileFunc func(path string) error) error {
	infos, err := ioutil.ReadDir(root)
	for i := 0; i < len(infos) && err == nil; i++ {
		if !infos[i].IsDir() {
			err = processFileFunc(filepath.ToSlash(filepath.Join(root, infos[i].Name())))
		}
	}
	return err
}

func eachFiles(root string, processFileFunc func(string) error) error {
	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			err = processFileFunc(filepath.ToSlash(path))
		}
		return err
	})
}

func cleanDotContents(root string, config *Content) error {
	return eachFiles(root, func(path string) error {
		return cleanDotContentFile(path, config)
	})
}

func cleanDotContentFile(path string, config *Content) error {
	if !matchString(path, config.FilesDotContent) {
		return nil
	}

	log.Printf("Cleaning file %s", path)
	inputLines, err := readLines(path)
	if err == nil {
		outputLines := filterLines(path, inputLines, config)
		err = writeLines(path, outputLines)
	}
	return err
}

func filterLines(path string, lines []string, config *Content) []string {
	var result []string
	for _, line := range lines {
		flag, processedLine := lineProcess(path, line, config)
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
	return cleanNamespaces(result, config)
}

func cleanNamespaces(lines []string, config *Content) []string {
	if !config.NamespacesSkipped {
		return lines
	}

	var result []string
	for _, line := range lines {
		if strings.HasPrefix(line, JcrRootPrefix) {
			var rootResult []string
			for _, part := range strings.Split(line, " ") {
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

func lineProcess(path string, line string, config *Content) (bool, string) {
	groups := regexp.MustCompile(ContentPropPattern).FindStringSubmatch(line)
	if groups == nil {
		return false, line
	} else if groups[1] == JcrMixinTypesProp {
		return normalizeMixins(path, line, groups[2], groups[3], config)
	} else if matchAnyRule(groups[1], path, config.PropertiesSkipped) {
		return true, groups[3]
	} else {
		return false, line
	}
}

func normalizeMixins(path string, line string, propValue string, lineSuffix string, config *Content) (bool, string) {
	normalizedValue := strings.Trim(propValue, "[]")
	var resultValues []string
	for _, value := range strings.Split(normalizedValue, ",") {
		if !matchAnyRule(value, path, config.MixinTypesSkipped) {
			resultValues = append(resultValues, value)
		}
	}
	if len(resultValues) == 0 {
		return true, lineSuffix
	} else {
		return false, strings.ReplaceAll(line, normalizedValue, strings.Join(resultValues, ","))
	}
}

func flattenFiles(root string, config *Content) error {
	return eachFiles(root, func(path string) error {
		return flattenFile(path, config)
	})
}

func flattenFile(path string, config *Content) error {
	if !matchString(path, config.FilesFlattened) {
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

func deleteFiles(root string, config *Content) error {
	err := eachParentFiles(root, func(parent string) error {
		return deleteFile(parent, func() bool {
			return matchAnyRule(parent, parent, config.FilesDeleted)
		})
	})
	if err == nil {
		err = eachFiles(root, func(path string) error {
			return deleteFile(path, func() bool {
				return matchAnyRule(path, path, config.FilesDeleted)
			})
		})
	}
	return err
}

func deleteBackupFiles(root string, config *Content) error {
	patterns := []string{".*" + config.ParentsBackupSuffix}
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
	infos, err := ioutil.ReadDir(root)
	for i := 0; i < len(infos) && err == nil; i++ {
		if infos[i].IsDir() {
			err = deleteEmptyDirs(filepath.ToSlash(filepath.Join(root, infos[i].Name())))
		}
	}
	if err == nil {
		infos, err = ioutil.ReadDir(root)
		if err == nil && len(infos) == 0 {
			log.Printf("Deleting empty directory %s", root)
			err = os.Remove(root)
		}
	}
	return err
}

func doParentsBackup(root string, config *Content) error {
	return eachParentFiles(root, func(parent string) error {
		return eachFilesInDir(parent, func(path string) error {
			if !strings.HasSuffix(path, config.ParentsBackupSuffix) {
				if err := backupFile(path, config, "Doing backup of parent file: %s"); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

func doRootBackup(root string, config *Content) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		if err = backupFile(root, config, "Doing backup of root file: %s"); err != nil {
			return err
		}
	}
	return eachFiles(root, func(path string) error {
		if matchString(path, config.FilesFlattened) {
			if err = backupFile(path, config, "Doing backup of file: %s"); err != nil {
				return err
			}
		}
		return nil
	})
}

func undoParentsBackup(root string, config *Content) error {
	return eachFilesInDir(root, func(path string) error {
		if strings.HasSuffix(path, config.ParentsBackupSuffix) {
			origin, _ := strings.CutSuffix(path, config.ParentsBackupSuffix)
			log.Printf("Undoing backup of parent file: %s", path)
			return os.Rename(path, origin)
		}
		return nil
	})
}

func cleanParents(root string, config *Content) error {
	return eachParentFiles(root, func(parent string) error {
		return eachFilesInDir(parent, func(path string) error {
			err := deleteFile(path, nil)
			if err == nil {
				err = cleanDotContentFile(path, config)
			}
			return err
		})
	})
}

func eachParentFiles(root string, processFileFunc func(string) error) error {
	parent := filepath.Dir(root)
	for parent != "" {
		if err := processFileFunc(filepath.ToSlash(parent)); err != nil {
			return err
		}
		if filepath.Base(parent) == JcrRoot {
			break
		}
		parent = filepath.Dir(parent)
	}
	return nil
}

func matchAnyRule(value string, path string, rules []PathRule) bool {
	result := false
	for i := 0; i < len(rules) && !result; i++ {
		result = matchRule(value, path, rules[i])
	}
	return result
}

func matchRule(value string, path string, rule PathRule) bool {
	return matchString(value, rule.Patterns) && !matchString(path, rule.ExcludedPaths) && (len(rule.IncludedPaths) == 0 || matchString(path, rule.IncludedPaths))
}

func matchString(value string, patterns []string) bool {
	result := false
	for i := 0; i < len(patterns) && !result; i++ {
		result, _ = regexp.MatchString(patterns[i], value)
	}
	return result
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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
	defer file.Close()

	for i := 0; i < len(lines) && err == nil; i++ {
		_, err = file.WriteString(lines[i] + "\n")
	}
	return nil
}

func backupFile(path string, config *Content, format string) error {
	source, err := os.Open(path)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(path + config.ParentsBackupSuffix)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err == nil {
		log.Printf(format, path)
	}
	return err
}
