package common

import "net/http"

type RequestFrame struct {
    ID      string            `json:"id"`
    Method  string            `json:"method"`
    URL     string            `json:"url"`
    Header  http.Header       `json:"header"`
    Body    []byte            `json:"body,omitempty"`
}

type ResponseFrame struct {
    ID         string      `json:"id"`
    StatusCode int         `json:"status_code"`
    Header     http.Header `json:"header"`
    Body       []byte      `json:"body,omitempty"`
    Error      string      `json:"error,omitempty"`
}
