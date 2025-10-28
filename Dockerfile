# ---- build ----
FROM golang:1.22-bookworm AS build
WORKDIR /src

# deps first (cache)
COPY go.mod ./
RUN go mod download

# source
COPY . .

# install swag and generate docs
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3
RUN swag init -g cmd/api/main.go -o internal/docs --parseInternal --parseDependency

# build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /bin/app ./cmd/api

# ---- run ----
FROM gcr.io/distroless/base-debian12
ENV PORT=8080
USER nonroot
COPY --from=build /bin/app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
