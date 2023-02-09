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

step "deploy AEM dispatcher"

# TODO <https://github.com/adobe/aem-project-archetype/issues/1043>
#sh aem/home/opt/sdk/dispatcher/bin/validate.sh dispatcher/src > aem/home/var/log/dispatcher-validate.log
#clc

docker tag "$(docker load --input "aem/home/opt/sdk/dispatcher/lib/dispatcher-publish-${ARCH}.tar.gz" | awk -v 'FS= ' '{print $3}')" "adobe/aem-ethos/dispatcher-publish:latest"
clc

docker compose up -d
clc
