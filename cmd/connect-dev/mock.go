package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// Explicit offline fixture only. It is not compiled into the Wails app.
func mockGateway() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := any(nil)
		switch r.URL.Path {
		case "/api/v1/auth/login", "/api/v1/auth/refresh":
			data = map[string]any{"access_token": "fixture-access", "refresh_token": "fixture-refresh", "expires_in": 3600}
		case "/api/v1/auth/logout":
			data = map[string]bool{"ok": true}
		case "/api/v1/user/profile":
			data = map[string]string{"email": "preview@example.com", "username": "测试用户"}
		case "/api/v1/groups/available":
			data = []any{map[string]any{"id": 1, "name": "离线测试分组", "platform": "anthropic", "status": "active", "allow_messages_dispatch": true}, map[string]any{"id": 2, "name": "OpenAI 测试分组", "platform": "openai", "status": "active"}}
		case "/api/v1/keys":
			data = map[string]any{"items": []any{map[string]any{"id": 1, "name": "离线测试 Key", "group_id": 1, "key": "fixture-not-a-real-key", "status": "active"}, map[string]any{"id": 2, "name": "OpenAI 测试 Key", "group_id": 2, "key": "fixture-openai-key", "status": "active"}}, "pages": 1}
		case "/api/v1/keys/1":
			data = map[string]any{"id": 1, "name": "离线测试 Key", "group_id": 1, "key": "fixture-not-a-real-key", "status": "active"}
		case "/api/v1/keys/2":
			data = map[string]any{"id": 2, "name": "OpenAI 测试 Key", "group_id": 2, "key": "fixture-openai-key", "status": "active"}
		case "/v1/models":
			if r.Header.Get("Authorization") == "Bearer fixture-openai-key" {
				json.NewEncoder(w).Encode(map[string]any{"data": []any{map[string]string{"id": "gpt-fixture"}}})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{map[string]string{"id": "fixture-model"}}})
			return
		default:
			w.WriteHeader(404)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data})
	}))
}
