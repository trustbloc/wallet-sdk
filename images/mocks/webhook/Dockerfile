#
# Copyright Avast Software. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

ARG GO_VER
ARG ALPINE_VER

FROM golang:${GO_VER}-alpine${ALPINE_VER} AS golang
RUN apk add --no-cache \
    bash \
    gcc \
    git \
    libtool \
    make \
    musl-dev;
COPY . $GOPATH/src/github.com/trustbloc/wallet-sdk/
WORKDIR $GOPATH/src/github.com/trustbloc/wallet-sdk
ENV EXECUTABLES go git

FROM golang AS wallet-sdk
ARG GO_TAGS
ARG GO_PROXY
RUN GO_TAGS=${GO_TAGS} GOPROXY=${GO_PROXY} make sample-webhook

FROM alpine:${ALPINE_VER} AS base
LABEL org.opencontainers.image.source=https://github.com/trustbloc/wallet-sdk
COPY --from=wallet-sdk /go/src/github.com/trustbloc/wallet-sdk/build/bin/webhook-server /usr/local/bin

ENTRYPOINT ["webhook-server"]
