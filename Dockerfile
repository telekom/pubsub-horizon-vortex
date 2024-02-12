# Copyright 2024 Deutsche Telekom IT GmbH
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.21-alpine AS build
ARG GOPROXY
ARG GONOSUMDB
ENV GOPROXY=$GOPROXY
ENV GONOSUMDB=$GONOSUMDB

USER root

WORKDIR /build
COPY . .
RUN apk add --no-cache build-base
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -extldflags=-static" -o vortex

FROM scratch
COPY --from=build /build/vortex vortex
ENTRYPOINT ["./vortex", "serve"]
