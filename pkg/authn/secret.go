package authn

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/render-oss/render-mcp-server/pkg/logging"
)

// secretHeader is the header Claude/Railway callers must send to reach /mcp.
// This mirrors the X-Webhook-Secret gate used on the other Railway MCP servers.
const secretHeader = "X-Webhook-Secret"

// SecretMiddleware rejects any request whose X-Webhook-Secret header does not
// match the MCP_SECRET environment variable. When MCP_SECRET is unset the gate
// is a no-op, so the server can be deployed first and locked down afterward —
// the same behavior as the other Railway MCP deployments.
//
// The comparison is constant-time to avoid leaking the secret via timing.
func SecretMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := os.Getenv("MCP_SECRET")

		// No secret configured -> open (deploy-first mode).
		if expected == "" {
			next.ServeHTTP(w, r)
			return
		}

		provided := r.Header.Get(secretHeader)
		if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			logging.Error("auth: rejected %s %s — bad or missing %s header", r.Method, r.URL.Path, secretHeader)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
