FROM scratch
ENTRYPOINT ["/aem"]
COPY aem /
WORKDIR /project
