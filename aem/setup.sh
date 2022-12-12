#!/usr/bin/env sh

. aem/up.sh

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

step "deploying APM"
wget https://github.com/wttech/APM/releases/download/apm-5.5.1/apm-all-5.5.1.zip -nc -O /tmp/apm-all-5.5.1.zip

aem package deploy --file /tmp/apm-all-5.5.1.zip
clc
