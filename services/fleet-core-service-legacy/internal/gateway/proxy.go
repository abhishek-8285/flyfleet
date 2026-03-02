package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

// NewServiceProxy returns a reverse proxy that strips prefix and forwards to target.
func NewServiceProxy(prefix, target string, logger *zap.Logger) (http.Handler, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	original := proxy.Director

	proxy.Director = func(req *http.Request) {
		original(req)
		req.URL.Path = joinPath(u.Path, strings.TrimPrefix(req.URL.Path, prefix))
		if req.URL.RawPath != "" {
			req.URL.RawPath = joinPath(u.Path, strings.TrimPrefix(req.URL.RawPath, prefix))
		}
		req.Host = u.Host
		req.Header.Set("X-FlyFleet-Monolith-Proxy", "fleet-core-service-legacy")
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		logger.Error("Proxy upstream error",
			zap.String("prefix", prefix),
			zap.String("target", target),
			zap.String("path", r.URL.Path),
			zap.Error(e))
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}

	return proxy, nil
}

func joinPath(basePath, reqPath string) string {
	if reqPath == "" {
		reqPath = "/"
	}
	switch {
	case strings.HasSuffix(basePath, "/") && strings.HasPrefix(reqPath, "/"):
		return basePath + strings.TrimPrefix(reqPath, "/")
	case !strings.HasSuffix(basePath, "/") && !strings.HasPrefix(reqPath, "/"):
		return basePath + "/" + reqPath
	default:
		return basePath + reqPath
	}
}
