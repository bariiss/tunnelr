package server

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"

	"github.com/coder/websocket"
)

// randomString creates 6 random bytes hex‑encoded (12 chars)
func randomString() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

func RegisterHandler(reg *Registry, baseDomain string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade to websocket
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // demo only; use proper TLS auth
		})
		if err != nil {
			log.Println("ws upgrade:", err)
			return
		}
		sub := r.URL.Query().Get("sub")
		if sub == "" {
			sub = randomString()
		}
		reg.Put(sub, &ConnEntry{Conn: c})
		fullURL := "https://" + sub + "." + baseDomain
		log.Printf("client registered subdomain %s -> %s", sub, fullURL)

		// keep connection open until closed
		ctx := r.Context()
		<-ctx.Done()
		reg.Delete(sub)
		c.Close(websocket.StatusNormalClosure, "server closing")
	}
}
