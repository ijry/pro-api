import { paginate, ok, clone, type PageParams } from './helpers'

import adminUser from './data/admin-user.json'
import userProfile from './data/user-profile.json'
import userWallet from './data/user-wallet.json'
import channels from './data/channels.json'
import adminTokens from './data/admin-tokens.json'
import userTokens from './data/user-tokens.json'
import models from './data/models.json'
import statsOverview from './data/stats-overview.json'
import statsTimeseries from './data/stats-timeseries.json'
import statsByModel from './data/stats-by-model.json'
import statsByChannel from './data/stats-by-channel.json'
import statsByUser from './data/stats-by-user.json'
import ledger from './data/ledger.json'
import usage from './data/usage.json'
import notices from './data/notices.json'
import logRequests from './data/log-requests.json'
import logErrors from './data/log-errors.json'
import logAudit from './data/log-audit.json'
import adminRecharges from './data/admin-recharges.json'
import userRecharges from './data/user-recharges.json'
import oauthBindings from './data/oauth-bindings.json'
import inviteSummary from './data/invite-summary.json'
import inviteInvitees from './data/invite-invitees.json'
import inviteRecords from './data/invite-records.json'
import groups from './data/groups.json'
import users from './data/users.json'
import userDetail from './data/user-detail.json'
import pricingRules from './data/pricing-rules.json'
import settings from './data/settings.json'
import accounts from './data/accounts.json'

export type MockMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE'
export type MockHandler = (method: MockMethod, url: string, params?: unknown) => unknown

export interface MockRoute {
  pattern: RegExp
  handler: MockHandler
  methods?: MockMethod[]
}

const writeOk: MockHandler = () => ok()

const channelById = (url: string) => {
  const m = url.match(/\/channels\/(\d+)/)
  const id = m ? Number(m[1]) : 0
  const found = (channels as any[]).find((c) => c.id === id)
  return found ? clone(found) : clone((channels as any[])[0])
}

const userTokenById = (url: string) => {
  const m = url.match(/\/(?:tokens|apikeys)\/([^/?]+)/)
  const id = m ? m[1] : ''
  const found = (userTokens as any[]).find((t) => t.id === id)
  return found ? clone(found) : clone((userTokens as any[])[0])
}

const publicModels = () => ({
  models: (models as any[]).map((m) => ({
    id: String(m.name ?? m.id),
    name: m.name,
    provider: m.family || 'Other',
    context_length: m.max_input_tokens,
    input_price_per_1k: m.default_input_ratio,
    output_price_per_1k: m.default_output_ratio,
  })),
})

const channelMappings: Record<number, any[]> = {
  1: [
    { id: 101, channel_id: 1, client_model: 'gpt-4o', upstream_model: 'gpt-4o', input_ratio: 5, output_ratio: 15, cached_ratio: 2.5, reasoning_ratio: null },
    { id: 102, channel_id: 1, client_model: 'gpt-4o-mini', upstream_model: 'gpt-4o-mini', input_ratio: 0.15, output_ratio: 0.6, cached_ratio: 0.075, reasoning_ratio: null },
    { id: 103, channel_id: 1, client_model: 'dall-e-3', upstream_model: 'dall-e-3', input_ratio: 40, output_ratio: 0, cached_ratio: null, reasoning_ratio: null },
  ],
  2: [
    { id: 201, channel_id: 2, client_model: 'claude-3-5-sonnet', upstream_model: 'claude-3-5-sonnet-20241022', input_ratio: 3, output_ratio: 15, cached_ratio: 0.3, reasoning_ratio: null },
    { id: 202, channel_id: 2, client_model: 'claude-3-opus', upstream_model: 'claude-3-opus-20240229', input_ratio: 15, output_ratio: 75, cached_ratio: 1.5, reasoning_ratio: null },
  ],
  3: [
    { id: 301, channel_id: 3, client_model: 'deepseek-chat', upstream_model: 'deepseek-chat', input_ratio: 0.14, output_ratio: 0.28, cached_ratio: null, reasoning_ratio: null },
    { id: 302, channel_id: 3, client_model: 'deepseek-reasoner', upstream_model: 'deepseek-reasoner', input_ratio: 0.55, output_ratio: 2.19, cached_ratio: 0.14, reasoning_ratio: 2.19 },
  ],
}

