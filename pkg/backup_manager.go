package pkg

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/osx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/sizex"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"github.com/wttech/aemc/pkg/common/timex"
)

const (
	// localBackupTempStaleAfter is how long a temporary archive may stay unmodified before it counts as abandoned.
	// An archive being written is modified continuously.
	localBackupTempStaleAfter = 15 * time.Minute
)

type BackupOpts struct {
	// Dir holds the backup archives.
	Dir string
	// MinFree is the free space that must remain on the disk of Dir after a backup is made,
	// as a percentage of the disk (e.g. '10%') or an absolute size (e.g. '50GB'); empty disables the check.
	MinFree string
	Prune   BackupPruneOpts
}

type BackupPruneOpts struct {
	// Auto prunes old backups right after a backup is made.
	Auto     bool
	KeepLast int
	// MaxSize caps the total size of backups (e.g. '100GB'); empty means no cap.
	MaxSize string
}

func newBackupOpts(cfg *viper.Viper) BackupOpts {
	return BackupOpts{
		Dir:     cfg.GetString("instance.local.backup.dir"),
		MinFree: cfg.GetString("instance.local.backup.min_free"),
		Prune: BackupPruneOpts{
			Auto:     cfg.GetBool("instance.local.backup.prune.auto"),
			KeepLast: cfg.GetInt("instance.local.backup.prune.keep_last"),
			MaxSize:  cfg.GetString("instance.local.backup.prune.max_size"),
		},
	}
}

func (o BackupOpts) PrunePolicy() (BackupPrunePolicy, error) {
	maxSize, err := sizex.Parse(o.Prune.MaxSize)
	if err != nil {
		return BackupPrunePolicy{}, fmt.Errorf("invalid backup prune max size: %w", err)
	}
	return BackupPrunePolicy{KeepLast: o.Prune.KeepLast, MaxSize: maxSize}, nil
}

// BackupManager makes, finds, restores and prunes local instance backups.
type BackupManager struct {
	manager *InstanceManager
}

func NewBackupManager(manager *InstanceManager) *BackupManager {
	return &BackupManager{manager: manager}
}

func (bm *BackupManager) opts() BackupOpts {
	return bm.manager.LocalOpts.Backup
}

func (bm *BackupManager) Dir() string {
	return bm.opts().Dir
}

// files returns all backup archives in the backup dir, including temporary ones.
func (bm *BackupManager) files() ([]BackupFile, error) {
	dir := bm.Dir()
	paths, err := pathx.GlobDir(dir, "*."+LocalInstanceBackupExtension)
	if err != nil {
		return nil, fmt.Errorf("cannot list instance backups in directory '%s': %w", dir, err)
	}
	var result []BackupFile
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) { // deleted meanwhile
				continue
			}
			return nil, fmt.Errorf("cannot read instance backup file '%s': %w", path, err)
		}
		result = append(result, newBackupFile(path, info))
	}
	return result, nil
}

func (bm *BackupManager) List() (*BackupList, error) {
	files, err := bm.files()
	if err != nil {
		return nil, err
	}
	return newBackupList(lo.Filter(files, func(b BackupFile, _ int) bool { return !b.Temp })), nil
}

// Newest finds the newest complete backup of an instance.
// An empty name matches backups of any name; an empty AEM version matches backups of any version.
func (bm *BackupManager) Newest(instanceName, name, aemVersion string) (*BackupFile, error) {
	files, err := bm.files()
	if err != nil {
		return nil, err
	}
	version := formatBackupVersion(aemVersion)
	matching := lo.Filter(files, func(b BackupFile, _ int) bool {
		return b.Recognized() && !b.Temp && b.Instance == instanceName &&
			(name == "" || b.Name == name) &&
			(version == "" || formatBackupVersion(b.AemVersion) == version)
	})
	if len(matching) == 0 {
		return nil, fmt.Errorf("no backup found in directory '%s' for instance '%s' (name '%s', AEM version '%s')", bm.Dir(), instanceName, name, aemVersion)
	}
	sortBackupsNewestFirst(matching)
	return &matching[0], nil
}

// NormalizeName turns a user-given backup name into the form used in file names.
func (bm *BackupManager) NormalizeName(name string) (string, error) {
	if name == "" {
		return "", nil
	}
	result := stringsx.Slug(name)
	if result == "" {
		return "", fmt.Errorf("backup name '%s' has no letters or digits", name)
	}
	return result, nil
}

