package pkg

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBackupManager(t *testing.T) (*BackupManager, string) {
	dir := t.TempDir()
	aem := DefaultAEM()
	im := aem.InstanceManager()
	im.LocalOpts.Backup = BackupOpts{Dir: dir}
	return im.Backup, dir
}

func writeTestBackup(t *testing.T, dir string, fileName string, size int, modified time.Time) string {
	path := filepath.Join(dir, fileName)
	require.NoError(t, os.WriteFile(path, make([]byte, size), 0644))
	require.NoError(t, os.Chtimes(path, modified, modified))
	return path
}

func TestBackupManagerPrune(t *testing.T) {
	a := assert.New(t)
	bm, dir := newTestBackupManager(t)
	now := time.Now()

	oldest := writeTestBackup(t, dir, "author-20260101000000.aemb.tar.zst", 10, now)
	writeTestBackup(t, dir, "author-6.5.21-20260102000000.aemb.tar.zst", 10, now)
	newest := writeTestBackup(t, dir, "author-20260103000000.aemb.tar.zst", 10, now)
	baseline := writeTestBackup(t, dir, "author-baseline-20250101000000.aemb.tar.zst", 10, now)
	foreign := writeTestBackup(t, dir, "my-backup.aemb.tar.zst", 10, now)
	staleTemp := writeTestBackup(t, dir, "tmp-author-20251231000000.aemb.tar.zst", 10, now.Add(-time.Hour))
	freshTemp := writeTestBackup(t, dir, "tmp-author-20260104000000.aemb.tar.zst", 10, now)

	list, err := bm.List()
	require.NoError(t, err)
	a.Equal(5, list.Total, "temporary archives are not listed")

	policy := BackupPrunePolicy{KeepLast: 2}
	preview, err := bm.PruneWith(policy, true, "")
	require.NoError(t, err)
	a.FileExists(oldest, "dry run deletes nothing")

	result, err := bm.PruneWith(policy, false, "")
	require.NoError(t, err)
	a.Equal(backupPaths(preview.Deleted), backupPaths(result.Deleted), "dry run lists exactly what a real run deletes")
	a.ElementsMatch([]string{oldest, staleTemp}, backupPaths(result.Deleted))
	a.Equal(uint64(20), result.Freed)

	a.NoFileExists(oldest)
	a.NoFileExists(staleTemp)
	for _, kept := range []string{newest, baseline, foreign, freshTemp} {
		a.FileExists(kept)
	}
}

func TestBackupManagerNewest(t *testing.T) {
	a := assert.New(t)
	bm, dir := newTestBackupManager(t)
	now := time.Now()

	writeTestBackup(t, dir, "author-6.5.21-20260101000000.aemb.tar.zst", 10, now)
	newest := writeTestBackup(t, dir, "author-20260102000000.aemb.tar.zst", 10, now)
	baseline := writeTestBackup(t, dir, "author-baseline-20250101000000.aemb.tar.zst", 10, now)
	writeTestBackup(t, dir, "tmp-author-20260103000000.aemb.tar.zst", 10, now)

	backup, err := bm.Newest("author", "", "")
	require.NoError(t, err)
	a.Equal(newest, backup.Path, "picked by timestamp, not by file name order")

	backup, err = bm.Newest("author", "baseline", "")
	require.NoError(t, err)
	a.Equal(baseline, backup.Path)

	backup, err = bm.Newest("author", "", "6.5.21")
	require.NoError(t, err)
	a.Contains(backup.Path, "6.5.21")

	_, err = bm.Newest("publish", "", "")
	a.Error(err)
}

func TestBackupManagerCheckSpace(t *testing.T) {
	a := assert.New(t)
	bm, dir := newTestBackupManager(t)
	author := bm.manager.NewLocalAuthor()
	writeTestBackup(t, dir, "author-20260101000000.aemb.tar.zst", 10, time.Now())
	file := filepath.Join(dir, "author-20260102000000.aemb.tar.zst")

	bm.manager.LocalOpts.Backup.MinFree = ""
	a.NoError(bm.CheckSpace(author.Local(), "", file), "empty min free disables the check")

	bm.manager.LocalOpts.Backup.MinFree = "0%"
	a.NoError(bm.CheckSpace(author.Local(), "", file))

	bm.manager.LocalOpts.Backup.MinFree = "100%"
	err := bm.CheckSpace(author.Local(), "", file)
	a.ErrorContains(err, "not enough disk space")
}

func TestBackupManagerCheckSpaceOfMany(t *testing.T) {
	a := assert.New(t)
	bm, dir := newTestBackupManager(t)
	author := bm.manager.NewLocalAuthor()
	publish := bm.manager.NewLocalPublish()
	writeTestBackup(t, dir, "author-20260101000000.aemb.tar.zst", 10, time.Now())
	writeTestBackup(t, dir, "publish-20260101000000.aemb.tar.zst", 10, time.Now())

	bm.manager.LocalOpts.Backup.MinFree = "0%"
	a.NoError(bm.checkSpace([]*LocalInstance{author.Local(), publish.Local()}, "", dir))

	bm.manager.LocalOpts.Backup.MinFree = "100%"
	err := bm.checkSpace([]*LocalInstance{author.Local(), publish.Local()}, "", dir)
	a.ErrorContains(err, "expected backup size 20 B")
}
