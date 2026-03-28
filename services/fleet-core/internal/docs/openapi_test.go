package docs

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestBuildOpenAPIJSON_FilterMethods(t *testing.T) {
	r := chi.NewRouter()

	// Register a route with Handle which registers ALL methods in chi
	r.Handle("/api/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Register standard methods explicitly
	r.Get("/api/get", func(w http.ResponseWriter, r *http.Request) {})
	r.Post("/api/post", func(w http.ResponseWriter, r *http.Request) {})

	spec, err := BuildOpenAPIJSON(r, "http://localhost:8084")
	assert.NoError(t, err)

	var doc openAPIDoc
	err = json.Unmarshal(spec, &doc)
	assert.NoError(t, err)

	// Check /api/test
	ops, ok := doc.Paths["/api/test"]
	assert.True(t, ok)

	// Should NOT contain connect
	assert.NotContains(t, ops, "connect")

	// Should contain standard ones
	assert.Contains(t, ops, "get")
	assert.Contains(t, ops, "post")
	assert.Contains(t, ops, "put")
	assert.Contains(t, ops, "delete")
	assert.Contains(t, ops, "patch")
	assert.Contains(t, ops, "options")
	assert.Contains(t, ops, "head")
	assert.Contains(t, ops, "trace")

	// Verify count - exactly 8 standard methods
	assert.Equal(t, 8, len(ops))
}
