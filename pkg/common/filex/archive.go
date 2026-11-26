package filex

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/mholt/archives"
	"github.com/wttech/aemc/pkg/common/pathx"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Archive writes the dir contents (or the file) to the archive, with the entries at its root, keeping file modes and symlinks.
// The format follows the file extension, e.g. 'zip', 'tar.gz' or 'tar.zst'.
func Archive(sourcePath, targetFile string) error {
	if !pathx.Exists(sourcePath) {
		return fmt.Errorf("cannot archive path '%s' to file '%s' as source path does not exist", sourcePath, targetFile)
	}
	ctx := context.Background()
	format, _, err := archives.Identify(ctx, archiveNameForFormat(targetFile), nil)
	if err != nil {
		return fmt.Errorf("cannot archive path '%s' to file '%s' as its format is not supported: %w", sourcePath, targetFile, err)
	}
	archiver, ok := format.(archives.Archiver)
	if !ok {
		return fmt.Errorf("cannot archive path '%s' to file '%s' as its format supports extracting only", sourcePath, targetFile)
	}
	if err := pathx.Ensure(filepath.Dir(targetFile)); err != nil {
		return err
	}
	root := pathx.Canonical(sourcePath) // the archive library names entries by trimming this prefix, so it must be clean
	if pathx.IsDir(sourcePath) {
		root += string(filepath.Separator) // the dir contents go to the archive root, not the dir itself
	}
	files, err := archives.FilesFromDisk(ctx, nil, map[string]string{root: ""})
	if err != nil {
		return fmt.Errorf("cannot read path '%s' to be archived to file '%s': %w", sourcePath, targetFile, err)
	}
	if err := writeArchive(ctx, archiver, files, targetFile); err != nil {
		return fmt.Errorf("cannot archive path '%s' to file '%s': %w", sourcePath, targetFile, err)
	}
	return nil
}

// archiveShortExtensions maps the short extensions of compressed tars, which the archive library identifies by content only,
// to the long ones it identifies by name too, so that archives can be written to them.
var archiveShortExtensions = map[string]string{".tgz": ".tar.gz", ".tbz2": ".tar.bz2", ".txz": ".tar.xz", ".tlz4": ".tar.lz4", ".tsz": ".tar.sz", ".tbr": ".tar.br"}

func archiveNameForFormat(file string) string {
	ext := strings.ToLower(filepath.Ext(file))
	if long, ok := archiveShortExtensions[ext]; ok {
		return strings.TrimSuffix(file, filepath.Ext(file)) + long
	}
	return file
}

func writeArchive(ctx context.Context, archiver archives.Archiver, files []archives.FileInfo, targetFile string) (err error) {
	file, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	buffered := bufio.NewWriterSize(file, 1<<20)
	if err := archiver.Archive(ctx, buffered, files); err != nil {
		return err
	}
	return buffered.Flush()
}

// ArchiveTempPrefix marks an archive file that is still being written; it is renamed to its target name when complete.
const ArchiveTempPrefix = "tmp-"

func ArchiveWithChanged(sourceDir, targetFile string) (bool, error) {
	if pathx.Exists(targetFile) {
		return false, nil
	}
	targetTmpFile := filepath.Dir(targetFile) + "/" + ArchiveTempPrefix + filepath.Base(targetFile)
	if err := pathx.DeleteIfExists(targetTmpFile); err != nil {
		return false, fmt.Errorf("cannot delete temporary archive file '%s': %w", targetTmpFile, err)
	}
	if err := Archive(sourceDir, targetTmpFile); err != nil {
		return false, err
	}
	if err := os.Rename(targetTmpFile, targetFile); err != nil {
		return false, fmt.Errorf("cannot move temporary archive file '%s' to target one '%s': %w", targetTmpFile, targetFile, err)
	}
	return true, nil
}

