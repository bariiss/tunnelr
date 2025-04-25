package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/coder/websocket"
)

// RegisterHandler handles WebSocket connections for registering subdomains and establishing tunnels
func RegisterHandler(reg *Registry, domain string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
			CompressionMode:    websocket.CompressionDisabled,
		})
		if err != nil {
			log.Println("ws upgrade:", err)
			return
		}

		sub := r.URL.Query().Get("sub")
		switch {
		case sub == "":
			sub = reg.uniqueSub(6) // ex: "a3f9kq"
		case reg.Has(sub):
			_ = conn.Close(websocket.StatusPolicyViolation, "subdomain already in use")
			return
		}

		reg.Put(sub, &ConnEntry{Conn: conn})
		fullHost := fmt.Sprintf("%s.%s", sub, domain)
		log.Printf("✅ tunnel registered: https://%s ↔︎ client", fullHost)

		if err := conn.Write(r.Context(), websocket.MessageText, []byte(fullHost)); err != nil {
			log.Println("handshake write:", err)
			reg.Delete(sub)
			conn.Close(websocket.StatusInternalError, "handshake failed")
			return
		}

		<-r.Context().Done()
		reg.Delete(sub)
		conn.Close(websocket.StatusNormalClosure, "client disconnected")
	}
}
