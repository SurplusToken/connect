package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRefreshAndSecretErrorRedaction(t *testing.T) {
	var refresh atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/login":
			fmt.Fprint(w, `{"code":0,"data":{"access_token":"old","refresh_token":"rotate","expires_in":1}}`)
		case "/api/v1/auth/refresh":
			refresh.Add(1)
			fmt.Fprint(w, `{"code":0,"data":{"access_token":"new","refresh_token":"rotated","expires_in":3600}}`)
		case "/api/v1/user/profile":
			if r.Header.Get("Authorization") != "Bearer new" {
				t.Error("expired token reused")
			}
			fmt.Fprint(w, `{"code":0,"data":{"email":"fixture@example.com","username":"nickname"}}`)
		default:
			w.WriteHeader(403)
			fmt.Fprint(w, `{"message":"fixture-secret"}`)
		}
	}))
	defer srv.Close()
	c := New()
	c.Base = srv.URL
	s, otp, e := c.Login(context.Background(), "fixture@example.com", "password")
	if e != nil || otp {
		t.Fatal(e)
	}
	saved := ""
	if e = s.Remember(func(token string) error { saved = token; return nil }); e != nil {
		t.Fatal(e)
	}
	u, e := s.Profile(context.Background())
	if e != nil || u.Nickname != "nickname" || refresh.Load() != 1 || saved != "rotated" {
		t.Fatal("refresh contract failed")
	}
	_, e = s.Groups(context.Background())
	if e == nil || strings.Contains(e.Error(), "fixture-secret") {
		t.Fatal("remote error exposed response")
	}
}
func TestNeverForwardCredentialAcrossRedirect(t *testing.T) {
	var reached atomic.Int32
	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached.Add(1) }))
	defer evil.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, evil.URL, http.StatusTemporaryRedirect)
	}))
	defer srv.Close()
	c := New()
	c.Base = srv.URL
	var out any
	e := c.request(context.Background(), "POST", "/api/v1/auth/login", "", map[string]string{"password": "fixture-secret"}, &out)
	if e == nil || reached.Load() != 0 {
		t.Fatal("redirect followed")
	}
}
func TestOTPDoesNotBecomeSessionBeforeVerification(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" {
			fmt.Fprint(w, `{"code":0,"data":{"requires_2fa":true,"temp_token":"challenge"}}`)
		} else {
			fmt.Fprint(w, `{"code":0,"data":{"access_token":"token","expires_in":3600,"user":{"email":"fixture@example.com"}}}`)
		}
	}))
	defer srv.Close()
	c := New()
	c.Base = srv.URL
	s, otp, e := c.Login(context.Background(), "fixture@example.com", "password")
	if e != nil || !otp || s.access != "" {
		t.Fatal("OTP flow not pending")
	}
	if e = s.CompleteOTP(context.Background(), "123456"); e != nil {
		t.Fatal(e)
	}
}
