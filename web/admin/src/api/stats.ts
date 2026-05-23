import { get } from './http'

export interface Overview {
  requests_today: number; revenue_today: number
  active_users: number; error_rate: number
  delta: { requests: number; revenue: number; users: number; error_rate: number }
}

export interface TimeseriesPoint { ts: string; requests: number; errors: number; quota: number }

export interface ByModelRow { model: string; requests: number; tokens_in: number; tokens_out: number; quota: number; errors: number }
export interface ByChannelRow { channel_id: number; channel_name: string; provider: string; requests: number; quota: number; errors: number }
export interface ByUserRow { user_id: number; username: string; requests: number; quota: number; errors: number }

export const statsApi = {
  overview: () => get<Overview>('/api/admin/stats/overview'),
  timeseries: (p: { from?: string; to?: string; granularity?: 'hour' | 'day'; metrics?: string }) =>
    get<{ points: TimeseriesPoint[] }>('/api/admin/stats/timeseries', p as Record<string, unknown>),
  byModel: (p: { order_by?: 'quota' | 'requests'; limit?: number; from?: string; to?: string }) =>
    get<{ rows: ByModelRow[] }>('/api/admin/stats/by_model', p as Record<string, unknown>),
  byChannel: (p: { order_by?: 'quota' | 'requests'; limit?: number; from?: string; to?: string }) =>
    get<{ rows: ByChannelRow[] }>('/api/admin/stats/by_channel', p as Record<string, unknown>),
  byUser: (p: { order_by?: 'quota' | 'requests'; limit?: number; from?: string; to?: string }) =>
    get<{ rows: ByUserRow[] }>('/api/admin/stats/by_user', p as Record<string, unknown>),
  exportURL: (p: Record<string, string>) => `/api/admin/stats/export?${new URLSearchParams({ format: 'csv', ...p }).toString()}`,
}
