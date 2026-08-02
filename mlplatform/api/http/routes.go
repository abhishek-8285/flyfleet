package httpapi

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

// HealthResponse is the /health payload.
type HealthResponse struct {
	Status string `json:"status"`
}

// RegisterHealth mounts a simple health endpoint.
func RegisterHealth(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(HealthResponse{Status: "ok"})
	})
}

// MetricsHandler returns the default Prometheus registry.
func MetricsHandler() *prometheus.Registry {
	return prometheus.DefaultRegisterer.(*prometheus.Registry)
}
