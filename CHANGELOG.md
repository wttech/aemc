# Changelog

Notable changes of AEM Compose, following [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and [Semantic Versioning](https://semver.org/).
How to move a project from one major version to the next is described in [UPGRADING.md](UPGRADING.md).

## [Unreleased]

### Changed

- **Breaking:** variables in the config file use the [Compose Specification interpolation](https://github.com/compose-spec/compose-spec/blob/main/12-interpolation.md) syntax (`${VAR:-default}`) instead of Go templates with `[[ ]]` delimiters, which IDEs could not parse or format as YAML ([#157](https://github.com/wttech/aemc/issues/157)). Variables are replaced after the file is parsed, in values only. See [UPGRADING.md](UPGRADING.md#variables-in-config-file).

### Added

- Built-in variables `AEM_PROJECT_DIR`, `AEM_OS`, `AEM_ARCH` and `AEM_ARCHIVE_EXT` in the config file.

### Removed

- Template functions in the config file (Sprig functions and `canonicalPath`).
