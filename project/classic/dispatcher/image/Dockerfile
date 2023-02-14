FROM httpd:2.4.55-alpine3.17

RUN apk update && apk add ca-certificates && update-ca-certificates && apk add openssl wget
RUN wget https://download.macromedia.com/dispatcher/download/dispatcher-apache2.4-linux-x86_64-4.3.5.tar.gz

# Project-specific source code
COPY etc.src/conf.d /etc/httpd/conf.d
COPY etc.src/conf.dispatcher.d /etc/httpd/conf.dispatcher.d

# Local environment-specific variables
COPY etc/httpd/conf.d/custom.conf /etc/httpd/conf.d/custom.conf

# Allow invalidating from any host
COPY etc/httpd/conf.dispatcher.d/cache/ams_author_invalidate_allowed.any /etc/httpd/conf.dispatcher.d/cache/ams_author_invalidate_allowed.any
COPY etc/httpd/conf.dispatcher.d/cache/ams_publish_invalidate_allowed.any /etc/httpd/conf.dispatcher.d/cache/ams_publish_invalidate_allowed.any

# Load dispatcher module
COPY etc/httpd/conf.modules.d/02-dispatcher.conf /etc/httpd/conf.modules.d/02-dispatcher.conf

# Deactivate mod-ssl dependencies
COPY etc/httpd/conf.d/proxy/mock.proxy /etc/httpd/conf.d/proxy/mock.proxy
COPY etc/httpd/conf.d/rewrites/xforwarded_forcessl_rewrite.rules /etc/httpd/conf.d/rewrites/xforwarded_forcessl_rewrite.rules

