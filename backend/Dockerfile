# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

ENV GOOS=linux
ENV GOARCH=arm64
ENV CGO_ENABLED=1

RUN apk add --no-cache build-base

# Create go cache
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download -x

# Build
COPY main.go .
COPY db/ db/
COPY routes/ routes/
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o splitter main.go

# Final stage
FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/splitter .
EXPOSE 8080
CMD ["./splitter"]
