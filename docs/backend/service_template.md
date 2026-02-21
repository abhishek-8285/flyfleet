# Standard Go Service Template

All new services in FlyFleet MUST follow this directory structure and coding standards.

```text
/cmd/server/main.go         # Entry point
/internal/
    /config/                # Configuration loading
    /database/              # DB connection and migration logic
    /events/                # NATS JetStream handlers
    /grpc/                  # gRPC server implementation
    /handler/               # HTTP handlers (if any)
    /middleware/            # Service-specific middleware
    /models/                # Domain models
    /repository/            # Data access layer
    /service/               # Business logic
/pkg/
    /common/                # Service-specific shared utilities
/migrations/                # SQL migrations (golang-migrate)
/proto/                     # Local proto copies or symlinks
go.mod
go.sum
Dockerfile
README.md
```

## Standards

1. **Protocol**: Primary communication MUST be gRPC. HTTP/JSON is only for BFF or public APIs.
2. **Database**: One database per service. No cross-service DB queries.
3. **Events**: Use NATS JetStream for async interaction.
4. **Context**: Pass `context.Context` through all layers.
5. **Logging**: Use standard Zap JSON logging.
6. **Error Handling**: Use wrapped errors with stack traces.
7. **mTLS**: Automatically enabled in gRPC server boilerplate.

```go
// Example main.go snippet
func main() {
    ctx := context.Background()
    cfg := config.Load()
    logger := logger.New(cfg.LogLevel)

    // DB setup
    db := database.Connect(cfg.DBURL)
    defer db.Close()

    // Repository & Service
    repo := repository.New(db)
    svc := service.New(repo, logger)

    // gRPC Server
    g := grpc.NewServer(svc, cfg.GRPCPort)
    g.Start(ctx)
}
```
