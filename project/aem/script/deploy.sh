#!/usr/bin/env sh

#PACKAGE_FILE=$(aem file find --file "all/target/*.all-*.zip" --output-value "file" 2>&1)
#PACKAGE_CHECKSUM_FILE="aem/home/build.md5"
#
#build_package() {
#  step "build AEM application"
#  step "check build progress using command 'tail -f aem/home/build.log'"
#  mvn clean package -l aem/home/build.log
#  clc
#}
#if [ "$PACKAGE_FILE" = "<undefined>" ] ; then
#  build_package
#else
#  PACKAGE_CHECKSUM_CURRENT=$(aem file checksum --path . --includes "pom.xml,all/**,core/**,ui.apps/**,ui.apps.structure/**,ui.config/**,ui.content/**" --output-value "checksum" 2>&1)
#  PACKAGE_CHECKSUM_PREVIOUS=$(cat "$PACKAGE_CHECKSUM_FILE" 2>/dev/null)
#  if [ "$PACKAGE_CHECKSUM_CURRENT" != "$PACKAGE_CHECKSUM_PREVIOUS" ]; then
#    step "changed AEM application detected"
#    build_package
#    echo "$PACKAGE_CHECKSUM_CURRENT" > "$PACKAGE_CHECKSUM_FILE"
#  else
#    step "build AEM application (skipped)"
#  fi
#fi

#step "deploy AEM application"
#aem package deploy --file "$PACKAGE_FILE"
#clc
