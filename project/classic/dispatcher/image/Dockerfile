FROM rockylinux:8

# Install required OS packages
RUN yum -y install httpd wget

# Download & install dispatcher module
RUN mkdir -p /tmp/dispatcher && cd /tmp/dispatcher && \
    wget https://download.macromedia.com/dispatcher/download/dispatcher-apache2.4-linux-aarch64-4.3.5.tar.gz -O dispatcher.tar.gz && \
    tar xzf dispatcher.tar.gz && \
    cp dispatcher-apache2.4-4.3.5.so /usr/lib64/httpd/modules/mod_dispatcher.so && \
    chmod 0755 /usr/lib64/httpd/modules/mod_dispatcher.so

# Project-specific source code
COPY src.origin /etc/httpd

# Local environment-specific variables
COPY src.override/conf.d/custom.conf /etc/httpd/conf.d/custom.conf

# Allow invalidating from any host
COPY src.override/conf.dispatcher.d/cache/ams_author_invalidate_allowed.any /etc/httpd/conf.dispatcher.d/cache/ams_author_invalidate_allowed.any
COPY src.override/conf.dispatcher.d/cache/ams_publish_invalidate_allowed.any /etc/httpd/conf.dispatcher.d/cache/ams_publish_invalidate_allowed.any

# Deactivate mod-ssl dependencies
COPY src.override/conf.d/proxy/mock.proxy /etc/httpd/conf.d/proxy/mock.proxy
COPY src.override/conf.d/rewrites/xforwarded_forcessl_rewrite.rules /etc/httpd/conf.d/rewrites/xforwarded_forcessl_rewrite.rules

# https://httpd.apache.org/docs/2.4/stopping.html#gracefulstop
STOPSIGNAL SIGWINCH

RUN bash -c 'ls -la /usr/lib64/httpd/modules'

COPY httpd-foreground /bin/httpd-foreground
EXPOSE 80
CMD ["/bin/httpd-foreground"]
