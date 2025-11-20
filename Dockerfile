FROM golang:1.25 AS builder

WORKDIR /app

# Copy go.mod and download dependencies
COPY go.mod ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application - output binary named 'main'
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w" \           
    -trimpath \                
    -o event-messenger .

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Install ca-certificates for HTTPS and SQLite runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*


# Copy binary from builder
COPY --from=builder /app/event-messenger .

# Copy templates and static files
COPY --from=builder /app/templates ./templates

# Create data directory for uploads and database
RUN mkdir -p /app/data/uploads

CMD ["./event-messenger"]

