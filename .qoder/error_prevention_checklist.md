# Error Prevention Checklist for Fleet Platform Development

## Pre-Commit Validation

### Go Code Validation
- [ ] `go mod tidy` - Dependencies are clean and up to date
- [ ] `gofumpt -w .` - Code is properly formatted
- [ ] `golangci-lint run --fast` - Static analysis passes
- [ ] `go test -short ./...` - Unit tests pass
- [ ] `go build ./...` - Code compiles without errors
- [ ] No vet warnings or errors
- [ ] Proper error handling in all functions
- [ ] Context is used appropriately for cancellation
- [ ] SQL queries use parameterized statements
- [ ] Struct tags are properly formatted

### TypeScript/React Validation
- [ ] `npm run lint` - ESLint passes without errors
- [ ] `npm run type-check` - TypeScript compilation succeeds
- [ ] `npm run test:ci` - Tests pass in CI environment
- [ ] No console errors or warnings in development
- [ ] Proper TypeScript interfaces defined
- [ ] React components have proper prop types
- [ ] Hooks are used correctly
- [ ] Async operations have proper error handling
- [ ] Components clean up subscriptions/effects
- [ ] Accessibility attributes are present where needed

### Kotlin/Android Validation
- [ ] `./gradlew ktlintCheck` - Code style validation passes
- [ ] `./gradlew detekt` - Static code analysis passes
- [ ] `./gradlew testDebugUnitTest` - Unit tests pass
- [ ] `./gradlew connectedDebugAndroidTest` - Instrumentation tests pass
- [ ] Proper null safety annotations
- [ ] Coroutines are used correctly
- [ ] ViewModel lifecycle is handled properly
- [ ] Permissions are requested appropriately
- [ ] Room database operations are on background threads
- [ ] Network calls handle offline scenarios

### Database Validation
- [ ] SQL migrations use `IF NOT EXISTS` for idempotency
- [ ] Proper foreign key constraints are in place
- [ ] Indexes exist for frequently queried columns
- [ ] `EXPLAIN ANALYZE` shows acceptable query performance
- [ ] Migration files have proper rollback procedures
- [ ] TimescaleDB hypertables configured correctly
- [ ] Compression policies set for time-series data
- [ ] Soft delete patterns implemented consistently
- [ ] UUID primary keys used appropriately
- [ ] Spatial data uses GEOGRAPHY type

### Security Validation
- [ ] No hardcoded secrets or credentials
- [ ] `gitleaks detect` finds no sensitive information
- [ ] `trivy fs .` shows no critical vulnerabilities
- [ ] `npm audit --production` shows no high severity issues
- [ ] Input validation on all API endpoints
- [ ] Proper authentication and authorization implemented
- [ ] SQL injection prevention measures in place
- [ ] XSS prevention in frontend code
- [ ] CSRF protection for state-changing operations
- [ ] Secure headers configured (CSP, HSTS, etc.)

### Docker/Infrastructure Validation
- [ ] `hadolint Dockerfile*` shows no critical issues
- [ ] Multi-stage builds used for Go applications
- [ ] `.dockerignore` files exclude unnecessary files
- [ ] Health checks implemented for all services
- [ ] Proper resource limits set in Docker Compose
- [ ] Environment variables used for configuration
- [ ] Secrets mounted properly, not hardcoded
- [ ] Base images are kept up to date
- [ ] Security scanning passes for container images

## Code Review Checklist

### Architecture & Design
- [ ] Follows clean architecture principles
- [ ] Proper separation of concerns
- [ ] SOLID principles applied appropriately
- [ ] Design patterns used correctly
- [ ] Code is maintainable and readable
- [ ] Appropriate use of interfaces/abstractions
- [ ] Business logic separated from infrastructure
- [ ] Proper error boundaries implemented
- [ ] Consistent naming conventions followed
- [ ] Code duplication minimized

