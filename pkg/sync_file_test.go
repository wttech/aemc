//go:build int_test

package pkg_test

import (
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/content"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetermineSyncFile(t *testing.T) {
	workDir := pathx.RandomDir(os.TempDir(), "sync_file")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := copyFiles("int_test_content/sync_file", workDir); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path, expected string
	}{
		{"/content/mysite/us/en/.content.xml", "/content/mysite/us/en/.content.xml"},
		{"/apps/mysite/components/helloworld/_cq_template/.content.xml", "/apps/mysite/components/helloworld/_cq_template/.content.xml"},
		{"/apps/mysite/components/helloworld/_cq_template.xml", "/apps/mysite/components/helloworld/_cq_template/.content.xml"},
		{"/apps/mysite/components/helloworld/_cq_editConfig.xml", "/apps/mysite/components/helloworld/_cq_editConfig.xml"},
		{"/apps/mysite/components/helloworld/helloworld.html", "/apps/mysite/components/helloworld/helloworld.html"},
		{"/conf/mysite/_sling_configs/com.mysite.pdfviewer.PdfViewerCaConfig/.content.xml", "/conf/mysite/_sling_configs/com.mysite.pdfviewer.PdfViewerCaConfig/.content.xml"},
		{"/conf/mysite/_sling_configs/.content.xml", "/conf/mysite/_sling_configs/.content.xml"},
		{"/content/mysite/us/en/.content.xml", "/content/mysite/us/en/.content.xml"},
	}
	for _, test := range tests {
		path := filepath.Join(workDir, content.JCRRoot, test.path)
		expected := filepath.Join(workDir, content.JCRRoot, test.expected)
		actual := pkg.DetermineSyncFile(workDir, path)
		if actual != expected {
			_, jcrPath, _ := strings.Cut(actual, content.JCRRoot)
			t.Errorf("DetermineSyncFile(%s) = %s; want %s", test.path, jcrPath, test.expected)
		}
	}
}
