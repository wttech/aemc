#!/usr/bin/env sh

. aem/script/undeploy.sh
. aem/script/down.sh

step "destroying instances"
aem instance destroy

