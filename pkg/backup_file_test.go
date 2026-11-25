package pkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackupFileNameRoundTrip(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	created := time.Date(2026, 2, 9, 14, 30, 22, 0, time.Local)
	name := formatBackupFileName(backupFileName{Instance: "author", Name: "clean-with-prereqs", AemVersion: "6.6 LTS", Created: created})
	a.Equal("author-clean-with-prereqs-6.6_lts-20260209143022.aemb.tar.zst", name)

	parsed, ok := parseBackupFileName(name)
	a.True(ok)
	a.Equal(backupFileName{Instance: "author", Name: "clean-with-prereqs", AemVersion: "6.6_lts", Created: created}, parsed)
}

func TestBackupFileNameParse(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	parsed, ok := parseBackupFileName("publish_preview-20260209143022.aemb.tar.zst")
	a.True(ok)
	a.Equal("publish_preview", parsed.Instance)
	a.Equal("", parsed.Name)
	a.Equal("", parsed.AemVersion)

	parsed, ok = parseBackupFileName("author-6.5.21-20260209143022.aemb.tar.zst")
	a.True(ok)
	a.Equal("", parsed.Name)
	a.Equal("6.5.21", parsed.AemVersion)

	// released versions wrote the reported version as is
	parsed, ok = parseBackupFileName("author-6.6 LTS-20260209143022.aemb.tar.zst")
	a.True(ok)
	a.Equal("6.6 LTS", parsed.AemVersion)
	a.Equal("6.6_lts", formatBackupVersion(parsed.AemVersion))

	parsed, ok = parseBackupFileName("author-<unknown>-20260209143022.aemb.tar.zst")
	a.True(ok)
	a.Equal("", parsed.Name)
	a.Equal("", parsed.AemVersion)

	parsed, ok = parseBackupFileName("author-baseline-20260209143022.aemb.tar.zst")
	a.True(ok)
	a.Equal("baseline", parsed.Name)

	for _, fileName := range []string{
		"my-backup.aemb.tar.zst",
		"author-2026.aemb.tar.zst",
		"20260209143022.aemb.tar.zst",
		"author-20260209143022.zip",
	} {
		_, ok = parseBackupFileName(fileName)
		a.False(ok, fileName)
	}
}
