package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
)

const (
	SessionCookieName = "aire_admin_session"

	argonTime    uint32 = 1
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32

	sessionTTL = 24 * time.Hour
)

// DeriveHashHex derives an Argon2id hash encoded as lowercase hex.
func DeriveHashHex(passphrase string, salt string) (string, error) {
	if salt == "" {
		return "", errors.New("salt cannot be empty")
	}
	var saltBytes []byte
	decoded, err := hex.DecodeString(salt)
	if err == nil && len(decoded) > 0 {
		saltBytes = decoded
	} else {
		saltBytes = []byte(salt)
	}

	key := argon2.IDKey([]byte(passphrase), saltBytes, argonTime, argonMemory, argonThreads, argonKeyLen)
	return hex.EncodeToString(key), nil
}

// GenerateRandomSaltHex returns a 16-byte cryptographically random salt encoded as hex.
func GenerateRandomSaltHex() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// Authenticator handles passphrase verification and HMAC-signed session cookies.
type Authenticator struct {
	expectedHashHex string
	salt            string
	sessionSecret   []byte
	nowFn           func() time.Time
}

// NewAuthenticator creates an Authenticator configured with Argon2id hash/salt and HMAC secret.
func NewAuthenticator(expectedHashHex, salt, sessionSecret string) (*Authenticator, error) {
	if strings.TrimSpace(expectedHashHex) == "" {
		return nil, errors.New("ADMIN_PASSPHRASE_HASH cannot be empty")
	}
	if strings.TrimSpace(salt) == "" {
		return nil, errors.New("ADMIN_PASSPHRASE_SALT cannot be empty")
	}
	if strings.TrimSpace(sessionSecret) == "" {
		return nil, errors.New("SESSION_SECRET cannot be empty")
	}
	return &Authenticator{
		expectedHashHex: strings.ToLower(strings.TrimSpace(expectedHashHex)),
		salt:            strings.TrimSpace(salt),
		sessionSecret:   []byte(sessionSecret),
		nowFn:           time.Now,
	}, nil
}

// VerifyPassphrase checks the provided phrase against the configured Argon2id hash in constant time.
func (a *Authenticator) VerifyPassphrase(phrase string) bool {
	if phrase == "" {
		return false
	}
	derivedHex, err := DeriveHashHex(phrase, a.salt)
	if err != nil {
		return false
	}
	if len(derivedHex) != len(a.expectedHashHex) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(derivedHex), []byte(a.expectedHashHex)) == 1
}

func (a *Authenticator) signExpiration(expUnix int64) string {
	payload := strconv.FormatInt(expUnix, 10)
	mac := hmac.New(sha256.New, a.sessionSecret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return payload + "." + sig
}

// IssueSessionCookie writes an HttpOnly HMAC-signed session cookie.
func (a *Authenticator) IssueSessionCookie(w http.ResponseWriter, secure bool) {
	exp := a.nowFn().Add(sessionTTL).Unix()
	token := a.signExpiration(exp)
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

// ClearSessionCookie invalidates the session cookie on the client.
func (a *Authenticator) ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

// ValidateSessionRequest verifies the session cookie signature and expiration.
func (a *Authenticator) ValidateSessionRequest(r *http.Request) bool {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return false
	}

	expUnix, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false
	}
	if a.nowFn().Unix() >= expUnix {
		return false
	}

	gotSig, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, a.sessionSecret)
	mac.Write([]byte(parts[0]))
	expectedSig := mac.Sum(nil)

	return hmac.Equal(gotSig, expectedSig)
}

// RequireAdminMiddleware rejects unauthenticated requests with a generic 401 response.
func (a *Authenticator) RequireAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.ValidateSessionRequest(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(domain.ErrorResponse{
				Error:   "unauthorized",
				Message: "Unauthorized",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// IPRateLimiter enforces a sliding-window attempt limit per client IP.
type IPRateLimiter struct {
	mu          sync.Mutex
	maxRequests int
	window      time.Duration
	attempts    map[string][]time.Time
	nowFn       func() time.Time
}

// NewIPRateLimiter constructs an IPRateLimiter allowing maxRequests within window.
func NewIPRateLimiter(maxRequests int, window time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		maxRequests: maxRequests,
		window:      window,
		attempts:    make(map[string][]time.Time),
		nowFn:       time.Now,
	}
}

// Allow reports whether a request from ip is permitted under the sliding window limit.
func (rl *IPRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.nowFn()
	cutoff := now.Add(-rl.window)

	// Opportunistic cleanup to keep memory bounded.
	if len(rl.attempts) > 1000 {
		for k, times := range rl.attempts {
			var valid []time.Time
			for _, t := range times {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.attempts, k)
			} else {
				rl.attempts[k] = valid
			}
		}
	}

	existing := rl.attempts[ip]
	pruned := make([]time.Time, 0, len(existing)+1)
	for _, t := range existing {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}

	if len(pruned) >= rl.maxRequests {
		rl.attempts[ip] = pruned
		return false
	}

	pruned = append(pruned, now)
	rl.attempts[ip] = pruned
	return true
}

// ExtractClientIP returns the IP address portion of r.RemoteAddr.
func ExtractClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return fmt.Sprintf("unknown")
}
