/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const c: DefineComponent<Record<string, unknown>, Record<string, unknown>, unknown>
  export default c
}

interface ImportMetaEnv {
  readonly VITE_DEMO_MOCK?: string
}
interface ImportMeta {
  readonly env: ImportMetaEnv
}
