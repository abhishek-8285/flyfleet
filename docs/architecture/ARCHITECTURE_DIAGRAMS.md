# FlyFleet Architecture Diagrams

Generated: 2026-03-27

---

## 1. High-Level System Architecture

```mermaid
graph TB
    subgraph "Frontend Layer"
        LP[Landing Page<br/>Next.js]
        DW[Dashboard Web<br/>React + Vite]
        MA[Mobile App<br/>React Native]
    end

    subgraph "API Layer"
        BFF[Web BFF<br/>Gateway]
        FC[Fleet Core Service<br/>Modular Monolith<br/>Go + Chi]
    end

    subgraph "Specialized Services"
        ML[Fleet ML Service<br/>Python/Go + ML Models]
    end

    subgraph "Data Layer"
        PG[(TimescaleDB<br/>PostgreSQL 15 HA)]
        REDIS[(Redis 7<br/>Cache & Sessions)]
        MINIO[(MinIO<br/>Object Storage)]
    end

    subgraph "Messaging Layer"
        NATS[NATS JetStream<br/>Event Streaming]
    end

    subgraph "Infrastructure"
        K8s[Kubernetes Cluster]
        MON[Observability<br/>Prometheus + Grafana]
        VAULT[Vault<br/>Secrets Mgmt]
    end

    LP --> BFF
    DW --> BFF
    MA --> BFF

    BFF --> FC
    FC --> PG
    FC --> REDIS
    FC --> NATS
    FC --> MINIO

    ML --> PG
    NATS -.-> ML

    FC --> MON
    NATS --> MON

    BFF --> K8s
    FC --> K8s
    PG --> K8s
    NATS --> K8s

    K8s --> VAULT
```

---

## 2. Fleet Core Service Internal Architecture (Modular Monolith)

```mermaid
graph TB
    subgraph "HTTP Layer"
        GW[APIBusGateway<br/>Middleware Chain]
        ROUTER[Chi Router<br/>Route Dispatcher]
    end

    subgraph "Domain Layer (27 Domains)"
        AUTH[Auth Domain]
        VEHICLE[Vehicle Domain]
        DRIVER[Driver Domain]
        TRACKING[Tracking Domain]
        TELEMETRY[Telemetry Domain]
        BILLING[Billing Domain]
        FINANCE[Finance Domain]
        INDIA[India Compliance]
        EMERGENCY[Emergency Domain]
        ALERT[Alert Domain]
        MAINTENANCE[Maintenance Domain]
        TRIP[Trip Domain]
        ANALYTICS[Analytics Domain]
        REPORTS[Reports Domain]
        GOVERNANCE[Governance Domain]
        SETTINGS[Settings Domain]
        INTEGRATIONS[Integrations Domain]
        COMPLIANCE[Compliance Domain]
        FUEL[Fuel Domain]
        GEOFENCING[Geofencing Domain]
        ROUTE[Route Domain]
        VIDEO[Video Domain]
        CUSTOMER[Customer Domain]
        DIAGNOSTICS[Diagnostics Domain]
        MONITORING[Monitoring Domain]
        ADMIN[Admin Support]
        PLATFORM[Platform Compat]
    end

    subgraph "Cross-Cutting Concerns"
        PLATFORM_LAYER[Platform Layer<br/>iam/, billing/, events/,<br/>gateway/, middleware/, metrics/]
        OBSERVABILITY[Observability<br/>Logging, Metrics, Traces]
    end

    subgraph "Infrastructure Layer"
        DB[(PostgreSQL<br/>Repository Pattern)]
        REDIS_CACHE[(Redis Cache)]
        EVENTBUS[NATS EventBus]
        MINIO_S3[(MinIO S3)]
    end

    GW --> ROUTER

    ROUTER --> AUTH
    ROUTER --> VEHICLE
    ROUTER --> DRIVER
    ROUTER --> TRACKING
    ROUTER --> TELEMETRY
    ROUTER --> BILLING
    ROUTER --> FINANCE
    ROUTER --> INDIA
    ROUTER --> EMERGENCY
    ROUTER --> ALERT
    ROUTER --> MAINTENANCE
    ROUTER --> TRIP
    ROUTER --> ANALYTICS
    ROUTER --> REPORTS
    ROUTER --> GOVERNANCE
    ROUTER --> SETTINGS
    ROUTER --> INTEGRATIONS
    ROUTER --> COMPLIANCE
    ROUTER --> FUEL
    ROUTER --> GEOFENCING
    ROUTER --> ROUTE
    ROUTER --> VIDEO
    ROUTER --> CUSTOMER
    ROUTER --> DIAGNOSTICS
    ROUTER --> MONITORING
    ROUTER --> ADMIN
    ROUTER --> PLATFORM

    AUTH --> PLATFORM_LAYER
    VEHICLE --> PLATFORM_LAYER
    DRIVER --> PLATFORM_LAYER
    TRACKING --> PLATFORM_LAYER

    PLATFORM_LAYER --> OBSERVABILITY

    AUTH --> DB
    VEHICLE --> DB
    DRIVER --> DB
    TRACKING --> DB

    AUTH --> REDIS_CACHE
    VEHICLE --> REDIS_CACHE

    TELEMETRY --> EVENTBUS
    TRACKING --> EVENTBUS
    ALERT --> EVENTBUS

    BILLING --> MINIO_S3
    REPORTS --> MINIO_S3

    style FC fill:#e1f5ff
    style GW fill:#f0f0f0
    style DB fill:#fff3e0
    style NATS fill:#f3e5f5
```

