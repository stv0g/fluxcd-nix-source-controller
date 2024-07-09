# SPDX-FileCopyrightText: 2020, 2021 The Flux authors
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

ARG GO_VERSION=1.22
ARG XX_VERSION=1.4.0

FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine as builder

# Copy the build utilities.
COPY --from=xx / /

ARG TARGETPLATFORM

WORKDIR /workspace

# copy modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache modules
RUN go mod download

# copy source code
COPY main.go main.go
COPY controllers/ controllers/

# build
ENV CGO_ENABLED=0
RUN xx-go build -a -o fluxcd-nix-source-controller main.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tini

COPY --from=builder /workspace/fluxcd-nix-source-controller /usr/local/bin/

RUN addgroup -S controller && adduser -S controller -G controller

USER controller

ENTRYPOINT [ "/sbin/tini", "--", "fluxcd-nix-source-controller" ]
