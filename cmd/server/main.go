package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bariiss/tunnelr/internal/server"
)

// main is the entry point for the tunnelr server, which handles WebSocket connections and proxies HTTP requests
func main() {
	baseDomain := os.Getenv("BASE_DOMAIN")
	if baseDomain == "" {
		baseDomain = "link.il1.nl"
	}
	reg := server.NewRegistry()
	mux := http.NewServeMux()
	mux.Handle("/register", server.RegisterHandler(reg, baseDomain))
	mux.Handle("/", server.ProxyHandler(reg))

	log.Println("tunnelr-server listening on :8095")
	if err := http.ListenAndServe(":8095", mux); err != nil {
		log.Fatal(err)
	}
}
