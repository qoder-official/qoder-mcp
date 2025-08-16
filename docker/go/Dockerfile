# --- Build Stage ---
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the entire server directory content.
# This ensures that go.mod, go.sum, and all source code are available.
COPY . .

# Build the server binary
# CGO_ENABLED=0 is important for a minimal, static binary
# -o /server places the output binary in the /server file
RUN CGO_ENABLED=0 go build -o /server ./cmd/github-mcp-server

# --- Final Stage ---
# Use a minimal image for the final container
FROM alpine:latest

# Set up a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy the server binary from the builder stage
COPY --from=builder /server /server

# Switch to the non-root user
USER appuser

# The command will be passed from the 'docker run' command,
# for example 'stdio'
CMD []
