# Stage 1: Build
FROM golang:1.20 AS builder

# Set the working directory inside the builder container
WORKDIR /codegirlai

# Copy Go module manifests and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . ./

# Build the Go application
RUN go build -o codegirlai main.go

# Stage 2: Runtime
FROM debian:bullseye-slim

# Set the working directory inside the runtime container
WORKDIR /codegirlai

# Copy the compiled binary from the builder stage
COPY --from=builder /codegirlai/codegirlai /codegirlai/codegirlai

# Copy the static directory
COPY --from=builder /codegirlai/static /codegirlai/static

# Expose port 80 for the application
EXPOSE 80

# Command to run the application
CMD ["/codegirlai/codegirlai", "-port", "80"]
