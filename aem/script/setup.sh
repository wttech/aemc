#!/usr/bin/env sh

. aem/script/up.sh

step "configuring replication agent on author instance"
PROPS="
enabled: true
transportUri: http://localhost:4503/bin/receive?sling:authRequestLogin=1
transportUser: admin
transportPassword: admin
userId: admin
"
echo "$PROPS" | aem repl agent setup -A --location "author" --name "publish"
clc

step "saving some node"
echo "foo: bar3" | aem repo node save -A --path "/content/foo"
clc

step "enabling CRX/DE"
aem osgi config save --pid "org.apache.sling.jcr.davex.impl.servlets.SlingDavExServlet" --input-string "alias: /crx/server"
clc

step "downloading APM"
# TODO impl this
#aem file download --url "https://github.com/wttech/APM/releases/download/apm-5.5.1/apm-all-5.5.1.zip" --file "${DOWNLOAD_DIR}/apm-all-5.5.1.zip"
#clc
downloadFile "https://github.com/wttech/APM/releases/download/apm-5.5.1/apm-all-5.5.1.zip" "${DOWNLOAD_DIR}/apm-all-5.5.1.zip"

step "deploying APM"
# TODO impl this
#aem package deploy --url "https://github.com/wttech/APM/releases/download/apm-5.5.1/apm-all-5.5.1.zip"
aem package deploy --file "${DOWNLOAD_DIR}/apm-all-5.5.1.zip"
clc
