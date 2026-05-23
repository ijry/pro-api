import { get } from './http'

export interface LogEntry {
  id: string
  model: string
  status: number
  cost_usd: number
  latency_ms: number
  prompt_tokens: number
  completion_tokens: number
  created_at: string
  trace_id?: string
}

export interface LogsResponse {
  items: LogEntry[]
  total: number
}

export const logApi = {
  list: (params: {
    page?: number
    page_size?: number
    from?: string
    to?: string
    model?: string
    status?: number
  } = {}) => {
    const q = new URLSearchParams()
    if (params.page) q.set('page', String(params.page))
    if (params.page_size) q.set('page_size', String(params.page_size))
    if (params.from) q.set('from', params.from)
    if (params.to) q.set('to', params.to)
    if (params.model) q.set('model', params.model)
    if (params.status) q.set('status', String(params.status))
    return get<LogsResponse>(`/api/user/logs/requests?${q}`)
  },
}
