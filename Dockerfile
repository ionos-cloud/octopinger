FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY manager-linux-amd64 /manager

USER 65532:65532

ENTRYPOINT ["/manager"]