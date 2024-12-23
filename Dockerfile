# Dockerfile
FROM golang:1.23-alpine

WORKDIR /app

# Install git and build essentials
RUN apk add --no-cache git build-base

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (specify the main package path)
RUN go build -o /app/bin/scheduler .

# Command to run the binary
CMD ["/app/bin/scheduler"]