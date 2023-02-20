FROM scratch

LABEL org.opencontainers.image.description="AEM Compose CLI - Universal tool to manage AEM instances everywhere!"
LABEL org.opencontainers.image.source="https://github.com/wttech/aemc"
LABEL org.opencontainers.image.vendor="Wunderman Thompson Technology"
LABEL org.opencontainers.image.authors="krystian.panek@wundermanthompson.com"
LABEL org.opencontainers.image.licenses="Apache-2.0"

ENTRYPOINT ["/aem"]
COPY aem /
WORKDIR /project
