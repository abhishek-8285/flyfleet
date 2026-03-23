package control

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type RouteInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type UpstreamCount struct {
	Service string `json:"service"`
	Target  string `json:"target"`
	Count   int    `json:"count"`
	Error   string `json:"error,omitempty"`
}

type UpstreamHealth struct {
	Service          string `json:"service"`
	Target           string `json:"target"`
	Reachable        bool   `json:"reachable"`
	StatusCode       int    `json:"status_code,omitempty"`
	OpenAPIAvailable bool   `json:"openapi_available"`
	OpenAPICount     int    `json:"openapi_count,omitempty"`
	Error            string `json:"error,omitempty"`
}

func LocalRoutes(routes chi.Routes) ([]RouteInfo, error) {
	items := make([]RouteInfo, 0, 512)
	seen := map[string]struct{}{}
	err := chi.Walk(routes, func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		path := normalizePath(route)
		method = strings.ToUpper(strings.TrimSpace(method))
		if path == "" || method == "" || !isBusinessMethod(method) {
			return nil
		}

		// chi route walkers often emit both "/x" and "/x/*" for handlers mounted on
		// a prefix; treat them as one API surface in inventory output.
		path = strings.TrimSuffix(path, "/*")
		if path == "" {
			path = "/"
		}

		key := method + " " + path
		if _, ok := seen[key]; ok {
			return nil
		}
		seen[key] = struct{}{}

		items = append(items, RouteInfo{
			Method: strings.ToUpper(method),
			Path:   path,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Path == items[j].Path {
			return items[i].Method < items[j].Method
		}
		return items[i].Path < items[j].Path
	})
	return items, nil
}

func UniqueMethodPathCount(items []RouteInfo) int {
	uniq := map[string]struct{}{}
	for _, it := range items {
		uniq[it.Method+" "+it.Path] = struct{}{}
	}
	return len(uniq)
}

func CountUpstreamOpenAPI(ctx context.Context, serviceName, target string) UpstreamCount {
	out := UpstreamCount{Service: serviceName, Target: target}
	client := &http.Client{Timeout: 2 * time.Second}
	candidates := []string{
		strings.TrimRight(target, "/") + "/swagger/openapi.json",
		strings.TrimRight(target, "/") + "/openapi.json",
	}
	for _, u := range candidates {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			out.Error = err.Error()
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			out.Error = err.Error()
			continue
		}
		func() {
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				out.Error = fmt.Sprintf("status %d at %s", resp.StatusCode, u)
				return
			}
			var payload struct {
				Paths map[string]map[string]interface{} `json:"paths"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				out.Error = err.Error()
				return
			}
			total := 0
			for _, methods := range payload.Paths {
				for method := range methods {
					switch strings.ToLower(method) {
					case "get", "post", "put", "patch", "delete", "options", "head":
						total++
					}
				}
			}
			out.Count = total
			out.Error = ""
		}()
		if out.Error == "" {
			return out
		}
	}
	return out
}

func CheckUpstreamHealth(ctx context.Context, serviceName, target string) UpstreamHealth {
	out := UpstreamHealth{
		Service: serviceName,
		Target:  target,
	}
	client := &http.Client{Timeout: 2 * time.Second}

	healthCandidates := []string{
		strings.TrimRight(target, "/") + "/health",
		strings.TrimRight(target, "/"),
	}
	for _, u := range healthCandidates {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			out.Error = err.Error()
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			out.Error = err.Error()
			continue
		}
		func() {
			defer resp.Body.Close()
			out.Reachable = resp.StatusCode >= 200 && resp.StatusCode < 500
			out.StatusCode = resp.StatusCode
		}()
		if out.Reachable {
			break
		}
	}

	count := CountUpstreamOpenAPI(ctx, serviceName, target)
	if count.Error == "" {
		out.OpenAPIAvailable = true
		out.OpenAPICount = count.Count
	} else if out.Error == "" {
		out.Error = count.Error
	}

	return out
}

func normalizePath(p string) string {
	path := strings.TrimSpace(p)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func isBusinessMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
