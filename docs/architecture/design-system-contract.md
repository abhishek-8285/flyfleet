# Design System Contract (Web + Mobile)

## Purpose
Provide one token baseline and one interaction-state baseline across web and mobile surfaces.

## Source of Truth
1. Canonical tokens: `packages/common/design/design_tokens.json`
2. Mobile token mirror: `apps/mobile/shared/design/design_tokens.json`
3. Web CSS variables: `apps/dashboard-web/src/styles/design-tokens.css`

## Required Token Families
1. color
2. spacing
3. radius
4. typography
5. state

## Required Interaction States
1. loading
2. error
3. empty

## Accessibility Baseline
1. Color tokens must support contrast checks in implementation.
2. Loading/error/empty states must include descriptive text.
3. Focus and keyboard behavior must remain deterministic across critical flows.

## Enforcement
CI gate:
1. `scripts/check_design_system_foundation.py`
2. `.github/workflows/design-system-foundation.yml`
