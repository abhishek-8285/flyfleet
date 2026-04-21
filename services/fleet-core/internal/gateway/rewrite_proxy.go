package gateway

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// NewRewriteProxy forwards requests to the same monolith host while rewriting prefix.
func NewRewriteProxy(fromPrefix, toPrefix string, logger *zap.Logger) http.HandlerFunc {
	client := &http.Client{}
	fromPrefix = strings.TrimSuffix(fromPrefix, "/")
	toPrefix = strings.TrimSuffix(toPrefix, "/")

	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, fromPrefix) {
			http.NotFound(w, r)
			return
		}

		hops, _ := strconv.Atoi(r.Header.Get("X-FlyFleet-Rewrite-Hop"))
		if hops > 3 {
			http.Error(w, "rewrite loop detected", http.StatusBadGateway)
			return
		}

		tail := strings.TrimPrefix(r.URL.Path, fromPrefix)
		if tail == "" {
			tail = "/"
		}
		targetPath := joinPath(toPrefix, tail)

		targetURL := &url.URL{
			Scheme:   "http",
			Host:     "127.0.0.1:8084", // Use internal loopback port
			Path:     targetPath,
			RawQuery: r.URL.RawQuery,
		}

		req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL.String(), r.Body)
		if err != nil {
			http.Error(w, "failed to build rewrite request", http.StatusInternalServerError)
			return
		}
		req.Header = r.Header.Clone()
		propagateIdentityHeaders(req, r)
		req.Header.Set("X-FlyFleet-Rewrite-Hop", strconv.Itoa(hops+1))
		req.Header.Set("X-FlyFleet-Rewrite-From", fromPrefix)
		req.Header.Set("X-FlyFleet-Rewrite-To", toPrefix)

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Rewrite proxy upstream error", zap.String("from", fromPrefix), zap.String("to", toPrefix), zap.String("path", r.URL.Path), zap.Error(err))
			http.Error(w, "upstream unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for k, vals := range resp.Header {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}
}

func propagateIdentityHeaders(outgoing *http.Request, incoming *http.Request) {
	if outgoing.Header.Get("X-Organization-ID") != "" && outgoing.Header.Get("X-User-ID") != "" {
		return
	}

	authHeader := strings.TrimSpace(incoming.Header.Get("Authorization"))
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenString == "" {
		return
	}

	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return
	}

	if outgoing.Header.Get("X-Organization-ID") == "" {
		if tenantID, ok := claims["tenant_id"].(string); ok && strings.TrimSpace(tenantID) != "" {
			outgoing.Header.Set("X-Organization-ID", tenantID)
		}
	}

	if outgoing.Header.Get("X-User-ID") == "" {
		if userID, ok := claims["user_id"].(string); ok && strings.TrimSpace(userID) != "" {
			outgoing.Header.Set("X-User-ID", userID)
			return
		}
		if subject, ok := claims["sub"].(string); ok && strings.TrimSpace(subject) != "" {
			outgoing.Header.Set("X-User-ID", subject)
		}
	}
}
