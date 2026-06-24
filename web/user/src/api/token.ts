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

interface ApiTokenView {
  id: number | string
  name: string
  key_prefix?: string
  prefix?: string
  status: 0 | 1 | 2 | 'enabled' | 'disabled'
  quota_used: number
  quota_limit: number | null
  allowed_models?: string[]
  allowed_ips?: string[]
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

interface ApiCreateTokenResponse {
  view: ApiTokenView
  plaintext_key: string
  warning?: string
}

interface ApiListTokensResponse {
  items?: ApiTokenView[]
  total?: number
}

function normalizeStatus(status: ApiTokenView['status']): TokenView['status'] {
  if (status === 'enabled' || status === 0) return 'enabled'
  return 'disabled'
}

function normalizeToken(token: ApiTokenView): TokenView {
  return {
    id: String(token.id),
    name: token.name,
    prefix: token.prefix ?? token.key_prefix ?? '',
    status: normalizeStatus(token.status),
    quota_used: token.quota_used,
    quota_limit: token.quota_limit,
    allowed_models: token.allowed_models ?? [],
    allowed_ips: token.allowed_ips ?? [],
    rpm_limit: token.rpm_limit,
    tpm_limit: token.tpm_limit,
    expires_at: token.expires_at,
    last_used_at: token.last_used_at,
    created_at: token.created_at,
  }
}

function normalizeList(resp: ApiListTokensResponse): ListTokensResponse {
  const items = Array.isArray(resp.items) ? resp.items : []
  return {
    items: items.map(normalizeToken),
    total: resp.total ?? items.length,
  }
}

function normalizeCreate(resp: ApiCreateTokenResponse): CreateTokenResponse {
  return {
    view: normalizeToken(resp.view),
    plaintext_key: resp.plaintext_key,
    warning: resp.warning,
  }
}

export const tokenApi = {
  list: async (page = 1, size = 20) =>
    normalizeList(await get<ApiListTokensResponse>(`/api/user/apikeys?page=${page}&size=${size}`)),
  create: async (p: CreateTokenParams) =>
    normalizeCreate(await post<ApiCreateTokenResponse>('/api/user/apikeys', p)),
  update: async (id: string, p: Partial<CreateTokenParams>) =>
    normalizeToken(await patch<ApiTokenView>(`/api/user/apikeys/${id}`, p)),
  revoke: (id: string) => del(`/api/user/apikeys/${id}`),
  regenerate: async (id: string) =>
    normalizeCreate(await post<ApiCreateTokenResponse>(`/api/user/apikeys/${id}/regenerate`)),
}
