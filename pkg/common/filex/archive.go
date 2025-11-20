package filex

import (
	"bytes"
	"fmt"
	"github.com/mholt/archiver/v3"
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/pathx"
	"io"
	"os"
	"path/filepath"
)

func Archive(sourcePath, targetFile string) error {
	if !pathx.Exists(sourcePath) {
		return fmt.Errorf("cannot archive path '%s' to file '%s' as source path does not exist", sourcePath, targetFile)
	}
	err := pathx.Ensure(filepath.Dir(targetFile))
	if err != nil {
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
	err = archiver.Archive(sourcePaths, targetFile)
	if err != nil {
		return fmt.Errorf("cannot archive dir '%s' to file '%s': %w", sourcePath, targetFile, err)
	}
	return nil
}

func ArchiveWithChanged(sourceDir, targetFile string) (bool, error) {
	if pathx.Exists(targetFile) {
		return false, nil
	}
	targetTmpFile := filepath.Dir(targetFile) + "/tmp-" + filepath.Base(targetFile)
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

func Unarchive(sourceFile string, targetDir string) error {
	if !pathx.Exists(sourceFile) {
		return fmt.Errorf("cannot unarchive file '%s' to dir '%s' as source file does not exist", sourceFile, targetDir)
	}
	if err := pathx.Ensure(targetDir); err != nil {
		return err
	}
	if err := archiver.Unarchive(sourceFile, targetDir); err != nil {
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
