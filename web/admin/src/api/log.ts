import { get, type Page } from './http'

export interface RequestLog {
  id: number; created_at: string
  user_id: number; token_id: number
  group_id: number | null
  event_type: 0 | 1
  client_model: string; upstream_model: string
  channel_id: number | null
  protocol: string
  endpoint: string; ip: string
  status_code: number; latency_ms: number; ttft_ms: number; stream: boolean
  input_tokens: number; output_tokens: number
  cached_tokens: number; reasoning_tokens: number
  total_quota: number
  billing_input_ratio: number; billing_output_ratio: number; billing_group_ratio: number
  error_code: number; error_msg: string
  trace_id: string
}

export interface ErrorLog {
  id: number; created_at: string
  user_id: number | null; token_id: number | null; channel_id: number | null
  error_code: number; error_type: string
  stack: string; context: Record<string, unknown>
  trace_id: string
}

export interface AuditLog {
  id: number; created_at: string
  actor_id: number | null; actor_role: number
  action: string; target_type: string; target_id: number | null
  before: unknown; after: unknown
  ip: string
}

export interface LogFilter {
  from?: string; to?: string
  user_id?: number; token_id?: number
  model?: string; channel_id?: number
  event_type?: 0 | 1; status_code?: number; error_code?: number
  page?: number; size?: number
}

export const logApi = {
  requests: (p: LogFilter) => get<Page<RequestLog>>('/api/admin/logs/requests', p as Record<string, unknown>),
  errors: (p: LogFilter) => get<Page<ErrorLog>>('/api/admin/logs/errors', p as Record<string, unknown>),
  audit: (p: { actor_id?: number; action?: string; target_type?: string; target_id?: number; from?: string; to?: string; page?: number; size?: number }) =>
    get<Page<AuditLog>>('/api/admin/logs/audit', p as Record<string, unknown>),
}
