# TunnelR

<p align="center">A lightweight, self-hosted HTTP tunneling service written in Go</p>

<p align="center">
  <img src="https://img.shields.io/badge/go-%231.24.2-blue" alt="Go version">
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
```

Your local server will be accessible at `https://<subdomain>.link.il1.nl` where `<subdomain>` is either your chosen subdomain or a randomly assigned one.

## How It Works

1. **Server Registration**: Client establishes a WebSocket connection to the tunnel server
2. **Subdomain Assignment**: Server assigns a subdomain (random or requested)
3. **Tunnel Establishment**: Persistent WebSocket connection serves as the tunnel
4. **Request Forwarding**: Incoming HTTP requests to the subdomain are sent through the WebSocket to the client
5. **Local Processing**: Client processes the request locally and returns the response
6. **Response Delivery**: Server delivers the response to the original requester

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

## Architecture

TunnelR uses a client-server architecture:

- **Server Component**: Handles subdomain registration, connection management, and request proxying
- **Client Component**: Maintains the WebSocket connection and forwards requests to the local service

## Docker Deployment

TunnelR is designed to work with Traefik for easy deployment with automatic HTTPS. The included `docker-compose.yml` provides a complete setup:

```yaml
services:
  traefik:
    # Traefik configuration for SSL/TLS termination
    # ...
  
  tunnelr:
    image: ghcr.io/bariiss/tunnelr:latest
    environment:
      - BASE_DOMAIN=link.il1.nl
    # Labels for Traefik integration
```

To use the provided setup:

1. Create a network for Traefik: `docker network create traefik_proxy`
2. Update the Cloudflare API token and email in `docker-compose.yml`
3. Run `docker compose up -d`

## DNS Configuration

For TunnelR to work properly, you'll need to configure your DNS settings with Cloudflare:

1. Add an A record for `link.il1.nl` pointing to your server's IP address
2. Add a wildcard A record for `*.link.il1.nl` also pointing to the same server IP address

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
6. After creating the token, copy it and replace `<YOUR_CLOUDFLARE_API_TOKEN>` in your `docker-compose.yml`
7. Also replace `<YOUR_EMAIL>` with the email associated with your Cloudflare account

This token allows Traefik to automatically verify domain ownership by creating temporary DNS records when requesting Let's Encrypt certificates.

## Customizing for Your Own Domain

The code and Docker Compose configuration are designed to be easily adaptable for your own domain:

1. Replace all instances of `link.il1.nl` with your own domain in the `docker-compose.yml` file
2. Update your DNS provider with appropriate A records for your domain and wildcard subdomains
3. Update the Cloudflare API token and email in the configuration

The included Traefik configuration automatically handles SSL certificate issuance and renewal through Let's Encrypt using Cloudflare DNS challenge for verification.

## License

MIT License

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.