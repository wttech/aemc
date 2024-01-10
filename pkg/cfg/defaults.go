package cfg

import (
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/instance"
	"time"
)

func (c *Config) setDefaults() {
	v := viper.New()
	c.viper = v

	v.SetDefault("log.level", "info")
	v.SetDefault("log.timestamp_format", "2006-01-02 15:04:05")
	v.SetDefault("log.full_timestamp", true)

	v.SetDefault("base.tmp_dir", common.TmpDir)
	v.SetDefault("base.tool_dir", common.ToolDir)
	v.SetDefault("base.cache_dir", common.CacheDir)

	v.SetDefault("input.format", fmtx.YML)
	v.SetDefault("input.file", common.STDIn)

	v.SetDefault("output.format", fmtx.Text)
	v.SetDefault("output.no_color", color.NoColor)
	v.SetDefault("output.value", common.OutputValueAll)
	v.SetDefault("output.log.file", common.LogFile)
	v.SetDefault("output.log.mode", OutputLogConsole)

	v.SetDefault("java.home_dir", "")
	v.SetDefault("java.version_constraints", ">= 11, < 12")
	v.SetDefault("java.download.url", c.tplString("https://github.com/adoptium/temurin11-binaries/releases/download/jdk-11.0.18%2B10/OpenJDK11U-jdk_[[.Arch]]_[[.Os]]_hotspot_11.0.18_10.[[.ArchiveExt]]"))
	v.SetDefault("java.download.replacements", map[string]string{"darwin": "mac", "x86_64": "x64", "amd64": "x64", "386": "x86-32", "arm64": "x64", "aarch64": "x64"})

	v.SetDefault("instance.processing_mode", instance.ProcessingAuto)

	v.SetDefault("instance.http.timeout", time.Minute*10)
	v.SetDefault("instance.http.debug", false)
	v.SetDefault("instance.http.disable_warn", true)
	v.SetDefault("instance.http.ignore_ssl_errors", true)

	v.SetDefault("instance.check.skip", false)
	v.SetDefault("instance.check.warmup", time.Second*1)
	v.SetDefault("instance.check.interval", time.Second*6)
	v.SetDefault("instance.check.done_threshold", 4)
	v.SetDefault("instance.check.installer.state", true)
	v.SetDefault("instance.check.installer.pause", true)

	v.SetDefault("instance.check.await_strict", true)
	v.SetDefault("instance.check.await_started.timeout", time.Minute*30)
	v.SetDefault("instance.check.await_stopped.timeout", time.Minute*10)

	v.SetDefault("instance.check.reachable.skip", false)
	v.SetDefault("instance.check.reachable.timeout", time.Second*3)

	v.SetDefault("instance.check.bundle_stable.skip", false)
	v.SetDefault("instance.check.bundle_stable.symbolic_names_ignored", []string{})

	v.SetDefault("instance.check.event_stable.skip", false)
	v.SetDefault("instance.check.event_stable.received_max_age", time.Second*5)
	v.SetDefault("instance.check.event_stable.topics_unstable", []string{"org/osgi/framework/ServiceEvent/*", "org/osgi/framework/FrameworkEvent/*", "org/osgi/framework/BundleEvent/*"})
	v.SetDefault("instance.check.event_stable.details_ignored", []string{"*.*MBean", "org.osgi.service.component.runtime.ServiceComponentRuntime", "java.util.ResourceBundle"})

	v.SetDefault("instance.check.component_stable.skip", false)
	v.SetDefault("instance.check.component_stable.pids.include", []string{})
	v.SetDefault("instance.check.component_stable.pids.exclude", []string{})
	v.SetDefault("instance.check.component_stable.pids.match", map[string]string{})

	v.SetDefault("instance.check.installer.skip", false)
	v.SetDefault("instance.check.installer.state", true)
	v.SetDefault("instance.check.installer.pause", true)

	v.SetDefault("instance.check.path_ready.timeout", time.Second*10)

	v.SetDefault("instance.check.login_page.skip", false)
	v.SetDefault("instance.check.login_page.path", "/libs/granite/core/content/login.html")
	v.SetDefault("instance.check.login_page.status_code", 200)
	v.SetDefault("instance.check.login_page.contained_text", "QUICKSTART_HOMEPAGE")

	v.SetDefault("instance.local.tool_dir", common.ToolDir)
	v.SetDefault("instance.local.unpack_dir", common.VarDir+"/instance")
	v.SetDefault("instance.local.override_dir", common.DefaultDir+"/"+common.VarDirName+"/instance")

	v.SetDefault("instance.local.quickstart.dist_file", common.LibDir+"/{aem-sdk,cq-quickstart}-*.{zip,jar}")
	v.SetDefault("instance.local.quickstart.license_file", common.LibDir+"/license.properties")

	v.SetDefault("instance.local.await_strict", true)
	v.SetDefault("instance.local.service_mode", false)

	v.SetDefault("instance.local.oak_run.download_url", "https://repo1.maven.org/maven2/org/apache/jackrabbit/oak-run/1.44.0/oak-run-1.44.0.jar")
	v.SetDefault("instance.local.oak_run.store_path", "crx-quickstart/repository/segmentstore")

	v.SetDefault("instance.status.timeout", time.Millisecond*500)

	v.SetDefault("instance.package.upload_optimized", true)

	v.SetDefault("instance.package.install_recursive", true)
	v.SetDefault("instance.package.install_save_threshold", 1024)
	v.SetDefault("instance.package.install_ac_handling", "")
	v.SetDefault("instance.package.install_dependency_handling", "required")

	v.SetDefault("instance.package.install_html.enabled", false)
	v.SetDefault("instance.package.install_html.strict", true)
	v.SetDefault("instance.package.install_html.console", false)

	v.SetDefault("instance.package.snapshot_deploy_skipping", true)
	v.SetDefault("instance.package.snapshot_ignored", false)
	v.SetDefault("instance.package.snapshot_patterns", []string{"**/*-SNAPSHOT.zip"})

	v.SetDefault("instance.package.toggled_workflows", []string{})

	v.SetDefault("instance.repo.property_change_ignored", []string{"jcr:created", "cq:lastModified", "transportPassword"})

	v.SetDefault("instance.osgi.shutdown_delay", time.Second*3)
	v.SetDefault("instance.osgi.bundle.install.start", true)
	v.SetDefault("instance.osgi.bundle.install.start_level", 20)
	v.SetDefault("instance.osgi.bundle.install.refresh_packages", true)

	v.SetDefault("instance.ssl.setup_timeout", time.Second*30)

	v.SetDefault("instance.crypto.key_bundle_symbolic_name", "com.adobe.granite.crypto.file")
	v.SetDefault("instance.replication.bundle_symbolic_name", "com.day.cq.cq-replication")

	v.SetDefault("instance.workflow.lib_root", "/libs/settings/workflow/launcher")
	v.SetDefault("instance.workflow.config_root", "/conf/global/settings/workflow/launcher")
	v.SetDefault("instance.workflow.toggle_retry_delay", time.Second*10)
	v.SetDefault("instance.workflow.toggle_retry_timeout", time.Minute*5)

	v.SetDefault("content.clean.files_deleted", []any{
		map[string]any{
			"patterns": []string{
				"**/.vlt",
				"**/.vlt*.tmp",
				"**/install/*.jar",
			},
		},
	})
	v.SetDefault("content.clean.files_flattened", []string{
		"**/_cq_design_dialog/.content.xml",
		"**/_cq_dialog/.content.xml",
		"**/_cq_htmlTag/.content.xml",
		"**/_cq_template/.content.xml",
	})
	v.SetDefault("content.clean.properties_skipped", []any{
		map[string]any{
			"patterns":       []string{"jcr:uuid"},
			"excluded_paths": []string{"**/home/users/*", "**/home/groups/*"},
		},
		map[string]any{
			"patterns":       []string{"cq:lastModified*"},
			"excluded_paths": []string{"**/content/experience-fragments/*"},
		},
		map[string]any{
			"patterns":       []string{"dam:sha1", "dam:size"},
			"included_paths": []string{"**/content/dam/*.svg/*"},
		},
		map[string]any{
			"patterns": []string{
				"jcr:lastModified*",
				"jcr:created*",
				"jcr:isCheckedOut",
				"cq:lastReplicat*",
				"cq:lastRolledout*",
				"dam:extracted",
				"dam:assetState",
				"dc:modified",
				"*_x0040_*",
				"cq:name",
				"cq:parentPath",
				"dam:copiedAt",
				"dam:parentAssetID",
				"dam:relativePath",
			},
		},
	})

	v.SetDefault("content.clean.mixin_types_skipped", []any{
		map[string]any{
			"patterns": []string{
				"cq:ReplicationStatus",
				"mix:versionable",
			},
		},
	})
	v.SetDefault("content.clean.namespaces_skipped", true)
	v.SetDefault("content.clean.parents_backup_enabled", true)
}
