package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/bariiss/tunnelr/internal/client"
	"github.com/coder/websocket"
)

// main is the entry point for the tunnelr client, which establishes a WebSocket connection to the server and forwards local traffic
func main() {
	port := flag.Int("port", 8080, "local port to expose")
	serverURL := flag.String("server", "wss://link.il1.nl/register", "tunnel server URL")
	sub := flag.String("sub", "", "desired sub‑domain (optional)")
	flag.Parse()

	// build final URL
	u, _ := url.Parse(*serverURL)
	q := u.Query()
	if *sub != "" {
		q.Set("sub", *sub)
	}
	u.RawQuery = q.Encode()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Use explicit options with compression disabled to prevent RSV bit issues
	dialOptions := &websocket.DialOptions{
		CompressionMode: websocket.CompressionDisabled,
	}

	conn, _, err := websocket.Dial(ctx, u.String(), dialOptions)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}

	// first message from server is the allocated host
	_, hostBytes, err := conn.Read(ctx)
	if err != nil {
		log.Fatalf("handshake read: %v", err)
	}
	publicURL := "https://" + string(hostBytes)
	log.Printf("🆕 public URL → %s", publicURL)

	fwd := client.NewForwarder(conn, fmt.Sprintf("127.0.0.1:%d", *port))
	log.Printf("✅ connected — forwarding http://127.0.0.1:%d", *port)

	if err := fwd.Serve(ctx); err != nil {
		log.Println("forwarder stopped:", err)
	}
}
