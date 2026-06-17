package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// OAuthServer implements a minimal OAuth 2.1 authorization server that bridges
// to Nextcloud app tokens. The access token is the base64-encoded username:app-password.
type OAuthServer struct {
	Issuer string // e.g. https://mcp.h-i-r.net

	mu    sync.Mutex
	codes map[string]*authCode // code -> credentials
}

type authCode struct {
	username string
	password string
	clientID string
	redirect string
	challenge string
	expires  time.Time
}

func NewOAuthServer(issuer string) *OAuthServer {
	return &OAuthServer{
		Issuer: issuer,
		codes:  make(map[string]*authCode),
	}
}

// Metadata serves the OAuth authorization server metadata document.
func (o *OAuthServer) Metadata(w http.ResponseWriter, r *http.Request) {
	meta := map[string]any{
		"issuer":                 o.Issuer,
		"authorization_endpoint": o.Issuer + "/authorize",
		"token_endpoint":         o.Issuer + "/token",
		"response_types_supported":             []string{"code"},
		"grant_types_supported":                []string{"authorization_code"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"none"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}

// Authorize handles the authorization endpoint.
// GET shows a login form, POST processes credentials and redirects with a code.
func (o *OAuthServer) Authorize(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	codeChallenge := r.URL.Query().Get("code_challenge")
	state := r.URL.Query().Get("state")

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html><head><title>Nextcloud MCP Login</title>
<style>
body { font-family: sans-serif; max-width: 400px; margin: 80px auto; }
input { display: block; width: 100%; padding: 8px; margin: 8px 0; box-sizing: border-box; }
button { padding: 10px 20px; margin-top: 8px; }
</style></head><body>
<h2>Nextcloud MCP Authorization</h2>
<p>Enter your Nextcloud username and app password.</p>
<form method="POST">
<input type="hidden" name="client_id" value="` + clientID + `">
<input type="hidden" name="redirect_uri" value="` + redirectURI + `">
<input type="hidden" name="code_challenge" value="` + codeChallenge + `">
<input type="hidden" name="state" value="` + state + `">
<input type="text" name="username" placeholder="Nextcloud username" required>
<input type="password" name="password" placeholder="App password" required>
<button type="submit">Authorize</button>
</form></body></html>`))
		return
	}

	// POST: process form
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	clientID = r.FormValue("client_id")
	redirectURI = r.FormValue("redirect_uri")
	codeChallenge = r.FormValue("code_challenge")
	state = r.FormValue("state")

	if username == "" || password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}

	// Generate auth code
	codeBytes := make([]byte, 32)
	rand.Read(codeBytes)
	code := hex.EncodeToString(codeBytes)

	o.mu.Lock()
	o.codes[code] = &authCode{
		username:  username,
		password:  password,
		clientID:  clientID,
		redirect:  redirectURI,
		challenge: codeChallenge,
		expires:   time.Now().Add(5 * time.Minute),
	}
	o.mu.Unlock()

	// Redirect back with code
	sep := "?"
	if idx := len(redirectURI); idx > 0 {
		for _, c := range redirectURI {
			if c == '?' {
				sep = "&"
				break
			}
		}
	}
	location := redirectURI + sep + "code=" + code
	if state != "" {
		location += "&state=" + state
	}
	log.Printf("OAuth authorize: user=%s, redirecting to %s", username, redirectURI)
	http.Redirect(w, r, location, http.StatusFound)
}

// Token handles the token exchange endpoint.
func (o *OAuthServer) Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	grantType := r.FormValue("grant_type")
	code := r.FormValue("code")

	if grantType != "authorization_code" || code == "" {
		writeOAuthError(w, "invalid_request", "invalid grant_type or missing code")
		return
	}

	o.mu.Lock()
	ac, ok := o.codes[code]
	if ok {
		delete(o.codes, code)
	}
	o.mu.Unlock()

	if !ok || time.Now().After(ac.expires) {
		writeOAuthError(w, "invalid_grant", "code expired or invalid")
		return
	}

	// The access token is base64(username:password)
	token := base64.StdEncoding.EncodeToString([]byte(ac.username + ":" + ac.password))

	log.Printf("OAuth token exchange: user=%s", ac.username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   86400,
	})
}

func writeOAuthError(w http.ResponseWriter, code, desc string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{
		"error":             code,
		"error_description": desc,
	})
}
