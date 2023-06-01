![AEM Compose Logo](https://github.com/wttech/aemc-ansible/raw/main/docs/logo-with-text.png)
[![WTT Logo](https://github.com/wttech/aemc-ansible/raw/main/docs/wtt-logo.png)](https://www.wundermanthompson.com/service/technology)

[![Apache License, Version 2.0, January 2004](https://github.com/wttech/aemc-ansible/raw/main/docs/apache-license-badge.svg)](http://www.apache.org/licenses/)

**AEM Compose**

Universal tool to manage AEM instances everywhere!

# Developer setup

## Prerequisites

1. Install Go: <https://go.dev/doc/install>,
2. Set up shell, append lines *~/.zshrc* with content below then restart IDE/terminals,

```shell
export GOPATH="$HOME/go"
export PATH="$GOPATH/bin:$PATH"
```

## Building & OS-wide installation

Ensure having installed [Go](https://go.dev/dl/) then:

### Manual installation (recommended)

Use this method to develop comfortably the tool.

1. Clone repository: `git clone git@github.com:wttech/aemc.git`
2. Enter cloned directory and run command: `make`*

*When using Git Bash on Windows, you will first need to add `make` to your Git Bash installation:
1. Go to [ezwinports](https://sourceforge.net/projects/ezwinports/files/).
2. Download `make-x.x.x-without-guile-w32-bin.zip` (get the newest version without guile).
3. Extract zip.
4. Copy the contents to your `Git\mingw64\` merging the folders, but do NOT overwrite/replace any existing files.

### Go installation

Use this method to check particular commit/version of the tool.

- latest released version: `go install github.com/wttech/aemc/cmd/aem@latest`,
- specific released version: `go install github.com/wttech/aemc/cmd/aem@v1.1.9`,
- recently committed version: `go install github.com/wttech/aemc/cmd/aem@main`,

After installing AEM CLI by one of above methods now instruct the [wrapper script](pkg/project/common/aemw) to use it by running the following command:

```shell
export AEM_CLI_VERSION=installed
```

To start using again version defined in wrapper file, simply unset the environment variable:

```shell
unset AEM_CLI_VERSION
```

## Releasing

Simply run script:

```shell
sh release.sh <major.minor.patch>
```

It will:

* bump version is source files automatically,
* commit changes,
* push release tag that will initiate [release workflow](.github/workflows/release-perform.yml).
