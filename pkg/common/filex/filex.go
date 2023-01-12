package filex

import (
	"crypto/md5"
	"fmt"
	"github.com/codingsince1985/checksum"
	"github.com/wttech/aemc/pkg/common/pathx"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

func ChecksumPath(path string, ignorePatterns []string) (string, error) {
	dir, err := pathx.IsDirStrict(path)
	if err != nil {
		return "", err
	}
	if dir {
		return ChecksumDir(path, ignorePatterns)
	} else {
		return ChecksumFile(path)
	}
}
func ChecksumFile(file string) (string, error) {
	return checksum.MD5sum(file)
}

func ChecksumDir(dir string, ignorePatterns []string) (string, error) {
	hash := md5.New()
	pathMatcher := pathx.NewMatcher(ignorePatterns)
	if err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && pathMatcher.Match(path) {
			fileSum, err := ChecksumFile(path)
			if err != nil {
				return err
			}
			_, _ = io.WriteString(hash, fileSum)
		}
		return nil
	}); err != nil {
		return "", fmt.Errorf("cannot find files to calculate checksum of directory '%s': %w", dir, err)
	}
	dirSum := fmt.Sprintf("%x", hash.Sum(nil))
	return dirSum, nil
}
