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
// and provides methods to manage these connections.
func NewRegistry() *Registry {
	return &Registry{conns: make(map[string]*ConnEntry)}
}

// Has checks if a subdomain is already registered in the registry.
// It returns true if the subdomain exists, false otherwise.
func (r *Registry) Has(sub string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.conns[sub]
	return ok
}

// Put adds or updates a connection entry for a given subdomain in the registry.
// It locks the registry for writing to ensure thread safety.
func (r *Registry) Put(sub string, c *ConnEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns[sub] = c
}

// Get retrieves a connection entry for a given subdomain from the registry.
// It locks the registry for reading to ensure thread safety.
func (r *Registry) Get(sub string) (*ConnEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.conns[sub]
	return c, ok
}

// Delete removes a connection entry for a given subdomain from the registry.
// It locks the registry for writing to ensure thread safety.
func (r *Registry) Delete(sub string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conns, sub)
}

// uniqueSub generates a unique subdomain string of length n that is not already registered in the registry.
// It uses a loop to ensure the generated subdomain is unique.
func (r *Registry) uniqueSub(n int) string {
	for {
		s := randomString(n)
		if !r.Has(s) {
			return s
		}
	}
}
