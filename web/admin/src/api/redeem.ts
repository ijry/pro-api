import { get, post, type Page } from './http'

export interface RedeemCode {
  id: number; code_prefix: string
  amount_quota: number; batch_no: string
  status: 0 | 1 | 2
  used_by: number | null; used_at: string | null
  expires_at: string | null
  created_by: number; created_at: string
}

export interface BatchGenerateInput {
  count: number; amount_quota: number
  batch_no: string
  expires_at?: string | null
  reason?: string
}

export const redeemApi = {
  list: (p: { batch_no?: string; status?: 0 | 1 | 2; page?: number; size?: number }) =>
    get<Page<RedeemCode>>('/api/admin/payments/redeem_codes', p as Record<string, unknown>),
  batchGenerate: (b: BatchGenerateInput) =>
    post<{ batch_no: string; count: number; codes: string[] }>('/api/admin/payments/redeem_codes/batch', b),
  disableOne: (id: number, reason?: string) =>
    post<{ ok: true }>(`/api/admin/payments/redeem_codes/${id}/disable`, { reason }),
  disableMany: (ids: number[], reason?: string) =>
    post<{ disabled: number }>('/api/admin/payments/redeem_codes/disable', { ids, reason }),
  exportURL: (p: { batch_no?: string; status?: 0 | 1 | 2 }) =>
    `/api/admin/payments/redeem_codes/export?${new URLSearchParams(p as Record<string, string>).toString()}`,
}
