# Landing Page Refactoring Plan

## Current State

**File:** `apps/dashboard-web/src/pages/Landing.tsx`  
**Size:** 1,073 lines  
**Issues:**
- Large monolithic component (1,073 lines)
- 100+ magic numbers for styling
- Mixed concerns (data fetching, UI rendering, state management)
- No unit tests
- Complex rendering logic with inline conditionals

## Target Architecture

### Component Breakdown

```
src/components/landing/
├── TopBar.tsx              # Top info bar (sales, security badge)
├── Header.tsx              # Navigation header
├── Hero.tsx                # Hero section with CTA buttons
├── TrustedBy.tsx           # Brand logos section
├── SuccessStories.tsx      # Customer success cards
├── Features.tsx            # Platform features grid
├── Apps.tsx                # Mobile apps showcase
├── Pricing.tsx             # Pricing tiers with toggle
├── Security.tsx            # Security features
├── HardwareAddons.tsx      # Hardware products
├── FAQ.tsx                 # FAQ accordion
├── Timeline.tsx            # Company roadmap
├── CTA.tsx                 # Bottom CTA section
├── Footer.tsx              # Footer links
└── index.ts                # Barrel exports

src/hooks/
└── useLandingData.ts       # Custom hook for data fetching
```

## Refactoring Steps (TDD Approach)

### Phase 1: Setup & Tests (RED)
1. Create test infrastructure for landing components
2. Write tests for each planned component
3. Define interfaces/types for props

### Phase 2: Extract Components (GREEN)
1. Extract `TopBar` component
2. Extract `Header` component
3. Extract `Hero` component
4. Extract section components (TrustedBy, SuccessStories, etc.)
5. Extract `Footer` component

### Phase 3: Create Custom Hook
1. Create `useLandingData` hook
2. Move data fetching logic from component
3. Add error handling and loading states

### Phase 4: Extract Constants
1. Create `landingConstants.ts` with styled values
2. Replace magic numbers with named constants
3. Define theme-aware values

### Phase 5: Refactor Main Component
1. Simplify Landing.tsx to orchestrator component
2. Use extracted components
3. Target: <200 lines

## Acceptance Criteria

- [ ] All existing tests pass (E2E tests)
- [ ] New unit tests achieve 80%+ coverage
- [ ] Landing.tsx < 200 lines
- [ ] Each component < 150 lines
- [ ] Zero magic numbers (all named constants)
- [ ] TypeScript types for all props
- [ ] No console.log in production code
- [ ] Visual regression tests pass

## Constants to Extract

```typescript
// Layout constants
const LANDING_MAX_WIDTH = '1440px';
const LANDING_CONTENT_MAX_WIDTH = '1280px';
const LANDING_PADDING_X = 'px-4 md:px-10';

// Typography
const HERO_TITLE_SIZE = 'text-4xl md:text-6xl lg:text-7xl';
const SECTION_TITLE_SIZE = 'text-3xl md:text-4xl';

// Spacing
const SECTION_PADDING_Y = 'py-20';
const HERO_PADDING_Y = 'py-16 md:py-24';

// Colors (semantic)
const PRIMARY_SHADOW = 'shadow-primary/25';
const SURFACE_DARK = '#111a22';
```

## Testing Strategy

### Unit Tests (Vitest + React Testing Library)
- Component rendering
- Props validation
- User interactions (theme toggle, pricing toggle)
- Conditional rendering

### Integration Tests
- Data fetching flow
- Error states
- Loading states

### E2E Tests (Playwright)
- Visual regression
- Full page load
- Navigation links

## Rollback Plan

If issues arise:
1. Keep original Landing.tsx as `Landing.old.tsx`
2. Test new implementation in isolation
3. Swap only after full validation
