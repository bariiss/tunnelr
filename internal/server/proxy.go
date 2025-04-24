package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bariiss/tunnelr/internal/common"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

// ProxyHandler creates an HTTP handler that proxies incoming requests through the WebSocket tunnel to the client
func ProxyHandler(reg *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		parts := strings.Split(host, ".")
		sub := parts[0]
		entry, ok := reg.Get(sub)
		if !ok || entry.Conn == nil {
			http.NotFound(w, r)
			return
		}

		// read body (could be empty)
		body, _ := io.ReadAll(r.Body)
		reqFrame := common.RequestFrame{
			ID:     uuid.NewString(),
			Method: r.Method,
			URL:    r.URL.String(),
			Header: r.Header,
			Body:   body,
		}
		// send
		data, _ := json.Marshal(reqFrame)
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := entry.Conn.Write(ctx, websocket.MessageText, data); err != nil {
			log.Println("write to client:", err)
			http.Error(w, "tunnel write error", http.StatusBadGateway)
			return
		}

		// wait for response
		_, respBytes, err := entry.Conn.Read(ctx)
		if err != nil {
			log.Println("read from client:", err)
			http.Error(w, "tunnel read error", http.StatusGatewayTimeout)
			return
		}
		var resp common.ResponseFrame
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			http.Error(w, "invalid response", http.StatusInternalServerError)
			return
		}
		if resp.Error != "" {
			http.Error(w, resp.Error, http.StatusBadGateway)
			return
		}
		// copy headers
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(resp.Body)
	}
}
