![AEM Compose Logo](https://github.com/wttech/aemc-ansible/raw/main/docs/logo-with-text.png)
[![WTT Logo](https://github.com/wttech/aemc-ansible/raw/main/docs/wtt-logo.png)](https://www.wundermanthompson.com/service/technology)

[![Last Release Version](https://img.shields.io/github/v/release/wttech/aemc?color=lightblue&label=Last%20Release)](https://github.com/wttech/aemc/releases)
![Go Version](https://img.shields.io/github/go-mod/go-version/wttech/aemc)
[![Apache License, Version 2.0, January 2004](https://github.com/wttech/aemc-ansible/raw/main/docs/apache-license-badge.svg)](http://www.apache.org/licenses/)

**AEM Compose**

Universal tool to manage AEM instances everywhere!

- Reusable core designed to handle advanced dev-ops operations needed to manage AEM instances
- Various distributions based on core for context-specific use cases:
  - [*CLI*](#cli---overview) - for developer workstations, shell scripting
  - [*Ansible Collection/Modules*](#ansible-collection) - for managing higher AEM environments
- Fast & lightweight
- No dependencies - usable on all operating systems and architectures

# Table of Contents

* [Distributions](#distributions)
  * [CLI - Overview](#cli---overview)
  * [CLI - AEM Project Quickstart](#cli---aem-project-quickstart)
  * [CLI - Building &amp; installing from source](#cli---building--installing-from-source)
  * [Ansible Collection](#ansible-collection)
  * [Go Scripting](#go-scripting)
* [Dependencies](#dependencies)
* [Configuration](#configuration)
  * [Generating default configuration](#generating-default-configuration)
  * [Configuration precedence](#configuration-precedence)
  * [Performance optimization](#performance-optimization)
* [Contributing](#contributing)
* [License](#license)

# Distributions

## CLI - Overview

Provides complete set of commands to comfortably work with CRX packages, OSGi configurations, repository nodes and more.

Key assumptions:

- Idempotent and fast
- Rich configuration options
- Self-describing, both machine & human-readable 
- Multiple input & output formats (text/yaml/json)

Main features:

- easy & declarative setup of:
  - JDK (isolated, version tied to project)
  - AEM instances (run modes, JVM & start opts, env & secret vars, Sling props, custom admin password)
  - OSGi (configurations, bundles, components)
  - replication agents 
  - any repository nodes
- deploying AEM packages with:
  - automatic workflow toggling - avoiding DAM asset renditions regeneration
  - advanced snapshot handling - avoiding redeploying the same package by checksum verification
  - customizable instance health checking
- building AEM packages with:
  - source code change detection - avoiding rebuilding application when it is not needed 
- making AEM instance backups (with restoring)
  - advanced archive format to speed up performance and storage efficiency ([ZSTD](https://github.com/facebook/zstd) used by default)
  - instance state aware - stopping, archiving then starting again AEM instances automatically (if needed)

Worth knowing:

- On Windows use it with [Git Bash](https://gitforwindows.org/) ([CMD](https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/cmd) and [PowerShell](https://learn.microsoft.com/en-us/powershell/scripting/overview) are not supported nor tested)

![CLI Screenshot](docs/cli-demo.gif)

## CLI - AEM Project Quickstart

Run one of the commands below according to the AEM version you are using to initialize the AEM Compose tool in your project (e.g existing one or generated from [Adobe AEM Project Archetype](https://github.com/adobe/aem-project-archetype#usage)):

- Cloud/AEMaaCS/SDK/202x.y.zzzzz

  ```shell
  curl https://raw.githubusercontent.com/wttech/aemc/main/project/cloud/init.sh | sh
  ```

- Classic/AMS/On-Prem/6.5.x

  ```shell
  curl https://raw.githubusercontent.com/wttech/aemc/main/project/classic/init.sh | sh
  ```

After successful initialization, remember to always use the tool via wrapper script in the following way:

```shell
sh aemw [command]
```

For example:

```shell
sh aemw version
```

## CLI - Building & installing from source

Ensure having installed [Go](https://go.dev/dl/) then run command:

- latest released version: `go install github.com/wttech/aemc/cmd/aem@latest`,
- specific released version: `go install github.com/wttech/aemc/cmd/aem@v1.0.1`,
- recently committed version: `go install github.com/wttech/aemc/cmd/aem@main`,

Use installed version of the tool instead of the one defined in file *aem/api.sh* by running the following command:

```shell
export AEM_CLI_VERSION=installed
```

To start using again version from wrapper file, simply unset the environment variable:

```shell
unset AEM_CLI_VERSION
```

## Ansible Collection

See a separate project based on AEM Compose: <https://github.com/wttech/aemc-ansible>

## Go Scripting

Consider implementing any application on top of AEM Compose API like using snippet below:

File: *aem.go*

```go
package main

import "fmt"
import "os"
import aemc "github.com/wttech/aemc/pkg"

func main() {
    aem := aemc.NewAem()
    instance := aem.InstanceManager().NewLocalAuthor()
    changed, err := instance.PackageManager().DeployWithChanged("/tmp/my-package.zip")
    if err != nil {
        fmt.Printf("cannot deploy package: %s\n", err)
        os.Exit(1)
    }
    if changed {
      aem.InstanceManager().AwaitStartedOne(instance)
    }
    fmt.Printf("package deployed properly\n")
    os.Exit(0)
}
```

Then to run application use command:

```shell
go run aem.go
```

# Dependencies

This tool is written in Go. Go applications are very often self-sufficient which means that they are not relying on platform-specific libraries/dependencies. 
The only requirement is to use proper tool binary distribution for each operating system and architecture.
Check out [releases page](https://github.com/wttech/aemc/releases) to review available binary distributions.

# Configuration

## Generating default configuration

To start working with tool run command:

```shell
aem config init
```

It will produce default configuration file named *aem.yml*. 
Correct the `dist_file`, `license_file`, `unpack_dir` properties to provide essential files to be able to launch AEM instances.

```yml
# AEM instances to work with
instance:

  # Defined by single value (only remote)
  config_url: ''

  # Defined strictly with full details (local or remote)
  config:
    local_author:
      http_url: http://127.0.0.1:4502
      user: admin
      password: admin
      run_modes: [ local ]
    local_publish:
      http_url: http://127.0.0.1:4503
      user: admin
      password: admin
      run_modes: [ local ]

  # Filters for defined
  filter:
    id: ''
    author: false
    publish: false

  # Tuning performance & reliability
  # 'auto'     - for more than 1 local instances - 'serial', otherwise 'parallel'
  # 'parallel' - for working with remote instances
  # 'serial'   - for working with local instances
  processing_mode: auto

  # State checking
  check:
    # Time to wait before first state checking (to avoid false-positives)
    warmup: 1s
    # Time to wait for next state checking
    interval: 5s
    # Number of successful check attempts that indicates end of checking
    done_threshold: 3
    # Wait only for those instances whose state has been changed internally (unaware of external changes)
    await_strict: true
    # Bundle state tracking
    bundle_stable:
      symbolic_names_ignored: []
    # OSGi events tracking
    event_stable:
      # Topics indicating that instance is not stable
      topics_unstable:
        - "org/osgi/framework/ServiceEvent/*"
        - "org/osgi/framework/FrameworkEvent/*"
        - "org/osgi/framework/BundleEvent/*"
      # Ignored service names to handle known issues
      details_ignored:
        - "*.*MBean"
        - "org.osgi.service.component.runtime.ServiceComponentRuntime"
        - "java.util.ResourceBundle"
      received_max_age: 5s
    # Sling Installer tracking
    installer:
      # JMX state checking
      state: true
      # Pause Installation nodes checking
      pause: true

  # Managed locally (set up automatically)
  local:
    # Current runtime dir (Sling launchpad, JCR repository)
    unpack_dir: "aem/home/data/instance"
    # Archived runtime dir (AEM backup files '*.aemb.zst')
    backup_dir: "aem/home/data/backup"

    # Source files
    quickstart:
      # AEM SDK ZIP or JAR
      dist_file: 'aem/home/lib/{aem-sdk,cq-quickstart}-*.{zip,jar}'
      # AEM License properties file
      license_file: "aem/home/lib/license.properties"

  # Package Manager
  package:
    # Force re-uploading/installing of snapshot AEM packages (just built / unreleased)
    snapshot_patterns: [ "**/*-SNAPSHOT.zip" ]
    # Use checksums to avoid re-deployments when snapshot AEM packages are unchanged
    snapshot_deploy_skipping: true

  # OSGi Framework
  osgi:
    bundle:
      install:
        start: true
        start_level: 20
        refresh_packages: true

# Java options used to launch AEM instances
java:
  # Java JRE/JDK location
  home_dir: {{ .Env.JAVA_HOME }}
  # Validate if following Java version constraints are met
  version_constraints: ">= 11, < 12"

# AEM application build
app:
  # Exclude the following paths when determining if the build should be executed or not
  sources_ignored:
    - "**/aem/home/**"
    - "**/.*"
    - "**/.*/**"
    - "!.content.xml"
    - "**/target"
    - "**/target/**"
    - "**/build"
    - "**/build/**"
    - "**/dist"
    - "**/dist/**"
    - "**/generated"
    - "**/generated/**"
    - "package-lock.json"
    - "**/package-lock.json"
    - "*.log"
    - "*.tmp"
    - "**/node_modules"
    - "**/node_modules/**"
    - "**/node"
    - "**/node/**"

base:
  # Location of temporary files (downloaded AEM packages, etc)
  tmp_dir: aem/home/tmp

log:
  level: info
  timestamp_format: "2006-01-02 15:04:05"
  full_timestamp: true

input:
  format: yml
  file: STDIN

output:
  format: text
  file: aem/home/aem.log
```

After instructing tool where the AEM instances files are located then, finally, instances may be created and launched:

```shell
aem instance create
aem instance start
```

## Configuration precedence

All configuration options specified in file *aem.yml* could be overridden by environment variables.
Simply add prefix `AEM_` then each level of nested YAML object join with `_` and lowercase the name of each object.

For example: `instance.local.quickstart.dist_file` could be overridden by environment variable `AEM_INSTANCE_LOCAL_QUICKSTART_DIST_FILE`

Also note that some configuration options may be ultimately overridden by CLI flags, like `--output-format`.

## Performance optimization

By default, fail-safe options are in use. However, to achieve maximum performance of the tool, considering setting these options:

```shell
export AEM_OUTPUT_MODE=none
export AEM_INSTANCE_PROCESSING_MODE=parallel
```

# Contributing

Issues reported or pull requests created will be very appreciated.

1. Fork plugin source code using a dedicated GitHub button.
2. See [development guide](DEVELOPMENT.md)
3. Do code changes on a feature branch created from *main* branch.
4. Create a pull request with a base of *main* branch.

# License

**AEM Compose** is licensed under the [Apache License, Version 2.0 (the "License")](https://www.apache.org/licenses/LICENSE-2.0.txt)
