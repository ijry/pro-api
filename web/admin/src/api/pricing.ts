import { get, post, patch, del, type Page } from './http'

export type PricingScope = 'global' | 'group' | 'model' | 'group_model'

export interface PricingRule {
  id: number
  scope: PricingScope
  group_id: number | null
  model: string | null
  input_ratio: number | null
  output_ratio: number | null
  cached_ratio: number | null
  reasoning_ratio: number | null
  priority: number
  status: 0 | 1
  created_at: string; updated_at: string
}

export interface PricingInput {
  scope: PricingScope
  group_id?: number | null
  model?: string | null
  input_ratio?: number | null
  output_ratio?: number | null
  cached_ratio?: number | null
  reasoning_ratio?: number | null
  priority?: number
  status?: 0 | 1
}

export const pricingApi = {
  list: (p: { scope?: PricingScope; group_id?: number; model?: string; page?: number; size?: number }) =>
    get<Page<PricingRule>>('/api/admin/pricing/rules', p as Record<string, unknown>),
  create: (b: PricingInput) => post<PricingRule>('/api/admin/pricing/rules', b),
  patch: (id: number, b: Partial<PricingInput>) => patch<PricingRule>(`/api/admin/pricing/rules/${id}`, b),
  remove: (id: number) => del<{ ok: true }>(`/api/admin/pricing/rules/${id}`),
}
