import { get, patch, del, type Page } from './http'

export interface Token {
  id: number; user_id: number; name: string
  key_prefix: string
  quota_limit: number | null; quota_used: number
  allowed_models: string[]; allowed_ips: string[]
  rpm_limit: number; tpm_limit: number
  expires_at: string | null; last_used_at: string | null
  status: 0 | 1 | 2
  created_at: string
}

export interface TokenListParams {
  page?: number; size?: number; keyword?: string
  user_id?: number; status?: 0 | 1 | 2
}

export interface TokenPatch {
  name?: string
  quota_limit?: number | null
  clear_quota_limit?: boolean
  allowed_models?: string[]
  allowed_ips?: string[]
  rpm_limit?: number; tpm_limit?: number
  expires_at?: string | null
  clear_expires_at?: boolean
  status?: 0 | 1
}

export const tokenApi = {
  list: (p: TokenListParams) => get<Page<Token>>('/api/admin/apikeys', p as Record<string, unknown>),
  patch: (id: number, b: TokenPatch) => patch<Token>(`/api/admin/apikeys/${id}`, b),
  remove: (id: number) => del<{ status: string }>(`/api/admin/apikeys/${id}`),
}
