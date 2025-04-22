package server

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

const alphanum = "abcdefghijklmnopqrstuvwxyz0123456789"

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

// RegisterHandler with nhooyr‑style keep‑alive
func RegisterHandler(reg *Registry, baseDomain string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Println("ws upgrade:", err)
			return
		}

		sub := r.URL.Query().Get("sub")
		if sub == "" {
			sub = randomString(6)
		}
		reg.Put(sub, &ConnEntry{Conn: conn})

		// -------- keep‑alive (ping) -----------
		go func(s string) {
			tick := time.NewTicker(25 * time.Second)
			defer tick.Stop()

			for {
				select {
				case <-r.Context().Done():
					return
				case <-tick.C:
					// Ping ve 5 sn timeout
					ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
					err := conn.Ping(ctx)
					cancel()
					if err != nil {
						log.Printf("ping failed (%s): %v", s, err)
						reg.Delete(s)
						conn.Close(websocket.StatusNormalClosure, "ping timeout")
						return
					}
				}
			}
		}(sub)
		// --------------------------------------

		fullHost := fmt.Sprintf("%s.%s", sub, baseDomain)
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
