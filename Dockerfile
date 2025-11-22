# ---- Builder Stage ----
# Use the official Go image as a builder image
FROM golang:1.25-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
# -ldflags="-w -s" strips debug information and symbols, reducing the binary size
# CGO_ENABLED=0 disables CGO for a statically linked binary
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /go-trendboard ./cmd/trendboard

# ---- Final Stage ----
# Use a lightweight Alpine image for the final image
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /go-trendboard /app/go-trendboard

# (Optional) Copy default configuration templates that the binary might need
# COPY dashboard.tpl ./

# Set the user to a non-root user for better security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# The entrypoint of the container
ENTRYPOINT ["/app/go-trendboard"]

# Default command can be overridden
CMD ["--help"]
