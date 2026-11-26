# Changelog

Notable changes of AEM Compose, following [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and [Semantic Versioning](https://semver.org/).
How to move a project from one major version to the next is described in [UPGRADING.md](UPGRADING.md).

## [Unreleased]

### Changed

- **Breaking:** variables in the config file use the [Compose Specification interpolation](https://github.com/compose-spec/compose-spec/blob/main/12-interpolation.md) syntax (`${VAR:-default}`) instead of Go templates with `[[ ]]` delimiters, which IDEs could not parse or format as YAML ([#157](https://github.com/wttech/aemc/issues/157)). Variables are replaced after the file is parsed, in values only. See [UPGRADING.md](UPGRADING.md#variables-in-config-file).
- **Breaking:** dotenv files (`.env`, `.env.local`, `local.env`) no longer override variables already set in the environment, as in [Twelve-Factor App](https://12factor.net/config), Docker Compose and mise. See [UPGRADING.md](UPGRADING.md#dotenv-files).
- `aem instance backup perform` stops the running instances together, archives them honouring the instance processing mode and starts them together again, so it waits for one startup instead of one per instance; disk space is checked for all instances up front and old backups are pruned once at the end. `restore` honours the processing mode too ([#258](https://github.com/wttech/aemc/issues/258)).
- **Breaking:** supportive tools are downloaded to `aem/home/opt/vendor` (`base.tool_dir`), so `aem vendor clean` no longer deletes the AEM Compose and Task binaries of the wrapper scripts, which failed on Windows ([#322](https://github.com/wttech/aemc/issues/322)). See [UPGRADING.md](UPGRADING.md#tool-directory).

### Added

- Command `aem config migrate` converting 2.x config files to the 3.x syntax; reading a 2.x config file stops with a hint to run it. Removed in 4.0.
- Config file *aem.yml* in the project root, looked up before *aem/default/etc/aem.yml* ([#309](https://github.com/wttech/aemc/issues/309)).
- Global flags `-C, --cd` to run for a project in another directory and `--config` to choose the config file ([#98](https://github.com/wttech/aemc/issues/98)).
- Built-in variables `AEM_PROJECT_DIR`, `AEM_OS`, `AEM_ARCH` and `AEM_ARCHIVE_EXT` in the config file.

### Removed

- Template functions in the config file (Sprig functions and `canonicalPath`).
- Unused setting `instance.local.tool_dir`.
- **Breaking:** setting `instance.local.backup_dir`, replaced by `instance.local.backup.dir`.
- **Breaking:** flag `--template` of `aem config value`; read a value with `--key` or filter all values with `aem config values --output-query`.
