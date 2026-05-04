# FlyFleet Agent Operating Protocol (agent.md)

This document defines the strict rules that the AI Agent (Antigravity) MUST follow when modifying the FlyFleet codebase. Failure to adhere to these rules results in technical debt and system instability.

## 1. Database & Schema Integrity
- **Unique Migrations**: Every new migration must have a unique, sequential numeric prefix. Duplicate numbers (e.g., two files starting with `064_`) are FORBIDDEN. Use 'b' suffixes (e.g., `064b_...`) if a slot is already taken.
- **Idempotency**: Every SQL statement must be idempotent. Use `CREATE TABLE IF NOT EXISTS`, `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`, and `DROP TABLE IF EXISTS`.
- **Identity Consistency**: Use `tenants` (table) and `tenant_id` (column). Always create the `organizations` view in `bootstrap_db.sql` for legacy compatibility.
- **Master Setup Protocol**: A fresh setup MUST follow this sequence:
  1. `DROP` and `CREATE` the database.
  2. Run `bootstrap_db.sql` (Foundational tables & views).
  3. Run `go run cmd/migrate/main.go up` (Application logic).
- **No Sync Short-circuiting**: Do not use `main.go` sync functions to "fix" schema issues. Fix the schema through migration files first.

## 2. Code Quality & Consistency
- **Unified Context**: Strictly use global React Contexts (e.g., `ThemeContext`) for shared UI state. Local state overrides for global properties are not allowed.
- **Premium Aura**: All UI changes must adhere to the Tailwind v4 \"Premium Aura\" guidelines (modern typography, harmonious palettes, no default colors).
- **Zero-Placeholder Policy**: Never use placeholders for images or data. Use `generate_image` or actual seed data.

## 3. Verification & Completion
- **Fresh Setup Validation**: A task involving database changes is not complete until it has been verified to work on a clean database setup (drop/recreate).
- **Native Readiness**: All backend changes must be verified using `go build` rather than `go run` to ensure package resolution and workspace permissions are correct.
- **Multi-Route Check**: Global feature changes must be verified across at least 2 distinct routes to ensure persistence and state consistency.

## 4. Communication Protocol
- **Technical Rigor**: Prioritize architectural stability over immediate \"feature unblocking.\" If a task requires a refactor of existing debt, propose the refactor before patching.
- **No Fluff**: Keep responses focused on technical implementation and verification.
