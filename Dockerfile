FROM alpine:latest as builder

RUN apk --update add ca-certificates

FROM scratch

ARG BINARY

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ${BINARY} /main

ENTRYPOINT ["/main"]