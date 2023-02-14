#!/usr/bin/env sh

step "build AEM application"
aem app build \
  --command "mvn clean package" \
  --sources "pom.xml,all,core,ui.apps,ui.apps.structure,ui.config,ui.content,ui.frontend,ui.tests" \
  --file "all/target/*.all-*.zip"
clc

step "deploy AEM application"
aem package deploy --file "all/target/*.all-*.zip"
clc

step "build AEM dispatcher"
(cd dispatcher/image && sh build.sh)
clc

step "deploy AEM dispatcher"
mkdir -p aem/home/var/dispatcher/httpd/logs aem/home/var/dispatcher/httpd/cache aem/home/var/dispatcher/httpd/htdocs && docker compose up -d
clc
