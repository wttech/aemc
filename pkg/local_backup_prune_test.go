package pkg

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func testBackup(path, instance, name string, created time.Time, size uint64) BackupFile {
	return BackupFile{Path: path, Instance: instance, Name: name, Created: created, Size: size}
}

func backupPaths(backups []BackupFile) []string {
	return lo.Map(backups, func(b BackupFile, _ int) string { return b.Path })
}

func TestPlanBackupPruneKeepLast(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	day := func(d int) time.Time { return time.Date(2026, 1, d, 0, 0, 0, 0, time.Local) }
	backups := []BackupFile{
		testBackup("a1", "author", "", day(1), 10),
		testBackup("a2", "author", "", day(2), 10),
		testBackup("a3", "author", "", day(3), 10),
		testBackup("ab", "author", "baseline", day(0), 10),
		testBackup("p1", "publish", "", day(1), 10),
		testBackup("other", "", "", time.Time{}, 10),
		{Path: "tmp", Instance: "author", Created: day(4), Temp: true},
	}

	plan := planBackupPrune(backups, BackupPrunePolicy{KeepLast: 2}, "")
	a.Equal([]string{"a1"}, backupPaths(plan.Delete))
	a.ElementsMatch([]string{"a3", "a2", "p1", "ab"}, backupPaths(plan.Keep))

	plan = planBackupPrune(backups, BackupPrunePolicy{KeepLast: 1}, "a2")
	a.Equal([]string{"a1"}, backupPaths(plan.Delete), "protected backup is kept")

	plan = planBackupPrune(backups, BackupPrunePolicy{}, "")
	a.Empty(plan.Delete, "empty policy keeps all")
}

func TestPlanBackupPruneMaxSize(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	day := func(d int) time.Time { return time.Date(2026, 1, d, 0, 0, 0, 0, time.Local) }
	backups := []BackupFile{
		testBackup("a1", "author", "", day(1), 10),
		testBackup("a2", "author", "", day(2), 10),
		testBackup("a3", "author", "", day(3), 10),
		testBackup("p1", "publish", "", day(1), 10),
	}

	plan := planBackupPrune(backups, BackupPrunePolicy{MaxSize: 25}, "")
	a.Equal([]string{"a2", "a1"}, backupPaths(plan.Delete), "oldest first, newest of each group stays")

	plan = planBackupPrune(backups, BackupPrunePolicy{MaxSize: 1}, "")
	a.ElementsMatch([]string{"a3", "p1"}, backupPaths(plan.Keep), "cap never deletes the newest of a group")
}
