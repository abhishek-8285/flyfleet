package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	port := "3005"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	// Termux Absolute Path Fix
	buildPath := "/data/data/com.termux/files/home/flyfleet/apps/dashboard-web/build"

	backendURL, _ := url.Parse("http://127.0.0.1:8081")
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 1. Proxy /api
		if strings.HasPrefix(r.URL.Path, "/api") {
			proxy.ServeHTTP(w, r)
			return
		}

		// 2. Resolve target file path
		cleanPath := filepath.Clean(r.URL.Path)
		if cleanPath == "/" {
			cleanPath = "/index.html"
		}
		
		targetPath := filepath.Join(buildPath, cleanPath)

		// 3. If file exists, serve it
		if stat, err := os.Stat(targetPath); err == nil && !stat.IsDir() {
			http.ServeFile(w, r, targetPath)
			return
		}

		// 4. Fallback to index.html for SPA
		http.ServeFile(w, r, filepath.Join(buildPath, "index.html"))
	})

	fmt.Printf("🚀 UI Server starting on http://localhost:%s\n", port)
	fmt.Printf("📁 Serving from: %s\n", buildPath)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
