package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Bridge is the native Wails service. It shares the exact same request pipeline
// as the development API, without relying on custom-scheme HTTP POST bodies.
type Bridge struct{ api *API }

func NewBridge(api *API) *Bridge { return &Bridge{api: api} }

type BridgeResponse struct {
	Status int             `json:"status"`
	Data   json.RawMessage `json:"data"`
}
type responseBuffer struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (b *responseBuffer) Header() http.Header  { return b.header }
func (b *responseBuffer) WriteHeader(code int) { b.status = code }
func (b *responseBuffer) Write(p []byte) (int, error) {
	if b.status == 0 {
		b.status = 200
	}
	return b.body.Write(p)
}
func (b *Bridge) Request(ctx context.Context, action, session, body string) (BridgeResponse, error) {
	if len(action) > 40 || strings.ContainsAny(action, "/?\\") || len(body) > 32<<10 {
		return BridgeResponse{}, errors.New("请求格式无效")
	}
	r, err := http.NewRequestWithContext(ctx, "POST", "http://wails.localhost/connect/"+action, strings.NewReader(body))
	if err != nil {
		return BridgeResponse{}, errors.New("请求格式无效")
	}
	r.Header.Set("X-Connect-Request", "1")
	r.Header.Set("X-Connect-Session", session)
	r.Header.Set("Origin", "wails://localhost")
	w := &responseBuffer{header: http.Header{}}
	b.api.ServeHTTP(w, r)
	return BridgeResponse{Status: w.status, Data: json.RawMessage(w.body.Bytes())}, nil
}
