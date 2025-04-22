package server

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"

	"nhooyr.io/websocket"
)

// 26 letters + 10 digits
const alphanum = "abcdefghijklmnopqrstuvwxyz0123456789"

// randomString returns an n‑char ID using [a‑z0‑9].
func randomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = alphanum[int(b[i])%len(alphanum)]
	}
	return string(b)
}

// RegisterHandler upgrades to WebSocket, assigns/records a sub‑domain,
// sends the chosen host back to the client, and keeps the tunnel alive.
func RegisterHandler(reg *Registry, baseDomain string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // remove once you add real TLS auth
		})
		if err != nil {
			log.Println("ws upgrade:", err)
			return
		}

		sub := r.URL.Query().Get("sub")
		if sub == "" {
			sub = randomString(6) // e.g. “f3a9xk”
		}
		reg.Put(sub, &ConnEntry{Conn: conn})

		fullHost := fmt.Sprintf("%s.%s", sub, baseDomain)
		log.Printf("✅ tunnel registered: https://%s ↔︎ client", fullHost)

		// Tell the client which host was allocated.
		if err := conn.Write(r.Context(), websocket.MessageText, []byte(fullHost)); err != nil {
			log.Println("handshake write:", err)
			reg.Delete(sub)
			conn.Close(websocket.StatusInternalError, "handshake failed")
			return
		}

		// Keep connection alive until client disconnects.
		<-r.Context().Done()
		reg.Delete(sub)
		conn.Close(websocket.StatusNormalClosure, "client disconnected")
	}
}
