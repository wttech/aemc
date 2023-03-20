FROM acme/aem/base:latest

ENV AEM_INSTANCE_CONFIG_LOCAL_AUTHOR_ACTIVE=true
ENV AEM_INSTANCE_CONFIG_LOCAL_PUBLISH_ACTIVE=false

ADD ./aem/home/lib /opt/aem/home/lib
