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

## Releasing

Simply run script:

```shell
sh release.sh <major.minor.patch>
```

It will:

* bump version is source files automatically,
* commit changes,
* push release tag that will initiate [release workflow](.github/workflows/release-perform.yml).
