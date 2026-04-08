# FlyFleet Cinematic Design System

## Overview

The FlyFleet landing page now features a 2.5D immersive experience using GSAP ScrollTrigger, depth-parallax, and scroll-driven animations. This document describes the system architecture and how to extend it.

---

## Core Technologies

- **GSAP 3.12.5** + ScrollTrigger — scroll animations, pinning, scrubbing
- **Framer Motion** — hover states, initial entry animations
- **Lenis** — buttery smooth scrolling (duration: 1.2s)
- **Tailwind CSS v4** — depth classes, utilities
- **CSS Custom Properties** — `--scene-height` for parallax distance

---

## Depth System (6 Layers)

Every scene element must have a depth class:

| Depth | Layer | Parallax | Blur | Scale | Use |
|-------|-------|----------|------|-------|-----|
| `depth-0` | Far background | 0.10x | 8px | 0.70 | Gradients, textures |
| `depth-1` | Atmosphere | 0.25x | 4px | 0.85 | Glow blobs, light |
| `depth-2` | Mid decorations | 0.50x | 0 | 1.00 | Floating icons, companions |
| `depth-3` | Main product | 0.80x | 0 | 1.05 | Hero truck, app screenshots |
| `depth-4` | Text / UI | 1.00x | 0 | 1.00 | Headlines, buttons |
| `depth-5` | Foreground FX | 1.20x | 0 | 1.10 | Sparkles, particles |

**Apply:** `<div class="depth-3" data-depth="3">...</div>`

---

## Scene Architecture

Every scrollable cinematic section follows this pattern:

```tsx
<section
  ref={sectionRef}
  className="scene relative"
  style={{ '--scene-height': '300vh' } as React.CSSProperties}
  data-scene="name"
>
  {/* Depth layers */}
  <div className="depth-0 layer">Background</div>
  <div className="depth-1 layer">Glow</div>
  <div className="depth-3 layer">Product</div>
  <div className="depth-4 layer">Text</div>
</section>
```

**Scene heights:**
- Hero: 200vh
- Simple sections: 150vh (auto)
- Pinned sections: 300vh
- Epic sequences: 400vh

---

## Common Animation Patterns

### 1. Parallax on Scene Entry/Exit

```tsx
useEffect(() => {
  const gsap = (window as any).gsap;
  if (!gsap) return;

  const ctx = gsap.context(() => {
    gsap.utils.toArray('.depth-0').forEach((layer: Element) => {
      gsap.to(layer, {
        y: -150,
        ease: 'none',
        scrollTrigger: {
          trigger: sectionRef.current,
          start: 'top bottom',
          end: 'bottom top',
          scrub: 1
        }
      });
    });
  }, sectionRef);

  return () => ctx.revert();
}, []);
```

### 2. Pinned Scroll Section

```tsx
// Pin a container while content scrolls
if (ScrollTrigger) {
  ScrollTrigger.create({
    trigger: sectionRef.current,
    start: 'top top',
    end: 'bottom top',
    pin: mapRef.current,
    pinSpacing: true,
    scrub: 1
  });
}
```

### 3. Scrub-Driven Animation

Animate properties based on scroll progress (0-100%):

```tsx
gsap.to('#efficiency-bar', {
  width: '94%',
  ease: 'none',
  scrollTrigger: {
    trigger: sectionRef.current,
    start: 'top 30%',
    end: 'bottom 40%',
    scrub: 1  // smooth 1s lag
  }
});
```

### 4. Count-Up Numbers

```tsx
const countEl = document.getElementById('vehicle-count');
if (countEl) {
  ScrollTrigger.create({
    trigger: sectionRef.current,
    start: 'top 50%',
    end: 'bottom 30%',
    onEnter: () => {
      gsap.to(countEl, {
        innerText: 128,
        duration: 2,
        snap: { innerText: 1 },
        ease: 'power1.out'
      });
    },
    onLeaveBack: () => gsap.set(countEl, { innerText: 0 })
  });
}
```

### 5. Clip-Path Birth Reveal

Reveal element from bottom (rising curtain):

```tsx
gsap.fromTo(element,
  { clipPath: 'inset(0 0 100% 0)' },
  {
    clipPath: 'inset(0)',
    ease: 'power3.inOut',
    scrollTrigger: {
      trigger: sectionRef.current,
      start: 'top 30%',
      end: 'bottom 60%',
      scrub: 1
    }
  }
);
```

---

## Component Best Practices

### 1. Hero Section

- Use **6 depth layers** (0-5)
- Hero product (depth-3) gets **elastic entrance** + **float-loop**
- Headline: **split-word** animation (words fly from sides)
- Companion items (depth-2): smaller, orbit hero
- On scroll exit: hero scales up, text fades out

**Example:** See `Hero.tsx`

### 2. Pinned Map/Command Center

- Section height: 300vh
- Map container pinned, text overlays scroll over
- Text moves upward with opacity fade
- Data bars scrub to fill
- Counters animate in
- Keep interactive elements (buttons) visible on top layer

