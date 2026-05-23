import { get, post, type Page } from './http'

export interface ManualRecharge {
  id: number; user_id: number; username: string
  amount_money: number; currency: 'CNY' | 'USD'
  amount_quota: number
  status: 0 | 1 | 2 | 3
  applicant_note: string
  reviewer_id: number | null; review_note: string
  reviewed_at: string | null
  created_at: string; updated_at: string
}

export const rechargeApi = {
  list: (p: { status?: 0 | 1 | 2 | 3; user_id?: number; from?: string; to?: string; page?: number; size?: number }) =>
    get<Page<ManualRecharge>>('/api/admin/payments/manual_recharges', p as Record<string, unknown>),
  detail: (id: number) => get<ManualRecharge>(`/api/admin/payments/manual_recharges/${id}`),
  approve: (id: number, review_note: string) =>
    post<ManualRecharge>(`/api/admin/payments/manual_recharges/${id}/approve`, { review_note }),
  reject: (id: number, review_note: string) =>
    post<ManualRecharge>(`/api/admin/payments/manual_recharges/${id}/reject`, { review_note }),
}
