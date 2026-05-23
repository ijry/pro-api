import { get, post, patch, del, type Page } from './http'

export interface ModelCatalog {
  id: number; name: string; family: 'chat' | 'embed' | 'image' | 'audio' | 'rerank'
  capabilities: string[]
  default_input_ratio: number; default_output_ratio: number
  default_cached_ratio: number | null; default_reasoning_ratio: number | null
  max_input_tokens: number
  status: 0 | 1
  created_at: string; updated_at: string
}

export interface ModelInput {
  name: string; family: ModelCatalog['family']
  capabilities: string[]
  default_input_ratio: number; default_output_ratio: number
  default_cached_ratio?: number | null; default_reasoning_ratio?: number | null
  max_input_tokens: number
  status?: 0 | 1
}

export const modelApi = {
  list: (p: { page?: number; size?: number; family?: string; status?: 0 | 1; keyword?: string }) =>
    get<Page<ModelCatalog>>('/api/admin/model_catalogs', p as Record<string, unknown>),
  create: (b: ModelInput) => post<ModelCatalog>('/api/admin/model_catalogs', b),
  patch: (id: number, b: Partial<ModelInput>) => patch<ModelCatalog>(`/api/admin/model_catalogs/${id}`, b),
  remove: (id: number) => del<{ ok: true }>(`/api/admin/model_catalogs/${id}`),
}
