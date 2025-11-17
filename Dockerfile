FROM golang:1.25

WORKDIR /app

# Copy go.mod and download dependencies
COPY go.mod ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application - output binary named 'main'
RUN go build -v -o /app/main .

RUN cd /app

# Run the binary
CMD ["./main"]