package server

import (
	"crypto/rand"
)

// 26 harf + 10 rakam
const alphanum = "abcdefghijklmnopqrstuvwxyz0123456789"

// randomString returns an n-character slug using [a-z0-9].
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
