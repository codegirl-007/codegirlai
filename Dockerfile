# Use the official Golang image as a base image
FROM golang:1.22.2

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the application code
COPY main.go ./

# Copy static files
COPY static ./static

# Expose the application port
EXPOSE 80

# Build the Go application
RUN go build -o codegirlai main.go

# Command to run the application
CMD ["./codegirlai"]