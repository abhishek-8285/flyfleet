package platform

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"github.com/flyfleet/mlplatform/internal/batch"
	"github.com/flyfleet/mlplatform/internal/collector"
	"github.com/flyfleet/mlplatform/internal/config"
	"github.com/flyfleet/mlplatform/internal/drift"
	"github.com/flyfleet/mlplatform/internal/eventbus"
	"github.com/flyfleet/mlplatform/internal/experiments"
	"github.com/flyfleet/mlplatform/internal/featurestore"
	"github.com/flyfleet/mlplatform/internal/inference"
	"github.com/flyfleet/mlplatform/internal/metrics"
	"github.com/flyfleet/mlplatform/internal/scoring"
	"github.com/flyfleet/mlplatform/internal/streaming"
	"github.com/flyfleet/mlplatform/internal/warehouse"
)

// Platform is the top-level container holding all components.
type Platform struct {
	Config      *config.Config
	Bus         eventbus.Bus
	Collector   *collector.Collector
	Features    featurestore.Store
	Warehouse   *warehouse.ClickHouseStore
	Engine      *inference.Engine
	Processor   *streaming.Processor
	Scorer      *scoring.Scorer
	Experiments *experiments.Manager
	Drift       *drift.Monitor
	Scheduler   *batch.Scheduler
	Metrics     *metrics.Metrics
	App         *fiber.App
}

// New constructs the entire platform from config.
func New(cfg *config.Config) (*Platform, error) {
	// Prometheus registry + metrics (constructed first so every component
	// can emit against it).
	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	m := metrics.NewMetrics(reg)

	// Event bus
	bus, err := eventbus.NewNATS(cfg.NATS.URL, cfg.NATS.Stream, m)
	if err != nil {
		return nil, fmt.Errorf("event bus: %w", err)
	}

	// Warehouse
	wh, err := warehouse.NewClickHouse(cfg.ClickHouse.DSN)
	if err != nil {
		return nil, fmt.Errorf("warehouse: %w", err)
	}

	// Feature store (Redis online + ClickHouse offline)
	fs := featurestore.NewRedisOnlineStore(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

	// Inference engine
	engine := inference.NewEngine(cfg.Inference.ModelDir, m)

	// Experiments
	expMgr := experiments.NewManager()

	// Drift monitor
	driftMon := drift.NewMonitor(cfg.Drift.Threshold, cfg.Drift.WindowSize)
	driftMon.OnAlert(func(feature string, score float64) {
		// The monitor reports the composite "model:feature" key.
		model, feat := feature, ""
		if parts := strings.SplitN(feature, ":", 2); len(parts) == 2 {
			model, feat = parts[0], parts[1]
		}
		m.DriftAlerts.WithLabelValues(model, feat).Inc()
	})

	// Collector
	coll := collector.New(bus, m)

	// Stream processor
	proc := streaming.NewProcessor(bus)

	// Scorer
	scorer := scoring.NewScorer(fs, engine, expMgr, bus, m)

	// Scheduler
	sched, err := batch.NewScheduler(cfg.Temporal.HostPort, cfg.Temporal.Namespace, cfg.Temporal.TaskQueue)
	if err != nil {
		return nil, fmt.Errorf("scheduler: %w", err)
	}

	// HTTP app
	app := fiber.New(fiber.Config{
		AppName: "mlplatform",
	})
	app.Use(recover.New())
	app.Use(cors.New())

	p := &Platform{
		Config:      cfg,
		Bus:         bus,
		Collector:   coll,
		Features:    fs,
		Warehouse:   wh,
		Engine:      engine,
		Processor:   proc,
		Scorer:      scorer,
		Experiments: expMgr,
		Drift:       driftMon,
		Scheduler:   sched,
		Metrics:     m,
		App:         app,
	}

	p.registerRoutes()
	p.registerStreamHandlers()

	return p, nil
}

func (p *Platform) registerRoutes() {
	// Event ingestion
	p.Collector.RegisterRoutes(p.App)

	// Scoring endpoint
	p.App.Post("/v1/score", func(c *fiber.Ctx) error {
		var req scoring.ScoreRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.ErrBadRequest
		}
		resp, err := p.Scorer.Score(c.Context(), &req)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(resp)
	})

	// Feature retrieval
	p.App.Get("/v1/features/:entityType/:entityID", func(c *fiber.Ctx) error {
		entityType := c.Params("entityType")
		entityID := c.Params("entityID")

		// ?features=acceptance_rate,avg_rating,total_rides
		// When omitted, fall back to every registered feature for the entity type.
		var featureNames []string
		if raw := c.Query("features"); raw != "" {
			featureNames = strings.Split(raw, ",")
		} else {
			defs, err := p.Features.ListFeatures(c.Context(), entityType)
			if err != nil {
				return fiber.ErrInternalServerError
			}
			featureNames = make([]string, len(defs))
			for i, d := range defs {
				featureNames[i] = d.Name
			}
		}

		fv, err := p.Features.GetOnlineFeatures(c.Context(), entityType, entityID, featureNames)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Cache hit/miss tracking
		result := "hit"
		if len(fv.Features) == 0 {
			result = "miss"
		}
		p.Metrics.FeatureCacheHits.WithLabelValues(entityType, result).Inc()

		return c.JSON(fiber.Map{
			"entity_type": fv.EntityType,
			"entity_id":   fv.EntityID,
			"features":    fv.Features,
			"count":       len(fv.Features),
		})
	})

	// Batch feature retrieval for ranking pipelines
	p.App.Post("/v1/features/batch", func(c *fiber.Ctx) error {
		var req struct {
			EntityType   string   `json:"entity_type"`
			EntityIDs    []string `json:"entity_ids"`
			FeatureNames []string `json:"feature_names"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.ErrBadRequest
		}

		type result struct {
			fv  *featurestore.FeatureVector
			err error
		}
		results := make([]result, len(req.EntityIDs))
		var wg sync.WaitGroup

		for i, id := range req.EntityIDs {
			wg.Add(1)
			go func(idx int, entityID string) {
				defer wg.Done()
				fv, err := p.Features.GetOnlineFeatures(c.Context(), req.EntityType, entityID, req.FeatureNames)
				results[idx] = result{fv, err}
			}(i, id)
		}
		wg.Wait()

		vectors := make([]*featurestore.FeatureVector, 0, len(req.EntityIDs))
		for _, r := range results {
			if r.err == nil && r.fv != nil {
				vectors = append(vectors, r.fv)
			}
		}

		return c.JSON(fiber.Map{
			"entity_type": req.EntityType,
			"vectors":     vectors,
			"count":       len(vectors),
		})
	})

	// Health
	p.App.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Metrics (Prometheus) — bridged from net/http via fiber/adaptor.
	p.App.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(p.Metrics.Registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})))

	// Model registry
	p.registerModelRoutes()
}

func (p *Platform) registerModelRoutes() {
	models := p.App.Group("/v1/models")

	models.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"models": p.Engine.List()})
	})

	models.Post("/load", func(c *fiber.Ctx) error {
		var req struct {
			Name     string   `json:"name"`
			Version  string   `json:"version"`
			Path     string   `json:"path"`
			Format   string   `json:"format"`
			Features []string `json:"features"`
			Activate bool     `json:"activate"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.ErrBadRequest
		}

		var model inference.Model
		var err error

		switch req.Format {
		case "onnx":
			model, err = inference.LoadONNX(req.Name, req.Version, req.Path, req.Features)
		case "xgboost":
			model, err = inference.LoadXGBoost(req.Name, req.Version, req.Path, req.Features)
		default:
			return c.Status(400).JSON(fiber.Map{"error": "format must be onnx or xgboost"})
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		if err := p.Engine.LoadModel(model, req.Activate); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		p.Metrics.ActiveModels.Inc()
		return c.Status(201).JSON(fiber.Map{
			"status":  "loaded",
			"name":    req.Name,
			"version": req.Version,
			"active":  req.Activate,
		})
	})

	models.Post("/swap", func(c *fiber.Ctx) error {
		var req struct {
			Name       string `json:"name"`
			NewVersion string `json:"new_version"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.ErrBadRequest
		}

		if err := p.Engine.HotSwap(req.Name, req.NewVersion); err != nil {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"status":  "swapped",
			"name":    req.Name,
			"version": req.NewVersion,
		})
	})

	models.Post("/rollback", func(c *fiber.Ctx) error {
		var req struct {
			Name string `json:"name"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.ErrBadRequest
		}

		prev, err := p.Engine.PreviousVersion(req.Name)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}

		if err := p.Engine.HotSwap(req.Name, prev); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"status": "rolled_back", "version": prev})
	})
}

