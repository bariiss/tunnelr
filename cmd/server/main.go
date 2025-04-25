package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bariiss/tunnelr/internal/server"
	"github.com/joho/godotenv" // + ekle
)

func main() {
	_ = godotenv.Load()

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		log.Fatal("DOMAIN environment variable is required")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		log.Fatal("SERVER_PORT environment variable is required")
	}
	addr := ":" + port

	reg := server.NewRegistry()
	mux := http.NewServeMux()
	mux.Handle("/register", server.RegisterHandler(reg, domain))
	mux.Handle("/", server.ProxyHandler(reg))

	log.Printf("tunnelr-server listening on %s (domain=%s)", addr, domain)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