**Example:** See `InteractiveIntelligence.tsx`

### 3. Floating Product Showcase

- Vertical stack layout (400vh)
- Each product has clip-path reveal
- Staggered timing: product1: 0-40%, product2: 40-80%
- Add depth-2 floating companions (SOS button, location pin)
- Exit: products scatter outward (x translation)

**Example:** See `AppShowcase.tsx`

### 4. Card Grid with Depth

- Cards = depth-2
- Parallax: y offset based on index
- Subtle floating loop: yoyo sine wave
- Hover: scale + shadow elevation
- Entrance: cascade with rotation

**Example:** See `Features.tsx`

---

## Accessibility

### Reduced Motion

All cinematic effects disabled when user prefers reduced motion:

```css
@media (prefers-reduced-motion: reduce) {
  * { animation-duration: 0.01ms !important; }
  .scene { height: auto !important; }
  .depth-* { transform: none !important; filter: none !important; }
  .clip-reveal { clip-path: inset(0) !important; }
}
```

### Mobile Touch

Reduce parallax intensity:

```css
@media (pointer: coarse) {
  .depth-0 { filter: blur(4px); transform: scale(0.85); }
  .depth-1 { filter: blur(2px); transform: scale(0.92); }
  .depth-2, .depth-3, .depth-4, .depth-5 { transform: none; }
}
```

---

## Performance Rules

1. **Animate only GPU properties:** `transform`, `opacity`, `clip-path`, `filter`
2. **Limit simultaneous animations:** < 40 active elements per frame
3. **Use `will-change`** on parallax elements (handled by `.depth-*` classes)
4. **Content visibility:** `.scene > * { content-visibility: auto; }`
5. **Cleanup:** `ctx.revert()` in `useEffect` return
6. **Check `if (gsap)`** before running animations

---

## Utility Classes

| Class | Purpose |
|-------|---------|
| `.scene` | Parallax container (200vh default) |
| `.layer` | Flex centering for depth content |
| `.depth-0` through `.depth-5` | Parallax multipliers + blur/scale |
| `.float-loop` | Persistent up/down float animation |
| `.clip-reveal` | Clip-path birth (use with GSAP) |
| `.glow-effect` | Radial gradient background glow |
| `.mobile-frame` | Pre-styled phone mockup frame |

---

## Adding New Cinematic Sections

1. **Import section in `page.tsx`** at appropriate position
2. **Create `<section className="scene" data-scene="...">`**
3. **Add 3+ depth layers** (min requirement)
4. **Choose scene height** (150-400vh based on complexity)
5. **Write `useEffect`** with GSAP context for scroll triggers
6. **Assign depth classes** to all animated elements
7. **Test reduced motion** (browser devtools → rendering → emulate CSS media)
8. **TypeScript:** `npx tsc --noEmit`, fix implicit any with `: Element` type

---

## Existing Cinematic Sections

| Section | File | Height | Pinned? | Key Effects |
|---------|------|--------|---------|-------------|
| Hero | `Hero.tsx` | 200vh | No | 6-layer parallax, split-text, float, scatter exit |
| Live Intelligence | `InteractiveIntelligence.tsx` | 300vh | Yes | Scrubbed bar fill, counter count-up, text slide |
| Field-Ready Apps | `AppShowcase.tsx` | 400vh | No | Clip-path reveal, companion floats, scatter exit |
| Complete Fleet Control | `Features.tsx` | 200vh | No | Parallax card stack, float, cascade entrance |
| ROI Calculator | `ROICalculator.tsx` | 300vh | Yes | Pinned results, scroll-driven number animation, slider interaction |
| Transparent Pricing | `Pricing.tsx` | 250vh | No | Floating pricing cards, parallax, hover elevation |
| Final CTA | `CTA.tsx` | 150vh | No | Depth layers, glow blobs, magnetic buttons, scroll entrance |

---

## Debugging Tips

**Parallax not working?**
- Confirm `.depth-*` classes applied
- Check `--scene-height` style present
- Verify GSAP loaded (`window.gsap` in console)
- Ensure ScrollTrigger registered

**Animations too fast/slow:**
- Adjust `scrub: 1` value (seconds of lag)
- Increase `--scene-height` for longer scroll distance

**Elements jitter on scroll:**
- Ensure `will-change` not overused (only on active layers)
- Check no layout shifts (all sizes fixed)

**Mobile issues:**
- Use `@media (pointer: coarse)` to reduce effects
- Test on actual device (emulator not perfect)

---

## Future Enhancements

- **Replace placeholder truck icon** with transparent PNG assets
- **Replace CSS mobile frames** with actual app screenshots
- **Add skew transform** to cards for "flying in" effect
- **Implement cursor microinteractions** (particle trails)
- **Add scroll progress indicator** in navbar
- **Add exit animations** for other sections (Pricing, CTA)

---

**Maintainer:** FlyFleet Design Team
**Last Updated:** 2026-03-29
**GSAP Version:** 3.12.5