func (bm *BackupManager) ProposeFile(li *LocalInstance, name string) string {
	aemVersion := ""
	if li.IsRunning() {
		aemVersion = li.instance.AemVersion()
	}
	fileName := formatBackupFileName(backupFileName{
		Instance:   li.Name(),
		Name:       name,
		AemVersion: aemVersion,
		Created:    time.Now(),
	})
	return pathx.Canonical(fmt.Sprintf("%s/%s", bm.Dir(), fileName))
}

type BackupMakeOpts struct {
	// File is the target archive; empty proposes one in the backup dir.
	File string
	// Name is the backup name (normalized); backups of the same instance and name replace each other when pruning.
	Name           string
	SkipSpaceCheck bool
}

// Make archives a created instance, stopping it for the time of archiving, and then prunes old backups if configured.
func (bm *BackupManager) Make(li *LocalInstance, opts BackupMakeOpts) (string, error) {
	file := opts.File
	if file == "" {
		file = bm.ProposeFile(li, opts.Name)
	}
	if !li.IsCreated() {
		return "", fmt.Errorf("%s > cannot make backup to file '%s' as instance not created", li.instance.IDColor(), file)
	}
	if !opts.SkipSpaceCheck {
		if err := bm.CheckSpace(li, opts.Name, file); err != nil {
			return "", err
		}
	}
	err := bm.whileStopped(li, func() error { return li.MakeBackup(file) })
	if err != nil {
		return "", err
	}
	if bm.opts().Prune.Auto {
		if _, err := bm.Prune(false, file); err != nil {
			return file, fmt.Errorf("%s > backup made to file '%s' but pruning old backups failed: %w", li.instance.IDColor(), file, err)
		}
	}
	return file, nil
}

type BackupUseOpts struct {
	// File is the archive to restore from; empty picks the newest backup of the instance.
	File string
	// Name narrows the picked backup to the given (normalized) name.
	Name          string
	DeleteCreated bool
}

// BackupResult is a backup made or used for an instance.
type BackupResult struct {
	Instance Instance `yaml:"instance" json:"instance"`
	File     string   `yaml:"file" json:"file"`
}

// Perform makes backups of the created instances like 'instance stop', then archiving, then 'instance start':
// the running ones are stopped together, archived honouring the instance processing mode, and started together again,
// so they wait for one startup instead of one per instance. Disk space is checked for all of them before any is stopped,
// and old backups are pruned once at the end.
func (bm *BackupManager) Perform(instances []Instance, opts BackupMakeOpts) ([]BackupResult, error) {
	created := lo.Filter(instances, func(i Instance, _ int) bool {
		if !i.Local().IsCreated() {
			log.Infof("%s > skipping making backup as it is not created", i.IDColor())
			return false
		}
		return true
	})
	if len(created) == 0 {
		return []BackupResult{}, nil
	}
	if !opts.SkipSpaceCheck {
		locals := lo.Map(created, func(i Instance, _ int) *LocalInstance { return i.Local() })
		if err := bm.checkSpace(locals, opts.Name, bm.Dir()); err != nil {
			return nil, err
		}
	}
	running := lo.Filter(created, func(i Instance, _ int) bool { return i.Local().IsRunning() })
	if _, err := bm.manager.Stop(running); err != nil {
		return nil, err
	}
	results, err := InstanceProcess(bm.manager.aem, created, func(i Instance) (BackupResult, error) {
		file := bm.ProposeFile(i.Local(), opts.Name)
		return BackupResult{Instance: i, File: file}, i.Local().MakeBackup(file)
	})
	if err != nil {
		return nil, err
	}
	if _, err := bm.manager.Start(running); err != nil {
		return results, err
	}
	if bm.opts().Prune.Auto {
		if _, err := bm.Prune(false, ""); err != nil {
			return results, fmt.Errorf("backups made but pruning old backups failed: %w", err)
		}
	}
	return results, nil
}

