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
		File   string `mapstructure:"file" yaml:"file"`
		Format string `mapstructure:"format" yaml:"format"`
	} `mapstructure:"output" yaml:"output"`

	Java struct {
		HomeDir            string `mapstructure:"home_dir" yaml:"home_dir"`
		VersionConstraints string `mapstructure:"version_constraints" yaml:"version_constraints"`
	}

	Instance struct {
		ConfigURL string `mapstructure:"config_url" yaml:"config_url"`

		Config map[string]struct {
			HTTPURL  string   `mapstructure:"http_url" yaml:"http_url"`
			User     string   `mapstructure:"user"`
			Password string   `mapstructure:"password"`
			JVMOpts  []string `mapstructure:"jvm_opts" yaml:"jvm_opts"`
			RunModes []string `mapstructure:"run_modes" yaml:"run_modes"`
			Version  string   `mapstructure:"version" yaml:"version"`
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
			AwaitStrict    bool `mapstructure:"await_strict" yaml:"await_strict"`
			AwaitUpTimeout struct {
				Duration time.Duration `mapstructure:"duration" yaml:"duration"`
			} `mapstructure:"await_up_timeout" yaml:"await_up_timeout"`
			AwaitDownTimeout struct {
				Duration time.Duration `mapstructure:"duration" yaml:"duration"`
			} `mapstructure:"await_down_timeout" yaml:"await_down_timeout"`
		} `mapstructure:"check" yaml:"check"`

		Local struct {
			UnpackDir string `mapstructure:"unpack_dir" yaml:"unpack_dir"`

			Quickstart struct {
				DistFile    string `mapstructure:"dist_file" yaml:"dist_file"`
				LicenseFile string `mapstructure:"license_file" yaml:"license_file"`
			} `mapstructure:"quickstart" yaml:"quickstart"`
		} `mapstructure:"local" yaml:"local"`

		Package struct {
			DeployAvoidance  bool     `mapstructure:"deploy_avoidance" yaml:"deploy_avoidance"`
			SnapshotPatterns []string `mapstructure:"snapshot_patterns" yaml:"snapshot_patterns"`
		} `mapstructure:"package" yaml:"package"`

		OSGi struct {
			Bundle struct {
				Install struct {
					Start           bool `mapstructure:"start" yaml:"start"`
					StartLevel      int  `mapstructure:"start_level" yaml:"start_level"`
					RefreshPackages bool `mapstructure:"refresh_packages" yaml:"refresh_packages"`
				} `mapstructure:"install" yaml:"install"`
			} `mapstructure:"bundle" yaml:"bundle"`
		} `mapstructure:"osgi" yaml:"osgi"`
	} `mapstructure:"instance" yaml:"instance"`
}
