# Upgrading

How to move a project to a new major version of AEM Compose. All changes are listed in [CHANGELOG.md](CHANGELOG.md).

## From 2.x to 3.x

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
