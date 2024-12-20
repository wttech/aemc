package content

import (
	"bufio"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	JCRRoot                  = "jcr_root"
	JCRContentFile           = ".content.xml"
	XmlFileSuffix            = ".xml"
	JCRMixinTypesProp        = "jcr:mixinTypes"
	JCRRootPrefix            = "<jcr:root"
	JCRContentNode           = "jcr:content"
	PropPattern              = "^\\s*([^ =]+)=\"([^\"]+)\"(.*)$"
	NamespacePattern         = "^\\w+:(\\w+)=\"[^\"]+\"$"
	FileWithNamespacePattern = "[\\\\/]_([a-zA-Z0-9]+)_[^\\\\/]+([\\\\/]\\.content)?\\.xml$"
)

var (
	propPatternRegex              *regexp.Regexp
	namespacePatternRegex         *regexp.Regexp
	fileWithNamespacePatternRegex *regexp.Regexp
)

func init() {
	propPatternRegex = regexp.MustCompile(PropPattern)
	namespacePatternRegex = regexp.MustCompile(NamespacePattern)
	fileWithNamespacePatternRegex = regexp.MustCompile(FileWithNamespacePattern)
}

type Editor struct {
	FilesDeleted      []PathRule
	FilesFlattened    []string
	PropertiesSkipped []PathRule
	MixinTypesSkipped []PathRule
	NamespacesSkipped bool
}

func NewEditor(config *cfg.Config) *Editor {
	cv := config.Values()

	return &Editor{
		FilesDeleted:      determinePathRules(cv.Get("content.clean.files_deleted")),
		FilesFlattened:    cv.GetStringSlice("content.clean.files_flattened"),
		PropertiesSkipped: determinePathRules(cv.Get("content.clean.properties_skipped")),
		MixinTypesSkipped: determinePathRules(cv.Get("content.clean.mixin_types_skipped")),
		NamespacesSkipped: cv.GetBool("content.clean.namespaces_skipped"),
	}
}

func (c Editor) Clean(path string) error {
	if pathx.IsDir(path) {
		log.Infof("cleaning directory '%s'", path)
		if err := c.cleanDotContents(path); err != nil {
			return err
		}
		if err := c.flattenFiles(path); err != nil {
			return err
		}
		if err := c.deleteFiles(path); err != nil {
			return err
		}
		if err := deleteEmptyDirs(path); err != nil {
			return err
		}
		log.Infof("cleaned directory '%s'", path)
		return nil
	}
	if pathx.IsFile(path) {
		log.Infof("cleaning file '%s'", path)
		if err := c.cleanDotContentFile(path); err != nil {
			return err
		}
		if err := c.flattenFile(path); err != nil {
			return err
		}
		if err := deleteEmptyDirs(filepath.Dir(path)); err != nil {
			return err
		}
		log.Infof("cleaned file '%s'", path)
		return nil
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

func (c Editor) cleanDotContents(root string) error {
	return eachFiles(root, func(path string) error {
		return c.cleanDotContentFile(path)
	})
}

func (c Editor) cleanDotContentFile(path string) error {
	if !strings.HasSuffix(path, XmlFileSuffix) {
		return nil
	}

	log.Infof("cleaning dot content file '%s'", path)
	inputLines, err := readLines(path)
	if err != nil {
		return err
	}
	outputLines := c.filterLines(path, inputLines)
	return writeLines(path, outputLines)
}

func (c Editor) filterLines(path string, lines []string) []string {
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
	return c.cleanNamespaces(path, result)
}

func (c Editor) cleanNamespaces(path string, lines []string) []string {
	if !c.NamespacesSkipped {
		return lines
	}

	var fileNamespace string
	groups := fileWithNamespacePatternRegex.FindStringSubmatch(path)
	if groups != nil {
		fileNamespace = groups[1]
	}

	var result []string
	for _, line := range lines {
		if strings.HasPrefix(line, JCRRootPrefix) {
			var rootResult []string
			for _, part := range strings.Split(line, " ") {
				groups = namespacePatternRegex.FindStringSubmatch(part)
				if groups == nil {
					rootResult = append(rootResult, part)
				} else {
					if lo.SomeBy(lines, func(line string) bool {
						return strings.Contains(line, groups[1]+":") || groups[1] == fileNamespace
					}) {
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

func (c Editor) lineProcess(path string, line string) (bool, string) {
	groups := propPatternRegex.FindStringSubmatch(line)
	if strings.TrimSpace(line) == "" {
		return true, ""
	} else if groups == nil {
		return false, line
	} else if groups[1] == JCRMixinTypesProp {
		return c.normalizeMixins(path, line, groups[2], groups[3])
	} else if matchAnyRule(groups[1], path, c.PropertiesSkipped) {
		return true, groups[3]
	}
	return false, line
}

func (c Editor) normalizeMixins(path string, line string, propValue string, lineSuffix string) (bool, string) {
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

func (c Editor) flattenFiles(root string) error {
	return eachFiles(root, func(path string) error {
		return c.flattenFile(path)
	})
}

func (c Editor) flattenFile(path string) error {
	if !matchString(path, c.FilesFlattened) {
		return nil
	}

	dest := filepath.Dir(path) + ".xml"
	if pathx.Exists(dest) {
		log.Infof("flattening file (override) '%s'", path)
	} else {
		log.Infof("flattening file '%s'", path)
	}
	return os.Rename(path, dest)
}

func (c Editor) deleteFiles(root string) error {
	return eachFiles(root, func(path string) error {
		return c.DeleteFile(path, func() bool {
			return matchAnyRule(path, path, c.FilesDeleted)
		})
	})
}

func (c Editor) DeleteDir(path string) error {
	if !pathx.Exists(path) {
		return nil
	}
	log.Infof("deleting directory '%s'", path)
	return os.RemoveAll(path)
}

func (c Editor) DeleteFile(path string, allowedFunc func() bool) error {
	if !pathx.Exists(path) || allowedFunc != nil && !allowedFunc() {
		return nil
	}
	log.Infof("deleting file '%s'", path)
	return os.Remove(path)
}

func deleteEmptyDirs(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if err = deleteEmptyDirs(filepath.Join(path, entry.Name())); err != nil {
				return err
			}
		}
	}
	entries, err = os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		log.Infof("deleting empty directory '%s'", path)
		if err = os.Remove(path); err != nil {
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
	fileStat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if fileStat.Size() > bufio.MaxScanTokenSize {
		size := fileStat.Size()
		buffer := make([]byte, size)
		scanner.Buffer(buffer, int(size))
	}
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

func IsPageContentFile(path string) bool {
	if pathx.IsDir(path) || !strings.HasSuffix(path, JCRContentFile) {
		return false
	}

	lines, err := readLines(path)
	if err != nil {
		return false
	}

	return lo.SomeBy(lines, func(line string) bool {
		return strings.Contains(line, "cq:PageContent")
	})
}
