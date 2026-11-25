package pkg

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/wttech/aemc/pkg/common"
)

// Temporary bridge for the released 'instance.local.backup_dir' setting, replaced by 'instance.local.backup.dir'.
// Remove in the next major version together with this file; then move the default to cfg/defaults.go.

const (
	localBackupDirKey       = "instance.local.backup.dir"
	localBackupDirLegacyKey = "instance.local.backup_dir"
	localBackupDirDefault   = common.VarDir + "/backup"
)

func localBackupDir(cfg *viper.Viper) string {
	if dir := cfg.GetString(localBackupDirKey); dir != "" {
		return dir
	}
	if dir := cfg.GetString(localBackupDirLegacyKey); dir != "" {
		log.Warnf("setting '%s' is deprecated, use '%s' instead", localBackupDirLegacyKey, localBackupDirKey)
		return dir
	}
	return localBackupDirDefault
}
