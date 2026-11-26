# Upgrading

How to move a project to a new major version of AEM Compose. All changes are listed in [CHANGELOG.md](CHANGELOG.md).

## From 2.x to 3.x

Run `aem config migrate` first. It converts the config files (`aem/home/etc/aem.yml` and the VCS-tracked one) from the template syntax to variables in place and lists the lines that need a manual change. Until then, AEM Compose refuses to read a 2.x config file. The command is removed in 4.0.

### Variables in config file

The config file (`aem/default/etc/aem.yml`, `aem/home/etc/aem.yml`) no longer is a Go template. Its values reference variables with the [Compose Specification interpolation](https://github.com/compose-spec/compose-spec/blob/main/12-interpolation.md) syntax, the one of Docker Compose and shells. The file is parsed as YAML first, so it stays valid YAML for IDEs.

| 2.x                                    | 3.x                                     |
|----------------------------------------|-----------------------------------------|
| `[[.Env.VAR]]`                         | `${VAR}`                                |
| `[[.Env.VAR \| default "value" ]]`     | `${VAR:-value}`                         |
| `[[.Env.VAR \| default true ]]`        | `${VAR:-true}`                          |
| `[[canonicalPath .Path "aem/home/tmp"]]` | `${AEM_PROJECT_DIR}/aem/home/tmp`    |
| `[[.Os]]`                              | `${AEM_OS}`                            |
| `[[.Arch]]`                            | `${AEM_ARCH}`                          |
| `[[.ArchiveExt]]`                      | `${AEM_ARCHIVE_EXT}`                   |
| a literal `$` in a value               | `$$`                                    |
| other Sprig functions (`upper`, `trim`, …) | not supported: compute the value in the environment (e.g. `.env`) |

Traps:

- A literal `$` in a value, e.g. in a password, must now be written as `$$`.
- An unquoted value is typed after the replacement (`${VAR:-true}` is a boolean, `${VAR:-4502}` a number); quote it to keep a string.
- Variables in comments are not replaced.

### Dotenv files

Variables set in the environment (by a shell, CI or mise) now win over the dotenv files. The order of the files is kept: `.env` over `.env.local` over `local.env`.

| Variable set in                    | 2.x winner   | 3.x winner  |
|------------------------------------|--------------|-------------|
| environment and any dotenv file    | dotenv file  | environment |
| `.env` and `local.env`             | `.env`       | `.env`      |

So `AEM_AUTHOR_HTTP_URL=http://localhost:4512 aem instance status` now uses that URL even when `local.env` sets another one.

### Removed settings and flags

| 2.x                                          | 3.x                                                             |
|----------------------------------------------|-----------------------------------------------------------------|
| `instance.local.backup_dir`                  | `instance.local.backup.dir`                                     |
| `aem config value --template '[[.log.level]]'` | `aem config value --key log.level`                            |
| `aem config value --template '...'` (other)  | `aem config values --output-format yml --output-query '<JMESPath>'` |

### Tool directory

Supportive tools (Java, Oak Run, Vault, unpacked AEM SDK) are downloaded to `aem/home/opt/vendor` instead of `aem/home/opt`.

- If your `aem.yml` sets `base.tool_dir: aem/home/opt`, change it to `aem/home/opt/vendor`.
- Before upgrading, run `aem vendor clean` with 2.x, or delete `aem/home/opt/{java,oak-run,sdk,vault}` afterwards; the tools are downloaded again on first use.
