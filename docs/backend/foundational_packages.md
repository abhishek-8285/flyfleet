# FlyFleet Foundational Packages

These packages provide standardized, production-safe patterns for all FlyFleet microservices.

## Observability

- [`logger`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/logger/logger.go): Zap-based JSON logging with automatic `trace_id`, `org_id`, and `vehicle_id` extraction from context.
- [`metrics`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/metrics/metrics.go): Prometheus metrics for HTTP/gRPC latency, error rates, and DB/NATS performance.
- [`tracer`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/tracer/tracer.go): OpenTelemetry-based distributed tracing.
- [`monitoring`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/monitoring/slo.go): SLO recording boilerplate and standard health checks.

## Resilience & Reliability

- [`resilience`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/resilience/resilience.go): Circuit breakers (`gobreaker`) and exponential retries (`backoff`) with production-safe defaults.
- [`idempotency`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/idempotency/idempotency.go): Redis-backed deduplication layer for critical operations.
- [`grpcclient`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/grpcclient/client.go): Standard gRPC client wrapper enforcing timeouts and resilience policies.
- [`warmup`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/warmup/warmup.go): Startup routine to warm DB and Redis connection pools.

## Data & Messaging

- [`database`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/database/standard.go): Standard connection pooling, `golang-migrate` integration, and Read/Write separation.
- [`timescale`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/database/timescale.go): TimescaleDB-specific hypertable and retention policy management.
- [`messaging`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/messaging/jetstream.go): NATS JetStream wrapper with durable consumers and idempotent replay guards.
- [`cache`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/cache/cache.go): Generic Redis caching with pattern-based invalidation.
- [`worker`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/worker/worker.go): Background job framework using `asynq`.

## Utilities

- [`flyfleet-gen`](file:///home/abhishek/Desktop/flyfleet/scripts/flyfleet-gen/main.go): CLI tool to generate standardized service templates.
- [`clocksync`](file:///home/abhishek/Desktop/flyfleet/packages/common/go/pkg/clocksync/clocksync.go): NTP synchronization checks for token/event consistency.
