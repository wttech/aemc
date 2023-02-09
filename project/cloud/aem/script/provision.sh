step "configuring replication agent for publishing content"
PROPS="
enabled: true
transportUri: http://127.0.0.1:4503/bin/receive?sling:authRequestLogin=1
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
