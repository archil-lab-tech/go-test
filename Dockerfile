# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-bookworm AS build
WORKDIR /src

# Static binary, let Go auto-fetch matching toolchain if needed
ENV CGO_ENABLED=0 GOTOOLCHAIN=auto

# ---- deps (cacheable) ----
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# ---- source ----
COPY . .

# Build ONLY the main package (your entrypoint is cmd/api)
ARG MAIN_PATH=./cmd/api
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -trimpath -ldflags="-s -w" -o /out/app ${MAIN_PATH}

# ---- minimal runtime ----
# Use base-debian12 (not static) so TLS roots are present for Mongo Atlas etc.
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=build /out/app /app/app
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/app"]
