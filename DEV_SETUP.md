# FlyFleet Development Setup & Troubleshooting Guide

This guide outlines the critical configuration and troubleshooting steps to maintain a stable development environment for the FlyFleet platform.

## 1. Backend: Fleet-Core Service

### Port Configuration
The `fleet-core` service is configured to run on port **8120**. 
- **Why?** Port 8109 is reserved/locked by the system's `nsjail` sandbox and cannot be bound by the application.
- **Auto-Cleanup:** On startup, the service automatically attempts to kill any orphaned processes on port 8120 using `fuser -k 8120/tcp`.

### Database Seeding
The database is automatically seeded on startup if no vehicle data is detected.
- **Forced Re-seed:** To force a clean re-seed, the service will truncate core tables if it detects existing but inconsistent data.
- **Seed Logic:** The service executes `migrations/seed_complete_dev_data.sql` as a single atomic block. 
- **Seeded Credentials:**
  - **Email:** `admin@delta.io`
  - **Password:** `Password123!`

### Critical Environment Variables
Ensure these are set in your `.env` file:
```bash
PORT=8120
JWT_SECRET=your_super_secret_jwt_key_here
VEHICLE_FIELD_ENCRYPTION_KEY=32_byte_hex_key_for_dev_only
DB_URL=postgres://user:pass@localhost:5432/flyfleet?sslmode=disable
```

## 2. Frontend: Dashboard Web

### Proxy Configuration
The Vite development server is configured to proxy API requests to the backend on port **8120**.
- **Config File:** `apps/dashboard-web/vite.config.ts`
- **Proxy Targets:** `/api`, `/auth`, `/ws`, `/api/v1/ws` are all mapped to `http://localhost:8120`.

### Content Security Policy (CSP)
The CSP in `vite.config.ts` includes `connect-src` for `http://localhost:8120`. If you change the backend port, you **must** update the CSP as well, or the browser will block API calls.

## 3. Common Troubleshooting

### "Address already in use" (Port 8120)
If the service fails to start despite the auto-cleanup:
1. Manually kill the process: `sudo fuser -k 8120/tcp`
2. Check for orphaned `air` or `go run` processes: `ps aux | grep fleet-core`

### Data Not Showing in Dashboard
1. Check the `fleet-core` logs (or `air.log`) for seeding success:
   - Look for: `✨ Dev data seeded successfully!`
2. Verify the browser console for CSP or Proxy errors.
3. Ensure the backend is actually listening on 8120: `curl http://localhost:8120/health` (if available) or check `netstat -tuln | grep 8120`.

### Database Migration Failures
The app uses a "Global Schema Harmonizer" on startup to fix common naming discrepancies (e.g., `org_id` vs `tenant_id`). If migrations fail:
1. Check `migrations/seed_complete_dev_data.sql` for constraint violations.
2. The harmonizer logs will show which tables were updated.

---
*Last Updated: May 2026*
