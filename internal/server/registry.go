package server

import (
    "sync"

    "nhooyr.io/websocket"
)

type ConnEntry struct {
    Conn *websocket.Conn
}

type Registry struct {
    mu   sync.RWMutex
    conns map[string]*ConnEntry
}

func NewRegistry() *Registry {
    return &Registry{conns: make(map[string]*ConnEntry)}
}

func (r *Registry) Put(sub string, c *ConnEntry) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.conns[sub] = c
}

func (r *Registry) Get(sub string) (*ConnEntry, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    c, ok := r.conns[sub]
    return c, ok
}

func (r *Registry) Delete(sub string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.conns, sub)
}
