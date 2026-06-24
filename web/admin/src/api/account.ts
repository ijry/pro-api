import { get, post, patch, del } from './http'

export interface QuotaWindow {
  remaining?: number | null
  total?: number | null
  reset_at?: string | null
  pct?: number | null
}

export interface Account {
  id: number
  name: string
  channel_id: number
  provider: 'anthropic' | 'openai'
  tier: string
  cred_type: 'oauth' | 'apikey' | 'token_pasted'
  email: string
  status: 0 | 1 | 2 | 3 | 4
  priority: number
  weight: number
  consec_failures: number
  refresh_token_valid: 0 | 1 | 2
  cooldown_until?: string | null
  last_used_at?: string | null
  last_success_at?: string | null
  last_failure_at?: string | null
  access_token_expires_at?: string | null
  quota_5h: QuotaWindow
  quota_week: QuotaWindow
}

export interface AccountDetail extends Account {
  import_source: string
  external_account_id: string
  created_at: string
  updated_at: string
}

export interface ListParams {
  channel_id?: number
  status?: number
}

export interface ListResp {
  items: Account[]
  total: number
}

export interface CreatePayload {
  channel_id: number
  name?: string
  provider?: string
  tier?: string
  format?: string
  text: string
  priority?: number
  weight?: number
  dry_run?: boolean
}

export interface ImportPayload {
  channel_id: number
  text?: string
  format?: string
  dry_run?: boolean
}

export interface ImportResp {
  total: number
  imported?: number
  preview?: Account[]
  errors?: { index: number; reason: string }[]
  account_ids?: number[]
}

export interface AccountEvent {
  id: number
  account_id: number
  event_type: string
  payload: unknown
  created_at: string
}

export interface EventsResp {
  items: AccountEvent[]
  total: number
  page: number
  size: number
}

export interface PeekResp {
  credentials: Record<string, unknown>
}

export interface StatsResp {
  totals: Record<string, number>
  by_provider: Record<string, number>
}

export interface OAuthStartPayload {
  provider: 'anthropic' | 'openai'
  channel_id: number
}

export interface OAuthStartResp {
  auth_url: string
  state: string
}

export const accountApi = {
  list:  (p: ListParams) => get<ListResp>('/api/admin/accounts', p as Record<string, unknown>),
  get:   (id: number) => get<AccountDetail>(`/api/admin/accounts/${id}`),
  create:(payload: CreatePayload) => post<{ id?: number; preview?: Account }>('/api/admin/accounts', payload),
  import:(payload: ImportPayload) => post<ImportResp>('/api/admin/accounts/import', payload),
  oauthStart: async (payload: OAuthStartPayload) =>
    (await post<{ data: OAuthStartResp }>('/api/admin/accounts/oauth/start', payload)).data,
  patch: (id: number, p: Partial<Account>) => patch<AccountDetail>(`/api/admin/accounts/${id}`, p),
  delete:(id: number) => del<{ id: number }>(`/api/admin/accounts/${id}`),
  enable:(id: number) => post<{ id: number; status: number }>(`/api/admin/accounts/${id}/enable`, {}),
  disable:(id: number) => post<{ id: number; status: number }>(`/api/admin/accounts/${id}/disable`, {}),
  test:   (id: number) => post<AccountDetail>(`/api/admin/accounts/${id}/test`, {}),
  refresh:(id: number) => post<{ id: number }>(`/api/admin/accounts/${id}/refresh_token`, {}),
  clearCooldown:(id: number) => post<{ id: number; status: number }>(`/api/admin/accounts/${id}/clear_cooldown`, {}),
  events: (id: number, page = 1, size = 20) => get<EventsResp>(`/api/admin/accounts/${id}/events`, { page, size }),
  peek:   (id: number) => get<PeekResp>(`/api/admin/accounts/${id}/credentials/peek`),
  stats:  () => get<StatsResp>('/api/admin/accounts/stats/overview'),
}
