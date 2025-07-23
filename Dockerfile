FROM golang:1.24-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy your entire project
COPY . .

# Build the application from cmd/mcp directory
WORKDIR /app/cmd/mcp
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mcp-server .

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/cmd/mcp/mcp-server .

# Make it executable
RUN chmod +x ./mcp-server

# Use Render's PORT environment variable
EXPOSE $PORT

# Set default environment variables for MCP
ENV MCP_SERVER_TYPE=sse
ENV MCP_SSE_PORT=${PORT:-8080}

# Run the server
CMD ["./mcp-server"]