// Unarchive extracts the archive to the dir, keeping file modes and symlinks. Directories get mode 0755.
// Entries that would end up outside the dir or be written through a symlink are refused ("zip slip"), which the
// archive library leaves to its callers.
func Unarchive(sourceFile string, targetDir string) error {
	if !pathx.Exists(sourceFile) {
		return fmt.Errorf("cannot unarchive file '%s' to dir '%s' as source file does not exist", sourceFile, targetDir)
	}
	if err := pathx.Ensure(targetDir); err != nil {
		return err
	}
	if err := extractArchive(sourceFile, newArchiveExtractor(targetDir)); err != nil {
		return fmt.Errorf("cannot unarchive file '%s' to dir '%s': %w", sourceFile, targetDir, err)
	}
	return nil
}

func extractArchive(sourceFile string, extractor *archiveExtractor) error {
	file, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer file.Close()
	ctx := context.Background()
	format, stream, err := archives.Identify(ctx, sourceFile, file)
	if err != nil {
		return fmt.Errorf("format is not supported: %w", err)
	}
	ex, ok := format.(archives.Extractor)
	if !ok {
		return fmt.Errorf("format '%s' is not an archive", format.Extension())
	}
	if _, isZip := format.(archives.Zip); isZip {
		if _, err := file.Seek(0, io.SeekStart); err != nil { // zip is read from its end, so it needs the file itself
			return err
		}
		stream = file
	}
	return ex.Extract(ctx, stream, extractor.handle)
}

func UnarchiveWithChanged(sourceFile, targetDir string) (bool, error) {
	if pathx.Exists(targetDir) {
		return false, nil
	}
	targetTmpDir := targetDir + ".tmp"
	if err := pathx.DeleteIfExists(targetTmpDir); err != nil {
		return false, fmt.Errorf("cannot delete unarchive temporary dir '%s': %w", targetTmpDir, err)
	}
	if err := Unarchive(sourceFile, targetTmpDir); err != nil {
		return false, err
	}
	if err := os.Rename(targetTmpDir, targetDir); err != nil {
		return false, fmt.Errorf("cannot move unarchived temporary dir '%s' to target one '%s': %w", targetTmpDir, targetDir, err)
	}
	return true, nil
}

// archiveExtractor writes archive entries under its dir. It creates every parent dir itself and refuses to go through
// a symlink, so an entry cannot reach outside the dir via a symlink extracted before it.
type archiveExtractor struct {
	dir      string
	safeDirs map[string]bool
}

func newArchiveExtractor(dir string) *archiveExtractor {
	dir = filepath.Clean(dir)
	return &archiveExtractor{dir: dir, safeDirs: map[string]bool{dir: true}}
}

// handle writes one archive entry; devices and pipes are skipped as they cannot be restored meaningfully.
func (e *archiveExtractor) handle(_ context.Context, f archives.FileInfo) error {
	switch {
	case f.IsDir():
		return e.dirEntry(f.NameInArchive)
	case f.Mode()&fs.ModeSymlink != 0:
		return e.symlinkEntry(f.NameInArchive, f.LinkTarget)
	case f.LinkTarget != "": // hard link in a tar
		return e.hardLinkEntry(f.NameInArchive, f.LinkTarget)
	case f.Mode().IsRegular():
		content, err := f.Open()
		if err != nil {
			return err
		}
		defer content.Close()
		return e.fileEntry(f.NameInArchive, f.Mode(), content)
	default:
		return nil
	}
}

func (e *archiveExtractor) ensureDir(dir string) error {
	if e.safeDirs[dir] {
		return nil
	}
	if !pathx.IsWithin(e.dir, dir) {
		return fmt.Errorf("dir '%s' is outside of dir '%s'", dir, e.dir)
	}
	if err := e.ensureDir(filepath.Dir(dir)); err != nil {
		return err
	}
	info, err := os.Lstat(dir)
	switch {
	case os.IsNotExist(err):
		if err := os.Mkdir(dir, 0755); err != nil && !os.IsExist(err) {
			return err
		}
	case err != nil:
		return err
	case info.Mode()&os.ModeSymlink != 0:
		return fmt.Errorf("refusing to extract through symlink '%s'", dir)
	case !info.IsDir():
		return fmt.Errorf("path '%s' is not a dir", dir)
	}
	e.safeDirs[dir] = true
	return nil
}

