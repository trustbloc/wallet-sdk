#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

ARG GO_VER
ARG ALPINE_VER

FROM golang:${GO_VER}-alpine${ALPINE_VER} as builder

RUN apk update && apk add git && apk add ca-certificates
RUN adduser -D -g '' appuser

COPY . $GOPATH/src/github.com/trustbloc/wallet-sdk/test/integration/webhook/
WORKDIR $GOPATH/src/github.com/trustbloc/wallet-sdk/test/integration/webhook/

RUN CGO_ENABLED=0 go build -o /usr/bin/webhook-server

FROM scratch

LABEL org.opencontainers.image.source https://github.com/trustbloc/wallet-sdk/test/integration/webhook

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /usr/bin/webhook-server /usr/bin/webhook-server
USER appuser

ENTRYPOINT ["/usr/bin/webhook-server"]
