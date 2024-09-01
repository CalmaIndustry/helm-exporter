# Use the official Golang image as a build environment
FROM golang:1.23-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o /helm-metrics-exporter

# Use a smaller base image for the final container
FROM alpine:3.14

# Install Helm in the final container
RUN apk add --no-cache \
    curl \
    bash \
    openssl \
    && curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Copy the compiled Go binary from the builder
COPY --from=builder /helm-metrics-exporter /helm-metrics-exporter

# Expose the port the app will run on
EXPOSE 2112

# Run the application
ENTRYPOINT ["/helm-metrics-exporter"]
