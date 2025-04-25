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

// ProxyHandler, which creates an HTTP handler that proxies incoming requests through the WebSocket tunnel to the client
// and returns the response back to the original requester.
func ProxyHandler(reg *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.Host, ".", 2)
		if len(parts) == 0 {
			http.NotFound(w, r)
			return
		}
		sub := parts[0]

		entry, ok := reg.Get(sub)
		if !ok || entry.Conn == nil {
			http.NotFound(w, r)
			return
		}

		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()

		reqID := uuid.NewString()
		reqFrame := common.RequestFrame{
			ID:     reqID,
			Method: r.Method,
			URL:    r.URL.String(),
			Header: r.Header,
			Body:   body,
		}

		data, _ := json.Marshal(reqFrame)

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		if err := entry.Conn.Write(ctx, websocket.MessageText, data); err != nil {
			log.Println("write to client:", err)
			http.Error(w, "tunnel write error", http.StatusBadGateway)
			return
		}

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
		if resp.ID != reqID { // yanlış eşleşme (çoklu uç için güvenlik)
			http.Error(w, "mismatched response id", http.StatusBadGateway)
			return
		}
		if resp.Error != "" {
			http.Error(w, resp.Error, http.StatusBadGateway)
			return
		}

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(resp.Body)
	}
}
