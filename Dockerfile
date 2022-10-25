FROM scionproto/docker-caps as caps

FROM gcr.io/distroless/static:nonroot
ARG BINARY

WORKDIR /
COPY --from=caps /bin/setcap /bin
COPY ${BINARY} /main

RUN setcap cap_net_raw+ep /main && rm /bin/setcap

USER 65532:65532

ENTRYPOINT ["/main"]