FROM gcr.io/distroless/static:nonroot
ARG BINARY

WORKDIR /
COPY ${BINARY} /main

ENTRYPOINT ["/main"]