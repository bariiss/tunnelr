package server

import (
	"sync"

	"github.com/coder/websocket"
)

type ConnEntry struct {
	Conn *websocket.Conn
}

type Registry struct {
	mu    sync.RWMutex
	conns map[string]*ConnEntry
}

// NewRegistry creates a new connection registry to track active WebSocket connections by subdomain
func NewRegistry() *Registry {
	return &Registry{conns: make(map[string]*ConnEntry)}
}

// Put adds or updates a connection entry for a given subdomain in the registry
func (r *Registry) Put(sub string, c *ConnEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns[sub] = c
}

// Get retrieves a connection entry for a given subdomain from the registry
func (r *Registry) Get(sub string) (*ConnEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.conns[sub]
	return c, ok
}

// Delete removes a connection entry for a given subdomain from the registry
func (r *Registry) Delete(sub string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conns, sub)
}
