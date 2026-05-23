import { get, post, patch, del } from './http'

export interface TokenView {
  id: string
  name: string
  prefix: string
  status: 'enabled' | 'disabled'
  quota_used: number
  quota_limit: number | null
  allowed_models: string[]
  allowed_ips: string[]
  rpm_limit: number
  tpm_limit: number
  expires_at: string | null
  last_used_at: string | null
  created_at: string
}

export interface CreateTokenParams {
  name: string
  quota_limit?: number | null
  allowed_models?: string[]
  allowed_ips?: string[]
  rpm_limit?: number
  tpm_limit?: number
  expires_at?: string | null
}

export interface CreateTokenResponse {
  view: TokenView
  plaintext_key: string
  warning?: string
}

export interface ListTokensResponse {
  items: TokenView[]
  total: number
}

export const tokenApi = {
  list: (page = 1, size = 20) => get<ListTokensResponse>(`/api/user/tokens?page=${page}&page_size=${size}`),
  create: (p: CreateTokenParams) => post<CreateTokenResponse>('/api/user/tokens', p),
  update: (id: string, p: Partial<CreateTokenParams>) => patch<TokenView>(`/api/user/tokens/${id}`, p),
  revoke: (id: string) => del(`/api/user/tokens/${id}`),
  regenerate: (id: string) => post<CreateTokenResponse>(`/api/user/tokens/${id}/regenerate`),
}
