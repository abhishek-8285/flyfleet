# Dead Code Audit & Cleanup

## Why we need this audit

The `fleet-core` module has grown to ~30 domain packages under `internal/domains/`,
but not every package is actively wired into the main binary. Dead/empty domains
cause real harm:

1. **Build time bloat** — Go compiles every `.go` file in a module, even unused
   packages. ~50K lines of dead code adds seconds to every build/test cycle.
2. **Cognitive noise** — New contributors can't tell which domains are real
   vs. abandoned, leading to wasted time and accidental dependencies.
3. **Maintenance debt** — Dead code never gets updated, so it bitrots and
   eventually fails to compile against newer versions of dependencies.

## What was found

Scanned all `internal/domains/*` directories for:
- **No `.go` files at all** — empty directory skeletons
- **Never imported anywhere** — no other package (including `main.go` or
  sibling domains) imports the package, checked via `grep -rl "domains/<name>/" .`

### Removed (confirmed dead)

| Domain | Files | Lines | Reason |
|--------|-------|-------|--------|
| `auth_deprecated` | 0 | 0 | Empty skeleton |
| `auth` | 15 | 5,405 | Never imported; auth handled by `iam` module |
| `internal/domains/tracking` | 35 | 8,140 | Never imported by `main.go`; active tracking lives in `internal/tracking/` |

Total removed: **~13,545 lines**

### Flagged (pending decision)

| Domain | Files | Lines | Notes |
|--------|-------|-------|-------|
| `document` | 0 | 0 | Empty directory (no .go files) |
| `alert` | 17 | 4,592 | Never imported |
| `analytics` | 37 | 8,702 | Never imported |
| `compliance` | 38 | 3,375 | Never imported |
| `diagnostics` | 48 | 2,802 | Never imported |
| `geofencing` | 12 | 2,774 | Never imported (note: active geofencing is in `internal/tracking/`) |
| `governance` | 26 | 3,134 | Never imported |
| `integrations` | 13 | 1,420 | Never imported |
| `maintenance` | 39 | 7,426 | Never imported |
| `platform_compat` | 7 | 1,372 | Never imported |
| `reports` | 11 | 1,473 | Never imported |
| `settings` | 19 | 4,264 | Never imported |
| `video` | 23 | 6,011 | Never imported |

Total flagged: **~47,345 lines**

## Why these domains are dead

The 12 flagged domains are orphaned because of an **incomplete code modernization**.
Investigation shows:

1. **Written with old patterns** — These domains use old-style HTTP handlers
   (`json.NewEncoder(w).Encode(...)`, `http.Error(...)`) instead of the current
   `platform/response` standard (`response.Success200`, `response.Error400`, etc.).

2. **A migration script was created** (`scripts/migrate_all_remaining.sh`) to
   automatically convert old handlers to the new response pattern. The script
   lists 16 handler files across the dead domains and was intended to be run
   on a local Termux environment (`cd /data/data/com.termux/files/home/flyfleet/...`).

3. **Migration was never completed** — The script exists in the repo but appears
   to reference a local-only filesystem path and was never committed/run on the
   main branch. The domains were never wired into `main.go`'s Chi router.

4. **Superseded by refactoring** — During the same period, parallel refactoring
   happened in the codebase:
   - `auth` → replaced by `iam` module (commit `135e72c` removed old auth)
   - `tracking` domain → replaced by top-level `internal/tracking/`
   - `dispatch` → rebuilt as real orchestrator (commit `5beca08`, `aacf03d`)
   - These 12 domains were caught in the gap — written before the modernization
     but never migrated to the new patterns and never wired in.

5. **No git stash or branch** contains completed migration work for these domains.
   Git stashes only cover auth/IAM hardening and UI changes.

### Verification details

Each domain was checked for imports using `grep -rl "domains/<name>/" . --include='*.go'`:

- **0 production Go imports** for all 12 domains
- **`integrations`**: 1 reference in `tests/fuel_integration_test.go` (build-tagged `//go:build integration`, not compiled in normal builds)
- **`alert`/`alerts`**: internal cross-references only (tests within the domain reference each other)
- **Remaining 10 domains**: only references found in `air.log` (dev server watch log, not code)

**The domains still compile and pass tests** — they're functional but orphaned, not bitrotting yet.

**Migration scripts** (`scripts/migrate_all_remaining.sh`, `scripts/migrate_remaining.sh`) are **untracked** local-only files (confirmed via `git ls-files`) pointing to Termux paths — never committed to git, never run on the main branch.

**Root cause: work-in-progress interrupted

These domains represent features that were **spec'd and partially built** but
abandoned mid-refactor when the codebase shifted to the new hexagonal architecture
with `platform/response`, the `iam` module, and the active tracking service.

### Actively used (no changes)

`admin_support`, `assignment`, `billing`, `customer`, `dispatch`, `driver`,
`emergency`, `finance`, `fuel`, `india`, `invoicing`, `iam`, `route`,
`trip`, `vehicle` — these are imported by `main.go` or by other active domains.

## How to verify

```bash
cd services/fleet-core
# Check if a domain is imported anywhere outside itself:
grep -rl "domains/<name>/" . --include='*.go' | grep -v "internal/domains/<name>/"

# Build the entire service:
go build -o /dev/null ./cmd/

# Run dispatch tests:
go test ./internal/domains/dispatch/... -count=1
```
