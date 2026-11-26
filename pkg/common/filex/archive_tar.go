package filex

import (
	"archive/tar"
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

const archiveBufferSize = 1 << 20

func writeTar(sourcePaths []string, targetFile string, compression string) (err error) {
	file, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	buffered := bufio.NewWriterSize(file, archiveBufferSize)

	var compressed io.WriteCloser
	switch compression {
	case "gz":
		compressed = gzip.NewWriter(buffered)
	case "zst":
		if compressed, err = zstd.NewWriter(buffered); err != nil {
			return err
		}
	default:
		compressed = nopWriteCloser{buffered}
	}
	tw := tar.NewWriter(compressed)
	for _, sourcePath := range sourcePaths {
		if err := walkArchived(sourcePath, func(path string, name string, info os.FileInfo, link string) error {
			return writeTarEntry(tw, path, name, info, link)
		}); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := compressed.Close(); err != nil {
		return err
	}
	return buffered.Flush()
}

func writeTarEntry(tw *tar.Writer, path string, name string, info os.FileInfo, link string) error {
	header, err := tar.FileInfoHeader(info, link)
	if err != nil {
		return fmt.Errorf("cannot make archive header of '%s': %w", path, err)
	}
	header.Name = name
	if info.IsDir() {
		header.Name += "/"
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("cannot write archive header of '%s': %w", path, err)
	}
	if header.Typeflag != tar.TypeReg {
		return nil
	}
	return copyFileTo(tw, path)
}

func readTar(sourceFile string, compression string, extractor *archiveExtractor) error {
	file, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer file.Close()
	buffered := bufio.NewReaderSize(file, archiveBufferSize)

	var decompressed io.Reader
	switch compression {
	case "gz":
		gz, err := gzip.NewReader(buffered)
		if err != nil {
			return err
		}
		defer gz.Close()
		decompressed = gz
	case "zst":
		zs, err := zstd.NewReader(buffered)
		if err != nil {
			return err
		}
		defer zs.Close()
		decompressed = zs
	case "bz2":
		decompressed = bzip2.NewReader(buffered)
	default:
		decompressed = buffered
	}
	tr := tar.NewReader(decompressed)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := readTarEntry(tr, header, extractor); err != nil {
			return fmt.Errorf("cannot extract entry '%s': %w", header.Name, err)
		}
	}
}

func readTarEntry(tr *tar.Reader, header *tar.Header, extractor *archiveExtractor) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return extractor.dirEntry(header.Name)
	case tar.TypeReg:
		return extractor.fileEntry(header.Name, header.FileInfo().Mode(), tr)
	case tar.TypeSymlink:
		return extractor.symlinkEntry(header.Name, header.Linkname)
	case tar.TypeLink:
		return extractor.hardLinkEntry(header.Name, header.Linkname)
	case tar.TypeXGlobalHeader:
		return nil // e.g. comments of archives made by 'git archive'
	default:
		return fmt.Errorf("entry type '%c' is not supported", header.Typeflag)
	}
}

// walkArchived visits the source path and everything under it, naming entries relative to its parent dir,
// so that the source path itself is at the archive root. Symlinks are archived as such, not followed.
func walkArchived(sourcePath string, visit func(path string, name string, info os.FileInfo, link string) error) error {
	baseDir := filepath.Dir(sourcePath)
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		link := ""
		if info.Mode()&os.ModeSymlink != 0 {
			if link, err = os.Readlink(path); err != nil {
				return err
			}
			link = filepath.ToSlash(link)
		}
		return visit(path, filepath.ToSlash(rel), info, link)
	})
}

func copyFileTo(w io.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(w, file); err != nil {
		return fmt.Errorf("cannot archive file '%s': %w", path, err)
	}
	return nil
}

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }
