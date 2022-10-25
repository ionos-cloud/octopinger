FROM gcr.io/distroless/static
ARG BINARY

WORKDIR /
COPY ${BINARY} /main

ENTRYPOINT ["/main"]