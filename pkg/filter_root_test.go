//go:build int_test

package pkg_test

import (
	"github.com/wttech/aemc/pkg"
	"testing"
)

func TestDetermineFilterRoot(t *testing.T) {
	tests := []struct {
		path, expected string
	}{
		{"/somepath/jcr_root/content/my_site", "/content/my_site"},
		{"/somepath/jcr_root/content/my_site/_cq_path", "/content/my_site/cq:path"},
		{"/somepath/jcr_root/content/my_site/_xmpBJ_path", "/content/my_site/xmpBJ:path"},
		{"/somepath/jcr_root/content/my_site/_s7sitecatalyst_path", "/content/my_site/s7sitecatalyst:path"},
		{"/somepath/jcr_root/content/my_site/adobe_dam%3apath", "/content/my_site/adobe_dam:path"},
		{"/somepath/jcr_root/content/my_site/_cq__sub_path_", "/content/my_site/cq:_sub_path_"},
		{"/somepath/jcr_root/content/my_site/__cq_path", "/content/my_site/_cq_path"},
		{"/somepath/jcr_root/content/my_site/__abc_path", "/content/my_site/_abc_path"},
		{"/somepath/jcr_root/content/my_site/__path", "/content/my_site/_path"},
		{"/somepath/jcr_root/content/my_site/___path", "/content/my_site/__path"},
		{"\\somepath\\jcr_root\\content\\my_site", "/content/my_site"},
		{"\\somepath\\jcr_root\\content\\my_site\\_cq_path", "/content/my_site/cq:path"},
		{"/somepath/jcr_root/content/my_app/_cq_dialog/.content.xml", "/content/my_app/cq:dialog"},
		{"/somepath/jcr_root/content/my_app/_cq_dialog.xml", "/content/my_app/cq:dialog"},
		{"/somepath/jcr_root/content/my_conf/workflow.xml", "/content/my_conf/workflow"},
		{"/somepath/jcr_root/content/dam/my_site/image.png", "/content/dam/my_site/image.png"},
		{"/somepath/jcr_root/conf/my_site/_sling_configs/com.config.ImageConfig", "/conf/my_site/sling:configs/com.config.ImageConfig"},
		{"/somepath/jcr_root/conf/my_site/_sling_configs/com.config.ImageConfig/_jcr_content", "/conf/my_site/sling:configs/com.config.ImageConfig/jcr:content"},
	}
	for _, test := range tests {
		actual := pkg.DetermineFilterRoot(test.path)
		if actual != test.expected {
			t.Errorf("DetermineFilterRoot(%s) = %s; want %s", test.path, actual, test.expected)
		}
	}
}
