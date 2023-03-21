![AEM Compose Logo](https://github.com/wttech/aemc-ansible/raw/main/docs/logo-with-text.png)
[![WTT Logo](https://github.com/wttech/aemc-ansible/raw/main/docs/wtt-logo.png)](https://www.wundermanthompson.com/service/technology)

[![Apache License, Version 2.0, January 2004](https://github.com/wttech/aemc-ansible/raw/main/docs/apache-license-badge.svg)](http://www.apache.org/licenses/)

# AEM Compose - Docker Example

Setup and launch AEM instances as Docker containers.

**Warning!** The purpose of this example is to demonstrate and experiment with AEM running in Docker. However, it is widely known that AEM runtime does not fit well into Docker architecture (e.g is not lightweight and stateless).

## Prerequisites

- Docker 20.x and higher
- AEM source files put into directory *src/aem/home/lib*
  - a) AEM SDK ZIP 
  - b) AEM On-Prem JAR and license file

# Usage 

  1. Build images using command:
        
      ```shell
      sh build-all.sh 
      ```
  2. Run containers using command:
    
      ```shell
      sh run-all.sh 
      ```
