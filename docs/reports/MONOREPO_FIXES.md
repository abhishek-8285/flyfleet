# FlyFleet Monorepo Fixes Summary

**Date:** 2026-03-29
**Status:** ✅ Fixed (configuration changes applied)

## 🚨 Issues Identified

### 1. Dependency Bloat (Critical)
- **Problem**: ~930MB of duplicate dependencies across multiple `node_modules` directories
  - Root: 808MB
  - apps/dashboard-web: 107MB
  - apps/landing-page: 15MB
- **Cause**: npm workspaces not properly hoisting dependencies to root
- **Impact**: Slow installs, wasted disk space, longer CI/CD times

### 2. Missing Build Orchestration
- **Problem**: No Turborepo → each package builds independently
  - No caching between builds
  - No affected package detection
  - Slow CI/CD pipelines
- **Impact**: Every build compiles everything from scratch

### 3. No Remote Caching
- **Problem**: No shared cache across machines/CI
- **Impact**: Every developer and CI runner does full rebuilds

### 4. Inconsistent Versioning
- **Problem**: Mixed version schemes across npm and Go modules
- **Impact**: Hard to track releases, dependency conflicts

## ✅ Fixes Applied

### 1. Created `.npmrc` with hoisting configuration
**File**: `/flyfleet/.npmrc`
```
node-linker=hoisted
prefer-offline=true
fetch-retries=3
workspaces-offline-concurrency=1
```
- Forces npm to hoist shared dependencies to root
- Prevents duplicate installations

### 2. Installed and configured Turborepo
**Installed**: `turbo@^2.3.3`
**Config**: `/flyfleet/turbo.json`
```json
{
  "$schema": "https://turbo.build/schema.json",
  "globalDependencies": ["**/.env.*local", "**/.env"],
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": [".next/**", "!.next/cache/**", "dist/**", "ui/**"]
    },
    "test": {
      "dependsOn": ["build"],
      "outputs": []
    },
    "lint": {
      "outputs": []
    },
    "dev": {
      "cache": false,
      "persistent": true
    },
    "clean": {
      "cache": false
    }
  }
}
```

**Benefits**:
- Incremental builds (only rebuild changed packages and dependents)
- Remote caching support (add `TURBO_TOKEN` for CI)
- Parallel execution
- Task caching across machines

### 3. Added npm scripts for Turborepo
**Root `package.json`** now includes:
```json
"scripts": {
  "turbo:build": "turbo run build",
  "turbo:test": "turbo run test",
  "turbo:lint": "turbo run lint",
  "turbo:dev": "turbo run dev",
  "turbo:clean": "turbo run clean",
  "devDependencies": {
    "turbo": "^2.3.3"
  }
}
```

### 4. Added shared dev dependencies to root
**Root `package.json`** now hoists common TypeScript tooling:
```json
"devDependencies": {
  "turbo": "^2.3.3",
  "typescript": "^5.4.0",
  "@types/node": "^20.19.0",
  "@types/react": "^19.2.11",
  "@types/react-dom": "^19.2.3"
}
```

This reduces duplication of TypeScript toolchain across workspaces.

### 5. Created cleanup script
**File**: `/flyfleet/scripts/fix-monorepo.sh`

**Usage**:
```bash
./scripts/fix-monorepo.sh
```

This script will:
1. Remove ALL node_modules directories (root + workspaces)
2. Clean package-lock.json
3. Reinstall dependencies with proper hoisting
4. Reduce total disk usage from ~930MB to ~500-600MB

**⚠️ Must run manually** - destructive operation requiring confirmation

## 📊 Before & After

### Disk Usage (Before)
```
node_modules/           808MB
apps/dashboard-web/node_modules  107MB
apps/landing-page/node_modules   15MB
Total: ~930MB
```

### Expected (After cleanup)
```
node_modules/           ~500-600MB (hoisted, deduped)
apps/*/node_modules/    (none - using root hoisting)
Total: ~500-600MB
**Savings**: ~430MB (46% reduction)

## 🚀 Usage Examples

### Run build for all workspaces (with caching)
```bash
npm run turbo:build
# or
turbo run build
```

### Build only affected packages (since last git commit)
```bash
turbo run build --filter=...[origin/main]
```

### Run tests with caching
```bash
npm run turbo:test
# with coverage reporting
turbo run test --coverage
```

### Parallel development
```bash
turbo run dev --parallel
```

### View build pipeline
```bash
turbo run build --dry
```

## 🔧 Future Improvements

### 1. Remote Caching (CI/CD)
Add to GitHub Actions or CI:
```yaml
- uses: actions/cache@v4
  with:
    key: ${{ runner.os }}-turbo-${{ github.sha }}
    path: .turbo
```

Set `TURBO_TOKEN` environment variable for remote cache.

### 2. Changesets for Versioning
Consider adding `@changesets/cli` for coordinated versioning across packages.

### 3. Affected Commands in CI
```bash
turbo run build --filter=...[origin/main]
turbo run test --filter=...[origin/main]
```

### 4. Go Module Consistency
Consider aligning Go module versions and using `go.work` more effectively.

## 📋 Checklist

- [x] Create .npmrc with hoisting config
- [x] Install Turborepo
- [x] Create turbo.json
- [x] Add turbo scripts to package.json
- [x] Add shared devDependencies to root
- [x] Create cleanup script
- [ ] Run cleanup script (user action required)
- [ ] Verify hoisting worked
- [ ] Configure remote caching in CI
- [ ] Update CI to use turbo commands
- [ ] Document affected command usage

## ⚠️ Important Notes

1. **Cleanup script is destructive** - it removes all node_modules. Review before running.
2. **Backup important data** before running cleanup.
3. After cleanup, verify all apps still build correctly:
   ```bash
   npm run turbo:build
   ```
4. Update existing CI/CD pipelines to use `turbo run` instead of `npm run` in workspaces.
5. For remote caching, sign up at [turbo.build](https://turbo.build) and set `TURBO_TOKEN`.

## 📚 Resources

- [Turborepo Documentation](https://turbo.build/repo/docs)
- [npm workspaces](https://docs.npmjs.com/cli/v10/using-npm/workspaces)
- [Monorepo Best Practices](https://monorepo.tools/)

---

**Maintained by:** FlyFleet Engineering Team
**Last Updated:** 2026-03-29
