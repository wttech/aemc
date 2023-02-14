package cfg

import (
	"github.com/spf13/viper"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/instance"
)

func setDefaults(v *viper.Viper) {
	v.SetDefault("log.level", "info")
	v.SetDefault("log.timestamp_format", "2006-01-02 15:04:05")
	v.SetDefault("log.full_timestamp", true)

	v.SetDefault("base.tmp_dir", common.TmpDir)

	v.SetDefault("input.format", fmtx.YML)
	v.SetDefault("input.file", common.STDIn)
	v.SetDefault("output.format", fmtx.Text)
	v.SetDefault("output.value", common.OutputValueAll)
	v.SetDefault("output.log.file", common.LogFile)
	v.SetDefault("output.log.console", true)

	v.SetDefault("instance.processing_mode", instance.ProcessingAuto)

	v.SetDefault("instance.check.warmup", "1s")
	v.SetDefault("instance.check.interval", "5s")
	v.SetDefault("instance.check.done_threshold", 5)
	v.SetDefault("instance.check.installer.state", true)
	v.SetDefault("instance.check.installer.pause", true)

	v.SetDefault("instance.check.await_strict", true)
	v.SetDefault("instance.check.await_started_timeout.duration", "20m")
	v.SetDefault("instance.check.await_stopped_timeout.duration", "10m")
	v.SetDefault("instance.check.reachable.timeout", "3s")
	v.SetDefault("instance.check.unreachable.timeout", "3s")

	v.SetDefault("instance.check.event_stable.received_max_age", "5s")
	v.SetDefault("instance.check.event_stable.topics_unstable", []string{
		"org/osgi/framework/ServiceEvent/*",
		"org/osgi/framework/FrameworkEvent/*",
		"org/osgi/framework/BundleEvent/*",
	})
	v.SetDefault("instance.check.event_stable.details_ignored", []string{
		"*.*MBean",
		"org.osgi.service.component.runtime.ServiceComponentRuntime",
		"java.util.ResourceBundle",
	})

	v.SetDefault("instance.package.snapshot_deploy_skipping", true)

	v.SetDefault("instance.repo.property_change_ignored", []string{
		"jcr:created",
		"cq:lastModified",
		"transportPassword",
	})
}
