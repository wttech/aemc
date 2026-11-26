package filex

import (
	"archive/tar"
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeTestTree(t *testing.T, dir string) {
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "crx-quickstart", "bin"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "crx-quickstart", "bin", "start"), []byte("#!/bin/sh\n"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "license.properties"), []byte("license"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "empty"), 0755))
	if runtime.GOOS != "windows" {
		require.NoError(t, os.Symlink("bin/start", filepath.Join(dir, "crx-quickstart", "start-link")))
	}
}

func TestArchiveRoundTrip(t *testing.T) {
	for _, ext := range []string{"zip", "tar", "tar.gz", "tgz", "aemb.tar.zst"} {
		t.Run(ext, func(t *testing.T) {
			source := filepath.Join(t.TempDir(), "source")
			writeTestTree(t, source)
			file := filepath.Join(t.TempDir(), "archive."+ext)
			require.NoError(t, Archive(source, file))

			target := filepath.Join(t.TempDir(), "target")
			require.NoError(t, Unarchive(file, target))

			content, err := os.ReadFile(filepath.Join(target, "license.properties"))
			require.NoError(t, err)
			require.Equal(t, "license", string(content), "dir contents are at the archive root")
			require.DirExists(t, filepath.Join(target, "empty"))
			if runtime.GOOS == "windows" {
				return
			}
			info, err := os.Stat(filepath.Join(target, "crx-quickstart", "bin", "start"))
			require.NoError(t, err)
			require.Equal(t, os.FileMode(0755), info.Mode().Perm(), "executable stays executable")
			link, err := os.Readlink(filepath.Join(target, "crx-quickstart", "start-link"))
			require.NoError(t, err)
			require.Equal(t, "bin/start", link)
		})
	}
}

func TestArchiveUnsupportedFormat(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source")
	writeTestTree(t, source)
	require.Error(t, Archive(source, filepath.Join(t.TempDir(), "archive.rar")))
}

func writeTestTar(t *testing.T, headers ...*tar.Header) string {
	file := filepath.Join(t.TempDir(), "evil.tar")
	f, err := os.Create(file)
	require.NoError(t, err)
	defer f.Close()
	tw := tar.NewWriter(f)
	for _, h := range headers {
		if h.Typeflag == tar.TypeReg {
			h.Size = int64(len("evil"))
		}
		require.NoError(t, tw.WriteHeader(h))
		if h.Typeflag == tar.TypeReg {
			_, err := tw.Write([]byte("evil"))
			require.NoError(t, err)
		}
	}
	require.NoError(t, tw.Close())
	return file
}

func TestUnarchiveRefusesEntriesLeavingTargetDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks need privileges on Windows")
	}
	cases := map[string][]*tar.Header{
		"parent path":      {{Name: "../evil.txt", Typeflag: tar.TypeReg, Mode: 0644}},
		"absolute path":    {{Name: "/tmp/evil.txt", Typeflag: tar.TypeReg, Mode: 0644}},
		"symlink outside":  {{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "../.."}},
		"absolute symlink": {{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "/tmp"}},
		"write through symlink": {
			{Name: "self", Typeflag: tar.TypeSymlink, Linkname: "."},
			{Name: "self/evil.txt", Typeflag: tar.TypeReg, Mode: 0644},
		},
		"hard link outside": {{Name: "link", Typeflag: tar.TypeLink, Linkname: "../evil.txt"}},
	}
	for name, headers := range cases {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			target := filepath.Join(parent, "target")
			require.Error(t, Unarchive(writeTestTar(t, headers...), target))
			require.NoFileExists(t, filepath.Join(parent, "evil.txt"))
		})
	}
}

func TestUnarchiveZipRefusesParentPath(t *testing.T) {
	file := filepath.Join(t.TempDir(), "evil.zip")
	f, err := os.Create(file)
	require.NoError(t, err)
	zw := zip.NewWriter(f)
	w, err := zw.Create("../evil.txt")
	require.NoError(t, err)
	_, err = w.Write([]byte("evil"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	require.NoError(t, f.Close())

	parent := t.TempDir()
	require.Error(t, Unarchive(file, filepath.Join(parent, "target")))
	require.NoFileExists(t, filepath.Join(parent, "evil.txt"))
}

func TestUnarchiveDirsWithoutPermissions(t *testing.T) {
	file := writeTestTar(t,
		&tar.Header{Name: "dir/", Typeflag: tar.TypeDir, Mode: 0},
		&tar.Header{Name: "dir/file.txt", Typeflag: tar.TypeReg, Mode: 0644},
	)
	target := filepath.Join(t.TempDir(), "target")
	require.NoError(t, Unarchive(file, target))
	require.FileExists(t, filepath.Join(target, "dir", "file.txt"), "dirs get 0755 whatever the archive says")
}
