package pkg

import (
	"testing"
)

func TestDetermineFilterRoot(t *testing.T) {
	tests := []struct {
		path, expected string
	}{
		{"/somepath/jcr_root/content/my_site", "/content/my_site"},
		{"/somepath/jcr_root/content/my_site/_cq_path", "/content/my_site/cq:path"},
		{"/somepath/jcr_root/content/my_site/__cq_path", "/content/my_site/_cq_path"},
		{"/somepath/jcr_root/content/my_site/__abc_path", "/content/my_site/_abc_path"},
		{"/somepath/jcr_root/content/my_site/__path", "/content/my_site/_path"},
		{"/somepath/jcr_root/content/my_site/___path", "/content/my_site/__path"},
		{"\\somepath\\jcr_root\\content\\my_site", "/content/my_site"},
		{"\\somepath\\jcr_root\\content\\my_site\\_cq_path", "/content/my_site/cq:path"},
		{"/somepath/jcr_root/content/my_site/.content.xml", "/content/my_site/jcr:content"},
		{"/somepath/jcr_root/content/my_app/_cq_dialog/.content.xml", "/content/my_app/cq:dialog"},
		{"/somepath/jcr_root/content/my_app/_cq_dialog.xml", "/content/my_app/cq:dialog"},
		{"/somepath/jcr_root/content/my_conf/workflow.xml", "/content/my_conf/workflow"},
		{"/somepath/jcr_root/content/my_app/__cq_dialog/.content.xml", "/content/my_app/_cq_dialog/jcr:content"},
	}
	for _, test := range tests {
		actual := DetermineFilterRoot(test.path)
		if actual != test.expected {
			t.Errorf("DetermineFilterRoot(%s) = %s; want %s", test.path, actual, test.expected)
		}
	}
}
