package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

const (
	HeaderTraceID   = "X-Trace-ID"
	HeaderRequestID = "X-Request-ID"
)

type traceIDKey struct{}

// TraceIDMiddleware guarantees a trace identifier for every request.
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := strings.TrimSpace(r.Header.Get(HeaderTraceID))
		if traceID == "" {
			traceID = strings.TrimSpace(r.Header.Get(HeaderRequestID))
		}
		if traceID == "" {
			traceID = strings.TrimSpace(chiMiddleware.GetReqID(r.Context()))
		}
		if traceID == "" {
			traceID = fmt.Sprintf("trace-%d", time.Now().UnixNano())
		}

		ctx := context.WithValue(r.Context(), traceIDKey{}, traceID)
		w.Header().Set(HeaderTraceID, traceID)
		w.Header().Set(HeaderRequestID, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value := ctx.Value(traceIDKey{})
	traceID, ok := value.(string)
	if !ok {
		return ""
	}
	return traceID
}
