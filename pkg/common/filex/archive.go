package filex

import (
	"fmt"
	"github.com/mholt/archiver/v3"
	"github.com/wttech/aemc/pkg/common/pathx"
	"os"
	"path/filepath"
)

func Archive(sourceDir, targetFile string) error {
	err := pathx.Ensure(filepath.Dir(targetFile))
	if err != nil {
		return err
	}
	err = archiver.Archive([]string{sourceDir}, targetFile)
	if err != nil {
		return fmt.Errorf("cannot archive dir '%s' to file '%s': %w", sourceDir, targetFile, err)
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
