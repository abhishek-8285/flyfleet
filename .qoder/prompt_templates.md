# AI Assistant Prompt Templates for Fleet Platform

These templates help ensure consistent and complete implementation when working with AI assistants.

## Template 1: Add New API Endpoint

```
Implement a new API endpoint for [FEATURE] in the fleet management system.

Requirements:
- HTTP Method: [GET/POST/PUT/DELETE]
- Path: /api/v1/[ENDPOINT]
- Authentication: [JWT/Device API Key]
- Authorization: [Role/Permission required]

Request Body:
```json
[PAYLOAD_SCHEMA]
```

Response:
```json
[RESPONSE_SCHEMA]
```

Database Changes Needed:
- [ ] New table? [TABLE_NAME]
- [ ] New columns? [COLUMN_DETAILS]
- [ ] New indexes? [INDEX_DETAILS]

Business Logic:
- [Describe what this endpoint should do]

Generate:
1. Database migration file
2. Entity struct
3. Repository methods with SQL
4. Service methods with business logic
5. HTTP handler with validation
6. Unit tests for service
7. Integration tests for handler
8. Update OpenAPI spec
```

## Template 2: Add New React Component

```
Create a React component for [FEATURE] in the web admin panel.

Requirements:
- Location: web-admin/src/components/[CATEGORY]/
- Props Interface: [DESCRIBE_PROPS]
- State Management: [Redux/Local State/React Query]
- Data Source: [API_ENDPOINT]

Features:
- [List UI features needed]
- [List interactions needed]
- [List error states to handle]

Styling:
- Use Tailwind CSS classes
- Follow existing design system
- Ensure responsive design

Generate:
1. TypeScript interfaces (types.ts)
2. Main component with logic
3. Custom hooks if needed (hooks.ts)
4. Storybook story for testing
5. Unit tests with React Testing Library
```

## Template 3: Add New Android Feature

```
Implement [FEATURE] in the Android driver app.

Requirements:
- Screen/Feature: [DESCRIBE]
- Architecture: MVVM with Hilt
- Data Layer: Room + Retrofit
- UI: Jetpack Compose

User Flow:
1. [Step 1]
2. [Step 2]
3. [Step 3]

Offline Support:
- [Describe what should work offline]
- [Describe sync strategy]

Permissions Needed:
- [List Android permissions]

Generate:
1. Data models (Entity, DTO)
2. Room Entity and DAO
3. Repository with sync logic
4. ViewModel with StateFlow
5. Compose screen with previews
6. Navigation setup
7. Instrumentation tests
```

## Template 4: Database Migration

```
Create a database migration for [CHANGE_TYPE] in the fleet management system.

Requirements:
- Migration Type: [CREATE/ALTER/DROP]
- Table: [TABLE_NAME]
- Purpose: [DESCRIBE_PURPOSE]

Changes:
- [Column additions/removals]
- [Index creations]
- [Constraint modifications]
- [Data transformations needed]

Performance Considerations:
- [Impact on query performance]
- [Indexing strategy]
- [Partitioning needs]

Generate:
1. Migration SQL file with rollback
2. Updated entity struct
3. Repository method updates
4. Service layer adjustments
5. API handler modifications
6. Migration test cases
7. Documentation updates
```

## Template 5: Performance Optimization

```
Optimize [COMPONENT/SERVICE] for better performance in the fleet platform.

Current Issues:
- [Describe current performance problems]
- [Metrics showing the issue]
- [User impact]

Optimization Goals:
- Response time: [TARGET_TIME]
- Throughput: [TARGET_RATE]
- Resource usage: [TARGET_USAGE]

Approaches to Consider:
- [Query optimization]
- [Caching strategy]
- [Algorithm improvement]
- [Resource pooling]
- [Asynchronous processing]

Generate:
1. Performance baseline measurements
2. Optimized implementation
3. Benchmark tests
4. Monitoring setup
5. Rollback strategy
6. Deployment plan
```

## Template 6: Security Enhancement

```
Implement security enhancement for [FEATURE/COMPONENT] in the fleet platform.

Security Requirements:
- [Authentication method]
- [Authorization scope]
- [Data protection needs]
- [Compliance requirements]

Threat Model:
- [Potential attack vectors]
- [Sensitive data handling]
- [Access control requirements]

Implementation:
- [Security controls to implement]
- [Validation and sanitization]
- [Encryption requirements]
- [Audit logging needs]

Generate:
1. Security analysis document
2. Implementation with security controls
3. Security tests
4. Penetration testing plan
5. Compliance documentation
6. Security monitoring setup
```