FROM --platform=linux/x86_64 rockylinux:8.7

ADD src /opt/aemc
WORKDIR /opt/aemc
ENTRYPOINT ["/bin/bash"]

ENV TERM=xterm
ENV AEM_INSTANCE_CONFIG_LOCAL_AUTHOR_ACTIVE=false
ENV AEM_INSTANCE_CONFIG_LOCAL_PUBLISH_ACTIVE=true

RUN sh taskw start
RUN sh taskw provision
