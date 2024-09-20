package test

import (
	"github.com/wttech/aemc/pkg"
	"testing"
)

func TestDetermineCleanFile(t *testing.T) {
	tests := []struct {
		path, expected string
	}{
		{"/somepath/jcr_root/content/my_file.xml", "/somepath/jcr_root/content/my_file.xml"},
		{"/somepath/jcr_root/content/_jcr_content/my_file.xml", "/somepath/jcr_root/content/_jcr_content/my_file.xml"},
		{"/somepath/jcr_root/content/_jcr_content/_cq_file.xml", "/somepath/jcr_root/content/_jcr_content/_cq_file/.content.xml"},
		{"/somepath/jcr_root/content/_cq_file.xml", "/somepath/jcr_root/content/_cq_file/.content.xml"},
		{"/somepath/jcr_root/content/_cq_file/.content.xml", "/somepath/jcr_root/content/_cq_file/.content.xml"},
		{"/somepath/jcr_root/content/.content.xml", "/somepath/jcr_root/content/.content.xml"},
	}
	for _, test := range tests {
		actual := pkg.DetermineCleanFile(test.path)
		if actual != test.expected {
			t.Errorf("DetermineCleanFile(%s) = %s; want %s", test.path, actual, test.expected)
		}
	}
}