const channelMappingsByUrl = (url: string) => {
  const id = Number((url.match(/\/channels\/(\d+)\/(?:model_mappings|mappings)/) || [])[1])
  return { items: clone(channelMappings[id] ?? []) }
}

const accountEventsByUrl = (url: string, params?: unknown) => {
  const id = Number((url.match(/\/accounts\/(\d+)\/events/) || [])[1])
  const base = [
    { id: id * 100 + 1, account_id: id, event_type: 'imported', payload: { source: 'demo' }, created_at: '2026-06-01T08:00:00Z' },
    { id: id * 100 + 2, account_id: id, event_type: 'test_ok', payload: { latency_ms: 318 }, created_at: '2026-06-02T09:30:00Z' },
    { id: id * 100 + 3, account_id: id, event_type: 'refreshed', payload: { token_expires_at: '2026-06-09T00:00:00Z' }, created_at: '2026-06-02T12:00:00Z' },
  ]
  return paginate(base, params as PageParams)
}

const accountCredentialsByUrl = (url: string) => {
  const id = Number((url.match(/\/accounts\/(\d+)\/credentials\/peek/) || [])[1])
  const found = (accounts as any[]).find((a) => a.id === id) ?? (accounts as any[])[0]
  return {
    credentials: {
      type: found.cred_type,
      api_key: found.cred_type === 'apikey' ? 'sk-demo-full-key-not-real' : undefined,
      access_token: found.cred_type !== 'apikey' ? 'demo-access-token-not-real' : undefined,
      refresh_token: found.cred_type !== 'apikey' ? 'demo-refresh-token-not-real' : undefined,
      note: 'Demo credentials only. No real secret is stored here.',
    },
  }
}

const userLogItem = (e: any) => ({
  id: String(e.id),
  model: e.client_model,
  status: e.status_code,
  cost_usd: e.total_quota / 100000,
  latency_ms: e.latency_ms,
  prompt_tokens: e.input_tokens,
  completion_tokens: e.output_tokens,
  created_at: e.created_at,
  trace_id: e.trace_id || undefined,
})

