# TunnelR

<p align="center">A lightweight, self-hosted HTTP tunneling service to expose local servers via custom subdomains with secure WebSocket transport</p>

<p align="center">
  <img src="https://img.shields.io/badge/go-%231.24.2-blue" alt="Go version">
  <img src="https://img.shields.io/badge/version-v0.3.5-orange" alt="Version">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
</p>

## Overview

TunnelR creates a secure tunnel to expose your local web servers to the internet through randomly generated or custom subdomains. Perfect for sharing work-in-progress features, testing webhooks, or demonstrating applications without deploying to production.

Built with Go, this minimal solution provides secure, efficient tunneling with Docker integration and Let's Encrypt certificate automation through Cloudflare DNS. The client preserves your configuration choices and makes connecting to your tunnel as simple as a single command.

With support for subdomain customization, secure WebSocket transport, and Docker secrets for credential management, TunnelR offers a complete solution for exposing local services with minimal setup overhead.

## Features

- 🚀 **Lightweight & Fast**: Minimal code footprint with efficient implementation
- 🔒 **Secure WebSocket Transport**: Communication secured via TLS
- 🌐 **Custom Subdomains**: Request specific subdomains or get randomly assigned ones
- 🔄 **HTTP(S) Protocol Support**: Tunnels HTTP and HTTPS traffic to your local service
- 🐳 **Docker Ready**: Easy deployment with Docker and Traefik integration
- 🔑 **Automatic SSL Certificates**: Wildcard certificates managed through Let's Encrypt
- 🌍 **Multi-Architecture Support**: ARM64 and AMD64 builds available
- 💾 **Configuration Persistence**: Save your settings in a local config file

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
# Using the pre-built client with default settings
tunnelr

# With a custom subdomain (as positional argument)
tunnelr myapp

# Specify a different local port
tunnelr -p 3000 myapp

# Forward to a different local address
tunnelr -t 192.168.1.100 -p 3000 myapp

# Specify a different tunnel server domain
tunnelr -d <DOMAIN> myapp
```

Your local server will be accessible at `https://<subdomain>.<DOMAIN>` where `<subdomain>` is either your chosen subdomain or a randomly assigned one.

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
| `DOMAIN` | Base domain for tunnel URLs | Your configured domain |
| `SERVER_PORT` | Port to listen on | `8095` |
| `CF_EMAIL` | Cloudflare account email | Required for SSL |

### Client Configuration

TunnelR client configuration is saved to `~/.config/tunnelr/config.yaml` and can be overridden with command-line flags.

#### Command-Line Arguments

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--config` | - | Config file path | `~/.config/tunnelr/config.yaml` |
| `--domain` | `-d` | Tunnel server domain | `<DOMAIN>` (or saved value) |
| `--port` | `-p` | Local port to forward traffic to | `8080` |
| `--target` | `-t` | Local host to forward to | `127.0.0.1` |

#### Positional Argument

The first positional argument is treated as the requested subdomain (optional). If not provided, a random subdomain will be assigned.

```bash
tunnelr [flags] [subdomain]
```

#### Configuration File

The client saves your domain preference to the config file. Example `config.yaml`:

```yaml
domain: <DOMAIN>
```

## Docker Deployment

TunnelR is designed to work with Traefik for easy deployment with automatic HTTPS. The included `docker-compose.yml` provides a complete setup with Docker secrets for secure credential management.

### Setup Steps

1. Create a network for Traefik (if not already created):
   ```bash
   docker network create traefik_proxy
   ```

2. Configure your environment variables and Cloudflare token:
   ```bash
   # Update .env file with your domain and Cloudflare email
   cat > .env << EOL
   DOMAIN=your.domain.com
   CF_EMAIL=your-cloudflare-email@example.com
   SERVER_PORT=8095
   EOL

   # Update your Cloudflare API token
   echo "your-cloudflare-api-token" > secrets/cf_token.txt
   ```

3. Launch the services:
   ```bash
   docker compose up -d
   ```

### Docker Compose Configuration

The `docker-compose.yml` file contains:

```yaml
services:
  traefik:
    image: traefik:v2.11
    container_name: traefik
    restart: always
    ports:
      - 80:80
      - 443:443
    env_file: .env
    environment:
      - CF_EMAIL
      - DOMAIN
    secrets:
      - cf_token
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./letsencrypt:/letsencrypt
    networks:
      - traefik_proxy
    entrypoint: ["/bin/sh", "-c",
      "export CF_DNS_API_TOKEN=$(cat /run/secrets/cf_token) && \
       traefik $$@",
      "--" ]
    command:
      # Traefik configuration for SSL/TLS
      - --api.dashboard=true
      - --providers.docker=true
      - --entrypoints.websecure.http.tls=true
      - --entrypoints.websecure.http.tls.certresolver=cloudflare
      - --entrypoints.websecure.http.tls.domains[0].main=${DOMAIN}
      - --entrypoints.websecure.http.tls.domains[0].sans=*.${DOMAIN}
      - --certificatesresolvers.cloudflare.acme.dnschallenge=true
      - --certificatesresolvers.cloudflare.acme.dnschallenge.provider=cloudflare
      # Additional configuration omitted for brevity
  
  tunnelr-server:
    image: ghcr.io/bariiss/tunnelr-server:latest
    container_name: tunnelr-server
    restart: always
    depends_on: [traefik]
    expose: [ "${SERVER_PORT}" ]
    env_file: .env
    environment:
      - SERVER_PORT
      - DOMAIN
    networks: [ traefik_proxy ]
    labels:
      # Traefik routing configuration
      traefik.enable: "true"
      traefik.http.services.tunnelr.loadbalancer.server.port: ${SERVER_PORT}
      traefik.http.routers.tunnelr.rule: Host(`${DOMAIN}`)
      traefik.http.routers.tunnelr-sub.rule: HostRegexp(`{subdomain:[a-z0-9]+}.${DOMAIN}`)
      # Additional labels omitted for brevity

networks:
  traefik_proxy:
    name: traefik_proxy

secrets:
  cf_token:
    file: ./secrets/cf_token.txt
```

## DNS Configuration and Cloudflare Setup

### DNS Records

For TunnelR to work properly, you'll need to configure your DNS settings with Cloudflare:

1. Add an A record for `<DOMAIN>` pointing to your server's IP address
2. Add a wildcard A record for `*.<DOMAIN>` also pointing to the same server IP address

> **Important**: Your domain must be managed through Cloudflare to use this setup. The DNS-01 challenge method used for obtaining wildcard certificates requires DNS provider API access.

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
6. Copy the token and save it to `secrets/cf_token.txt`

### Security Note

The Cloudflare API token is stored as a Docker secret, which is more secure than environment variables because:
- It's mounted as a file in the container instead of being part of the environment
- It's not exposed in Docker inspect commands
- It's not logged in container logs

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