// Restore restores the not created instances from their newest backups, honouring the instance processing mode.
func (bm *BackupManager) Restore(instances []Instance, name string) ([]BackupResult, error) {
	restorable := lo.Filter(instances, func(i Instance, _ int) bool {
		if i.Local().IsCreated() {
			log.Infof("%s > skipping using backup as it is already created", i.IDColor())
			return false
		}
		if _, err := bm.Newest(i.Local().Name(), name, ""); err != nil {
			log.Warnf("%s > skipping using backup as cannot find any: %s", i.IDColor(), err)
			return false
		}
		return true
	})
	if len(restorable) == 0 {
		return []BackupResult{}, nil
	}
	return InstanceProcess(bm.manager.aem, restorable, func(i Instance) (BackupResult, error) {
		file, err := bm.Use(i.Local(), BackupUseOpts{Name: name})
		return BackupResult{Instance: i, File: file}, err
	})
}

// Use restores an instance from a backup, stopping it for the time of restoring.
func (bm *BackupManager) Use(li *LocalInstance, opts BackupUseOpts) (string, error) {
	file := opts.File
	if file == "" {
		proposed, err := bm.proposeFileToUse(li, opts.Name)
		if err != nil {
			return "", err
		}
		file = proposed
	}
	err := bm.whileStopped(li, func() error { return li.UseBackup(file, opts.DeleteCreated) })
	if err != nil {
		return "", err
	}
	return file, nil
}

// proposeFileToUse picks the newest backup; for a running instance only the ones of its AEM version.
func (bm *BackupManager) proposeFileToUse(li *LocalInstance, name string) (string, error) {
	aemVersion := ""
	if li.IsRunning() {
		version, err := li.instance.status.AemVersion()
		if err != nil {
			return "", fmt.Errorf("%s > cannot determine AEM version to find backup file to use: %w", li.instance.IDColor(), err)
		}
		aemVersion = version
	}
	backup, err := bm.Newest(li.Name(), name, aemVersion)
	if err != nil {
		return "", fmt.Errorf("%s > no backup file found to use: %w", li.instance.IDColor(), err)
	}
	return backup.Path, nil
}

func (bm *BackupManager) whileStopped(li *LocalInstance, action func() error) error {
	running := li.IsRunning()
	if running {
		if err := li.StopAndAwait(); err != nil {
			return err
		}
	}
	if err := action(); err != nil {
		return err
	}
	if running {
		if err := li.StartAndAwait(); err != nil {
			return err
		}
	}
	return nil
}

// CheckSpace fails when making a backup of the instance to the file would leave less than the configured free space on its disk.
func (bm *BackupManager) CheckSpace(li *LocalInstance, name string, file string) error {
	return bm.checkSpace([]*LocalInstance{li}, name, filepath.Dir(file))
}

// checkSpace checks that the disk of dir has room for the next backups of all given instances together.
func (bm *BackupManager) checkSpace(lis []*LocalInstance, name string, dir string) error {
	minFreeValue := bm.opts().MinFree
	if minFreeValue == "" {
		return nil
	}
	if err := pathx.Ensure(dir); err != nil {
		return err
	}
	usage, err := osx.DiskUsageOf(dir)
	if err != nil {
		return err
	}
	minFree, err := sizex.ParseThreshold(minFreeValue, usage.Total)
	if err != nil {
		return fmt.Errorf("invalid backup min free space: %w", err)
	}
	var expected uint64
	for _, li := range lis {
		size, err := bm.estimateSize(li, name)
		if err != nil {
			return err
		}
		expected += size
	}
	if usage.Free < expected+minFree {
		ids := strings.Join(lo.Map(lis, func(li *LocalInstance, _ int) string { return li.instance.IDColor() }), ",")
		return fmt.Errorf("%s > not enough disk space to make backup in directory '%s': expected backup size %s, free %s, to stay free %s ('%s'); "+
			"free up space, e.g. with 'aem instance backup prune', or skip this check with '--skip-space-check'",
			ids, dir, humanize.Bytes(expected), humanize.Bytes(usage.Free), humanize.Bytes(minFree), minFreeValue)
	}
	return nil
}

// estimateSize guesses the size of the next backup: the newest backup of the same name, of any name, or else the instance dir size (an upper bound).
func (bm *BackupManager) estimateSize(li *LocalInstance, name string) (uint64, error) {
	if backup, err := bm.Newest(li.Name(), name, ""); err == nil {
		return backup.Size, nil
	}
	if backup, err := bm.Newest(li.Name(), "", ""); err == nil {
		return backup.Size, nil
	}
	size, err := pathx.Size(li.Dir())
	if err != nil {
		return 0, fmt.Errorf("%s > cannot estimate backup size: %w", li.instance.IDColor(), err)
	}
	return size, nil
}

