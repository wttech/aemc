package filex

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func writeZip(sourcePaths []string, targetFile string) (err error) {
	file, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	buffered := bufio.NewWriterSize(file, archiveBufferSize)
	zw := zip.NewWriter(buffered)
	for _, sourcePath := range sourcePaths {
		if err := walkArchived(sourcePath, func(path string, name string, info os.FileInfo, link string) error {
			return writeZipEntry(zw, path, name, info, link)
		}); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return buffered.Flush()
}

func writeZipEntry(zw *zip.Writer, path string, name string, info os.FileInfo, link string) error {
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("cannot make archive header of '%s': %w", path, err)
	}
	header.Name = name
	if info.IsDir() {
		header.Name += "/"
	} else {
		header.Method = zip.Deflate
	}
	w, err := zw.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("cannot write archive header of '%s': %w", path, err)
	}
	switch {
	case info.IsDir():
		return nil
	case link != "":
		_, err = io.WriteString(w, link) // zip keeps the symlink target as the entry content
		return err
	case info.Mode().IsRegular():
		return copyFileTo(w, path)
	default:
		return fmt.Errorf("cannot archive '%s' as its file type is not supported", path)
	}
}

func readZip(sourceFile string, extractor *archiveExtractor) error {
	zr, err := zip.OpenReader(sourceFile)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, entry := range zr.File {
		if err := readZipEntry(entry, extractor); err != nil {
			return fmt.Errorf("cannot extract entry '%s': %w", entry.Name, err)
		}
	}
	return nil
}

func readZipEntry(entry *zip.File, extractor *archiveExtractor) error {
	mode := entry.Mode()
	if mode.IsDir() || strings.HasSuffix(entry.Name, "/") || strings.HasSuffix(entry.Name, "\\") {
		return extractor.dirEntry(entry.Name)
	}
	content, err := entry.Open()
	if err != nil {
		return err
	}
	defer content.Close()
	if mode&os.ModeSymlink != 0 {
		target, err := io.ReadAll(io.LimitReader(content, 4096))
		if err != nil {
			return err
		}
		return extractor.symlinkEntry(entry.Name, string(target))
	}
	if !mode.IsRegular() {
		return fmt.Errorf("entry type '%s' is not supported", mode.Type())
	}
	return extractor.fileEntry(entry.Name, mode, content)
}
