# Stage 1: Build the Go application
FROM golang:alpine AS builder

WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -o app .

# Stage 2: Create a lightweight production image
FROM alpine:latest

# Install any necessary dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the built Go application from the previous stage
COPY --from=builder /app/app .
COPY index.html .

# Expose the port on which the application will listen
EXPOSE 8080

# Set the entry point for the container
ENTRYPOINT ["./app"]
