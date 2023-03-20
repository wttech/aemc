FROM --platform=linux/x86_64 rockylinux:8.7

ADD src /opt/aemc
WORKDIR /opt/aemc

RUN sh aemw instance init

ENTRYPOINT ["/bin/bash"]
