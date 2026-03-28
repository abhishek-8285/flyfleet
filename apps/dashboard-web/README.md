# FlyFleet Dashboard Web

React-based administrative dashboard for FlyFleet, built with Vite, TypeScript, and Tailwind CSS.

## Tech Stack

- **Vite 8** - Fast build tool and dev server
- **React 19** - UI library
- **TypeScript** - Type safety
- **Tailwind CSS** - Utility-first CSS framework
- **React Query** - Data fetching and caching
- **Redux Toolkit** - State management
- **Vitest** - Unit testing (migrated from Jest)

## Prerequisites

- Node.js 18+ (Node 20+ recommended)
- npm 9+

## Development

### Install dependencies

From the monorepo root:

```bash
npm install
```

### Run development server

```bash
npm run dev --workspace=@flyfleet/dashboard-web
# or
cd apps/dashboard-web
npm run dev
```

The dev server will start at http://localhost:3000 with API proxy to `http://localhost:8084`.

### Available scripts

| Script | Description |
|--------|-------------|
| `npm run dev` | Start Vite dev server with HMR |
| `npm run build` | Type-check and build for production (outputs to `dist/`) |
| `npm run preview` | Preview production build locally |
| `npm test` | Run unit tests once (Vitest) |
| `npm run test:watch` | Run tests in watch mode |
| `npm run lint` | Run ESLint |

## Testing

The project uses Vitest with React Testing Library. Tests are located alongside source files:

- `src/**/*.test.{ts,tsx}` - Unit tests
- `src/**/*.spec.{ts,tsx}` - Specification tests

Test setup file: `src/setupTests.ts` (includes `@testing-library/jest-dom`)

## Building for Production

```bash
npm run build --workspace=@flyfleet/dashboard-web
```

Outputs optimized static files to `apps/dashboard-web/dist/`. These can be served by any static file server or deployed to CDN/S3.

### Environment Variables

- `VITE_API_URL` - Backend API base URL (defaults to `/api/v1`)
- `REACT_APP_PUBLIC_POSTHOG_KEY` - PostHog analytics key
- `REACT_APP_PUBLIC_POSTHOG_HOST` - PostHog host (defaults to `https://us.i.posthog.com`)

At runtime, these are injected via Vite’s `define` config. For production, set them in your hosting environment.

## Migrated from Create React App

This project was originally created with Create React App and has been migrated to Vite. The migration included:

- Replaced `react-scripts` with Vite + `@vitejs/plugin-react`
- Updated test runner from Jest to Vitest
- Preserved all existing functionality and tests
- Maintained ESLint configuration via `eslint-config-react-app`

## Notes

- The `eslintConfig` still extends `react-app` and `react-app/jest` for compatibility. Consider migrating to standalone ESLint config in future.
- Some legacy tests may need updates for full Vitest compatibility.
- CSS modules and asset imports work via Vite’s built-in support.

---

**Last Updated:** 2026-03-28
**Maintainer:** FlyFleet Frontend Team
