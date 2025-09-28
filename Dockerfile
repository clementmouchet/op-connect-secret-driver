# Use a smaller base image for a smaller final image
FROM golang:1.25-alpine AS builder

ARG VERSION="dev"

WORKDIR /app

RUN mkdir -p /run/docker/plugins

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY main.go .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -ldflags="-w -X main.version=${VERSION}" \
    -a \
    -installsuffix cgo \
    -o /op-connect-secret-driver .

# Use a minimal image for the final stage
FROM gcr.io/distroless/static-debian12 AS runner

# Copy the binary from the builder stage
COPY --from=builder /op-connect-secret-driver /op-connect-secret-driver
COPY --from=builder /run/docker/plugins /run/docker/plugins

# Command to run the driver
CMD ["/op-connect-secret-driver"]