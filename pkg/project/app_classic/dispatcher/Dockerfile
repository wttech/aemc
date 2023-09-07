FROM --platform=linux/amd64 rockylinux:8
ENV DISPATCHER_VERSION=4.3.6
ENV DISP_ID docker

# Install required OS packages
RUN yum -y install httpd wget mod_ssl

# Download & install dispatcher module
RUN mkdir -p /tmp/dispatcher && cd /tmp/dispatcher && \
    wget https://download.macromedia.com/dispatcher/download/dispatcher-apache2.4-linux-x86_64-${DISPATCHER_VERSION}.tar.gz -O dispatcher.tar.gz && \
    tar xzf dispatcher.tar.gz && \
    cp dispatcher-apache2.4-${DISPATCHER_VERSION}.so /usr/lib64/httpd/modules/mod_dispatcher.so && \
    chmod 0755 /usr/lib64/httpd/modules/mod_dispatcher.so

# Project-specific source code
COPY home/src /etc/httpd

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
RUN mkdir -p /var/www/cache/author /var/www/cache/publish /mnt/var/www/default && \
    chown apache:apache -R /var/www /mnt/var/www

# Fix for 'SSLCertificateFile: file '/etc/pki/tls/certs/localhost.crt' does not exist or is empty'
RUN /usr/libexec/httpd-ssl-gencerts

# https://httpd.apache.org/docs/2.4/stopping.html#gracefulstop
STOPSIGNAL SIGWINCH

COPY docker/httpd-foreground /bin/httpd-foreground
EXPOSE 80
CMD ["/bin/httpd-foreground"]
