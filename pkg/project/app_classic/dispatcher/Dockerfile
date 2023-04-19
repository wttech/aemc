FROM rockylinux:8
ENV DISPATCHER_VERSION=4.3.5

# Install required OS packages
RUN yum -y install httpd wget mod_ssl

# Download & install dispatcher module
RUN mkdir -p /tmp/dispatcher && cd /tmp/dispatcher && \
    wget https://download.macromedia.com/dispatcher/download/dispatcher-apache2.4-linux-x86_64-${DISPATCHER_VERSION}.tar.gz -O dispatcher.tar.gz && \
    tar xzf dispatcher.tar.gz && \
    cp dispatcher-apache2.4-${DISPATCHER_VERSION}.so /usr/lib64/httpd/modules/mod_dispatcher.so && \
    chmod 0755 /usr/lib64/httpd/modules/mod_dispatcher.so

# Project-specific source code
COPY src /etc/httpd

# Recreate symlinks (needed when building on Windows)
RUN cd /etc/httpd/conf.d && rm -fr enabled_vhosts && mkdir enabled_vhosts && cd enabled_vhosts && \
    ln -s ../available_vhosts/aem_author.vhost aem_author.vhost && \
    ln -s ../available_vhosts/aem_flush.vhost aem_flush.vhost && \
    ln -s ../available_vhosts/aem_health.vhost aem_health.vhost && \
    ln -s ../available_vhosts/mysite_publish.vhost mysite_publish.vhost

RUN cd /etc/httpd/conf.dispatcher.d && rm -fr enabled_farms && mkdir enabled_farms && cd enabled_farms && \
    ln -s ../available_farms/000_ams_author_farm.any 000_ams_author_farm.any && \
    ln -s ../available_farms/999_ams_publish_farm.any 999_ams_publish_farm.any

# Local environment-specific variables
COPY docker/src/conf.d/custom.conf /etc/httpd/conf.d/custom.conf
COPY docker/src/conf.d/variables/default.vars /etc/httpd/conf.d/variables/default.vars

# Allow invalidating from any host
COPY docker/src/conf.dispatcher.d/cache/ams_author_invalidate_allowed.any /etc/httpd/conf.dispatcher.d/cache/ams_author_invalidate_allowed.any
COPY docker/src/conf.dispatcher.d/cache/ams_publish_invalidate_allowed.any /etc/httpd/conf.dispatcher.d/cache/ams_publish_invalidate_allowed.any

# Deactivate mod-ssl dependencies
COPY docker/src/conf.d/proxy/mock.proxy /etc/httpd/conf.d/proxy/mock.proxy
COPY docker/src/conf.d/rewrites/xforwarded_forcessl_rewrite.rules /etc/httpd/conf.d/rewrites/xforwarded_forcessl_rewrite.rules

# Initialize doc roots

RUN mkdir -p /var/www/localhost/author/cache /var/www/localhost/publish/cache /mnt/var/www/default

# Fix for 'SSLCertificateFile: file '/etc/pki/tls/certs/localhost.crt' does not exist or is empty'
RUN /usr/libexec/httpd-ssl-gencerts

# https://httpd.apache.org/docs/2.4/stopping.html#gracefulstop
STOPSIGNAL SIGWINCH

COPY docker/httpd-foreground /bin/httpd-foreground
EXPOSE 80
CMD ["/bin/httpd-foreground"]
