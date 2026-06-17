package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"go-nc-mcp-web/internal/nextcloud"
)

type contextKey struct{}

// ClientFromContext retrieves the authenticated Nextcloud client from the request context.
func ClientFromContext(ctx context.Context) *nextcloud.Client {
	c, _ := ctx.Value(contextKey{}).(*nextcloud.Client)
	return c
}

// NextcloudAppToken validates requests using either:
//   - Basic auth (username:app-token)
//   - Bearer token (base64-encoded username:app-token, from OAuth flow)
func NextcloudAppToken(ncHost string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, token, ok := r.BasicAuth()
		if !ok {
			// Try Bearer token (base64-encoded username:password from OAuth)
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				encoded := strings.TrimPrefix(auth, "Bearer ")
				decoded, err := base64.StdEncoding.DecodeString(encoded)
				if err == nil {
					parts := strings.SplitN(string(decoded), ":", 2)
					if len(parts) == 2 {
						username, token = parts[0], parts[1]
						ok = true
					}
				}
			}
		}

		if !ok || username == "" || token == "" {
			w.Header().Set("WWW-Authenticate", `Bearer realm="mcp"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		nc := nextcloud.NewClient(ncHost, username, token)

		var user struct {
			ID string `json:"id"`
		}
		err := nc.DoJSON("GET", "/ocs/v2.php/cloud/user?format=json", nil, &user)
		if err != nil {
			http.Error(w, "unauthorized: invalid nextcloud credentials", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKey{}, nc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
