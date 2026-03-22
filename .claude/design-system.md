# Pebblr Design System

Extracted from `web/src/styles/global.css` (Tailwind v4 theme). This is the reference for all UI work.

## Color Tokens

All colors are defined as Tailwind theme variables (`@theme`) and are available as Tailwind utility classes.

### Brand / Primary

| Token | Value | Usage |
|---|---|---|
| `primary` | `#00355f` | Primary brand color — headings, active nav, icons |
| `primary-container` | `#0f4c81` | Hover states, gradients |
| `primary-fixed` | `#d2e4ff` | Light background tints, badges |

### Secondary

| Token | Value | Usage |
|---|---|---|
| `secondary` | `#515f74` | Subdued text, dividers |
| `secondary-container` | `#d5e3fc` | Light secondary backgrounds |

### Tertiary (Success / Active)

| Token | Value | Usage |
|---|---|---|
| `tertiary` | `#003c27` | Dark green |
| `tertiary-container` | `#005539` | Success text, positive metrics |
| `tertiary-fixed` | `#6ffbbe` | Online status indicator |
| `tertiary-fixed-dim` | `#4edea3` | Dimmed success accent |

### Surface Hierarchy

Surfaces nest from lowest (white) to highest (darkest tint):

| Token | Value | Usage |
|---|---|---|
| `surface` | `#f7f9fb` | Page background |
| `surface-container-lowest` | `#ffffff` | Cards, panels |
| `surface-container-low` | `#f2f4f6` | Section backgrounds |
| `surface-container-high` | `#e6e8ea` | Toolbar buttons |
| `surface-container-highest` | `#e0e3e5` | Deepest container tint |

### Text

| Token | Value | Usage |
|---|---|---|
| `on-surface` | `#191c1e` | Primary body text |
| `on-surface-variant` | `#42474f` | Secondary / muted text |

### Semantic

| Token | Value | Usage |
|---|---|---|
| `error` | `#ba1a1a` | Error states, negative metrics |
| `error-container` | `#ffdad6` | Error backgrounds |

---

## Typography

### Fonts

- **Headline:** `Manrope` (weights 400–800) — use via `font-headline` class
- **Body:** `Inter` (weights 300–600) — default for all body text

Fonts are loaded from Google Fonts in `global.css`.

### Scale Conventions

- `text-3xl font-extrabold font-headline` — page metric values
- `text-[10px] font-bold uppercase tracking-wider` — labels / column headers
- `text-sm font-medium font-headline` — nav items
- `text-xs text-on-surface-variant` — secondary metadata

---

## Spacing Patterns

- **Page padding:** `p-8` (2rem all sides)
- **Card padding:** `p-6` (1.5rem)
- **Section gaps:** `space-y-8` between top-level sections
- **Card corners:** `rounded-xl` (0.75rem)
- **Shadow:** `shadow-sm` for cards, `shadow-md` on hover

---

## Component Patterns

### Cards

```tsx
<div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-50">
  ...
</div>
```

### Stat Card (metric display)

```tsx
<div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-50">
  <p className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mb-2">Label</p>
  <h2 className="text-3xl font-extrabold text-primary font-headline">Value</h2>
</div>
```

Primary variant: swap background to `bg-primary text-white`.

### Section Header

```tsx
<h1 className="text-3xl font-extrabold text-primary tracking-tight font-headline">Title</h1>
<p className="text-on-surface-variant">Subtitle</p>
```

### Buttons

**Primary action:**
```tsx
<button className="px-5 py-2.5 bg-primary text-white rounded-xl text-sm font-semibold hover:opacity-90 transition-opacity">
```

**Secondary action:**
```tsx
<button className="px-4 py-2 bg-surface-container-high text-on-surface font-semibold rounded-xl text-sm hover:bg-surface-container-highest transition-colors">
```

**Gradient action:**
```tsx
<button className="primary-gradient text-white font-semibold py-3 px-4 rounded-xl shadow-lg shadow-primary/20 hover:opacity-90 transition-opacity">
```

### Status Badges

```tsx
<span className="px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tight bg-primary-fixed text-primary">
  New
</span>
```

Status color map:
- `new` → `bg-primary-fixed text-primary`
- `scheduled/in_progress` → `bg-amber-100 text-amber-700`
- `done/won` → `bg-tertiary-container/10 text-tertiary-container`

### Glass Panel

```tsx
<div className="glass-panel p-6 rounded-xl border border-white/40 shadow-2xl">
```

Uses backdrop-filter blur for frosted glass effect. Use for floating overlays.

### Navigation Item (active state)

```tsx
<Link className="w-full flex items-center px-4 py-3 bg-blue-50 text-primary font-bold rounded-xl relative">
  <div className="absolute left-0 w-1 h-8 bg-primary rounded-r-full" />  {/* active indicator */}
  <Icon className="mr-3 w-5 h-5 text-primary" />
  <span className="font-headline text-sm font-medium">Label</span>
</Link>
```

### Progress Bar

```tsx
<div className="h-1.5 w-full bg-slate-100 rounded-full overflow-hidden">
  <div className="h-full bg-primary rounded-full" style={{ width: `${pct}%` }} />
</div>
```

---

## Layout Structure

```
flex min-h-screen overflow-hidden bg-surface
├── Sidebar (w-64, h-screen, border-r, bg-white)
└── div.flex-1.flex.flex-col.h-screen.overflow-hidden
    ├── TopBar (h-16, sticky, bg-white/80 backdrop-blur)
    └── main.flex-1.overflow-y-auto
        └── <page content with p-8>
```

---

## Animation

Uses `motion/react` for page transitions:

```tsx
<motion.div
  initial={{ opacity: 0, y: 20 }}
  animate={{ opacity: 1, y: 0 }}
  className="p-8 space-y-8"
>
```

Variants by page type:
- Dashboard: `y: 20` fade-in-up
- Leads: `x: 20` fade-in-right
- Calendar: `scale: 0.95` scale-in
