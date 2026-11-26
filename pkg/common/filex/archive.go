package filex

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/pathx"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// archiveFormat is chosen by the file extension: zip, tar, tar.gz (tgz), tar.zst (tzst) and, for unarchiving only, tar.bz2 (tbz2).
type archiveFormat struct {
	zip         bool
	compression string // for tar: "", "gz", "zst" or "bz2"
}

func archiveFormatOf(file string) (archiveFormat, error) {
	name := strings.ToLower(filepath.Base(file))
	switch {
	case strings.HasSuffix(name, ".zip"):
		return archiveFormat{zip: true}, nil
	case strings.HasSuffix(name, ".tar"):
		return archiveFormat{}, nil
	case strings.HasSuffix(name, ".tar.gz"), strings.HasSuffix(name, ".tgz"):
		return archiveFormat{compression: "gz"}, nil
	case strings.HasSuffix(name, ".tar.zst"), strings.HasSuffix(name, ".tzst"):
		return archiveFormat{compression: "zst"}, nil
	case strings.HasSuffix(name, ".tar.bz2"), strings.HasSuffix(name, ".tbz2"):
		return archiveFormat{compression: "bz2"}, nil
	}
	return archiveFormat{}, fmt.Errorf("archive format of file '%s' is not supported (zip, tar, tar.gz, tgz, tar.zst, tzst, tar.bz2, tbz2)", file)
}

// Archive writes the dir contents (or the file) to the archive, with the entries at its root, keeping file modes and symlinks.
func Archive(sourcePath, targetFile string) error {
	if !pathx.Exists(sourcePath) {
		return fmt.Errorf("cannot archive path '%s' to file '%s' as source path does not exist", sourcePath, targetFile)
	}
	format, err := archiveFormatOf(targetFile)
	if err != nil {
		return err
	}
	if format.compression == "bz2" {
		return fmt.Errorf("cannot archive path '%s' to file '%s' as bzip2 is supported for unarchiving only", sourcePath, targetFile)
	}
	if err := pathx.Ensure(filepath.Dir(targetFile)); err != nil {
		return err
	}
	var sourcePaths []string
	if pathx.IsDir(sourcePath) {
		sourceDirEntries, err := os.ReadDir(sourcePath)
		if err != nil {
			return fmt.Errorf("cannot read dir '%s' to be archived to file '%s': %w", sourcePath, targetFile, err)
		}
		sourcePaths = lo.Map(sourceDirEntries, func(e os.DirEntry, _ int) string {
			return pathx.Canonical(fmt.Sprintf("%s/%s", sourcePath, e.Name()))
		})
	} else {
		sourcePaths = []string{sourcePath}
	}
	if format.zip {
		err = writeZip(sourcePaths, targetFile)
	} else {
		err = writeTar(sourcePaths, targetFile, format.compression)
	}
	if err != nil {
		return fmt.Errorf("cannot archive path '%s' to file '%s': %w", sourcePath, targetFile, err)
	}
	return nil
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
// Entries that would end up outside the dir, write through a symlink or link outside it are refused ("zip slip").
func Unarchive(sourceFile string, targetDir string) error {
	if !pathx.Exists(sourceFile) {
		return fmt.Errorf("cannot unarchive file '%s' to dir '%s' as source file does not exist", sourceFile, targetDir)
	}
	format, err := archiveFormatOf(sourceFile)
	if err != nil {
		return err
	}
	if err := pathx.Ensure(targetDir); err != nil {
		return err
	}
	extractor := newArchiveExtractor(targetDir)
	if format.zip {
		err = readZip(sourceFile, extractor)
	} else {
		err = readTar(sourceFile, format.compression, extractor)
	}
	if err != nil {
		return fmt.Errorf("cannot unarchive file '%s' to dir '%s': %w", sourceFile, targetDir, err)
	}
	return nil
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

// symlinkEntry creates a symlink whose target, relative to the symlink, stays within the dir.
func (e *archiveExtractor) symlinkEntry(name string, target string) error {
	path, err := e.entryPath(name)
	if err != nil {
		return err
	}
	target = filepath.FromSlash(pathx.Normalize(target))
	if filepath.IsAbs(target) || filepath.VolumeName(target) != "" || !pathx.IsWithin(e.dir, filepath.Join(filepath.Dir(path), target)) {
		return fmt.Errorf("symlink '%s' points to '%s' outside of dir '%s'", name, target, e.dir)
	}
	return os.Symlink(target, path)
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
