package cfg

import "time"

// ConfigValues defines all available configuration options
type ConfigValues struct {
	Log struct {
		Level           string `mapstructure:"level" yaml:"level"`
		TimestampFormat string `mapstructure:"timestamp_format" yaml:"timestamp_format"`
		FullTimestamp   bool   `mapstructure:"full_timestamp" yaml:"full_timestamp"`
	} `mapstructure:"log" yaml:"log"`

	Base struct {
		TmpDir string `mapstructure:"tmp_dir" yaml:"tmp_dir"`
	}

	Input struct {
		File   string `mapstructure:"file" yaml:"file"`
		String string `mapstructure:"string" yaml:"string"`
		Format string `mapstructure:"format" yaml:"format"`
	} `mapstructure:"input" yaml:"input"`

	Output struct {
		Format  string `mapstructure:"format" yaml:"format"`
		NoColor bool   `mapstructure:"no_color" yaml:"no_color"`
		Value   string `mapstructure:"value" yaml:"value"`
		Log     struct {
			File string `mapstructure:"file" yaml:"file"`
			Mode string `mapstructure:"mode" yaml:"mode"`
		} `mapstructure:"log" yaml:"log"`
	} `mapstructure:"output" yaml:"output"`

	Java struct {
		VersionConstraints string `mapstructure:"version_constraints" yaml:"version_constraints"`
		HomeDir            string `mapstructure:"home_dir" yaml:"home_dir"`
		Download           struct {
			URL          string            `mapstructure:"url" yaml:"url"`
			Replacements map[string]string `mapstructure:"replacements" yaml:"replacements"`
		}
	}

	Instance struct {
		ConfigURL string `mapstructure:"config_url" yaml:"config_url"`

		Config map[string]struct {
			HTTPURL    string   `mapstructure:"http_url" yaml:"http_url"`
			User       string   `mapstructure:"user"`
			Password   string   `mapstructure:"password"`
			StartOpts  []string `mapstructure:"start_opts" yaml:"start_opts"`
			JVMOpts    []string `mapstructure:"jvm_opts" yaml:"jvm_opts"`
			RunModes   []string `mapstructure:"run_modes" yaml:"run_modes"`
			EnvVars    []string `mapstructure:"env_vars" yaml:"env_vars"`
			SecretVars []string `mapstructure:"secret_vars" yaml:"secret_vars"`
			SlingProps []string `mapstructure:"sling_props" yaml:"sling_props"`
			Version    string   `mapstructure:"version" yaml:"version"`
		} `mapstructure:"config" yaml:"config"`

		Filter struct {
			ID      string `mapstructure:"id" yaml:"id"`
			Author  bool   `mapstructure:"author" yaml:"author"`
			Publish bool   `mapstructure:"publish" yaml:"publish"`
		} `mapstructure:"filter" yaml:"filter"`

		ProcessingMode string `mapstructure:"processing_mode" yaml:"processing_mode"`

		Check struct {
			Warmup        time.Duration `mapstructure:"warmup" yaml:"warmup"`
			Interval      time.Duration `mapstructure:"interval" yaml:"interval"`
			DoneThreshold int           `mapstructure:"done_threshold" yaml:"done_threshold"`
			BundleStable  struct {
				SymbolicNamesIgnored []string `mapstructure:"symbolic_names_ignored" yaml:"symbolic_names_ignored"`
			} `mapstructure:"bundle_stable" yaml:"bundle_stable"`
			EventStable struct {
				ReceivedMaxAge time.Duration `mapstructure:"received_max_age" yaml:"received_max_age"`
				TopicsUnstable []string      `mapstructure:"topics_unstable" yaml:"topics_unstable"`
				DetailsIgnored []string      `mapstructure:"details_ignored" yaml:"details_ignored"`
			} `mapstructure:"event_stable" yaml:"event_stable"`
			Installer struct {
				State bool `mapstructure:"state" yaml:"state"`
				Pause bool `mapstructure:"pause" yaml:"pause"`
			} `mapstructure:"installer" yaml:"installer"`
			AwaitStrict         bool `mapstructure:"await_strict" yaml:"await_strict"`
			AwaitStartedTimeout struct {
				Duration time.Duration `mapstructure:"duration" yaml:"duration"`
			} `mapstructure:"await_started_timeout" yaml:"await_started_timeout"`
			AwaitStoppedTimeout struct {
				Duration time.Duration `mapstructure:"duration" yaml:"duration"`
			} `mapstructure:"await_stopped_timeout" yaml:"await_stopped_timeout"`
			Reachable struct {
				Timeout time.Duration `mapstructure:"timeout" yaml:"timeout"`
			} `mapstructure:"reachable" yaml:"reachable"`
			Unreachable struct {
				Timeout time.Duration `mapstructure:"timeout" yaml:"timeout"`
			} `mapstructure:"unreachable" yaml:"unreachable"`
		} `mapstructure:"check" yaml:"check"`

		Local struct {
			UnpackDir   string `mapstructure:"unpack_dir" yaml:"unpack_dir"`
			BackupDir   string `mapstructure:"backup_dir" yaml:"backup_dir"`
			OverrideDir string `mapstructure:"override_dir" yaml:"override_dir"`
			ToolDir     string `mapstructure:"tool_dir" yaml:"tool_dir"`
			ServiceMode bool   `mapstructure:"service_mode" yaml:"service_mode"`

			Quickstart struct {
				DistFile    string `mapstructure:"dist_file" yaml:"dist_file"`
				LicenseFile string `mapstructure:"license_file" yaml:"license_file"`
			} `mapstructure:"quickstart" yaml:"quickstart"`
			OakRun struct {
				DownloadURL string `mapstructure:"download_url" yaml:"download_url"`
				StorePath   string `mapstructure:"store_path" yaml:"store_path"`
			} `mapstructure:"oak_run" yaml:"oak_run"`
		} `mapstructure:"local" yaml:"local"`

		Status struct {
			Timeout time.Duration `mapstructure:"timeout" yaml:"timeout"`
		} `mapstructure:"status" yaml:"status"`

		Repo struct {
			PropertyChangeIgnored []string `mapstructure:"property_change_ignored" yaml:"property_change_ignored"`
		} `mapstructure:"repo" yaml:"repo"`

		Package struct {
			SnapshotDeploySkipping bool     `mapstructure:"snapshot_deploy_skipping" yaml:"snapshot_deploy_skipping"`
			SnapshotPatterns       []string `mapstructure:"snapshot_patterns" yaml:"snapshot_patterns"`
		} `mapstructure:"package" yaml:"package"`

		OSGi struct {
			ShutdownDelay time.Duration `mapstructure:"shutdown_delay" yaml:"shutdown_delay"`

			Bundle struct {
				Install struct {
					Start           bool `mapstructure:"start" yaml:"start"`
					StartLevel      int  `mapstructure:"start_level" yaml:"start_level"`
					RefreshPackages bool `mapstructure:"refresh_packages" yaml:"refresh_packages"`
				} `mapstructure:"install" yaml:"install"`
			} `mapstructure:"bundle" yaml:"bundle"`
		} `mapstructure:"osgi" yaml:"osgi"`

		Crypto struct {
			KeyBundleSymbolicName string `mapstructure:"key_bundle_symbolic_name" yaml:"key_bundle_symbolic_name"`
		} `mapstructure:"crypto" yaml:"crypto"`
	} `mapstructure:"instance" yaml:"instance"`

	Content struct {
		FilesDotContent      []string   `mapstructure:"files_dot_content" yaml:"files_dot_content"`
		FilesDeleted         []PathRule `mapstructure:"files_deleted" yaml:"files_deleted"`
		FilesFlattened       []string   `mapstructure:"files_flattened" yaml:"files_flattened"`
		PropertiesSkipped    []PathRule `mapstructure:"properties_skipped" yaml:"properties_skipped"`
		MixinTypesSkipped    []PathRule `mapstructure:"mixin_types_skipped" yaml:"mixin_types_skipped"`
		NamespacesSkipped    bool       `mapstructure:"namespaces_skipped" yaml:"namespaces_skipped"`
		ParentsBackupEnabled bool       `mapstructure:"parents_backup_enabled" yaml:"parents_backup_enabled"`
		ParentsBackupSuffix  string     `mapstructure:"parents_backup_suffix" yaml:"parents_backup_suffix"`
	} `mapstructure:"content" yaml:"content"`
}

type PathRule struct {
	Patterns      []string `mapstructure:"patterns" yaml:"patterns"`
	ExcludedPaths []string `mapstructure:"excluded_paths" yaml:"excluded_paths"`
	IncludedPaths []string `mapstructure:"included_paths" yaml:"included_paths"`
}
