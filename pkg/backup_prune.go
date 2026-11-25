package pkg

import (
	"sort"

	"github.com/samber/lo"
)

// BackupPrunePolicy says which backups to keep; zero values keep everything.
type BackupPrunePolicy struct {
	// KeepLast is the number of newest backups kept per group (instance + backup name); 0 keeps all.
	KeepLast int
	// MaxSize caps the total size of backups; the oldest are deleted first, but never the newest of a group. 0 means no cap.
	MaxSize uint64
}

func (p BackupPrunePolicy) Enabled() bool {
	return p.KeepLast > 0 || p.MaxSize > 0
}

type backupPrunePlan struct {
	Keep   []BackupFile
	Delete []BackupFile
}

// planBackupPrune decides which recognized, complete backups to delete. The backup at 'protectedPath'
// (e.g. the one just made) and the newest backup of each group are never deleted.
func planBackupPrune(backups []BackupFile, policy BackupPrunePolicy, protectedPath string) backupPrunePlan {
	candidates := lo.Filter(backups, func(b BackupFile, _ int) bool { return b.Recognized() && !b.Temp })
	groups := lo.GroupBy(candidates, func(b BackupFile) string { return b.group() })

	var keep, remove []BackupFile
	newestOfGroup := map[string]bool{}
	for _, group := range groups {
		sortBackupsNewestFirst(group)
		newestOfGroup[group[0].Path] = true
		for i, backup := range group {
			if policy.KeepLast > 0 && i >= policy.KeepLast && backup.Path != protectedPath {
				remove = append(remove, backup)
			} else {
				keep = append(keep, backup)
			}
		}
	}

	if policy.MaxSize > 0 {
		sortBackupsNewestFirst(keep)
		total := lo.SumBy(keep, func(b BackupFile) uint64 { return b.Size })
		for i := len(keep) - 1; i >= 0 && total > policy.MaxSize; i-- {
			backup := keep[i]
			if newestOfGroup[backup.Path] || backup.Path == protectedPath {
				continue
			}
			remove = append(remove, backup)
			total -= backup.Size
			keep = append(keep[:i], keep[i+1:]...)
		}
	}

	sortBackupsNewestFirst(keep)
	sortBackupsNewestFirst(remove)
	return backupPrunePlan{Keep: keep, Delete: remove}
}

func sortBackupsNewestFirst(backups []BackupFile) {
	sort.SliceStable(backups, func(i, j int) bool {
		if backups[i].Created.Equal(backups[j].Created) {
			return backups[i].Path > backups[j].Path
		}
		return backups[i].Created.After(backups[j].Created)
	})
}
