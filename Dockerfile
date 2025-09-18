# Build stage
FROM golang:alpine AS builder

# Install dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy source
COPY . .

# Build CLI
WORKDIR /app/cmd/pixelart
RUN go build -o /pixelart

# Final runtime stage
FROM alpine:3.20

# Install runtime deps (needed for image formats like PNG/JPEG)
RUN apk add --no-cache ca-certificates libjpeg-turbo-utils libpng

WORKDIR /app

# Copy binary
COPY --from=builder /pixelart /usr/local/bin/pixelart

# Default command
ENTRYPOINT ["pixelart"]
