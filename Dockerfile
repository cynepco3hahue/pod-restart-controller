FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ARG BIN_DIR=

ENV LANG=en_US.utf8

COPY ${BIN_DIR}/pod-restarter /usr/local/bin/pod-restarter
USER 1001

ENTRYPOINT [ "/usr/local/bin/pod-restarter" ]
