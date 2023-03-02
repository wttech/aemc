package filex

import (
	"crypto/md5"
	"fmt"
	"github.com/codingsince1985/checksum"
	cpy "github.com/otiai10/copy"
	"github.com/wttech/aemc/pkg/common/pathx"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func Write(path string, data []byte) error {
	err := pathx.Ensure(filepath.Dir(path))
	if err != nil {
		return err
	}
	err = os.WriteFile(path, data, 0755)
	if err != nil {
		return fmt.Errorf("cannot write to file '%s': %w", path, err)
	}
	return nil
}

func WriteString(path string, text string) error {
	err := pathx.Ensure(filepath.Dir(path))
	if err != nil {
		return err
	}
	return Write(path, []byte(text))
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

func ReadString(path string) (string, error) {
	bytes, err := Read(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func AppendString(path string, text string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot append to file '%s': %w", path, err)
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		return fmt.Errorf("cannot append to file '%s': %w", path, err)
	}
	return nil
}

func AmendString(file string, updater func(string) string) error {
	content, err := ReadString(file)
	if err != nil {
		return err
	}
	return WriteString(file, updater(content))
}

var fileCopyBufferSize = 4 * 1024 // 4 kB <https://stackoverflow.com/a/3034155>

func Copy(sourcePath, destinationPath string, overwrite bool) error {
	sourceStat, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	if !sourceStat.Mode().IsRegular() {
		return fmt.Errorf("cannot copy file from '%s' to '%s' as source does not exist (or is not a regular file)", sourcePath, destinationPath)
	}
	if err := pathx.Ensure(filepath.Dir(destinationPath)); err != nil {
		return err
	}
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	if !overwrite {
		_, err = os.Stat(destinationPath)
		if err == nil {
			return fmt.Errorf("cannot copy file from '%s' to '%s' as destination already exists", sourcePath, destinationPath)
		}
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

func CopyDir(src, dest string) error {
	exists, err := pathx.ExistsStrict(src)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("cannot copy dir from '%s' to '%s' as source does not exist", src, dest)
	}
	if err := pathx.Ensure(filepath.Dir(dest)); err != nil {
		return err
	}
	return cpy.Copy(src, dest)
}

func ChecksumPath(path string, pathIgnored []string) (string, error) {
	dir, err := pathx.IsDirStrict(path)
	if err != nil {
		return "", err
	}
	if dir {
		return ChecksumDir(path, pathIgnored)
	}
	return ChecksumFile(path)
}

func ChecksumPaths(paths []string, pathIgnored []string) (string, error) {
	hash := md5.New()
	for _, path := range paths {
		pathSum, err := ChecksumPath(path, pathIgnored)
		if err != nil {
			return "", err
		}
		_, _ = io.WriteString(hash, pathSum)
	}
	dirsSum := fmt.Sprintf("%x", hash.Sum(nil))
	return dirsSum, nil
}

func ChecksumFile(file string) (string, error) {
	return checksum.MD5sum(file)
}

func ChecksumDir(dir string, pathIgnored []string) (string, error) {
	hash := md5.New()
	pathIgnoreMatcher := pathx.NewIgnoreMatcher(pathIgnored)
	if err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && !pathIgnoreMatcher.Match(path) {
			fileSum, err := ChecksumFile(path)
			if err != nil {
				return err
			}
			_, _ = io.WriteString(hash, path)
			_, _ = io.WriteString(hash, fileSum)
		}
		return nil
	}); err != nil {
		return "", fmt.Errorf("cannot find files to calculate checksum of directory '%s': %w", dir, err)
	}
	dirSum := fmt.Sprintf("%x", hash.Sum(nil))
	return dirSum, nil
}

func Equals(file1 string, file2 string) (bool, error) {
	if !pathx.Exists(file1) || !pathx.Exists(file2) {
		return false, nil
	}
	sum1, err := ChecksumFile(file1)
	if err != nil {
		return false, err
	}
	sum2, err := ChecksumFile(file2)
	if err != nil {
		return false, err
	}
	return sum1 == sum2, nil
}
