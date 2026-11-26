# Changelog

Notable changes of AEM Compose, following [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and [Semantic Versioning](https://semver.org/).
How to move a project from one major version to the next is described in [UPGRADING.md](UPGRADING.md).

## [Unreleased]

## [3.0.0] - 2026-11-26

### Changed

- **Breaking:** `aem file archive` and `aem file unarchive` support zip, tar, tar.gz (tgz), tar.zst (tzst) and, for unarchiving, tar.bz2 (tbz2); rar, tar.xz, tar.lz4, tar.br and tar.sz are no longer supported.
- **Breaking:** `aem instance upgrade` refuses instances in service mode (`instance.local.service_mode`), as the upgrade boots AEM outside the service. See [UPGRADING.md](UPGRADING.md#instance-upgrade-in-service-mode).
- **Breaking:** variables in the config file use the [Compose Specification interpolation](https://github.com/compose-spec/compose-spec/blob/main/12-interpolation.md) syntax (`${VAR:-default}`) instead of Go templates with `[[ ]]` delimiters, which IDEs could not parse or format as YAML ([#157](https://github.com/wttech/aemc/issues/157)). Variables are replaced after the file is parsed, in values only. See [UPGRADING.md](UPGRADING.md#variables-in-config-file).
- **Breaking:** dotenv files (`.env`, `.env.local`, `local.env`) no longer override variables already set in the environment, as in [Twelve-Factor App](https://12factor.net/config), Docker Compose and mise. See [UPGRADING.md](UPGRADING.md#dotenv-files).
- `aem instance backup perform` stops the running instances together, archives them honouring the instance processing mode and starts them together again, so it waits for one startup instead of one per instance; disk space is checked for all instances up front and old backups are pruned once at the end. `restore` honours the processing mode too ([#258](https://github.com/wttech/aemc/issues/258)).
- **Breaking:** supportive tools are downloaded to `aem/home/opt/vendor` (`base.tool_dir`), so `aem vendor clean` no longer deletes the AEM Compose and Task binaries of the wrapper scripts, which failed on Windows ([#322](https://github.com/wttech/aemc/issues/322)). See [UPGRADING.md](UPGRADING.md#tool-directory).

### Added

- Projects run by [mise](https://mise.jdx.dev), the default of `aem project scaffold --runner mise|task`: `mise.toml` with the project tasks and pinned tools (AEM Compose, Java, Maven) locked with checksums in `mise.lock`, AEM Compose and Maven on the Java installed by mise, and `bin/mise` to run it where mise is not installed, e.g. on CI. No wrapper scripts. See [README](README.md#choosing-a-task-runner) for a comparison with Task.
- Command `aem config migrate` converting 2.x config files to the 3.x syntax; reading a 2.x config file stops with a hint to run it. Removed in 4.0.
- Config file *aem.yml* in the project root, looked up before *aem/default/etc/aem.yml* ([#309](https://github.com/wttech/aemc/issues/309)).
- Global flags `-C, --cd` to run for a project in another directory and `--config` to choose the config file ([#98](https://github.com/wttech/aemc/issues/98)).
- Built-in variables `AEM_PROJECT_DIR`, `AEM_OS`, `AEM_ARCH` and `AEM_ARCHIVE_EXT` in the config file.

- The wrapper scripts of projects run by Task (`aemw`, `taskw`) verify TLS, with a clear error and the `AEM_INSECURE_DOWNLOAD=true` opt-out behind TLS-inspecting proxies, and the checksums of the downloaded archives; they refuse unsupported platforms. Task 3.54.0; Maven Wrapper 3.3.4 without the binary `maven-wrapper.jar`.

- Project files for AEM as a Cloud Service and instance projects use Java 21 (Temurin 21.0.13), as current AEM SDKs do not start on Java 11; classic projects keep Java 11.

- **Breaking:** the dispatcher of app projects is reached at `publish.localhost` (and `author.localhost` on classic projects) instead of `publish`, `publish.aem.local` and `author.aem.local`. Names under `.localhost` resolve to the machine itself ([RFC 6761](https://www.rfc-editor.org/rfc/rfc6761)), so the `dispatcher:hosts` task, which edited the hosts file with administrator rights, is removed. See [UPGRADING.md](UPGRADING.md#dispatcher-domains).

- The `dispatcher:check` task of app projects retries for up to 5 minutes while the dispatcher warms up and follows redirects, e.g. to the home page, instead of one request after a fixed 3-second wait, which failed on cold starts, especially on Windows ([#353](https://github.com/wttech/aemc/issues/353)).

- Java 25 accepted for running AEM (`vendor.java.version_constraints`), as AEM as a Cloud Service supports it; the Java used by default stays the same.

### Security

- Dependencies with known vulnerabilities updated: `golang.org/x/crypto` 0.52.0, `golang.org/x/net` 0.55.0, `github.com/klauspost/compress` 1.18.7.
- Archives (instance backups, downloaded Java, AEM SDK, `aem file`) are handled with the Go standard library and zstd instead of the abandoned `github.com/mholt/archiver/v3`, which has unfixed path traversal vulnerabilities ([CVE-2025-3445](https://github.com/advisories/GHSA-7vpp-9cxj-q8gv), [CVE-2024-0406](https://github.com/advisories/GHSA-rhh4-rh7c-7r5v)). Entries that would land outside the target dir, also through symlinks, are refused, as they are when unpacking content packages. File modes and symlinks are kept; backups stay compatible both ways.

### Removed

- Template functions in the config file (Sprig functions and `canonicalPath`).
- Unused setting `instance.local.tool_dir`.
- **Breaking:** setting `instance.local.backup_dir`, replaced by `instance.local.backup.dir`.
- **Breaking:** flag `--template` of `aem config value`; read a value with `--key` or filter all values with `aem config values --output-query`.

### Fixed

- `aem instance upgrade` boots the instance once from the new Quickstart JAR and shuts it down again, as Adobe's [in-place upgrade](https://experienceleague.adobe.com/en/docs/experience-manager-65-lts/content/implementing/deploying/upgrading/in-place-upgrade) requires; starting it from the unpacked app JAR applied the upgrade only partly, e.g. AEM 6.5 LTS 6.6.2 to 6.6.3 left over a third of the bundles unresolved ([#355](https://github.com/wttech/aemc/issues/355)). The installed build is recorded first where Quickstart expects it, so AEM runs its upgrade tasks (e.g. content migration) in upgrade mode instead of install mode. The upgrade therefore takes minutes and needs the instance HTTP port free. If the boot fails, the instance stays outdated: `instance start` refuses it and `instance upgrade` runs again. See [README](README.md#upgrading-local-instances).
- The built-in Java version constraints, used when the config file does not set `vendor.java.version_constraints`, were malformed (`=>` instead of `>=`) so checking the Java version failed.
