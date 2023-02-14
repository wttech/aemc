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

step "configuring replication agent for flushing cached content"
PROPS="
enabled: true
transportUri: http://127.0.0.1/dispatcher/invalidate.cache
protocolHTTPHeaders:
  - CQ-Action: {action}
  - CQ-Handle: {path}
  - CQ-Path: {path}
  - Host: publish
"
echo "$PROPS" | aem repl agent setup -P --location "publish" --name "flush"
clc

step "enabling CRX/DE"
aem osgi config save --pid "org.apache.sling.jcr.davex.impl.servlets.SlingDavExServlet" --input-string "alias: /crx/server"
clc

step "deploying APM"
aem package deploy --url "https://github.com/wttech/APM/releases/download/apm-5.5.1/apm-all-5.5.1.zip"
clc

step "deploying AEM service pack (it may take a several minutes)"
aem package deploy --file "aem/home/lib/aem-service-pkg-6.5.*.0.zip"
clc
