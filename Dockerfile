# Use the official Golang image for building the application
FROM golang:1.20 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the application code
COPY . ./

# Build the Go application
RUN go build -o codegirlai main.go

# Use a minimal base image for running the application
FROM debian:bullseye-slim

# Set the working directory for the runtime
WORKDIR /app

# Copy the compiled Go binary from the builder
COPY --from=builder /app/codegirlai /app/codegirlai

# Copy static files
COPY static ./static

# Expose the port the application listens on
EXPOSE 9000

# Command to run the application
CMD ["/app/codegirlai"]