export const routes: MockRoute[] = [
  // ============ admin =============
  { pattern: /^\/api\/admin\/auth\/login$/,                handler: () => ({ user: clone(adminUser), session: { id: 'demo-session', expires_at: '2099-12-31T23:59:59Z' } }) },
  { pattern: /^\/api\/admin\/auth\/logout$/,               handler: writeOk },
  { pattern: /^\/api\/admin\/auth\/me$/,                   handler: () => clone(adminUser) },

  { pattern: /^\/api\/admin\/stats\/overview$/,            handler: () => clone(statsOverview) },
  { pattern: /^\/api\/admin\/stats\/timeseries$/,          handler: () => clone(statsTimeseries) },
  { pattern: /^\/api\/admin\/stats\/by_model$/,            handler: () => clone(statsByModel) },
  { pattern: /^\/api\/admin\/stats\/by_channel$/,          handler: () => clone(statsByChannel) },
  { pattern: /^\/api\/admin\/stats\/by_user$/,             handler: () => clone(statsByUser) },

  { pattern: /^\/api\/admin\/channels\/batch_test$/,        handler: () => ({ results: (channels as any[]).slice(0, 3).map((c: any) => ({ ok: true, latency_ms: 120 + Math.floor(c.id * 17), channel_id: c.id })) }) },
  { pattern: /^\/api\/admin\/channels\/\d+\/model_mappings$/, handler: (m, u, p) => m === 'GET' ? channelMappingsByUrl(u) : { items: (p as any)?.mappings ?? [] } },
  { pattern: /^\/api\/admin\/channels\/\d+\/mappings$/,    handler: (_m, u) => channelMappingsByUrl(u) },
  { pattern: /^\/api\/admin\/channels\/\d+\/test$/,        handler: () => ({ ok: true, latency_ms: 142 }) },
  { pattern: /^\/api\/admin\/channels\/\d+\/health$/,      handler: () => ({ state: 'closed', consec_fail: 0, opened_at: null }) },
  { pattern: /^\/api\/admin\/channels\/\d+\/?$/,           handler: (m, u) => m === 'GET' ? channelById(u) : ok() },
  { pattern: /^\/api\/admin\/channels$/,                   handler: (m, _u, p) => {
    if (m !== 'GET') return ok()
    const pp = p as PageParams & { group_id?: string | number }
    let list = channels as any[]
    if (pp?.group_id !== undefined && pp.group_id !== '') {
      const gid = Number(pp.group_id)
      list = list.filter(c => c.group_id === gid)
    }
    return paginate(list, pp)
  }},

  // API Keys (admin)
  { pattern: /^\/api\/admin\/apikeys\/\d+$/,               handler: (m, _u, _p) => m === 'GET' ? clone((adminTokens as any[])[0]) : ok() },
  { pattern: /^\/api\/admin\/apikeys$/,                    handler: (_m, _u, p) => paginate(adminTokens as any[], p as PageParams) },

  { pattern: /^\/api\/admin\/model_catalogs\/(\d+)$/,      handler: (m, u) => {
    const id = Number((u.match(/\/model_catalogs\/(\d+)/) || [])[1])
    const found = (models as any[]).find((m: any) => m.id === id) ?? (models as any[])[0]
    return m === 'GET' ? clone(found) : ok()
  }},
  { pattern: /^\/api\/admin\/model_catalogs$/,             handler: (m, _u, p) => m === 'GET' ? paginate(models as any[], p as PageParams) : ok() },

  { pattern: /^\/api\/admin\/logs\/requests$/,             handler: (_m, _u, p) => paginate(logRequests as any[], p as PageParams) },
  { pattern: /^\/api\/admin\/logs\/errors$/,               handler: (_m, _u, p) => paginate(logErrors as any[], p as PageParams) },
  { pattern: /^\/api\/admin\/logs\/audit$/,                handler: (_m, _u, p) => paginate(logAudit as any[], p as PageParams) },

  { pattern: /^\/api\/admin\/notices\/\d+\/(publish|unpublish)$/, handler: () => clone((notices as any[])[0]) },
  { pattern: /^\/api\/admin\/notices\/\d+$/,               handler: (m, _u, _p) => m === 'GET' ? clone((notices as any[])[0]) : ok() },
  { pattern: /^\/api\/admin\/notices$/,                    handler: (m, _u, p) => m === 'GET' ? paginate(notices as any[], p as PageParams) : ok() },

  { pattern: /^\/api\/admin\/payments\/manual_recharges\/\d+\/(approve|reject)$/, handler: () => clone((adminRecharges as any[])[0]) },
  { pattern: /^\/api\/admin\/payments\/manual_recharges\/\d+$/,                   handler: () => clone((adminRecharges as any[])[0]) },
  { pattern: /^\/api\/admin\/payments\/manual_recharges$/,                        handler: (_m, _u, p) => paginate(adminRecharges as any[], p as PageParams) },
  { pattern: /^\/api\/admin\/payments\/redeem_codes\/batch$/,                     handler: (_m, _u, p: any) => ({ batch_no: `BATCH-DEMO-${Date.now()}`, count: p?.count ?? 5, codes: Array.from({ length: p?.count ?? 5 }, (_, i) => `DEMO-${String(i + 1).padStart(4, '0')}`) }) },
  { pattern: /^\/api\/admin\/payments\/redeem_codes\/disable$/,                   handler: () => ({ disabled: 1 }) },
  { pattern: /^\/api\/admin\/payments\/redeem_codes\/\d+\/disable$/,              handler: writeOk },
  { pattern: /^\/api\/admin\/payments\/redeem_codes$/,                            handler: (_m, _u, p) => paginate([], p as PageParams) },
  { pattern: /^\/api\/admin\/payments\/redeem\/\d+$/,                             handler: writeOk },
  { pattern: /^\/api\/admin\/payments\/redeem$/,                                  handler: (_m, _u, p) => paginate([], p as PageParams) },

  // Groups (admin)
  { pattern: /^\/api\/admin\/groups\/\d+$/,                handler: (m, u) => {
    const id = Number((u.match(/\/groups\/(\d+)/) || [])[1])
    const found = (groups as any[]).find((g) => g.id === id) ?? (groups as any[])[0]
    return m === 'GET' ? clone(found) : ok()
  }},
  { pattern: /^\/api\/admin\/groups$/,                     handler: (m, _u, _p) => m === 'GET' ? { items: clone(groups) } : ok() },

  // Users (admin)
  { pattern: /^\/api\/admin\/users\/\d+\/quota$/,          handler: () => ({ ok: true, balance_after: 8650000 }) },
  { pattern: /^\/api\/admin\/users\/\d+\/reset_password$/, handler: () => ({ ok: true, temp_password: 'Demo@2026!' }) },
  { pattern: /^\/api\/admin\/users\/\d+$/,                 handler: (m, _u, _p) => m === 'GET' ? clone(userDetail) : ok() },
  { pattern: /^\/api\/admin\/users$/,                      handler: (_m, _u, p) => paginate(users as any[], p as PageParams) },

  // Pricing (admin)
  { pattern: /^\/api\/admin\/pricing\/rules\/\d+$/,        handler: (m, u) => {
    const id = Number((u.match(/\/rules\/(\d+)/) || [])[1])
    const found = (pricingRules as any[]).find((r) => r.id === id) ?? (pricingRules as any[])[0]
    return m === 'GET' ? clone(found) : ok()
  }},
  { pattern: /^\/api\/admin\/pricing\/rules$/,             handler: (m, _u, p) => m === 'GET' ? paginate(pricingRules as any[], p as PageParams) : ok() },

  // Accounts / 号池 (admin)
  { pattern: /^\/api\/admin\/accounts\/import$/,           handler: () => ({ imported: 0, preview: [], errors: [], account_ids: [] }) },
  { pattern: /^\/api\/admin\/accounts\/\d+\/events$/,      handler: (_m, u, p) => accountEventsByUrl(u, p) },
  { pattern: /^\/api\/admin\/accounts\/\d+\/credentials\/peek$/, handler: (_m, u) => accountCredentialsByUrl(u) },
  { pattern: /^\/api\/admin\/accounts\/\d+\/[a-z_]+$/,     handler: writeOk },
  { pattern: /^\/api\/admin\/accounts\/stats\/overview$/,  handler: () => ({ total: 7, active: 5, cooldown: 1, disabled: 1, error: 0 }) },
  { pattern: /^\/api\/admin\/accounts\/\d+$/,              handler: (m, u) => {
    const id = Number((u.match(/\/accounts\/(\d+)/) || [])[1])
    const found = (accounts as any[]).find((a) => a.id === id) ?? (accounts as any[])[0]
    return m === 'GET' ? clone(found) : ok()
  }},
  { pattern: /^\/api\/admin\/accounts$/,                   handler: (m, _u, p) => m === 'GET' ? paginate(accounts as any[], p as PageParams) : ok() },

  // Settings (admin)
  { pattern: /^\/api\/admin\/settings\/test_smtp$/,        handler: () => ({ ok: true, stubbed: true }) },
  { pattern: /^\/api\/admin\/settings\/[^/?]+$/,           handler: writeOk },
  { pattern: /^\/api\/admin\/settings$/,                   handler: () => clone(settings) },

  // Ratelimit (admin)
  { pattern: /^\/api\/admin\/ratelimit\/keys\/[^/?]+\/reset$/, handler: writeOk },
  { pattern: /^\/api\/admin\/ratelimit\/keys\/[^/?]+\/stats$/, handler: () => ({ key: 'demo:key', count: 42, window_seconds: 60, ttl_ms: 28000 }) },

  // ============ user =============
  { pattern: /^\/api\/user\/profile$/,                     handler: (m, _u, _p) => m === 'GET' ? clone(userProfile) : clone(userProfile) },
  { pattern: /^\/api\/user\/password$/,                    handler: writeOk },
  { pattern: /^\/api\/user\/oauth\/bindings(\/[^/?]+)?$/,  handler: () => clone(oauthBindings) },
  { pattern: /^\/api\/auth\/oauth\/github\/start$/,        handler: () => ({ redirect_url: '#demo' }) },

  { pattern: /^\/api\/user\/wallet\/ledger$/,              handler: (_m, _u, p) => paginate(ledger as any[], p as PageParams) },
  { pattern: /^\/api\/user\/wallet$/,                      handler: () => clone(userWallet) },

  // API Keys (user)
  { pattern: /^\/api\/user\/apikeys\/[^/?]+\/regenerate$/, handler: () => ({ view: clone((userTokens as any[])[0]), plaintext_key: 'sk-prx-demo-regenerated-xxxxxxxxxxxx' }) },
  { pattern: /^\/api\/user\/apikeys\/[^/?]+$/,             handler: (m, u, _p) => m === 'GET' ? userTokenById(u) : (m === 'POST' ? { view: clone((userTokens as any[])[0]), plaintext_key: 'sk-prx-demo-xxxxxxxxxxxx' } : ok()) },
  { pattern: /^\/api\/user\/apikeys(\?.*)?$/,              handler: (m, _u, _p) => m === 'GET' ? clone({ items: userTokens, total: (userTokens as any[]).length }) : ({ view: clone((userTokens as any[])[0]), plaintext_key: 'sk-prx-demo-new-xxxxxxxxxxxx' }) },

  { pattern: /^\/api\/user\/usage(\?.*)?$/,                handler: () => clone(usage) },

  { pattern: /^\/api\/user\/payment\/manual\/[^/?]+\/cancel$/, handler: writeOk },
  { pattern: /^\/api\/user\/payment\/manual\/[^/?]+$/,         handler: () => clone((userRecharges as any).items[0]) },
  { pattern: /^\/api\/user\/payment\/manual(\?.*)?$/,          handler: (m) => m === 'GET' ? clone(userRecharges) : clone((userRecharges as any).items[0]) },
  { pattern: /^\/api\/user\/payment\/redeem$/,                 handler: () => ({ granted_usd: 5, code: 'DEMO-REDEEM-XXXX' }) },

  { pattern: /^\/api\/user\/logs\/requests(\?.*)?$/,           handler: (_m, _u, p) => paginate((logRequests as any[]).map(userLogItem), p as PageParams) },

  { pattern: /^\/api\/user\/notices\/unread_count$/,           handler: () => ({ count: 1 }) },
  { pattern: /^\/api\/user\/notices\/[^/?]+\/read$/,           handler: writeOk },
  { pattern: /^\/api\/user\/notices\/[^/?]+$/,                 handler: () => clone((notices as any[])[0]) },
  { pattern: /^\/api\/user\/notices(\?.*)?$/,                  handler: (_m, _u, p) => paginate(notices as any[], p as PageParams) },

  { pattern: /^\/api\/user\/invites\/me$/,                   handler: () => clone(inviteSummary) },
  { pattern: /^\/api\/user\/invites\/invitees(\?.*)?$/,      handler: (_m, _u, p) => paginate(inviteInvitees as any[], p as PageParams) },
  { pattern: /^\/api\/user\/invites\/records(\?.*)?$/,       handler: (_m, _u, p) => paginate(inviteRecords as any[], p as PageParams) },

  { pattern: /^\/api\/public\/groups(\?.*)?$/,               handler: () => ({ items: (groups as any[]).filter((g) => g.status === 0) }) },
  { pattern: /^\/api\/public\/models(\?.*)?$/,               handler: () => clone(publicModels()) },
  { pattern: /^\/api\/public\/notices(\?.*)?$/,              handler: (_m, _u, p) => paginate(notices as any[], p as PageParams) },
]

export function routeMock(method: MockMethod, url: string, params?: unknown):
  { matched: boolean; data: unknown } {
  const path = url.split('?')[0]
  for (const r of routes) {
    if (r.methods && !r.methods.includes(method)) continue
    if (r.pattern.test(path) || r.pattern.test(url)) {
      return { matched: true, data: r.handler(method, url, params) }
    }
  }
  return { matched: false, data: null }
}

export type { PageParams }
