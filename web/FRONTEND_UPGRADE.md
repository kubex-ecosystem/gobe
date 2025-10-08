# Frontend Upgrade - Kubex Brand v0.0.2 (Light-First)

> **Freedom, Engineered.** · **Independent by Design.**

## Overview

GoBE frontend has been upgraded to align with Kubex Brand Visual Specification v0.0.2, introducing a modern light-first aesthetic while maintaining the technical DNA of the ecosystem.

## What Changed

### 🎨 Visual Identity

**Before**: Dark-first with neon tech aesthetic
**After**: Light-first with soft, professional gradients

- **New Tagline**: "Freedom, Engineered" (replacing "Code Fast. Own Everything")
- **New Subtitle**: "Independent by Design • No Artificial Borders"
- **Footer Message**: Emphasizes interoperability and freedom philosophy

### 🎨 Color Palette Migration

#### Light Mode (Default)
```css
Background:    #f9fafb (ice)
Surface:       #ffffff (white)
Text:          #111827 (graphite) - headings
               #334155 (slate)    - body
Primary:       #06b6d4 (cyan 500)
Accent:        #a855f7 (lilac 500)
Borders:       #e2e8f0 (slate 200)
```

#### Dark Mode (Optional via `prefers-color-scheme: dark`)
```css
Background:    #0a0f14
Surface:       #0f1419
Text:          #e5f2f2 - headings
               #cbd5e1 - body
Borders:       #1e293b
```

### ✨ Typography

**Headings**: `Exo 2` / `Orbitron` (geometric/futuristic sans)
- Font-weight: 700 (semibold)
- Tight tracking for H1

**Body**: `Inter` (system sans)
- Line-height: 1.6 (comfortable reading)

**Code/Mono**: `IBM Plex Mono`
- Used in API endpoints and code snippets

**Loaded via Google Fonts**:
```html
<link href="https://fonts.googleapis.com/css2?family=Exo+2:wght@500;600;700;800&family=Inter:wght@400;500;600;700&family=Orbitron:wght@500;600;700;800&family=IBM+Plex+Mono:wght@400;500;600&display=swap" rel="stylesheet">
```

### 🌟 Design Elements

#### Soft Neon Glows
- **Before**: 15-25% opacity (aggressive)
- **After**: ≤10% opacity (subtle, professional)

```css
--glow-cyan: rgba(0, 136, 255, 0.08)
--glow-lilac: rgba(124, 77, 255, 0.08)
```

#### Hex Grid Background
- **Pattern**: Subtle hex grid with `rgba(0, 76, 153, 0.05)` stroke
- **Animation**: Gentle float (20s ease-in-out)
- **Opacity**: 5% in light mode, 15% in dark mode

#### Shadows
```css
--shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.1)
--shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1)
--shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.1)
--shadow-hover: 0 12px 24px rgba(6, 182, 212, 0.15)
```

### 🎯 Components Updated

#### Cards
- **Background**: White surface with soft borders
- **Border**: `#e2e8f0` (slate-200)
- **Shadow**: `sm` (subtle elevation)
- **Hover**: Lifts with cyan border and enhanced shadow
- **Top accent**: 3px gradient bar on hover

#### Badges ("Powered by Kubex")
```css
Border:     1px solid rgba(6, 182, 212, 0.2)
Background: #ecfeff (cyan-50)
Text:       #0e7490 (cyan-700)
Border-radius: full (pill shape)
Shadow: sm
```

#### Buttons
- **Primary**: Cyan-to-lilac gradient (`--gradient-hero`)
- **Secondary**: Transparent with cyan border
- **Hover**: Lifts with enhanced shadow and subtle background tint

#### Links
- **Default**: `#111827` (text-head)
- **Underline**: `rgba(6, 182, 212, 0.4)` (cyan 40% opacity)
- **Hover**: `#0891b2` (cyan-600)

### ♿ Accessibility (WCAG AA)

✅ **Contrast Ratios**:
- Body text: ≥ 4.5:1
- Headings: ≥ 3:1
- Focus rings: 2px cyan `#06b6d4` with 2px offset

✅ **Motion Preferences**:
```css
@media (prefers-reduced-motion: reduce) {
    * {
        animation-duration: 0.01ms !important;
        animation-iteration-count: 1 !important;
        transition-duration: 0.01ms !important;
    }
}
```

✅ **Dark Mode Support**: Automatic via `prefers-color-scheme: dark`

## Files Modified

```
web/
├── index.html       # Updated taglines, added fonts, new inline tokens
├── style.css        # Complete token migration + dark mode
└── FRONTEND_UPGRADE.md  # This document
```

## Token Map (v0.0.1 → v0.0.2)

| Old Token                      | New Token                     |
|--------------------------------|-------------------------------|
| `--bg-dark: #0d1117`          | `--bg-base: #f9fafb`         |
| `--text-primary: #e6edf3`     | `--text-head: #111827`       |
| `--accent-color: #00ff88`     | `--primary-cyan: #06b6d4`    |
| `--border-subtle: rgba(255,...)` | `--border-muted: #e2e8f0` |
| `--glow (15-25%)`             | `--glow-cyan (8%)`           |

## Migration Checklist

✅ Replace base tokens (bg, surface, text, border)
✅ Reduce glows to ≤10%
✅ Update components: Badge, Card, Link, Section
✅ Validate contrast (WCAG AA)
✅ Maintain logo/icon palette (no saturation change)
✅ Add dark mode with token parity
✅ Add motion preference support
✅ Load geometric fonts (Exo 2, Orbitron)
✅ Test build (✅ **Passed**)

## Philosophy Alignment

The new design embodies Kubex's matured philosophy:

- **Freedom, Engineered**: Technical excellence without lock-in
- **Independent by Design**: No vendor dependencies
- **Interoperability as Diplomacy**: Talk to every system, depend on none

The light-first aesthetic appeals equally to garage hackers and enterprise CTOs, reflecting Kubex's "grown up" positioning while preserving its open, rebellious DNA.

## Build Info

- **Version**: 1.3.5
- **Build Status**: ✅ Success
- **Platform**: linux/amd64
- **Go Version**: 1.25.1

---

**© 2025 Kubex Ecosystem** — All co-authors, human and artificial.
