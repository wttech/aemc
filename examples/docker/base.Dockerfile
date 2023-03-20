FROM rockylinux:8.7

RUN mkdir -p /opt/aem && \
    cd /opt/aem && \
    curl https://raw.githubusercontent.com/wttech/aemc/main/project-init.sh | sh && \
    sh aemw init --project-kind instance

ADD ./aem/home/lib /opt/aem/home/lib

ENTRYPOINT ["/bin/bash"]
WORKDIR /opt/aem