### Performance Considerations
- [ ] Database queries optimized with proper indexes
- [ ] N+1 query problems identified and resolved
- [ ] Caching strategies implemented where appropriate
- [ ] Connection pooling configured correctly
- [ ] Memory usage optimized
- [ ] Proper pagination for large datasets
- [ ] Asynchronous operations used appropriately
- [ ] Resource cleanup implemented
- [ ] Performance benchmarks meet requirements
- [ ] Load testing conducted for critical paths

### Testing Requirements
- [ ] Minimum 80% code coverage achieved
- [ ] Unit tests cover business logic
- [ ] Integration tests cover API endpoints
- [ ] Edge cases and error conditions tested
- [ ] Mocks used appropriately in unit tests
- [ ] Test data is realistic and comprehensive
- [ ] Tests are fast and reliable
- [ ] No flaky tests present
- [ ] Property-based testing for complex logic
- [ ] Contract testing for API compatibility

### Documentation & Communication
- [ ] Code comments explain complex logic
- [ ] Public APIs documented with OpenAPI/Swagger
- [ ] README files updated if needed
- [ ] Breaking changes clearly communicated
- [ ] Migration guides provided when needed
- [ ] Inline documentation accurate and helpful
- [ ] Diagrams updated for architecture changes
- [ ] Changelog entries added
- [ ] Pull request description is comprehensive
- [ ] Review comments addressed properly

## Deployment Readiness Checklist

### Pre-Deployment
- [ ] All tests passing in CI/CD pipeline
- [ ] Security scans completed with no critical issues
- [ ] Performance benchmarks meet requirements
- [ ] Database migrations tested in staging
- [ ] Rollback procedures documented and tested
- [ ] Monitoring and alerting configured
- [ ] Backup procedures verified
- [ ] Capacity planning completed
- [ ] Incident response procedures updated
- [ ] Stakeholder communication plan ready

### Post-Deployment
- [ ] Health checks passing in production
- [ ] Key metrics within expected ranges
- [ ] No critical alerts triggered
- [ ] User acceptance testing completed
- [ ] Performance monitoring shows expected results
- [ ] Error rates within acceptable thresholds
- [ ] Resource utilization within limits
- [ ] Backup jobs running successfully
- [ ] Logs flowing to central logging system
- [ ] Tracing working for critical user flows

## Common Error Prevention Patterns

### Database Related
```
❌ DON'T: SELECT * FROM large_table
✅ DO: SELECT specific_columns FROM table WHERE indexed_condition

❌ DON'T: Use string concatenation for SQL queries
✅ DO: Use parameterized queries with placeholders

❌ DON'T: Forget foreign key constraints
✅ DO: Add proper foreign key relationships with cascading

❌ DON'T: Skip migration testing
✅ DO: Test migrations in staging environment first
```

### API Related
```
❌ DON'T: Return raw database errors to clients
✅ DO: Map errors to user-friendly messages

❌ DON'T: Skip input validation
✅ DO: Validate all inputs at API boundary

❌ DON'T: Ignore rate limiting
✅ DO: Implement proper rate limiting per client/IP

❌ DON'T: Hardcode API versions
✅ DO: Use versioned API endpoints
```

### Frontend Related
```
❌ DON'T: Block UI during API calls
✅ DO: Show loading states and handle errors gracefully

❌ DON'T: Store sensitive data in localStorage
✅ DO: Use secure storage mechanisms

❌ DON'T: Skip accessibility considerations
✅ DO: Implement proper ARIA labels and keyboard navigation

❌ DON'T: Make components too tightly coupled
✅ DO: Use proper state management and props drilling
```

### Mobile Related
```
❌ DON'T: Perform network calls on main thread
✅ DO: Use background threads for I/O operations

❌ DON'T: Ignore offline scenarios
✅ DO: Implement proper offline-first architecture

❌ DON'T: Request permissions without explanation
✅ DO: Explain why permissions are needed before requesting

❌ DON'T: Store sensitive data insecurely
✅ DO: Use encrypted storage for sensitive information
```