---

## 3. Domain Internal Structure (Hexagonal Pattern)

```mermaid
graph LR
    subgraph "Domain: Example (Auth)"
        H[Handlers<br/>HTTP Adapters]
        S[Service<br/>Application Logic]
        R[Repository<br/>Persistence Adapter]
        M[Models/DTOs<br/>Domain Objects]
    end

    subgraph "Infrastructure"
        DB_EXT[(PostgreSQL<br/>External)]
        REDIS_EXT[(Redis<br/>External)]
    end

    H --> S
    S --> R
    R --> DB_EXT
    S --> M
    H --> M
    M --> S

    H -.-> REDIS_EXT
    S -.-> EVENTBUS[NATS EventBus]
    S -.-> MINIO[MinIO S3]
    S -.-> EXTERNAL[Third-Party APIs]
    end

    style H fill:#e8f5e8
    style S fill:#e3f2fd
    style R fill:#fff3e0
    style M fill:#f3e5f5
```

---

## 4. Deployment Architecture

```mermaid
graph TB
    subgraph "Kubernetes Cluster: flyfleet"
        subgraph "Namespace: flyfleet"
            subgraph "Deployments"
                FC_DEP[fleet-core-service<br/>3 replicas<br/>8080 → API/Handlers]
                ML_DEP[fleet-ml-service<br/>1 replica<br/>8000 → ML endpoint]
            end

            subgraph "Stateful Services"
                PG_HA[(TimescaleDB HA<br/>PostgreSQL 15)]
                REDIS_HA[(Redis HA Cluster)]
                NATS_CLUST[NATS JetStream<br/>Cluster Mode]
                MINIO_CLUST[MinIO Cluster<br/>Object Storage]
            end

            subgraph "Services (ClusterIP)"
                FC_SVC[Fleet Core Service<br/>LoadBalancer]
                ML_SVC[ML Service]
                PG_SVC[PostgreSQL Service]
                REDIS_SVC[Redis Service]
                NATS_SVC[NATS Service]
            end

            subgraph "Observability"
                PROM[Prometheus<br/>Metrics Collection]
                GRAFANA[Grafana<br/>Dashboards]
                TEMPO[Tempo<br/>Traces]
                LOKI[Loki<br/>Logs]
            end

            subgraph "Ingress & Security"
                INGRESS[Ingress Controller<br/>TLS Termination]
                AUTH_PROXY[Auth Proxy<br/>OAuth/OIDC]
                WAF[WAF<br/>Rate Limiting]
            end
        end

        subgraph "External"
            INTERNET[Internet / Users]
            GITHUB[GitHub Actions<br/>CI/CD]
            VAULT[Vault<br/>Secrets Mgmt]
        end
    end

    INTERNET --> INGRESS
    INGRESS --> WAF
    WAF --> AUTH_PROXY
    AUTH_PROXY --> FC_SVC

    FC_DEP --> PG_SVC
    FC_DEP --> REDIS_SVC
    FC_DEP --> NATS_SVC
    FC_DEP --> MINIO_CLUST

    ML_DEP --> PG_SVC

    FC_SVC --> PG_HA
    FC_SVC --> REDIS_HA
    FC_SVC --> NATS_CLUST
    FC_SVC --> MINIO_CLUST

    FC_DEP --> PROM
    FC_DEP --> LOKI
    NATS_CLUST --> PROM

    PROM --> GRAFANA
    TEMPO --> GRAFANA
    LOKI --> GRAFANA

    VAULT --> FC_DEP
    VAULT --> PG_HA
    VAULT --> REDIS_HA

    GITHUB --> INGRESS

    style FC_DEP fill:#ffebee
    style PG_HA fill:#fff3e0
    style NATS_CLUST fill:#f3e5f5
    style REDIS_HA fill:#e8f5e8
```

