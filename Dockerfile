# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN go build -o main

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
