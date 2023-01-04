package cfg

import (
	"github.com/spf13/viper"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/instance"
)

func setDefaults(v *viper.Viper) {
	v.SetDefault("log.level", "info")
	v.SetDefault("log.timestamp_format", "2006-01-02 15:04:05")
	v.SetDefault("log.full_timestamp", true)

	v.SetDefault("base.tmp_dir", "aem/home/tmp")

	v.SetDefault("input.format", fmtx.YML)
	v.SetDefault("input.file", InputStdin)
	v.SetDefault("output.format", fmtx.Text)
	v.SetDefault("output.file", OutputFileDefault)

	v.SetDefault("instance.processing_mode", instance.ProcessingAuto)
}
