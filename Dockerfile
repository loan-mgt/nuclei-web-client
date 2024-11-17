# Base image for building
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o main .

# Runtime image with updated glibc
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y docker.io && apt-get clean

WORKDIR /app

COPY --from=builder /app/main .
COPY ./public ./public

CMD ["./main"]
