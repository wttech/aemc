FROM --platform=linux/x86_64 rockylinux:8.7

WORKDIR /opt/aemc

RUN curl https://raw.githubusercontent.com/wttech/aemc/main/project-init.sh | sh && \
    sh aemw init --project-kind instance

ADD aem/home/lib /opt/aemc/aem/home/lib

RUN sh aemw instance init

ENTRYPOINT ["/bin/bash"]
