package pkg

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/timex"
	"github.com/wttech/aemc/pkg/instance"
)

// Backup file name convention: '<instance>[-<name>][-<aem version>]-<timestamp>.aemb.tar.zst'.
// Instance names never contain '-' (role and classifier are joined with '_'), names are slugs (no '.'),
// and AEM versions always contain a '.', which makes every part recognizable.
const (
	localBackupNamePartSep = "-"
)

var localBackupVersionInvalidChars = regexp.MustCompile(`[^a-z0-9.]+`)

// BackupFile is a local instance backup archive with the details encoded in its file name.
// Instance is empty when the file does not follow the backup file name convention.
type BackupFile struct {
	Path       string    `yaml:"path" json:"path"`
	Instance   string    `yaml:"instance" json:"instance"`
	Name       string    `yaml:"name" json:"name"`
	AemVersion string    `yaml:"aem_version" json:"aemVersion"`
	Created    time.Time `yaml:"created" json:"created"`
	Size       uint64    `yaml:"size" json:"size"`
	Modified   time.Time `yaml:"modified" json:"modified"`
	Temp       bool      `yaml:"temp" json:"temp"`
}

func (b BackupFile) Recognized() bool {
	return b.Instance != ""
}

// group identifies backups that replace each other: the same instance and the same backup name.
func (b BackupFile) group() string {
	return b.Instance + "/" + b.Name
}

func newBackupFile(path string, info os.FileInfo) BackupFile {
	result := BackupFile{
		Path:     path,
		Size:     uint64(info.Size()),
		Modified: info.ModTime(),
	}
	fileName, temp := strings.CutPrefix(filepath.Base(path), filex.ArchiveTempPrefix)
	result.Temp = temp
	parsed, ok := parseBackupFileName(fileName)
	if ok {
		result.Instance = parsed.Instance
		result.Name = parsed.Name
		result.AemVersion = parsed.AemVersion
		result.Created = parsed.Created
	}
	return result
}

type backupFileName struct {
	Instance   string
	Name       string
	AemVersion string
	Created    time.Time
}

func formatBackupFileName(f backupFileName) string {
	parts := []string{f.Instance}
	if f.Name != "" {
		parts = append(parts, f.Name)
	}
	if version := formatBackupVersion(f.AemVersion); version != "" {
		parts = append(parts, version)
	}
	parts = append(parts, timex.FileTimestamp(f.Created))
	return strings.Join(parts, localBackupNamePartSep) + "." + LocalInstanceBackupExtension
}

// formatBackupVersion turns a reported AEM version into a single file name part, e.g. '6.6 LTS' to '6.6_lts'.
func formatBackupVersion(version string) string {
	if version == "" || version == instance.AemVersionUnknown {
		return ""
	}
	return strings.Trim(localBackupVersionInvalidChars.ReplaceAllString(strings.ToLower(version), "_"), "_")
}

func parseBackupFileName(fileName string) (backupFileName, bool) {
	base, ok := strings.CutSuffix(fileName, "."+LocalInstanceBackupExtension)
	if !ok {
		return backupFileName{}, false
	}
	parts := strings.Split(base, localBackupNamePartSep)
	if len(parts) < 2 || parts[0] == "" {
		return backupFileName{}, false
	}
	created, err := timex.ParseFileTimestamp(parts[len(parts)-1])
	if err != nil {
		return backupFileName{}, false
	}
	result := backupFileName{Instance: parts[0], Created: created}
	middle := parts[1 : len(parts)-1]
	if len(middle) > 0 {
		last := middle[len(middle)-1]
		if strings.Contains(last, ".") || last == instance.AemVersionUnknown {
			if last != instance.AemVersionUnknown {
				result.AemVersion = last
			}
			middle = middle[:len(middle)-1]
		}
	}
	result.Name = strings.Join(middle, localBackupNamePartSep)
	return result, true
}
