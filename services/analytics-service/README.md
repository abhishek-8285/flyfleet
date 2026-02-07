# Analytics Service

The Analytics Service is a microservice that provides AI/ML analytics capabilities for the FlyFleet platform. It handles predictive modeling, fleet analytics, driver behavior analysis, and predictive maintenance insights.

## Features

### Predictive Modeling
- Create and manage ML models for various analytics tasks
- Support for different model types (regression, classification, clustering)
- Model versioning and lifecycle management
- Model performance tracking and evaluation

### Prediction Results
- Store and retrieve prediction results
- Confidence scoring and uncertainty quantification
- Anomaly detection capabilities
- Feature importance analysis

### Analytics Jobs
- Batch and real-time analytics processing
- Job scheduling and monitoring
- Progress tracking and error handling
- Flexible job parameterization

### Fleet Analytics
- Comprehensive fleet performance metrics
- Fuel efficiency analysis
- Maintenance cost tracking
- Safety score calculation
- Trip and routing analytics

### Driver Behavior Analytics
- Driver performance scoring
- Aggressive driving detection
- Safety and efficiency metrics
- Behavioral trend analysis
- Improvement recommendations

### Predictive Maintenance Analytics
- Vehicle component risk assessment
- Maintenance scheduling recommendations
- Failure prediction modeling
- Cost estimation and optimization
- Parts inventory management

## Architecture

The service follows a clean architecture pattern with the following layers:

- **Models**: Data structures and domain entities
- **Repository**: Data access layer with PostgreSQL integration
- **Service**: Business logic and validation layer
- **Handlers**: HTTP API endpoints and request handling

## Database Schema

The service uses PostgreSQL with the following main tables:

- `predictive_models` - ML model definitions and metadata
- `prediction_results` - Individual prediction outcomes
- `analytics_jobs` - Analytics job definitions and status
- `fleet_analytics` - Fleet-level performance metrics
- `driver_behavior_analytics` - Driver performance analysis
- `predictive_maintenance_analytics` - Vehicle maintenance insights

## API Endpoints

### Health Check
```
GET /health
```

### Predictive Models
```
POST /predictive-models
GET /predictive-models
GET /predictive-models/{id}
PUT /predictive-models/{id}
DELETE /predictive-models/{id}
```

### Prediction Results
```
POST /prediction-results
GET /prediction-results
GET /prediction-results/{id}
```

### Analytics Jobs
```
POST /analytics-jobs
GET /analytics-jobs
GET /analytics-jobs/{id}
PUT /analytics-jobs/{id}
DELETE /analytics-jobs/{id}
```

### Fleet Analytics
```
POST /fleet-analytics
GET /fleet-analytics
GET /fleet-analytics/{id}
PUT /fleet-analytics/{id}
DELETE /fleet-analytics/{id}
```

### Driver Behavior Analytics
```
POST /driver-behavior-analytics
GET /driver-behavior-analytics
GET /driver-behavior-analytics/{id}
PUT /driver-behavior-analytics/{id}
DELETE /driver-behavior-analytics/{id}
```

### Predictive Maintenance Analytics
```
POST /predictive-maintenance-analytics
GET /predictive-maintenance-analytics
GET /predictive-maintenance-analytics/{id}
PUT /predictive-maintenance-analytics/{id}
DELETE /predictive-maintenance-analytics/{id}
```

## Configuration

The service can be configured using environment variables:

- `PORT` - HTTP server port (default: 8080)
- `DATABASE_URL` - PostgreSQL connection string
- `LOG_LEVEL` - Logging level (debug, info, warn, error)

## Deployment

### Docker

```bash
docker build -t analytics-service .
docker run -p 8080:8080 -e DATABASE_URL=your_db_url analytics-service
```

### Docker Compose

```yaml
analytics-service:
  build: ./services/analytics-service
  ports:
    - "8080:8080"
  environment:
    - DATABASE_URL=postgres://user:pass@db:5432/analytics_db
  depends_on:
    - postgres
```

## Database Migrations

Run migrations using your preferred migration tool:

```bash
# Apply migrations
psql -f migrations/001_create_analytics_tables.sql
```

## Development

### Prerequisites
- Go 1.21+
- PostgreSQL 13+
- Docker (optional)

### Setup
1. Clone the repository
2. Install dependencies: `go mod download`
3. Set up PostgreSQL database
4. Run migrations
5. Start the service: `go run main.go`

### Testing
```bash
go test ./...
```

## Monitoring and Observability

The service includes:
- Health check endpoints
- Structured logging
- Error handling and recovery
- Performance metrics collection

## Security

- Input validation and sanitization
- SQL injection prevention through parameterized queries
- Proper error handling to avoid information leakage
- Secure database connection handling

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.