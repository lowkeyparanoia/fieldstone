# ChampIQ on Sim — Vite + React + TypeScript

A real **Vite + React 18 + TypeScript** build of the ChampIQ product UI (the
design-system UI kit, ported off the Babel-CDN prototype). Dark-violet, calm-technical
ChampIQ design language; Pixie mascot; node/edge canvas; the 8-page IA + the Pixie
Copilot with the live **zerolang** guardrail.

## Run
```bash
npm install
npm run dev      # http://localhost:5174
npm run build    # production build (dist/)
```
Log in on the gate (pick a client / KG namespace) → the rail routes the pages:
Dashboard · ChampGraph · Canvas · Sequences · Personalize · Inbox · Prospects ·
Deliverability · Analytics · Settings.

## Layout
```
index.html              ← Vite entry → /src/main.tsx
src/main.tsx            ← imports design-system CSS + the app module
src/styles.css          ← design-system entry (@imports tokens/*)
src/tokens/*            ← color / type / spacing / motion tokens
src/champiq-app.jsx     ← the kit primitives + 14 screens (one module; React via imports)
public/assets/pixie/*   ← Pixie sprite PNGs
```

## Notes
- The 14 authored screens shared a global scope in the prototype; here they are
  concatenated into one ES module with `React`/`ReactDOM` imported at the top — the
  authored JSX is preserved verbatim (no per-screen rewrite), so behaviour matches the
  prototype while compiling under Vite/esbuild.
- Flexbox `min-width:0` is applied to the content columns so panels shrink instead of
  clipping on narrower viewports (the prototype's horizontal-overflow bug).
- No backend: mock data, hash-linkable routes, localStorage for client/accent/auth.
