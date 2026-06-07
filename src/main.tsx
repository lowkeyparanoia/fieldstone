// ChampIQ web entry. Loads the design-system tokens, then the app module
// (kit + 14 screens) which self-mounts into #root.
import './styles.css'
// @ts-expect-error — JSX app module compiled by Vite/esbuild (no TS types needed)
import './champiq-app.jsx'