func (p *Platform) registerStreamHandlers() {
	pipeline := streaming.NewFeaturePipeline(p.Features, p.Warehouse)

	// Real-time feature computation
	p.Processor.On("driver.*", "feature-compute", pipeline.HandleDriverEvent)

	// Log all events to warehouse
	p.Processor.On(">", "warehouse-sink", func(ctx context.Context, evt *eventbus.Event) error {
		return p.Warehouse.WriteEvents(ctx, []warehouse.EventRow{{
			ID:        evt.ID,
			Type:      evt.Type,
			Source:    evt.Source,
			Subject:   evt.Subject,
			Timestamp: evt.Timestamp,
			Data:      string(evt.Data),
			Metadata:  evt.Metadata,
			TraceID:   evt.TraceID,
		}})
	})

	// Feed drift monitor
	p.Processor.On("predictions", "drift-monitor", func(ctx context.Context, evt *eventbus.Event) error {
		// Parse prediction and observe feature values
		return nil
	})
}

// Start boots all components.
func (p *Platform) Start(ctx context.Context) error {
	// Run migrations
	if err := p.Warehouse.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	// Start stream processors
	if err := p.Processor.Start(ctx); err != nil {
		return fmt.Errorf("streaming: %w", err)
	}

	// Schedule batch jobs
	if err := p.Scheduler.ScheduleRecurring(ctx); err != nil {
		log.Warn().Err(err).Msg("scheduler (non-fatal)")
	}

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", p.Config.Server.Host, p.Config.Server.Port)
	log.Info().Str("addr", addr).Msg("starting platform")
	return p.App.Listen(addr)
}

// Shutdown gracefully stops everything.
func (p *Platform) Shutdown() {
	log.Info().Msg("shutting down")
	p.App.Shutdown()
	p.Bus.Close()
	p.Scheduler.Close()
	p.Processor.Wait()
}
