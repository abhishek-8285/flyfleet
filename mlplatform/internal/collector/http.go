package collector

import (
	"encoding/json"
	"io"

	"github.com/gofiber/fiber/v2"
)

type IngestRequest struct {
	EventType string            `json:"event_type"`
	Source    string            `json:"source"`
	Subject   string            `json:"subject"`
	Data      json.RawMessage   `json:"data"`
	Metadata  map[string]string `json:"metadata"`
}

// RegisterRoutes mounts the collector HTTP API.
func (c *Collector) RegisterRoutes(app *fiber.App) {
	api := app.Group("/v1/events")

	api.Post("/", func(ctx *fiber.Ctx) error {
		body, err := io.ReadAll(ctx.Request().BodyStream())
		if err != nil {
			return fiber.ErrBadRequest
		}

		var req IngestRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return fiber.ErrBadRequest
		}

		if err := c.Ingest(ctx.Context(), req.EventType, req.Source, req.Subject, req.Data); err != nil {
			return fiber.ErrInternalServerError
		}

		return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "accepted"})
	})

	// Batch ingest
	api.Post("/batch", func(ctx *fiber.Ctx) error {
		var reqs []IngestRequest
		if err := ctx.BodyParser(&reqs); err != nil {
			return fiber.ErrBadRequest
		}

		for _, req := range reqs {
			_ = c.Ingest(ctx.Context(), req.EventType, req.Source, req.Subject, req.Data)
		}

		return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"accepted": len(reqs),
		})
	})
}
