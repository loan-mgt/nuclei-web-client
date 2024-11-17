# Stage 1: Build the application
FROM golang:1.23-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Install necessary dependencies (git and other packages)
RUN apk add --no-cache git

# Copy the go.mod and go.sum files first for caching dependencies
COPY go.mod go.sum ./

# Download the Go dependencies
RUN go mod download

# Copy the entire project into the container (make sure Dockerfile is in the root)
COPY . .

# Set the working directory to the folder containing the main.go file (cmd/server)
WORKDIR /app/cmd/server

# Build the Go application
RUN go build -o main .

# Stage 2: Run the application
FROM alpine:latest

# Install necessary dependencies for running the app
RUN apk add --no-cache ca-certificates

# Set the working directory for the runtime container
WORKDIR /app

# Copy the compiled Go binary from the builder image
COPY --from=builder /app/cmd/server/main .

# Copy the public folder into the runtime container
COPY ./public /app/public

# Expose the port the application will run on
EXPOSE 8080

# Command to run the application
CMD ["./main"]
