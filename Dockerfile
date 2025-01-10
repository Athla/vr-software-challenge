# Build stage
FROM golang:1.23-alpine AS build

# Install build dependencies
RUN apk add --no-cache gcc musl-dev git pkgconfig build-base bash zlib-dev zstd-dev libc-dev librdkafka-dev

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 \
    CGO_LDFLAGS="-L/usr/lib -lrdkafka" \
    CGO_CFLAGS="-I/usr/include/librdkafka" \
    go build -o main cmd/api/main.go

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache librdkafka-dev ca-certificates

WORKDIR /app
COPY --from=build /app/main .
COPY docs/ docs/

EXPOSE 8080
CMD ["./main"]
