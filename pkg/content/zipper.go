package content

import (
	"archive/zip"
	"fmt"
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
	err = compress(sourcePath, targetFile)
	if err != nil {
		return fmt.Errorf("cannot archive dir '%s' to file '%s': %w", sourcePath, targetFile, err)
	}
	return nil
}

func Unarchive(sourceFile string, targetDir string) error {
	if !pathx.Exists(sourceFile) {
		return fmt.Errorf("cannot unarchive file '%s' to dir '%s' as source file does not exist", sourceFile, targetDir)
	}
	if err := pathx.Ensure(targetDir); err != nil {
		return err
	}
	if err := extract(sourceFile, targetDir); err != nil {
		return fmt.Errorf("cannot unarchive file '%s' to dir '%s': %w", sourceFile, targetDir, err)
	}
	return nil
}

func compress(src string, dest string) error {
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	zipWriter := zip.NewWriter(destFile)
	defer zipWriter.Close()

	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func extract(src string, dest string) error {
	zipReader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		filePath := filepath.Join(dest, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		inFile, err := file.Open()
		if err != nil {
			return err
		}
		defer inFile.Close()

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, inFile)
		if err != nil {
			return err
		}
	}

	return nil
}
