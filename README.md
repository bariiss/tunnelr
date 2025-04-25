# TunnelR

<p align="center">A lightweight, self-hosted HTTP tunneling service to expose local servers via custom subdomains with secure WebSocket transport</p>

<p align="center">
  <img src="https://img.shields.io/badge/go-%231.24.2-blue" alt="Go version">
  <img src="https://img.shields.io/badge/version-v0.2.1-orange" alt="Version">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
</p>

## Overview

TunnelR creates a secure tunnel to expose your local web servers to the internet through randomly generated or custom subdomains. Perfect for sharing work-in-progress features, testing webhooks, or demonstrating applications without deploying to production.

## Features

- 🚀 **Lightweight & Fast**: Minimal code footprint with efficient implementation
- 🔒 **Secure WebSocket Transport**: Communication secured via TLS
- 🌐 **Custom Subdomains**: Request specific subdomains or get randomly assigned ones
- 🔄 **HTTP(S) Protocol Support**: Tunnels HTTP and HTTPS traffic to your local service
- 🐳 **Docker Ready**: Easy deployment with Docker and Traefik integration
- 🔑 **Automatic SSL Certificates**: Wildcard certificates managed through Let's Encrypt
- 🌍 **Multi-Architecture Support**: ARM64 and AMD64 builds available

## Quick Start

### Running the Server

```bash
# Using Docker Compose (recommended)
docker compose up -d

# Or build and run directly
go build -o tunnelr-server ./cmd/server
./tunnelr-server
```

### Connecting a Client

```bash
# Using the pre-built client
go run ./cmd/client/tunnelr -port 8080 -server wss://link.il1.nl/register

# With a custom subdomain
go run ./cmd/client/tunnelr -port 8080 -server wss://link.il1.nl/register -sub myapp

# Download and use a pre-built binary from the releases page
./tunnelr-darwin-arm64 -port 3000 -server wss://link.il1.nl/register -sub myapp
```

Your local server will be accessible at `https://<subdomain>.link.il1.nl` where `<subdomain>` is either your chosen subdomain or a randomly assigned one.

## How It Works

1. **Server Registration**: Client establishes a WebSocket connection to the tunnel server
2. **Subdomain Assignment**: Server assigns a subdomain (random or requested)
3. **Tunnel Establishment**: Persistent WebSocket connection serves as the tunnel
4. **Request Forwarding**: Incoming HTTP requests to the subdomain are sent through the WebSocket to the client
5. **Local Processing**: Client processes the request locally and returns the response
6. **Response Delivery**: Server delivers the response to the original requester

## Detailed Technical Architecture

### Protocol

TunnelR uses a simple JSON-based message protocol for communication between the server and client:

- **RequestFrame**: Sent from server to client when a HTTP request arrives
  ```json
  {
    "id": "unique-request-id",
    "method": "GET",
    "url": "/path",
    "header": { "User-Agent": ["Browser"] },
    "body": [base64-encoded-bytes]
  }
  ```

- **ResponseFrame**: Sent from client to server after processing the request
  ```json
  {
    "id": "unique-request-id",
    "status_code": 200,
    "header": { "Content-Type": ["application/json"] },
    "body": [base64-encoded-bytes],
    "error": "optional error message"
  }
  ```

### Security Considerations

- All traffic is encrypted using TLS (HTTPS)
- Each tunnel gets a unique subdomain
- WebSocket connection is authenticated at registration
- Random subdomain generation uses cryptographically secure random numbers

## Configuration

### Server Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `BASE_DOMAIN` | Base domain for tunnel URLs | `link.il1.nl` |
| `SERVER_PORT` | Port to listen on | `8095` |

### Client Arguments

| Flag | Description | Default |
|------|-------------|---------|
| `-port` | Local port to forward traffic to | `8080` |
| `-server` | Tunnel server WebSocket URL | `wss://link.il1.nl/register` |
| `-sub` | Custom subdomain (optional) | Random string |

## Docker Deployment

