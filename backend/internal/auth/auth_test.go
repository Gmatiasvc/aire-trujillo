package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthenticatorPassphraseAndCookies(t *testing.T) {
	phrase := "Caminas por el mismo camino pero no estas en ningun lugar porque dejaste ir si no queria estar en ningun lugar"
	salt := "616972655f7472756a696c6c6f5f73616c74"

	hashHex, err := DeriveHashHex(phrase, salt)
	if err != nil {
		t.Fatalf("DeriveHashHex error: %v", err)
	}

	a, err := NewAuthenticator(hashHex, salt, "test-session-secret")
	if err != nil {
		t.Fatalf("NewAuthenticator error: %v", err)
	}

	if !a.VerifyPassphrase(phrase) {
		t.Fatalf("expected valid passphrase to verify")
	}
	if a.VerifyPassphrase("wrong phrase") {
		t.Fatalf("expected invalid passphrase to fail")
	}
	if a.VerifyPassphrase("") {
		t.Fatalf("expected empty passphrase to fail")
	}

	// Issue cookie and validate request
	rec := httptest.NewRecorder()
	a.IssueSessionCookie(rec, false)
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if !c.HttpOnly || c.Name != SessionCookieName {
		t.Fatalf("unexpected cookie attributes: %+v", c)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/nodes", nil)
	req.AddCookie(c)
	if !a.ValidateSessionRequest(req) {
		t.Fatalf("expected issued cookie to validate")
	}

	// Tampered cookie must fail
	badReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/nodes", nil)
	badReq.AddCookie(&http.Cookie{Name: SessionCookieName, Value: c.Value + "00"})
	if a.ValidateSessionRequest(badReq) {
		t.Fatalf("expected tampered cookie to fail validation")
	}

	// Expired session must fail
	currentTime := time.Now()
	a.nowFn = func() time.Time { return currentTime.Add(25 * time.Hour) }
	if a.ValidateSessionRequest(req) {
		t.Fatalf("expected expired cookie to fail validation")
	}
	a.nowFn = time.Now

	// Middleware rejects unauthorized with generic 401
	protected := a.RequireAdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	unauthRec := httptest.NewRecorder()
	protected.ServeHTTP(unauthRec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/nodes", nil))
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthRec.Code)
	}

	authRec := httptest.NewRecorder()
	protected.ServeHTTP(authRec, req)
	if authRec.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid cookie, got %d", authRec.Code)
	}
}

func TestIPRateLimiter(t *testing.T) {
	now := time.Unix(1700000000, 0)
	rl := NewIPRateLimiter(3, time.Minute)
	rl.nowFn = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if !rl.Allow("127.0.0.1") {
			t.Fatalf("expected attempt %d to be allowed", i+1)
		}
	}
	if rl.Allow("127.0.0.1") {
		t.Fatalf("expected 4th attempt within window to be blocked")
	}
	if !rl.Allow("10.0.0.2") {
		t.Fatalf("expected different IP to be allowed")
	}

	// Advance past window
	now = now.Add(61 * time.Second)
	if !rl.Allow("127.0.0.1") {
		t.Fatalf("expected IP to be allowed again after window expires")
	}
}
