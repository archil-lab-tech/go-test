# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-bookworm AS build
WORKDIR /src

# Allow the go tool to auto-install matching toolchains if needed
ENV CGO_ENABLED=0 GOTOOLCHAIN=auto

# deps first for better caching
COPY go.mod go.sum ./
RUN go mod download

# app source
COPY . .

# build (adjust ./cmd/server if your main.go is elsewhere)
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -ldflags="-s -w" -o /out/app ./...

# small runtime image
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /out/app /app/app
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/app"]
