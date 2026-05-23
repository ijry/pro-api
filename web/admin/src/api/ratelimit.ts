import { get, post } from './http'

export interface RateLimitStats { key: string; count: number; window_seconds: number; ttl_ms: number }

export const ratelimitApi = {
  stats: (key: string) => get<RateLimitStats>(`/api/admin/ratelimit/keys/${encodeURIComponent(key)}/stats`),
  reset: (key: string) => post<{ ok: true }>(`/api/admin/ratelimit/keys/${encodeURIComponent(key)}/reset`, {}),
}
