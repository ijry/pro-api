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
  const m = url.match(/\/tokens\/([^/?]+)/)
  const id = m ? m[1] : ''
  const found = (userTokens as any[]).find((t) => t.id === id)
  return found ? clone(found) : clone((userTokens as any[])[0])
}

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

  { pattern: /^\/api\/admin\/channels\/\d+\/mappings$/,    handler: () => [] },
  { pattern: /^\/api\/admin\/channels\/\d+\/?$/,           handler: (m, u) => m === 'GET' ? channelById(u) : ok() },
  { pattern: /^\/api\/admin\/channels$/,                   handler: (m, _u, p) => m === 'GET' ? paginate(channels as any[], p as PageParams) : ok() },

  { pattern: /^\/api\/admin\/tokens\/\d+$/,                handler: (m, _u, _p) => m === 'GET' ? clone((adminTokens as any[])[0]) : ok() },
  { pattern: /^\/api\/admin\/tokens$/,                     handler: (_m, _u, p) => paginate(adminTokens as any[], p as PageParams) },

  { pattern: /^\/api\/admin\/model_catalogs\/\d+$/,        handler: (m, _u, _p) => m === 'GET' ? clone((models as any[])[0]) : ok() },
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

{ pattern: /^\/api\/admin\/users\/\d+\/quota$/,          handler: () => ({ ok: true, balance_after: 8650000 }) },

  // 兜底:其他 admin 子模块返回空(分页或数组),避免白屏
  { pattern: /^\/api\/admin\/(pricing|ratelimit|groups?|users|settings|redeem|payments)/, handler: (_m, _u, p) => paginate([], p as PageParams) },

  // ============ user =============
  { pattern: /^\/api\/user\/profile$/,                     handler: (m, _u, _p) => m === 'GET' ? clone(userProfile) : clone(userProfile) },
  { pattern: /^\/api\/user\/password$/,                    handler: writeOk },
  { pattern: /^\/api\/user\/oauth\/bindings(\/[^/?]+)?$/,  handler: () => clone(oauthBindings) },
  { pattern: /^\/api\/auth\/oauth\/github\/start$/,        handler: () => ({ redirect_url: '#demo' }) },

  { pattern: /^\/api\/user\/wallet\/ledger$/,              handler: (_m, _u, p) => paginate(ledger as any[], p as PageParams) },
  { pattern: /^\/api\/user\/wallet$/,                      handler: () => clone(userWallet) },

  { pattern: /^\/api\/user\/tokens\/[^/?]+\/regenerate$/,  handler: () => ({ view: clone((userTokens as any[])[0]), plaintext_key: 'sk-prx-demo-regenerated-xxxxxxxxxxxx' }) },
  { pattern: /^\/api\/user\/tokens\/[^/?]+$/,              handler: (m, u, _p) => m === 'GET' ? userTokenById(u) : (m === 'POST' ? { view: clone((userTokens as any[])[0]), plaintext_key: 'sk-prx-demo-xxxxxxxxxxxx' } : ok()) },
  { pattern: /^\/api\/user\/tokens(\?.*)?$/,               handler: (m, _u, _p) => m === 'GET' ? clone({ items: userTokens, total: (userTokens as any[]).length }) : ({ view: clone((userTokens as any[])[0]), plaintext_key: 'sk-prx-demo-new-xxxxxxxxxxxx' }) },

  { pattern: /^\/api\/user\/usage(\?.*)?$/,                handler: () => clone(usage) },

  { pattern: /^\/api\/user\/payment\/manual\/[^/?]+\/cancel$/, handler: writeOk },
  { pattern: /^\/api\/user\/payment\/manual\/[^/?]+$/,         handler: () => clone((userRecharges as any).items[0]) },
  { pattern: /^\/api\/user\/payment\/manual(\?.*)?$/,          handler: (m) => m === 'GET' ? clone(userRecharges) : clone((userRecharges as any).items[0]) },

  { pattern: /^\/api\/user\/notices/,                      handler: () => clone(notices) },
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
