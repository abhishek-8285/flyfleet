# Hardcoded Values Fix Plan

## Priority 1: Critical Security Issues (Fix Immediately)

### 1. Database Credentials
**Files affected:** All service `main.go` files and some test files
**Action:** 
- Replace hardcoded DB connections with environment variables
- Use consistent pattern across all services

**Example fix pattern:**
```go
// Before
db, err := sql.Open("postgres", "postgres://user:password@localhost:5432/dbname")

// After  
dbHost := os.Getenv("DB_HOST")
dbPort := os.Getenv("DB_PORT")
dbUser := os.Getenv("DB_USER")
dbPass := os.Getenv("DB_PASSWORD")
dbName := os.Getenv("DB_NAME")

dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName)
db, err := sql.Open("postgres", dsn)
```

### 2. API Keys and Secrets
**Files affected:**
- `/services/microservices/web-bff/cmd/main.go` (lines 118-119)
- `/services/video-service/internal/handlers/video_handlers.go` (line 311)

**Action:**
- Move to environment variables or secure vault
- Never commit secrets to code repository

### 3. Email Addresses
**Files affected:**
- Multiple support handlers in web-bff
- Notification service files

**Action:**
- Move to configuration or database
- Use template system for notifications

## Priority 2: Configuration Issues

### Hosts and Ports
**Pattern:** Replace all hardcoded `localhost:port` with env vars
**Services affected:** All services
**Variables needed:**
- `DB_HOST`, `DB_PORT`
- `SERVICE_PORT`
- `API_BASE_URL`

## Priority 3: Code Quality Improvements

### Magic Numbers
**Action:** Create named constants for all magic numbers

**Example:**
```go
// Before
if duration > 3600 {
    // logic
}

// After
const MaxTripDurationSeconds = 3600
if duration > MaxTripDurationSeconds {
    // logic
}
```

### Hardcoded Strings
**Action:** Move to configuration or constants
- Error messages
- Default values
- Business rule strings

## Implementation Steps

### Phase 1: Framework Setup (1-2 days)
1. Create shared configuration package
2. Define standard environment variable names
3. Add configuration validation

### Phase 2: Service-by-Service Refactoring (5-7 days)
1. **Web BFF** - Highest issue count (26 issues)
2. **Trip Service** - Many magic numbers (13 issues)  
3. **Auth Service** - Configuration issues (12 issues)
4. **Remaining services** - 3-5 issues each

### Phase 3: Testing and Validation (2-3 days)
1. Update all test files
2. Verify environment variable loading
3. Add configuration validation tests

## Environment Variables to Define

### Database
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=flyfleet
DB_PASSWORD=secure_password
DB_NAME=flyfleet_db
```

### Service Configuration
```
SERVICE_PORT=8080
API_BASE_URL=http://localhost:8080
LOG_LEVEL=info
```

### Secrets (Store securely)
```
JWT_SECRET_KEY=
ENCRYPTION_KEY=
THIRD_PARTY_API_KEY=
```

## Validation Checklist

Before merging changes:
- [ ] All database connections use env vars
- [ ] No hardcoded passwords/secrets in code
- [ ] All magic numbers replaced with constants
- [ ] Configuration loading tested
- [ ] Environment variable documentation updated
- [ ] Tests updated to use configuration
- [ ] Docker/CI environments updated

## Risk Mitigation

1. **Backward compatibility:** Support both env vars and current hardcoded values temporarily
2. **Rollback plan:** Keep current configuration as fallback
3. **Testing:** Comprehensive test coverage before deployment
4. **Documentation:** Update all deployment guides
