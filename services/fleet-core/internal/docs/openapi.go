package docs

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

type openAPIDoc struct {
	OpenAPI string                         `json:"openapi"`
	Info    openAPIInfo                    `json:"info"`
	Servers []openAPIServer                `json:"servers"`
	Paths   map[string]map[string]openAPIOp `json:"paths"`
}

type openAPIInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

type openAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type openAPIOp struct {
	Summary     string                      `json:"summary"`
	OperationID string                      `json:"operationId"`
	Tags        []string                    `json:"tags,omitempty"`
	Responses   map[string]openAPIResponse `json:"responses"`
}

type openAPIResponse struct {
	Description string `json:"description"`
}

// BuildOpenAPIJSON creates a minimal OpenAPI document from the live chi router.
func BuildOpenAPIJSON(routes chi.Routes, serverURL string) ([]byte, error) {
	doc, err := buildLocalDoc(routes, serverURL)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(doc, "", "  ")
}

// BuildCombinedOpenAPIJSON returns local routes plus upstream OpenAPI routes mounted at /api/edge/{service}.
func BuildCombinedOpenAPIJSON(routes chi.Routes, serverURL string, upstreams map[string]string) ([]byte, error) {
	doc, err := buildLocalDoc(routes, serverURL)
	if err != nil {
		return nil, err
	}

	for service, target := range upstreams {
		paths := fetchUpstreamPaths(service, target)
		for p, methods := range paths {
			mountedPath := "/api/edge/" + service + normalizePath(p)
			if _, ok := doc.Paths[mountedPath]; !ok {
				doc.Paths[mountedPath] = map[string]openAPIOp{}
			}
			for m := range methods {
				method := strings.ToLower(m)
				// Filter unsupported methods from upstream as well for safety
				mu := strings.ToUpper(method)
				if mu != "GET" && mu != "POST" && mu != "PUT" && mu != "DELETE" && mu != "PATCH" && mu != "OPTIONS" && mu != "HEAD" && mu != "TRACE" {
					continue
				}

				doc.Paths[mountedPath][method] = openAPIOp{
					Summary:     fmt.Sprintf("UPSTREAM %s %s %s", service, strings.ToUpper(method), p),
					OperationID: operationID(method, "/api/edge/"+service+normalizePath(p)),
					Tags:        []string{"upstream-" + service},
					Responses: map[string]openAPIResponse{
						"200": {Description: "Success"},
					},
				}
			}
		}
	}

	return json.MarshalIndent(doc, "", "  ")
}

func buildLocalDoc(routes chi.Routes, serverURL string) (openAPIDoc, error) {
	doc := openAPIDoc{
		OpenAPI: "3.0.3",
		Info: openAPIInfo{
			Title:       "FlyFleet Monolith API",
			Version:     "1.0.0",
			Description: "Auto-generated from chi route registrations in fleet-core-service-legacy.",
		},
		Servers: []openAPIServer{
			{
				URL:         serverURL,
				Description: "Monolith service",
			},
		},
		Paths: map[string]map[string]openAPIOp{},
	}

	err := chi.Walk(routes, func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if route == "" || method == "" {
			return nil
		}

		// OpenAPI 3.0.3 only supports: get, put, post, delete, options, head, patch, trace.
		// Methods like 'CONNECT' are rejected by standard validators.
		m := strings.ToUpper(method)
		if m != "GET" && m != "POST" && m != "PUT" && m != "DELETE" && m != "PATCH" && m != "OPTIONS" && m != "HEAD" && m != "TRACE" {
			return nil
		}

		path := normalizePath(route)
		opMethod := strings.ToLower(method)

		if _, ok := doc.Paths[path]; !ok {
			doc.Paths[path] = map[string]openAPIOp{}
		}
		doc.Paths[path][opMethod] = openAPIOp{
			Summary:     fmt.Sprintf("%s %s", strings.ToUpper(method), path),
			OperationID: operationID(method, path),
			Tags:        []string{tagForPath(path)},
			Responses: map[string]openAPIResponse{
				"200": {Description: "Success"},
			},
		}
		return nil
	})
	if err != nil {
		return doc, err
	}

	// Keep deterministic operation ordering by recreating path method maps sorted by key.
	for path, ops := range doc.Paths {
		keys := make([]string, 0, len(ops))
		for k := range ops {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sortedOps := make(map[string]openAPIOp, len(keys))
		for _, k := range keys {
			sortedOps[k] = ops[k]
		}
		doc.Paths[path] = sortedOps
	}

	return doc, nil
}

func fetchUpstreamPaths(serviceName, target string) map[string]map[string]interface{} {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil
	}
	client := &http.Client{Timeout: 2 * time.Second}
	candidates := []string{
		strings.TrimRight(target, "/") + "/swagger/openapi.json",
		strings.TrimRight(target, "/") + "/openapi.json",
	}
	for _, candidate := range candidates {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, candidate, nil)
		if err != nil {
			cancel()
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			cancel()
			continue
		}
		var payload struct {
			Paths map[string]map[string]interface{} `json:"paths"`
		}
		func() {
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				return
			}
			_ = json.NewDecoder(resp.Body).Decode(&payload)
		}()
		cancel()
		if len(payload.Paths) > 0 {
			return payload.Paths
		}
	}
	_ = serviceName
	return nil
}

func normalizePath(p string) string {
	path := strings.TrimSpace(p)
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = strings.ReplaceAll(path, "/*", "")
	if path == "" {
		return "/"
	}
	return path
}

func operationID(method, route string) string {
	sanitized := strings.Trim(route, "/")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "{", "")
	sanitized = strings.ReplaceAll(sanitized, "}", "")
	if sanitized == "" {
		sanitized = "root"
	}
	return strings.ToLower(method) + "_" + sanitized
}

func tagForPath(route string) string {
	trim := strings.Trim(route, "/")
	if trim == "" {
		return "root"
	}
	parts := strings.Split(trim, "/")
	if len(parts) > 1 && parts[0] == "api" && parts[1] == "v1" && len(parts) > 2 {
		return parts[2]
	}
	return parts[0]
}