// Prune deletes the backups that the configured policy does not keep, and abandoned temporary archives.
func (bm *BackupManager) Prune(dryRun bool, protectedPath string) (*BackupPruneResult, error) {
	policy, err := bm.opts().PrunePolicy()
	if err != nil {
		return nil, err
	}
	return bm.PruneWith(policy, dryRun, protectedPath)
}

func (bm *BackupManager) PruneWith(policy BackupPrunePolicy, dryRun bool, protectedPath string) (*BackupPruneResult, error) {
	files, err := bm.files()
	if err != nil {
		return nil, err
	}
	plan := planBackupPrune(files, policy, protectedPath)
	staleTemps := lo.Filter(files, func(b BackupFile, _ int) bool {
		return b.Temp && time.Since(b.Modified) > localBackupTempStaleAfter
	})
	result := &BackupPruneResult{
		DryRun:  dryRun,
		Kept:    plan.Keep,
		Deleted: append(plan.Delete, staleTemps...),
	}
	result.Freed = lo.SumBy(result.Deleted, func(b BackupFile) uint64 { return b.Size })
	for _, backup := range result.Deleted {
		if dryRun {
			log.Infof("would delete backup file '%s' (%s)", backup.Path, humanize.Bytes(backup.Size))
			continue
		}
		log.Infof("deleting backup file '%s' (%s)", backup.Path, humanize.Bytes(backup.Size))
		if err := pathx.DeleteIfExists(backup.Path); err != nil {
			return result, fmt.Errorf("cannot delete backup file '%s': %w", backup.Path, err)
		}
	}
	return result, nil
}

type BackupPruneResult struct {
	DryRun  bool         `yaml:"dry_run" json:"dryRun"`
	Kept    []BackupFile `yaml:"kept" json:"kept"`
	Deleted []BackupFile `yaml:"deleted" json:"deleted"`
	Freed   uint64       `yaml:"freed" json:"freed"`
}

func (r BackupPruneResult) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblMap("stats", "stat", "value", map[string]any{
		"dry run": r.DryRun,
		"kept":    len(r.Kept),
		"deleted": len(r.Deleted),
		"freed":   humanize.Bytes(r.Freed),
	}))
	bs.WriteString("\n")
	bs.WriteString(backupFilesTable("deleted", r.Deleted))
	return bs.String()
}

type BackupList struct {
	Total int          `yaml:"total" json:"total"`
	Files []BackupFile `yaml:"files" json:"files"`
}

func newBackupList(files []BackupFile) *BackupList {
	sort.SliceStable(files, func(i, j int) bool {
		if files[i].group() != files[j].group() {
			return files[i].group() < files[j].group()
		}
		return files[i].Created.After(files[j].Created)
	})
	return &BackupList{Total: len(files), Files: files}
}

func (fl BackupList) MarshalText() string {
	bs := bytes.NewBufferString("")
	bs.WriteString(fmtx.TblMap("stats", "stat", "value", map[string]any{
		"total": len(fl.Files),
		"size":  humanize.Bytes(lo.SumBy(fl.Files, func(file BackupFile) uint64 { return file.Size })),
	}))
	bs.WriteString("\n")
	instanceSizes := map[string]any{}
	for instanceName, files := range lo.GroupBy(fl.Files, func(b BackupFile) string { return b.Instance }) {
		instanceSizes[lo.Ternary(instanceName == "", "<unrecognized>", instanceName)] = fmt.Sprintf("%d (%s)", len(files), humanize.Bytes(lo.SumBy(files, func(b BackupFile) uint64 { return b.Size })))
	}
	bs.WriteString(fmtx.TblMap("instances", "instance", "backups", instanceSizes))
	bs.WriteString("\n")
	bs.WriteString(backupFilesTable("list", fl.Files))
	return bs.String()
}

func backupFilesTable(caption string, files []BackupFile) string {
	return fmtx.TblRows(caption, false, []string{"instance", "name", "aem version", "created", "size", "path"}, lo.Map(files, func(file BackupFile, _ int) map[string]any {
		created := ""
		if !file.Created.IsZero() {
			created = timex.Human(file.Created)
		}
		return map[string]any{
			"instance":    file.Instance,
			"name":        file.Name,
			"aem version": file.AemVersion,
			"created":     created,
			"size":        humanize.Bytes(file.Size),
			"path":        file.Path,
		}
	}))
}