---

## 5. Data Flow: Telemetry Ingestion

```mermaid
sequenceDiagram
    participant Device as IoT Device/GPS
    participant Ingestion as Tracking/Ingestion Domain
    participant NATS as NATS JetStream
    participant Telemetry as Telemetry Domain
    participant DB as TimescaleDB
    participant Cache as Redis Cache

    Device->>Ingestion: POST /api/v1/telemetry<br/>(GPS, OBD, Sensor data)
    Ingestion->>NATS: Publish to "telemetry.raw" stream
    NATS-->>Ingestion: Ack
    Ingestion-->>Device: 202 Accepted

    Note over NATS,Telemetry: Asynchronous Processing

    NATS->>Telemetry: Subscribe to "telemetry.raw"
    Telemetry->>Telemetry: Validate & Transform
    Telemetry->>DB: Insert time-series data
    Telemetry->>Cache: Update latest position
    Telemetry->>NATS: Publish to "telemetry.processed"
    NATS-->>Telemetry: Ack

    Note over Telemetry,Alert: Event-Driven Alerts

    NATS->>Alert: Subscribe to "telemetry.processed"
    Alert->>Alert: Evaluate rules (geofence, speed)
    alt Violation Detected
        Alert->>Cache: Store alert
        Alert->>NATS: Publish to "alerts.new"
        NATS->>Notification: Push notification
    end
```

---

## 6. API Gateway Request Flow

```mermaid
flowchart TD
    Start[Client Request] --> Gateway{APIBusGateway Middleware}

    Gateway --> Log[1. Log Request & Metrics]
    Log --> Tenant{2. Tenant Context Check}

    Tenant -->|Public Route| AuthCheck[3. Skip Auth]
    Tenant -->|Private Route| ValidateJWT[3. Validate JWT<br/>JWKS]

    AuthCheck --> Router[4. Chi Router<br/>Route to Domain Handler]
    ValidateJWT --> Router

    Router --> DomainHandler[Domain Handler<br/>(e.g., AuthHandlers)]

    DomainHandler --> Service[Service Layer<br/>Business Logic]
    Service --> Repo[Repository<br/>Data Access]
    Repo --> DB[(Database)]

    Service --> EventBus[NATS EventBus<br/>Publish Events]
    Service --> Cache[Redis Cache<br/>Read/Write]

    Repo --> Result[Return Results]
    EventBus --> Async[Async Processing]
    Cache --> FastReturn[Fast Response]

    Result --> DomainHandler
    DomainHandler --> Gateway
    Gateway --> Response[HTTP Response]

    Response --> End[Client]

    style Gateway fill:#e3f2fd
    style Service fill:#f3e5f5
    style Repo fill:#fff3e0
    style DB fill:#ffebee
```

---

## 7. Multi-Tenancy Isolation