// entryPath returns where the entry goes and prepares its parent dir; an existing file or symlink there is removed.
func (e *archiveExtractor) entryPath(name string) (string, error) {
	path, err := pathx.JoinWithin(e.dir, name)
	if err != nil {
		return "", err
	}
	if path == e.dir {
		return "", fmt.Errorf("entry '%s' names the target dir itself", name)
	}
	if err := e.ensureDir(filepath.Dir(path)); err != nil {
		return "", err
	}
	if info, err := os.Lstat(path); err == nil && !info.IsDir() {
		if err := os.Remove(path); err != nil {
			return "", err
		}
	}
	return path, nil
}

func (e *archiveExtractor) dirEntry(name string) error {
	path, err := pathx.JoinWithin(e.dir, name)
	if err != nil {
		return err
	}
	return e.ensureDir(path)
}

func (e *archiveExtractor) fileEntry(name string, mode os.FileMode, content io.Reader) error {
	path, err := e.entryPath(name)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode.Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, content); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Chmod(mode.Perm()); err != nil { // the exact mode from the archive, not narrowed by the umask
		_ = file.Close()
		return err
	}
	return file.Close()
}

// symlinkEntry creates a symlink as archived, also an absolute one or one pointing outside the dir (e.g. an instance dir
// linked to another disk); nothing is ever extracted through it, as ensureDir refuses to go through symlinks.
func (e *archiveExtractor) symlinkEntry(name string, target string) error {
	path, err := e.entryPath(name)
	if err != nil {
		return err
	}
	return os.Symlink(filepath.FromSlash(target), path)
}

func (e *archiveExtractor) hardLinkEntry(name string, target string) error {
	path, err := e.entryPath(name)
	if err != nil {
		return err
	}
	targetPath, err := pathx.JoinWithin(e.dir, target)
	if err != nil {
		return fmt.Errorf("hard link '%s': %w", name, err)
	}
	return os.Link(targetPath, path)
}

// UnarchiveMakeself extracts tar.gz archive from a Makeself self-extracting shell script.
// Makeself (https://makeself.io/) creates self-extracting archives by embedding a tar.gz archive after a shell header.
// This function finds the gzip magic bytes (0x1f 0x8b) and extracts everything after it.
func UnarchiveMakeself(scriptPath string, targetDir string) error {
	file, err := os.Open(scriptPath)
	if err != nil {
		return fmt.Errorf("cannot open Makeself script file '%s': %w", scriptPath, err)
	}
	defer file.Close()

	// Read entire file to find embedded gzip archive
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("cannot read Makeself script file '%s': %w", scriptPath, err)
	}

	// Find gzip magic bytes (1f 8b) which mark the start of embedded tar.gz archive
	gzipMagic := []byte{0x1f, 0x8b}
	archiveStart := bytes.Index(data, gzipMagic)
	if archiveStart == -1 {
		return fmt.Errorf("cannot find gzip archive in Makeself script file '%s' (missing gzip magic bytes 0x1f 0x8b)", scriptPath)
	}

	// Extract archive data to temporary file as sibling of the script file
	tmpFile, err := os.CreateTemp(filepath.Dir(scriptPath), filepath.Base(scriptPath)+"-*.tar.gz")
	if err != nil {
		return fmt.Errorf("cannot create temporary archive file in dir '%s' for Makeself extraction: %w", filepath.Dir(scriptPath), err)
	}
	tmpArchive := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpArchive)

	if err := os.WriteFile(tmpArchive, data[archiveStart:], 0644); err != nil {
		return fmt.Errorf("cannot write extracted Makeself archive data to temporary file '%s': %w", tmpArchive, err)
	}

	// Unpack the tar.gz using standard archiver
	if err := Unarchive(tmpArchive, targetDir); err != nil {
		return fmt.Errorf("cannot unarchive extracted Makeself tar.gz '%s' to dir '%s': %w", tmpArchive, targetDir, err)
	}

	return nil
}
