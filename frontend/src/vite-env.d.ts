/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  readonly VITE_CARTO_API_KEY?: string
  readonly VITE_CARTO_TILE_URL_LIGHT?: string
  readonly VITE_CARTO_TILE_URL_DARK?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
