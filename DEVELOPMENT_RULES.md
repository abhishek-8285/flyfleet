# FlyFleet Development Rules & Best Practices

To maintain a "Premium Aura" experience and ensure system robustness, all development must follow these rules.

## 1. Global State Preference
- **Rule**: User preferences (Theme, Language, Sidebar state) MUST be synchronized with the database when a user is authenticated.
- **Implementation**: Use local storage as a fallback/initial state, but trigger an API update on change and fetch preference on login.
- **Goal**: Cross-browser and cross-device consistency.

## 2. Environment Readiness (Agent Safety)
- **Rule**: Before initiating autonomous verification (e.g., browser testing), ensure the target service is healthy and reachable.
- **Verification**: Use `lsof -i :PORT`, `curl -I`, or `netstat` to confirm the server is listening.
- **Goal**: Prevent infinite loops and "stuck" autonomous agents.

## 3. Unified Context Architecture
- **Rule**: UI components must strictly use global Contexts (`ThemeContext`, etc.) instead of local state for shared properties.
- **Goal**: Single source of truth and predictable UI behavior.

## 4. Multi-Route Verification
- **Rule**: Any global feature change must be verified across at least 2 distinct routes and 1 navigation event.
- **Goal**: Catch persistence race conditions during Next.js route transitions.

## 5. Tailwind v4 Dark Mode
- **Rule**: Use the selector strategy for dark mode (mapping `dark` variant to `.dark` class).
- **Implementation**: Ensure `@variant dark (&:where(.dark, .dark *));` is present in the global CSS.
