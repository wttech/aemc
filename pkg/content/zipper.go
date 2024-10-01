/*
This is a simple library for archiving files. It allows you to compress a single directory into a zip archive file
and extract files from a zip archive file. The library creates zip archive files that are compatible
with AEM package file specifications.

This library is limited to basic zip operations, whereas github.com/mholt/archiver/v3 provides more advanced features
but may not always produce zip archive files compatible with AEM package file specifications. Additionally,
when using github.com/mholt/archiver/v3, extracted directories may have insufficient permissions,
requiring additional handling.
*/

package content

import (
	zipper "archive/zip"
	"fmt"
	"github.com/wttech/aemc/pkg/common/pathx"
	"io"
	"os"
	"path/filepath"
)

func Zip(sourcePath, targetFile string) error {
	if !pathx.Exists(sourcePath) {
		return fmt.Errorf("cannot zip path '%s' to file '%s' as source path does not exist", sourcePath, targetFile)
	}
	err := pathx.Ensure(filepath.Dir(targetFile))
	if err != nil {
		return err
	}
	err = zip(sourcePath, targetFile)
	if err != nil {
		return fmt.Errorf("cannot zip dir '%s' to file '%s': %w", sourcePath, targetFile, err)
	}
	return nil
}

func Unzip(sourceFile string, targetDir string) error {
	if !pathx.Exists(sourceFile) {
		return fmt.Errorf("cannot unzip file '%s' to dir '%s' as source file does not exist", sourceFile, targetDir)
	}
	if err := pathx.Ensure(targetDir); err != nil {
		return err
	}
	if err := unzip(sourceFile, targetDir); err != nil {
		return fmt.Errorf("cannot unzip file '%s' to dir '%s': %w", sourceFile, targetDir, err)
	}
	return nil
}

func zip(src string, dest string) error {
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	zipWriter := zipper.NewWriter(destFile)
	defer zipWriter.Close()

	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zipper.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if header.Name == "." {
			return nil
		}
		header.Name = pathx.Normalize(header.Name)

		if info.IsDir() {
			header.Name += "/"
			header.Method = zipper.Store
		} else {
			header.Method = zipper.Deflate
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

func unzip(src string, dest string) error {
	zipReader, err := zipper.OpenReader(src)
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
