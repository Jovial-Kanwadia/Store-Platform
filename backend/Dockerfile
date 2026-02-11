# Build Stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download
# Copy source code
COPY . .
# Build static binary - CGO_ENABLED=0 is critical for alpine/distroless
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api cmd/api/main.go

# Run Stage (Distroless for security - no shell, no root)
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/api /app/api
# Security Best Practice: Run as non-root
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/api"]