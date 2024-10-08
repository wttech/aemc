//go:build int_test

package pkg_test

import (
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/content"
	"os"
	"path/filepath"
	"testing"
)

func TestDetermineFilterRoot(t *testing.T) {
	workDir := pathx.RandomDir(os.TempDir(), "filter_root")
	defer func() { _ = pathx.DeleteIfExists(workDir) }()
	if err := copyFiles("int_test_content/repo", workDir); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path, expected string
	}{
		//{"/content/my_site", "/content/my_site"},
		//{"/content/my_site/_cq_path", "/content/my_site/cq:path"},
		//{"/content/my_site/_xmpBJ_path", "/content/my_site/xmpBJ:path"},
		//{"/content/my_site/_s7sitecatalyst_path", "/content/my_site/s7sitecatalyst:path"},
		//{"/content/my_site/adobe_dam%3apath", "/content/my_site/adobe_dam:path"},
		//{"/content/my_site/_cq__sub_path_", "/content/my_site/cq:_sub_path_"},
		//{"/content/my_site/__cq_path", "/content/my_site/_cq_path"},
		//{"/content/my_site/__abc_path", "/content/my_site/_abc_path"},
		//{"/content/my_site/__path", "/content/my_site/_path"},
		{"/content/my_site/___path", "/content/my_site/__path"},
		{"content\\my_site", "/content/my_site"},
		{"content\\my_site\\_cq_path", "/content/my_site/cq:path"},
		{"/content/my_app/_cq_dialog/.content.xml", "/content/my_app/cq:dialog"},
		{"/content/my_app/_cq_dialog.xml", "/content/my_app/cq:dialog"},
		{"/content/my_conf/workflow.xml", "/content/my_conf/workflow"},
		{"/content/dam/my_site/image.png", "/content/dam/my_site/image.png"},
		{"/conf/my_site/_sling_configs/com.config.ImageConfig", "/conf/my_site/sling:configs/com.config.ImageConfig"},
		{"/conf/my_site/_sling_configs/com.config.ImageConfig/_jcr_content", "/conf/my_site/sling:configs/com.config.ImageConfig/jcr:content"},
		{"/apps/mysite/components/helloworld/_cq_template/.content.xml", "/apps/mysite/components/helloworld/cq:template"},
		{"/apps/mysite/components/helloworld/_cq_template.xml", "/apps/mysite/components/helloworld/cq:template"},
		{"/apps/mysite/components/helloworld/_cq_editConfig.xml", "/apps/mysite/components/helloworld/cq:editConfig"},
		{"/apps/mysite/components/helloworld/helloworld.html", "/apps/mysite/components/helloworld/helloworld.html"},
		{"/conf/mysite/_sling_configs/com.mysite.pdfviewer.PdfViewerCaConfig", "/conf/mysite/sling:configs/com.mysite.pdfviewer.PdfViewerCaConfig"},
		{"/conf/mysite/_sling_configs/com.mysite.pdfviewer.PdfViewerCaConfig/.content.xml", "/conf/mysite/sling:configs/com.mysite.pdfviewer.PdfViewerCaConfig/jcr:content"},
		{"/conf/mysite/_sling_configs/.content.xml", "/conf/mysite/sling:configs"},
		{"/conf/mysite/_sling_configs", "/conf/mysite/sling:configs"},
		{"/content/mysite/us/en/.content.xml", "/content/mysite/us/en/jcr:content"},
	}
	for _, test := range tests {
		path := filepath.Join(workDir, content.JCRRoot, test.path)
		actual := pkg.DetermineFilterRoot(path)
		if actual != test.expected {
			t.Errorf("DetermineFilterRoot(%s) = %s; want %s", test.path, actual, test.expected)
		}
	}
}
