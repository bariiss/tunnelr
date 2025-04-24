package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bariiss/tunnelr/internal/common"
	"github.com/coder/websocket"
)

type Forwarder struct {
	conn       *websocket.Conn
	localAddr  string
	httpClient *http.Client
}

// NewForwarder creates a new forwarder instance that handles WebSocket communication with the tunnel server
func NewForwarder(c *websocket.Conn, localAddr string) *Forwarder {
	return &Forwarder{
		conn:      c,
		localAddr: localAddr,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Serve starts the forwarding service, continuously reading WebSocket messages from the server
func (f *Forwarder) Serve(ctx context.Context) error {
	for {
		_, data, err := f.conn.Read(ctx)
		if err != nil {
			return err
		}
		go f.handleFrame(ctx, data)
	}
}

// handleFrame processes an incoming WebSocket frame, forwards the request to the local server, and returns the response
func (f *Forwarder) handleFrame(ctx context.Context, data []byte) {
	var req common.RequestFrame
	if err := json.Unmarshal(data, &req); err != nil {
		log.Println("invalid frame:", err)
		return
	}

	// build local request
	localURL := "http://" + f.localAddr + req.URL
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, localURL, bytes.NewReader(req.Body))
	if err != nil {
		log.Println("new req:", err)
		return
	}
	httpReq.Header = req.Header

	resp, err := f.httpClient.Do(httpReq)
	var respFrame common.ResponseFrame
	respFrame.ID = req.ID
	if err != nil {
		respFrame.StatusCode = 502
		respFrame.Error = err.Error()
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		respFrame.StatusCode = resp.StatusCode
		respFrame.Header = resp.Header
		respFrame.Body = body
	}

	bytesData, _ := json.Marshal(respFrame)
	err = f.conn.Write(ctx, websocket.MessageText, bytesData)
	if err != nil {
		log.Println("write response error:", err)
	}
}