TunnelR is designed to work with Traefik for easy deployment with automatic HTTPS. The included `docker-compose.yml` provides a complete setup:

```yaml
services:
  traefik:
    image: traefik:v2.11
    # SSL/TLS termination and routing configuration
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./letsencrypt:/letsencrypt
    environment:
      CF_DNS_API_TOKEN: ${CF_DNS_API_TOKEN}
  
  tunnelr-server:
    image: ghcr.io/bariiss/tunnelr-server:latest
    environment:
      - BASE_DOMAIN=link.il1.nl
      - SERVER_PORT=8095
    # Traefik labels for routing
```

To use the provided setup:

1. Create a network for Traefik: `docker network create traefik_proxy`
2. Create a `.env` file with your Cloudflare credentials:
   ```
   CF_DNS_API_TOKEN=your_cloudflare_api_token
   DOMAIN=your.domain.com
   EMAIL=your@email.com
   SERVER_PORT=8095
   ```
3. Run `docker compose up -d`

## DNS Configuration

For TunnelR to work properly, you'll need to configure your DNS settings with Cloudflare:

1. Add an A record for `link.il1.nl` pointing to your server's IP address
2. Add a wildcard A record for `*.link.il1.nl` also pointing to the same server IP address

> **Important**: Your domain must be managed through Cloudflare to use this setup. The DNS-01 challenge method used for obtaining wildcard SSL certificates requires DNS provider API access, which this configuration implements using Cloudflare. Without Cloudflare DNS management, you won't be able to automatically obtain wildcard certificates with Let's Encrypt.

### Setting Up Cloudflare API Token

For automated SSL certificate management, you'll need a Cloudflare API token:

1. Log in to your Cloudflare account
2. Navigate to "My Profile" > "API Tokens"
3. Click "Create Token"
4. Either:
   - Use the "Edit zone DNS" template and select your specific zone, or
   - Create a custom token with the following permissions:
     - Zone > DNS > Edit
     - Zone > Zone > Read
5. Restrict the token to the specific zone (domain) you're using
6. Add the token to your `.env` file

## Customizing for Your Own Domain

The code and Docker Compose configuration are designed to be easily adaptable for your own domain:

1. Replace all instances of `link.il1.nl` with your own domain in the configuration files
2. Update your DNS provider with appropriate A records for your domain and wildcard subdomains
3. Update the Cloudflare API token and email in the `.env` file

## Available Client Binaries

Pre-built binaries are available for various platforms:

- **macOS**: `tunnelr-darwin-amd64`, `tunnelr-darwin-arm64` (Apple Silicon)
- **Linux**: Various architectures including ARM, MIPS, and RISC-V
- **Windows**: `tunnelr-windows-amd64.exe`, `tunnelr-windows-arm64.exe`

Download the latest release from the [GitHub Releases page](https://github.com/bariiss/tunnelr/releases).

## Building from Source

```bash
# Build the server
go build -o tunnelr-server ./cmd/server

# Build the client
go build -o tunnelr ./cmd/client/tunnelr

# Cross-compile for different platforms
GOOS=darwin GOARCH=arm64 go build -o tunnelr-darwin-arm64 ./cmd/client/tunnelr
GOOS=windows GOARCH=amd64 go build -o tunnelr-windows-amd64.exe ./cmd/client/tunnelr
```

## Use Cases

- **Webhook Testing**: Receive webhooks from third-party services on your local machine
- **Demo Environments**: Quickly share your development work with clients or team members
- **Mobile App Development**: Test mobile apps against a local backend
- **Remote Collaboration**: Enable team members to access your local development environment
- **Microservice Development**: Expose local microservices for testing with external services

## Limitations

- Not designed for high-traffic production use
- Single WebSocket connection per subdomain can be a bottleneck
- No built-in authentication for accessing exposed services

## License

MIT License

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

- **Bug Reports**: Include detailed steps to reproduce
- **Feature Requests**: Explain the use case and expected behavior
- **Pull Requests**: Reference issues and include tests where possible