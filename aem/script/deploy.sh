#!/usr/bin/env sh

# TODO <https://github.com/wttech/aemc/issues/7>

step "building AEM application"
mvn clean package
clc

step "deploying AEM application"
aem package deploy --file all/target/mysite.all-1.0.0-SNAPSHOT.zip
clc
