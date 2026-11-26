package cfg

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/pathx"
	"regexp"
	"strings"
)

// Temporary bridge for config files written for AEM Compose 2.x, which were Go templates with '[[ ]]' delimiters.
// Remove in the next major version together with this file and the 'aem config migrate' command.

var (
	legacyTemplatePattern = regexp.MustCompile(`\[\[.*?\]\]`)
	legacyReplacements    = []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`\[\[\s*\.Env\.(\w+)\s*\|\s*default\s+"([^"]*)"\s*\]\]`), "$${$1:-$2}"},
		{regexp.MustCompile(`\[\[\s*\.Env\.(\w+)\s*\|\s*default\s+([^\s\]]+)\s*\]\]`), "$${$1:-$2}"},
		{regexp.MustCompile(`\[\[\s*\.Env\.(\w+)\s*\]\]`), "$${$1}"},
		{regexp.MustCompile(`\[\[\s*canonicalPath\s+\.Path\s+"([^"]*)"\s*\]\]`), "$${" + VarProjectDir + "}/$1"},
		{regexp.MustCompile(`\[\[\s*\.Path\s*\]\]`), "$${" + VarProjectDir + "}"},
		{regexp.MustCompile(`\[\[\s*\.Os\s*\]\]`), "$${" + VarOs + "}"},
		{regexp.MustCompile(`\[\[\s*\.Arch\s*\]\]`), "$${" + VarArch + "}"},
		{regexp.MustCompile(`\[\[\s*\.ArchiveExt\s*\]\]`), "$${" + VarArchiveExt + "}"},
	}
	legacySettingHints = []struct {
		pattern *regexp.Regexp
		hint    string
	}{
		{regexp.MustCompile(`^\s*backup_dir\s*:`), "setting 'instance.local.backup_dir' is removed, use 'instance.local.backup.dir'"},
		{regexp.MustCompile(`^\s*tool_dir\s*:\s*['"]?aem/home/opt['"]?\s*$`), "set 'base.tool_dir' to 'aem/home/opt/vendor'"},
	}
)

// legacyTemplateUsed tells if a config file uses the 2.x template syntax outside comments.
func legacyTemplateUsed(content []byte) bool {
	for _, line := range strings.Split(string(content), "\n") {
		if legacyTemplatePattern.MatchString(stripYAMLComment(line)) {
			return true
		}
	}
	return false
}

// stripYAMLComment cuts a comment off a line; a '#' starts a comment at the line start or after a whitespace.
func stripYAMLComment(line string) string {
	for i, r := range line {
		if r == '#' && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t') {
			return line[:i]
		}
	}
	return line
}

// LegacyMigration is the result of converting a 2.x config file to the 3.x syntax.
type LegacyMigration struct {
	Content string `yaml:"-" json:"-"`
	// Unresolved lists the lines left for a manual change, e.g. using template functions without a 3.x counterpart.
	Unresolved []string `yaml:"unresolved" json:"unresolved"`
}

// MigrateLegacy converts the 2.x template syntax to variables: a literal '$' becomes '$$', known template expressions become '${...}'.
func MigrateLegacy(content string) LegacyMigration {
	result := LegacyMigration{Unresolved: []string{}}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.ReplaceAll(line, "$", "$$")
		for _, r := range legacyReplacements {
			line = r.pattern.ReplaceAllString(line, r.replacement)
		}
		if legacyTemplatePattern.MatchString(line) {
			result.Unresolved = append(result.Unresolved, fmt.Sprintf("line %d: template expression without a counterpart: %s", i+1, strings.TrimSpace(line)))
		}
		lines[i] = line
	}
	result.Content = strings.Join(lines, "\n")
	result.Unresolved = append(result.Unresolved, legacySettingIssues(result.Content)...)
	return result
}

// legacySettingIssues lists the lines using settings of 2.x that need a manual change.
func legacySettingIssues(content string) []string {
	result := []string{}
	for i, line := range strings.Split(content, "\n") {
		for _, h := range legacySettingHints {
			if h.pattern.MatchString(line) {
				result = append(result, fmt.Sprintf("line %d: %s", i+1, h.hint))
			}
		}
	}
	return result
}

// LegacyFile is the config file skipped because it uses the 2.x template syntax; empty if none.
func (c *Config) LegacyFile() string {
	return c.legacyFile
}

// LegacyFileMigration is the result of migrating one config file.
type LegacyFileMigration struct {
	File       string   `yaml:"file" json:"file"`
	Changed    bool     `yaml:"changed" json:"changed"`
	Unresolved []string `yaml:"unresolved" json:"unresolved"`
}

// MigrateLegacyFiles converts the existing config files, the personal and the VCS-tracked one, to the 3.x syntax in place.
func MigrateLegacyFiles() ([]LegacyFileMigration, error) {
	files := lo.Filter(lo.Uniq([]string{File(), TemplateFile()}), func(f string, _ int) bool { return pathx.Exists(f) })
	results := []LegacyFileMigration{}
	for _, file := range files {
		content, err := filex.ReadString(file)
		if err != nil {
			return results, fmt.Errorf("cannot read config file '%s': %w", file, err)
		}
		result := LegacyFileMigration{File: file, Unresolved: legacySettingIssues(content)}
		if legacyTemplateUsed([]byte(content)) {
			migration := MigrateLegacy(content)
			if err := filex.WriteString(file, migration.Content); err != nil {
				return results, fmt.Errorf("cannot write config file '%s': %w", file, err)
			}
			result.Changed = true
			result.Unresolved = migration.Unresolved
		}
		results = append(results, result)
	}
	return results, nil
}