```mermaid
graph TB
    subgraph "Request Flow"
        REQ[Incoming Request]
        MIDDLEWARE[Tenant Middleware]
        CONTEXT[org_id in Context]
        HANDLER[Domain Handler]
        DB_QUERY[SQL Query with org_id]
    end

    subgraph "Database Row-Level Isolation"
        subgraph "Tenant A"
            ROW1[Vehicle<br/>org_id: 'a1b2c3']
            ROW2[Driver<br/>org_id: 'a1b2c3']
        end

        subgraph "Tenant B"
            ROW3[Vehicle<br/>org_id: 'd4e5f6']
            ROW4[Driver<br/>org_id: 'd4e5f6']
        end
    end

    subgraph "Query Examples"
        Q1[SELECT * FROM vehicles<br/>WHERE org_id = $1<br/>AND id = $2]
        Q2[SELECT * FROM drivers<br/>WHERE org_id = $1]
    end

    REQ --> MIDDLEWARE
    MIDDLEWARE --> CONTEXT
    CONTEXT --> HANDLER
    HANDLER --> DB_QUERY

    DB_QUERY --> ROW1
    DB_QUERY -.-> ROW3

    style MIDDLEWARE fill:#ffcccb
    style DB_QUERY fill:#ffebcc
```

---

## 8. Technology Stack Map

```mermaid
graph LR
    subgraph "Backend"
        GO[Go 1.26]
        CHI[Chi Router v5]
        PGX[pgx v5<br/>PostgreSQL Driver]
        NATSGO[NATS.go v1.48]
        REDISGO[redis/go-redis/v9]
        JWT[golang-jwt/jwt/v5]
        ZAP[go.uber.org/zap]
        OTEL[OpenTelemetry SDK]
    end

    subgraph "Frontend"
        REACT[React 19]
        VITE[Vite 5]
        TS[TypeScript 4.9]
        RTK[Redux Toolkit]
        QS[React Query v5]
        ZUST[Zustand]
        TAIL[TailwindCSS 3.4]
        LM[Leaflet/MapLibre]
        FM[framer-motion]
    end

    subgraph "Data"
        TSDB[TimescaleDB<br/>PostgreSQL 15]
        REDIS[Redis 7]
        NATSJS[NATS JetStream 2.10]
        MINIOS3[MinIO S3]
    end

    subgraph "Infra"
        K8S[Kubernetes]
        DOCKER[Docker]
        HELM[Helm Charts]
        TERRAFORM[Terraform]
    end

    subgraph "Observability"
        PROM[Prometheus]
        GRAFANA[Grafana]
        JAEGER[Jaeger/Tempo]
        LOKI[Loki]
    end

    subgraph "CI/CD"
        GH[GitHub Actions]
        PRE_COMMIT[Pre-commit Hooks]
        TEST[Testcontainers]
    end

    GO --> CHI
    CHI --> PGX
    CHI --> NATSGO
    CHI --> REDISGO
    CHI --> JWT
    CHI --> ZAP
    CHI --> OTEL

    REACT --> VITE
    VITE --> TS
    TS --> RTK
    TS --> QS
    TS --> ZUST
    TS --> TAIL
    TS --> LM
    TS --> FM

    GO --> TSDB
    GO --> REDIS
    GO --> NATSJS
    GO --> MINIOS3

    DOCKER --> K8S
    K8S --> HELM
    TERRAFORM --> K8S

    OTEL --> PROM
    PROM --> GRAFANA
    OTEL --> JAEGER
    ZAP --> LOKI

    GH --> DOCKER
    PRE_COMMIT --> GO
    PRE_COMMIT --> TS
    TEST --> GO
```

---

## 9. Proto/API Contracts

