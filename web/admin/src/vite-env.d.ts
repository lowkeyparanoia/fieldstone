/// <reference types="vite/client" />

// Without this reference Vite's ambient types are missing and
// import.meta.env fails to typecheck.
interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
}
interface ImportMeta {
  readonly env: ImportMetaEnv
}
