# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-bookworm AS build
WORKDIR /src

ENV CGO_ENABLED=0 GOTOOLCHAIN=auto

# deps
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# build
COPY . .
# build ONLY your main package (adjust path if needed)
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -ldflags="-s -w" -o /out/app ./cmd/api

# stage with CA certificates
FROM debian:bookworm-slim AS certs
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && update-ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# minimal runtime + CA roots
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
# copy CA bundle so TLS works (MongoDB Atlas requires TLS)
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/app /app/app
ENTRYPOINT ["/app/app"]
