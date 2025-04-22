# TunnelR – tiny Go self‑hosted tunnel

## Quick start (Dev)

```bash
# server
docker compose up -d

# client
go run ./cmd/client -port 8080 -server wss://tunnel.example.com/register
```

Open https://<random>.link.il1.nl in your browser and you’ll be proxied to localhost:8080 🎉