```mermaid
graph TB
    subgraph "Proto Definitions<br/>(proto/ directory)"
        A_AUTH[auth/v1/auth.proto]
        A_VEHICLE[vehicle/v1/vehicle.proto]
        A_DRIVER[driver/v1/driver.proto]
        A_TRACKING[tracking/v1/tracking.proto]
        A_BILLING[billing/v1/billing.proto]
        A_INDIA[india/v1/india.proto]
        A_ALERT[alert/v1/alert.proto]
        A_ANALYTICS[analytics/v1/analytics.proto]
        A_COMPLIANCE[compliance/v1/compliance.proto]
        A_FLEET[fleet/v1/fleet.proto]
        A_MAINT[maintenance/v1/maintenance.proto]
        A_REPORTS[reports/v1/reports.proto]
        A_ADMIN[admin/v1/admin.proto]
    end

    subgraph "Generated Code (Go)"
        GEN_AUTH[github.com/flyfleet/proto/auth/v1]
        GEN_VEHICLE[github.com/flyfleet/proto/vehicle/v1]
        GEN_DRIVER[github.com/flyfleet/proto/driver/v1]
        GEN_TRACKING[github.com/flyfleet/proto/tracking/v1]
    end

    subgraph "Generated Code (TypeScript)"
        TS_CLIENTS[@flyfleet/api-clients]
    end

    A_AUTH --> GEN_AUTH
    A_VEHICLE --> GEN_VEHICLE
    A_DRIVER --> GEN_DRIVER
    A_TRACKING --> GEN_TRACKING
    A_BILLING --> GEN_BILLING
    A_INDIA --> GEN_INDIA

    GEN_AUTH --> AUTH_HANDLERS[Auth Handlers]
    GEN_VEHICLE --> VEHICLE_HANDLERS[Vehicle Handlers]
    GEN_DRIVER --> DRIVER_HANDLERS[Driver Handlers]
    GEN_TRACKING --> TRACKING_HANDLERS[Tracking Handlers]

    AUTH_HANDLERS --> TS_CLIENTS
    VEHICLE_HANDLERS --> TS_CLIENTS
    DRIVER_HANDLERS --> TS_CLIENTS
    TRACKING_HANDLERS --> TS_CLIENTS

    TS_CLIENTS --> DASHBOARD[Dashboard Web]
    TS_CLIENTS --> MOBILE[Mobile App]

    style A_AUTH fill:#e8f5e8
    style GEN_AUTH fill:#e3f2fd
    style TS_CLIENTS fill:#f3e5f5
```

---

## 10. Shared Packages Structure

```mermaid
graph TB
    subgraph "packages/common/go"
        subgraph "logging"
            LOG[zap Logger<br/>Wrapper & Utilities]
        end
        subgraph "errors"
            ERR[Error Types<br/>Domain Errors]
        end
        subgraph "middleware"
            MW[Common Middleware<br/>CORS, Recovery, Metrics]
        end
        subgraph "config"
            CFG[Config Loader<br/>Viper-based]
        end
        subgraph "database"
            DB_U[Database Utilities<br/>Connection Pooling]
        end
        subgraph "observability"
            OBS[OpenTelemetry<br/>Integration]
        end
        subgraph "security"
            SEC[JWT, Encryption<br/>Password Hashing]
        end
    end

    subgraph "packages/config"
        subgraph "branding"
            BRAND[Whitelabel Branding<br/>Colors, Logos, Themes]
        end
        subgraph "localization"
            LOC[Multi-language<br/>Support]
        end
    end

    FC[fleet-core-service] --> COMMON
    FC --> CONFIG

    WEB[dashboard-web] --> CONFIG

    COMMON --> LOG
    COMMON --> ERR
    COMMON --> MW
    COMMON --> CFG
    COMMON --> DB_U
    COMMON --> OBS
    COMMON --> SEC

    CONFIG --> BRAND
    CONFIG --> LOC

    style COMMON fill:#e3f2fd
    style CONFIG fill:#fff3e0
    style FC fill:#ffebee
```

---

## Notes

- Diagrams represent architecture as of March 2026 (based on codebase exploration)
- Some components (e.g., auth-service, ingestion-service separate deployments) appear to be legacy K8s configs
- Current primary deployment is fleet-core-service monolith with 3 replicas
- Event-driven architecture via NATS JetStream enables loose coupling
- Modular monolith allows easy future extraction of domains to microservices
