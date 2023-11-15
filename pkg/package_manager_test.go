package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPackageManager_interpretFail(t *testing.T) {
	type fields struct {
		instance               *Instance
		UploadOptimized        bool
		InstallRecursive       bool
		InstallHTMLEnabled     bool
		InstallHTMLConsole     bool
		InstallHTMLStrict      bool
		SnapshotDeploySkipping bool
		SnapshotIgnored        bool
		SnapshotPatterns       []string
		ToggledWorkflows       []string
	}
	type args struct {
		message string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"no message", fields{}, args{}, "unexpected status: "},
		{"generic message", fields{}, args{"generic message"}, "unexpected status: generic message"},
		{"inaccessible value", fields{}, args{"Inaccessible value"}, "probably no disk space left (server respond with 'Inaccessible value')"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := &PackageManager{
				instance:               tt.fields.instance,
				UploadOptimized:        tt.fields.UploadOptimized,
				InstallRecursive:       tt.fields.InstallRecursive,
				InstallHTMLEnabled:     tt.fields.InstallHTMLEnabled,
				InstallHTMLConsole:     tt.fields.InstallHTMLConsole,
				InstallHTMLStrict:      tt.fields.InstallHTMLStrict,
				SnapshotDeploySkipping: tt.fields.SnapshotDeploySkipping,
				SnapshotIgnored:        tt.fields.SnapshotIgnored,
				SnapshotPatterns:       tt.fields.SnapshotPatterns,
				ToggledWorkflows:       tt.fields.ToggledWorkflows,
			}
			assert.Equalf(t, tt.want, pm.interpretFail(tt.args.message), "interpretFail(%v)", tt.args.message)
		})
	}
}
