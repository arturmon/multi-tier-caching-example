# Use the official Golang image
FROM golang:1.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go source code
COPY . .

# Build the Go application
RUN go mod tidy && go get -u && go build -o server && ls -l /app

# Use a lightweight image for the final container
FROM debian:bookworm-slim

# Add user
RUN adduser --system --no-create-home --group serveruser

# Create working directory
RUN mkdir -p /server && chown serveruser:serveruser /server

WORKDIR /server

# Copy the compiled Go binary
COPY --from=builder /app/server /server/server

# Ensure the binary is executable
RUN chmod +x /server/server

# Switch to non-root user
USER serveruser

# Run the application
ENTRYPOINT ["/server/server"]