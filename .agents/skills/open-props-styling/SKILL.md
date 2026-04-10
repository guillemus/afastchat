---
name: open-props-styling
description:
    Use this skill whenever the user asks for UI styling, CSS tweaks, layout polish,
    spacing/color/typography changes, or new visual components in afastchat.com. Prefer the repo's
    existing Open Props setup in public/styles.css instead of ad hoc values, Tailwind-style
    thinking, or adding another styling system. Reach for raw Open Props tokens directly for
    spacing, colors, radii, shadows, typography, and motion.
license: MIT
metadata:
    author: OpenAI
    version: '0.1.0'
---

# Open Props styling for afastchat.com

Use this skill to keep UI work consistent with the repo's current CSS approach.

## Start here

1. Read `public/styles.css` before editing styles.
2. Assume Open Props should come from:

```css
@import 'https://unpkg.com/open-props@1.7.23/open-props.min.css';
```

3. Keep styling in CSS unless the user explicitly wants inline styles or a different system.

## Default approach

### Prefer raw Open Props tokens

Reach for Open Props values directly in component styles:

- spacing: `--size-*`, `--size-fluid-*`
- colors: `--gray-*`, `--stone-*`, `--red-*`, `--orange-*`, `--green-*`, `--blue-*`
- radius: `--radius-*`
- shadows: `--shadow-*`
- font sizes: `--font-size-*`
- font weights: `--font-weight-*`
- motion: `--ease-*`, `--animation-*`

Good:

```css
.card {
    padding: var(--size-4);
    border-radius: var(--radius-3);
    background: var(--gray-11);
    box-shadow: var(--shadow-2);
}
```

Avoid one-off hardcoded values when an Open Props token is already close enough.

### Reuse existing semantic vars when they already exist

If `public/styles.css` already defines app-level variables like `--bg-primary` or `--text-muted`,
keep using them where that preserves meaning.

Only add new semantic variables when a value is repeated, domain-specific, or clearly improves
readability. Do not create aliases for one-off values.

### Match the repo's current structure

Fit changes into the existing stylesheet shape when possible:

- `@layer reset`
- `@layer tokens`
- `@layer layout`
- `@layer components`

Extend the nearest existing section instead of scattering styles across new files unless the user
asks for a refactor.

## Styling heuristics

- Prefer balanced spacing over dense layouts.
- Use subtle radius and shadow changes before large visual effects.
- Favor neutral Open Props colors for surfaces and text; use accent colors intentionally.
- Keep contrast readable, especially on dark surfaces.
- Prefer `font-size` and spacing scales from Open Props over custom pixel tuning.
- If a component needs hover/focus states, make them visible but restrained.

## Avoid

- Adding Tailwind, CSS frameworks, or another token library just for styling work.
- Mixing many hardcoded hex values with Open Props tokens.
- Replacing meaningful semantic app variables with raw tokens everywhere.
- Creating large design-system abstractions for a small local tweak.

## Working pattern

When asked to style or polish UI:

1. Inspect nearby markup and current CSS conventions.
2. Reuse existing semantic vars first.
3. Fill gaps with raw Open Props tokens.
4. Keep the patch small and consistent with surrounding code.
5. If you introduce a new repeated visual concept, add a semantic variable in `:root` backed by an
   Open Props token.

## Examples

**Example: button polish**

Prefer:

```css
.button {
    padding: var(--size-2) var(--size-3);
    border-radius: var(--radius-2);
    font-size: var(--font-size-0);
    font-weight: var(--font-weight-5);
    transition: filter 0.15s;
}
```

Over:

```css
.button {
    padding: 7px 13px;
    border-radius: 5px;
    font-size: 13px;
    font-weight: 600;
}
```

**Example: repeated app color**

If an accent appears in multiple places, this is reasonable:

```css
:root {
    --accent-brand: var(--orange-7);
}
```

Then use `var(--accent-brand)` in components.
