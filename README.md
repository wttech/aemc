![AEM Compose Logo](https://github.com/wttech/aemc-ansible/raw/main/docs/logo-with-text.png)
[![WTT Logo](https://github.com/wttech/aemc-ansible/raw/main/docs/wtt-logo.png)](https://www.wundermanthompson.com/service/technology)

[![GitHub All Releases](https://img.shields.io/github/downloads/wttech/aemc/total)](https://github.com/wttech/aemc/releases)
[![Last Release Version](https://img.shields.io/github/v/release/wttech/aemc?color=lightblue&label=Last%20Release)](https://github.com/wttech/aemc/releases)
![Go Version](https://img.shields.io/github/go-mod/go-version/wttech/aemc)
[![Apache License, Version 2.0, January 2004](https://github.com/wttech/aemc-ansible/raw/main/docs/apache-license-badge.svg)](http://www.apache.org/licenses/)

**AEM Compose**

AEMC is a versatile tool for managing Adobe Experience Manager (AEM) instances. With a comprehensive set of commands, it simplifies tasks such as working CRX packages, OSGi configurations, JCR repository nodes, and more. Its seamless integration with Ansible enhances automation capabilities.

- Reusable core designed to handle advanced dev-ops operations needed to manage AEM instances
- Various distributions based on core for context-specific use cases:
  - [*CLI*](#cli---overview) - for developer workstations, shell scripting
    - [AEM Project Quickstart](https://github.com/wttech/aemc#cli---aem-project-quickstart) - add development environment automation to the existing AEM projects
    - [Docker Example](examples/docker) - for experiments only
  - [*Terraform AEM Provider*](https://github.com/wttech/terraform-provider-aem) - set up with ease higher environments (VMs and AEM instances) in the cloud (AWS, Azure, GCP, etc.)
  - [*Pulumi AEM Provider*](https://github.com/wttech/pulumi-aem-native) - working the same as Terraform AEM Provider but with Pulumi
  - [*Ansible Collection/Modules*](#ansible-collection) - for managing higher AEM environments
    - [Packer Example](https://github.com/wttech/aemc-ansible/tree/main/examples/packer) - starting point for baking AWS EC2 image using Ansible
    - [Local Example](https://github.com/wttech/aemc-ansible/tree/main/examples/local) - development & testing sandbox for AEM Compose project
- Fast & lightweight
- Usable on all operating systems and architectures without pre-installing any dependencies

# References

* Blog Posts
  * [Get your AEM together with AEM Compose!](https://wttech.blog/blog/2023/get-your-aem-together-with-aem-compose/) by [Krystian Panek](mailto:krystian.panek@wundermanthompson.com)
  * [Turning IDE into an AEM development powerhouse](https://wttech.blog/blog/2024/turning-ide-into-an-aem-development-powerhouse/) by [Krystian Panek](mailto:krystian.panek@wundermanthompson.com) and [Dominik Przybył](mailto:dominik.przybyl@wundermanthompson.com)
* Talk at AdaptTo 2023 Conference - [Get Your AEM Together: AEM Compose, the Ultimate DevEx Tool](https://www.youtube.com/watch?v=EH4ubsxNpbs) by [Tomasz Sobczyk](mailto:tomasz.sobczyk@wundermanthompson.com) & [Krystian Panek](mailto:krystian.panek@wundermanthompson.com)

[![AdaptTo 2023 Video](docs/adapto-to-video.png)](https://www.youtube.com/watch?v=EH4ubsxNpbs)

# Table of Contents

* [Distributions](#distributions)
  * [CLI - Overview](#cli---overview)
  * [CLI - Screenshots](#cli---screenshots)
  * [CLI - AEM Project Quickstart](#cli---aem-project-quickstart)
  * [IaaC Providers](#iaac-providers)
  * [Ansible Collection](#ansible-collection)
  * [Go Scripting](#go-scripting)
* [Dependencies](#dependencies)
* [Configuration](#configuration)
  * [Generating default configuration](#generating-default-configuration)
  * [Configuration precedence](#configuration-precedence)
  * [Context-specific customization](#context-specific-customization)
    * [Improving performance](#improving-performance)
    * [Increasing verbosity](#increasing-verbosity)
    * [Installing content packages](#installing-content-packages)
    * [Installing packages with troubleshooting](#installing-packages-with-troubleshooting)
* [Examples](#examples)
  * [Replication agents](#replication-agents)
  * [SSL by Default](#ssl-by-default)
  * [Global Trust Store](#global-trust-store)
* [Troubleshooting](#troubleshooting)
* [Contributing](#contributing)
* [Authors](#authors)
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
  - Crypto Support, Trust Store, SSL by Default
  - replication agents 
  - any repository nodes
- deploying AEM packages with:
  - efficiently upload with optimized memory usage (constant and independent of package size)
  - automatic workflow toggling - avoiding DAM asset renditions regeneration
  - advanced snapshot handling - avoiding redeploying the same package by checksum verification
  - customizable instance health checking
- building AEM packages with:
  - source code change detection - avoiding rebuilding application when it is not needed (handled by [Taskfile](https://taskfile.dev/usage/#prevent-unnecessary-work))
- making AEM instance backups (with restoring)
  - advanced archive format to speed up performance and storage efficiency ([ZSTD](https://github.com/facebook/zstd) used by default)
  - instance state aware - stopping, archiving then starting again AEM instances automatically (if needed)
- content synchronization between running instance and local file system

Worth knowing:

- On Windows use it with [Git Bash](https://gitforwindows.org/) ([CMD](https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/cmd) and [PowerShell](https://learn.microsoft.com/en-us/powershell/scripting/overview) are not supported nor tested)

## CLI - Screenshots

Help command:

<img src="docs/cli-help-home.png" alt="CLI Help Home" width="600"/>

Instance commands:

<img src="docs/cli-help-instance.png" alt="CLI Help Instance" width="320"/>

Package commands:

<img src="docs/cli-help-pkg.png" alt="CLI Help Package" width="320"/>

Setup task:

- AEM instances are provisioned only when the configuration is changed
- Maven AEM application build executed only when there are code changes
- AEM dispatcher set up using Docker Compose, redeployed only when there are configuration changes as well
- Health checks executed to confirm the stability of the environment as a whole

<img src="docs/cli-demo-short.gif" alt="CLI Demo"/>

## CLI - AEM Project Quickstart

Supported project types:

- with structure based on [Adobe AEM Project Archetype](https://github.com/adobe/aem-project-archetype#usage), compatibility:

  | AEM Compose (init) | AEM Project Archetype (tested) |
  |--------------------|--------------------------------|
  | >= 1.2.0           | 41, 42, 43, 47, 48             |

- with any type of structure, however afterwards only a little customizations in *Taskfile.yml* need to be done to reflect configuration related to built AEM application artifact path and AEM dispatcher files location
- empty folder; the project kind will be needed to be specified explicitly during initialization

---

Run command below to initialize the AEM Compose tool in your project:

```shell
curl https://raw.githubusercontent.com/wttech/aemc/main/project-init.sh | sh
```

and then:

```shell
sh aemw init
```

After successful initialization, remember to always use the tool via wrapper script in the following way:

```shell
sh aemw [command]
```

For example:

```shell
sh aemw version
```

Project initialization sets up ready-to-use tasks powered by [Task tool](https://taskfile.dev/) which are aggregating one or many AEM Compose CLI commands into useful procedures.

To list all available tasks, run:

```shell
sh taskw --list
```

For example:

```shell
sh taskw setup
```

Some tasks like `aem:build` may accept parameters.
For example, to build AEM application with:

- Applying frontend development mode Maven profile
- Unit tests skipping 
- UI tests skipping

Simply run command with appending [task variable](https://taskfile.dev/usage/#variables) to the end:

```shell
sh taskw aem:build AEM_BUILD_ARGS="-PfedDev -DskipTests -pl '!ui.tests'"
```

## IaaC Providers

The tool is designed to be used in Infrastructure as a Code (IaaC) solutions such as [Terraform](https://www.terraform.io/) or [Pulumi](https://www.pulumi.com/).

See separate projects based on AEM Compose:

- Terraform Provider: <https://github.com/wttech/terraform-provider-aem>
- Pulumi Provider: <https://github.com/wttech/pulumi-aem>

## Ansible Collection

The tool is designed to be used in Ansible playbooks and roles.

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
    aem := aemc.DefaultAEM()
    instance, _ := aem.InstanceManager().NewByIDAndURL("remote_author", "http://admin:admin@192.168.1.2:4502")
    changed, err := instance.PackageManager().DeployWithChanged("/tmp/my-package.zip")
    if err != nil {
        fmt.Printf("cannot deploy package: %s\n", err)
        os.Exit(1)
    }
    if changed {
      if err := aem.InstanceManager().AwaitStartedOne(*instance); err != nil {
		  fmt.Printf("instance not stable after deploying package: %s\n", err)
		  os.Exit(1)
      }
    }
    fmt.Println("package deployed properly")
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

## Defaults

The tool tries to make configuration as much explicit as it could be to allow customization in an easy way.

Below are listed all available configuration options ([source](pkg/project/instance/aem/default/etc/aem.yml)):

```shell
# AEM instances to work with
instance:

  # Full details of local or remote instances
  config:
    local_author:
      active: [[.Env.AEM_AUTHOR_ACTIVE | default true ]]
      http_url: [[.Env.AEM_AUTHOR_HTTP_URL | default "http://127.0.0.1:4502" ]]
      user: [[.Env.AEM_AUTHOR_USER | default "admin" ]]
      password: [[.Env.AEM_AUTHOR_PASSWORD | default "admin" ]]
      run_modes: [ local ]
      jvm_opts:
        - -server
        - -Djava.awt.headless=true
        - -Djava.io.tmpdir=[[canonicalPath .Path "aem/home/tmp"]]
        - -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=[[.Env.AEM_AUTHOR_DEBUG_ADDR | default "0.0.0.0:14502" ]]
        - -Duser.language=en
        - -Duser.country=US
        - -Duser.timezone=UTC
      start_opts: []
      secret_vars:
        - ACME_SECRET=value
      env_vars:
        - ACME_VAR=value
      sling_props: []
    local_publish:
      active: [[.Env.AEM_PUBLISH_ACTIVE | default true ]]
      http_url: [[.Env.AEM_PUBLISH_HTTP_URL | default "http://127.0.0.1:4503" ]]
      user: [[.Env.AEM_PUBLISH_USER | default "admin" ]]
      password: [[.Env.AEM_PUBLISH_PASSWORD | default "admin" ]]
      run_modes: [ local ]
      jvm_opts:
        - -server
        - -Djava.awt.headless=true
        - -Djava.io.tmpdir=[[canonicalPath .Path "aem/home/tmp"]]
        - -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=[[.Env.AEM_PUBLISH_DEBUG_ADDR | default "0.0.0.0:14503" ]]
        - -Duser.language=en
        - -Duser.country=US
        - -Duser.timezone=UTC
      start_opts: []
      secret_vars:
        - ACME_SECRET=value
      env_vars:
        - ACME_VAR=value
      sling_props: []

  # Tuning performance & reliability
  # 'auto'     - for more than 1 local instances - 'serial', otherwise 'parallel'
  # 'parallel' - for working with remote instances
  # 'serial'   - for working with local instances
  processing_mode: auto

  # HTTP client settings
  http:
    timeout: 10m
    debug: false
    disable_warn: true

  # State checking
  check:
    # Time to wait before first state checking (to avoid false-positives)
    warmup: 1s
    # Time to wait for next state checking
    interval: 6s
    # Number of successful check attempts that indicates end of checking
    done_threshold: 4
    # Max time to wait for the instance to be healthy after executing the start script or e.g deploying a package
    await_started:
      timeout: 30m
    # Max time to wait for the instance to be stopped after executing the stop script
    await_stopped:
      timeout: 10m
    # Max time in which socket connection to instance should be established
    reachable:
      timeout: 3s
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
    # OSGi components state tracking
    component_stable:
      skip: false
      pids:
        include: ['com.day.crx.packaging.*', 'org.apache.sling.installer.*']
        exclude: ['org.apache.sling.installer.hc.*', 'org.apache.sling.installer.core.impl.console.*']
        match:
          "disabled": []
          "no config": []
          "unsatisfied (reference)": []
          "satisfied": []
    # Sling Installer tracking
    installer:
      # JMX state checking
      state: true
      # Pause Installation nodes checking
      pause: true

  # Managed locally (set up automatically)
  local:
    # Wait only for those instances whose state has been changed internally (unaware of external changes)
    await_strict: true

    # Current runtime dir (Sling launchpad, JCR repository)
    unpack_dir: "aem/home/var/instance"
    # Archived runtime dir (AEM backup files '*.aemb.zst')
    backup_dir: "aem/home/var/backup"

    # Oak Run tool options (offline instance management)
    oak_run:
      download_url: "https://repo1.maven.org/maven2/org/apache/jackrabbit/oak-run/1.44.0/oak-run-1.44.0.jar"
      store_path: "crx-quickstart/repository/segmentstore"

    # Source files
    quickstart:
      # AEM SDK ZIP or JAR
      dist_file: "aem/home/lib/{aem-sdk,cq-quickstart}-*.{zip,jar}"
      # AEM License properties file
      license_file: "aem/home/lib/license.properties"

  # Status discovery (timezone, AEM version, etc)
  status:
    timeout: 500ms

  # JCR Repository
  repo:
    property_change_ignored:
      # AEM assigns them automatically
      - "jcr:created"
      - "cq:lastModified"
      # AEM encrypts it right after changing by replication agent setup command
      - "transportPassword"

  # CRX Package Manager
  package:
    # Force re-uploading/installing of snapshot AEM packages (just built / unreleased)
    snapshot_patterns: [ "**/*-SNAPSHOT.zip" ]
    # Use checksums to avoid re-deployments when snapshot AEM packages are unchanged
    snapshot_deploy_skipping: true
    # Disable following workflow launchers for a package deployment time only
    toggled_workflows: [/libs/settings/workflow/launcher/config/asset_processing_on_sdk_*,/libs/settings/workflow/launcher/config/update_asset_*,/libs/settings/workflow/launcher/config/dam_*]
    # Also sub-packages
    install_recursive: true
    # Use slower HTML endpoint for deployments but with better troubleshooting
    install_html:
      enabled: false
      # Print HTML directly to console instead of writing to file
      console: false
      # Fail on case 'installed with errors'
      strict: true
    # Number of changes after which the commit to the repository is performed
    install_save_threshold: 1024
    # Allows to relax dependency handling if needed
    install_dependency_handling: required
    # Controls how 'rep:policy' nodes are handled during import
    install_ac_handling: ''

  # OSGi Framework
  osgi:
    bundle:
      install:
        start: true
        start_level: 20
        refresh_packages: true

  # Crypto Support
  crypto:
    key_bundle_symbolic_name: com.adobe.granite.crypto.file

  # Workflow Manager
  workflow:
    launcher:
      lib_root: /libs/settings/workflow/launcher
      config_root: /conf/global/settings/workflow/launcher
      toggle_retry:
        timeout: 10m
        delay: 10s

java:
  # Require following versions before e.g running AEM instances
  version_constraints: ">= 11, < 12"

  # Pre-installed local JDK dir
  # a) keep it empty to download open source Java automatically for current OS and architecture
  # b) set it to absolute path or to env var '[[.Env.JAVA_HOME]]' to indicate where closed source Java like Oracle is installed
  home_dir: ""

  # Auto-installed JDK options
  download:
    # Source URL with template vars support
    url: "https://github.com/adoptium/temurin11-binaries/releases/download/jdk-11.0.18%2B10/OpenJDK11U-jdk_[[.Arch]]_[[.Os]]_hotspot_11.0.18_10.[[.ArchiveExt]]"
    # Map source URL template vars to be compatible with Adoptium Java
    replacements:
      # Var 'Os' (GOOS)
      "darwin": "mac"
      # Var 'Arch' (GOARCH)
      "x86_64": "x64"
      "amd64": "x64"
      "386": "x86-32"
      # enforce non-ARM Java as some AEM features are not working on ARM (e.g Scene7)
      "arm64": "x64"
      "aarch64": "x64"

vault:
  "download_url": "https://repo1.maven.org/maven2/org/apache/jackrabbit/vault/vault-cli/3.7.2/vault-cli-3.7.2-bin.tar.gz"

base:
  # Location of temporary files (downloaded AEM packages, etc)
  tmp_dir: aem/home/tmp
  # Location of supportive tools (downloaded Java, OakRun, unpacked AEM SDK)
  tool_dir: aem/home/opt

log:
  level: info
  timestamp_format: "2006-01-02 15:04:05"
  full_timestamp: true

input:
  format: yml
  file: STDIN

output:
  format: text
  log:
    # File path of logs written especially when output format is different than 'text'
    file: aem/home/var/log/aem.log
    # Controls where outputs and logs should be written to when format is 'text' (console|file|both)
    mode: console

# Content clean options
content:
  clean:
    # File patterns to be deleted
    files_deleted:
      - patterns:
          - "**/.vlt"
          - "**/.vlt*.tmp"
          - "**/install/*.jar"
    # File patterns to be flattened
    files_flattened:
      - "**/_cq_design_dialog/.content.xml"
      - "**/_cq_dialog/.content.xml"
      - "**/_cq_htmlTag/.content.xml"
      - "**/_cq_template/.content.xml"
    # Property patterns to be skipped, removed from cleaned file
    properties_skipped:
      - patterns: "jcr:uuid"
        excluded_paths: [ "**/home/users/*", "**/home/groups/*" ]
      - patterns: "cq:lastModified*"
        excluded_paths: [ "**/content/experience-fragments/*" ]
      - patterns: [ "dam:sha1", "dam:size" ]
        included_paths: [ "**/content/dam/*.svg/*" ]
      - patterns:
          - "jcr:lastModified*"
          - "jcr:created*"
          - "jcr:isCheckedOut"
          - "cq:lastReplicat*"
          - "cq:lastRolledout*"
          - "dam:extracted"
          - "dam:assetState"
          - "dc:modified"
          - "*_x0040_*"
          - "cq:name"
          - "cq:parentPath"
          - "dam:copiedAt"
          - "dam:parentAssetID"
          - "dam:relativePath"
    # Mixin type patterns to be skipped, removed from cleaned file
    mixin_types_skipped:
      - patterns:
          - "cq:ReplicationStatus"
          - "mix:versionable"
    # Unused namespaces to be skipped, removed from cleaned file
    namespaces_skipped: true
```

Note that environment variables may be injected in any part of config file. 
Environment variables could be defined in one or many [dotenv files](https://github.com/wttech/aemc/blob/main/pkg/common/osx/osx.go#L32).

## Overriding default configuration

By default, the VCS-tracked file is loaded (*aem/default/etc*).
However, occasionally developers might want to override the default config file and load a VCS-ignored file instead (*aem/home/etc/aem.yml*).

To do so, run the command:

```shell
sh aemw config init
```

## Configuration precedence

All configuration options specified in file *aem.yml* could be overridden by environment variables.
Simply add prefix `AEM_` then each level of nested YAML object join with `_` and lowercase the name of each object.

For example: `instance.local.quickstart.dist_file` could be overridden by environment variable `AEM_INSTANCE_LOCAL_QUICKSTART_DIST_FILE`

Also note that some configuration options may be ultimately overridden by CLI flags, like `--output-format`.

## Context-specific customization

By default, fail-safe options are in use. However, consider using the configuration options listed below to achieve a more desired tool experience. 

### Improving performance

  ```shell
  export AEM_INSTANCE_PROCESSING_MODE=parallel
  ```

  This setting will significantly reduce command execution time. Although be aware that when deploying heavy AEM packages like Service Packs on the same machine in parallel, a heavy load could be observed, which could lead to unpredicted AEM CMS and AEM Compose tool behavior.

### Increasing verbosity

  ```shell
  export AEM_OUTPUT_VALUE=ALL
  ```

  Setting this environment variable will instruct the tool to request from the AEM instance descriptive information about the recently executed command subject.
  For example, if a recently executed command was `sh aemw package deploy my-package.zip -A` the AEM Compose tool after doing the actual package deployment will request from CRX Package Manager the exact information about just deployed package.
  This feature is beneficial for clarity and debugging purposes.

### Installing content packages

To install larger AEM packages that may include content pages, assets, and more, you can adjust the HTTP timeout setting. 
By default, the timeout is set to `10m`, but you have the option to increase it (e.g., to `3h`) or disable it completely (by using `0`).

To set the timeout for a single AEMC command, use the following syntax:

```shell
AEM_INSTANCE_HTTP_TIMEOUT=0 sh aemw package deploy --file my-package.zip
```

It's important to be aware that AEMaaCS also has its own timeout for requests made to the Package Manager UI. For detailed information, please refer to the [documentation](https://experienceleague.adobe.com/docs/experience-manager-cloud-service/content/implementing/developer-tools/package-manager.html?lang=en#aemaacs-packages).

To skip the instance health check after running a command that would normally trigger it, `AEM_INSTANCE_CHECK_SKIP` environment variable can be used. This can save a lot of time when deploying a lot of packages sequentially.

To skip the instance health check for a single AEMC command, use the following syntax:

```shell
AEM_INSTANCE_CHECK_SKIP=true sh aemw package deploy --file my-package.zip
```

### Installing packages with troubleshooting

Starting from version 1.4.0 (see [#177](https://github.com/wttech/aemc/pull/177)), AEMC now supports AEM package installations using an HTML report serving endpoint, similar to CRX Package Manager's. While this method may result in slightly extended installation times, it provides a comprehensive HTML report detailing the package installation process.

This new feature offers two distinct modes for leveraging its benefits:

1. Saving HTML report to file:

   ```shell
   AEM_INSTANCE_PACKAGE_INSTALL_HTML_ENABLED=true sh aemw package deploy --file my-package.zip
   ```

2. Direct console output of HTML report:

   ```shell
   AEM_INSTANCE_PACKAGE_INSTALL_HTML_ENABLED=true AEM_INSTANCE_PACKAGE_INSTALL_HTML_CONSOLE=true sh aemw package deploy --file my-package.zip
   ```

# Examples

## Replication agents

1. Configuring publish agent on AEM author:

    ```shell
    PROPS="
    enabled: true
    transportUri: http://localhost:4503/bin/receive?sling:authRequestLogin=1
    transportUser: admin
    transportPassword: admin
    userId: admin
    "
    echo "$PROPS" | sh aemw repl agent setup -A --location "author" --name "publish"
    ```

2. Configuring flush agent on AEM publish:

    ```shell
    PROPS="
    enabled: true
    transportUri: http://localhost/dispatcher/invalidate.cache
    protocolHTTPHeaders:
    - 'CQ-Action: {action}'
    - 'CQ-Handle: {path}'
    - 'CQ-Path: {path}'
    - 'Host: flush'
    "
    echo "$PROPS" | sh aemw repl agent setup -P --location "publish" --name "flush"
    ```
   If needed, update `localhost` to the value on which AEM dispatcher is available, e.g.`localhost:8080`.

## SSL by Default

AEM Compose supports *SSL by Default* feature of AEM.

This feature requires:
- certificate file in PEM format
- private key file in DER or PEM format
- password for keystore (has to be the same for each invocation of the tool)
- password for truststore (can be different for each invocation of the tool)
- hostname for HTTPS connector (used by AEM to check if the setup was successful; has to be reachable by AEM)
- port for HTTPS connector

To set up *SSL by Default*, run:

```shell
sh aemw ssl setup \
   --instance-author
   --keystore-password password \
   --truststore-password password1 \
   --certificate-file localhost.crt \
   --private-key-file localhostprivate.key \
   --https-hostname localhost \
   --https-port 8443

sh aemw ssl setup \
   --instance-publish
   --keystore-password password \
   --truststore-password password1 \
   --certificate-file localhost.crt \
   --private-key-file localhostprivate.key \
   --https-hostname localhost \
   --https-port 9443
```

Above example uses the self-signed certificate created as per the AEM docs.

See the reference documentation: [AEM 6.5 > Administering Guide > SSL by Default](https://experienceleague.adobe.com/docs/experience-manager-65/administering/security/ssl-by-default.html?lang=en)

For local environment remember to set different port numbers for author and publish instances.

## Global Trust Store

AEM Compose supports managing the trust store of AEM instances.

This feature supports:

- creation of the general trust store if it does not exist
    ```shell
    sh aemw gts create --password PASSWORD_HERE -A
    ```

- getting the general trust store status
    ```shell
    sh aemw gts status -A
    ```

- adding a certificate to the general trust store
    ```shell
    sh aemw gts certificate add --path <path> -A
    ```
  
This command will add a certificate to the general trust store only if not exists in trust store and will return the alias of the certificate.
Command `certificate add` supports certificates in PEM and DER formats.

- reading a certificate from the general trust store (by alias)
    ```shell
    sh aemw gts certificate read --alias <alias> -A
    ```

- removing a certificate from the general trust store (by alias)
    ```shell
    sh aemw gts certificate remove --alias <alias> -A
    ```


# Troubleshooting

## Migration from Gradle plugins

If you're migrating from [Gradle AEM Plugin](https://github.com/wttech/gradle-aem-plugin) you may run into issue with setting up dispatcher.
In case of error when running task dispatcher:start or dispatcher:setup:

```
2024-01-22 16:37:23 dispatcher  | 2024/01/22 15:37:23 Dispatcher configuration validation failed:
2024-01-22 16:37:23 dispatcher  | 2024/01/22 15:37:23   /mnt/dev/src/conf.d/variables/default.vars: lstat /etc/httpd.extra: no such file or directory
2024-01-22 16:37:23 dispatcher  |
2024-01-22 16:37:23 dispatcher  | ERROR Mon Jan 22 15:37:23 UTC 2024 Configuration invalid, please fix and retry,
2024-01-22 16:37:23 dispatcher  |               Line numbers reported are correct for your configuration files.
```

Check if there is a broken symlink/file in `dispatcher/src/variables/default.vars`. If the file exists, delete it and
run the task again.

# Contributing

Issues reported or pull requests created will be very appreciated.

1. Fork plugin source code using a dedicated GitHub button.
2. See [development guide](DEVELOPMENT.md)
3. Do code changes on a feature branch created from *main* branch.
4. Create a pull request with a base of *main* branch.

# Authors

- Creator, owner, and maintainer: [Krystian Panek](mailto:krystian.panek@wundermanthompson.com)
- Consultancy: [Tomasz Sobczyk](mailto:tomasz.sobczyk@wundermanthompson.com), [Maciej Majchrzak](mailto:maciej.majchrzak@wundermanthompson.com)
- Contributors: [&lt;see all&gt;](https://github.com/wttech/aemc/graphs/contributors)

# License

**AEM Compose** is licensed under the [Apache License, Version 2.0 (the "License")](https://www.apache.org/licenses/LICENSE-2.0.txt)
