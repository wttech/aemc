package filex

import (
	"fmt"
	"github.com/codingsince1985/checksum"
	"github.com/mholt/archiver/v3"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func Write(path string, text string) error {
	err := pathx.Ensure(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("cannot ensure path '%s': %w", path, err)
	}
	err = os.WriteFile(path, []byte(text), 0755)
	if err != nil {
		return fmt.Errorf("cannot write to file '%s': %w", path, err)
	}
	return nil
}

func Read(path string) ([]byte, error) {
	if !pathx.Exists(path) {
		return nil, fmt.Errorf("cannot read file as it does not exist at path '%s'", path)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file '%s': %w", path, err)
	}
	return bytes, nil
}

var fileCopyBufferSize = 4 * 1024 // 4 kB <https://stackoverflow.com/a/3034155>

func Copy(sourcePath, destinationPath string) error {
	sourceStat, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	if !sourceStat.Mode().IsRegular() {
		return fmt.Errorf("cannot copy file from '%s' to '%s' as source does not exist (or is not a regular file)", sourcePath, destinationPath)
	}
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	_, err = os.Stat(destinationPath)
	if err == nil {
		return fmt.Errorf("cannot copy file from '%s' to '%s' as destination already exists", sourcePath, destinationPath)
	}
	destination, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destination.Close()
	buf := make([]byte, fileCopyBufferSize)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return err
}

func ChecksumPath(path string, dirIncludes []string, dirExcludes []string) (string, error) {
	dir, err := pathx.IsDirStrict(path)
	if err != nil {
		return "", err
	}
	if dir {
		return ChecksumDir(path, dirIncludes, dirExcludes)
	} else {
		return ChecksumFile(path)
	}
}
func ChecksumFile(file string) (string, error) {
	return checksum.MD5sum(file)
}

func ChecksumDir(dir string, includes []string, excludes []string) (string, error) {
	var tarFiles []string
	err := filepath.WalkDir(dir, func(dirPath string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if (len(includes) == 0 || stringsx.MatchSomePattern(dirPath, includes)) && !stringsx.MatchSomePattern(dirPath, excludes) {
			tarFiles = append(tarFiles, dirPath)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("cannot find files to calculate checksum of directory '%s': %w", dir, err)
	}
	dirTarFile := filepath.Dir(dir) + "/" + filepath.Base(dir) + ".tar"
	defer func() {
		_ = pathx.DeleteIfExists(dirTarFile)
	}()
	sort.Strings(tarFiles)
	/* TODO remove it
	for _, f := range tarFiles {
		println(f)
	}
	*/
	// TODO calculate hashes separately; dump to file and compare what is changing
	// TODO maybe dedicated command aem file track --dir '.' --checksum-file 'aem/home/build.log --output-value 'changed' ?
	// TODO or: aem file updated --input-paths 'core/**,ui.apps/**' --output-file all/target/all-*.zip --checksum-dir=aem/home/build (output file optional)

	err = archiver.Archive(tarFiles, dirTarFile)
	if err != nil {
		return "", fmt.Errorf("cannot make a temporary archive to calculate checksum of directory '%s': %w", dir, err)
	}
	return ChecksumFile(dirTarFile)
}
