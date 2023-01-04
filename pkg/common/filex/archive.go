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
	if pathx.Exists(targetTmpFile) {
		err := pathx.Delete(targetTmpFile)
		if err != nil {
			return false, fmt.Errorf("cannot delete temporary archive file '%s': %w", targetTmpFile, err)
		}
	}
	err := Archive(sourceDir, targetTmpFile)
	if err != nil {
		return false, err
	}
	err = os.Rename(targetTmpFile, targetFile)
	if err != nil {
		return false, fmt.Errorf("cannot move temporary archive file '%s' to target one '%s': %w", targetTmpFile, targetFile, err)
	}
	return true, nil
}

func Unarchive(sourceFile string, targetDir string) error {
	err := pathx.Ensure(targetDir)
	if err != nil {
		return err
	}
	err = archiver.Unarchive(sourceFile, targetDir)
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
	if pathx.Exists(targetTmpDir) {
		err := pathx.Delete(targetTmpDir)
		if err != nil {
			return false, fmt.Errorf("cannot delete unarchive temporary dir '%s': %w", targetTmpDir, err)
		}
	}
	err := Unarchive(sourceFile, targetTmpDir)
	if err != nil {
		return false, err
	}
	err = os.Rename(targetTmpDir, targetDir)
	if err != nil {
		return false, fmt.Errorf("cannot move unarchived temporary dir '%s' to target one '%s': %w", targetTmpDir, targetDir, err)
	}
	return true, nil
}
