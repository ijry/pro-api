import { get, post } from './http'

export interface ManualRechargeOrder {
  id: string
  amount_cny: number
  remark: string
  status: 'pending' | 'approved' | 'rejected' | 'cancelled'
  reject_reason?: string
  granted_usd?: number
  created_at: string
  updated_at: string
}

export interface ManualRechargeResponse {
  items: ManualRechargeOrder[]
  total: number
}

export const rechargeApi = {
  create: (amount_cny: number, remark: string) =>
    post<ManualRechargeOrder>('/api/user/payment/manual', { amount_cny, remark }),
  list: (page = 1, size = 20) =>
    get<ManualRechargeResponse>(`/api/user/payment/manual?page=${page}&page_size=${size}`),
  get: (id: string) => get<ManualRechargeOrder>(`/api/user/payment/manual/${id}`),
  cancel: (id: string) => post(`/api/user/payment/manual/${id}/cancel`),
}
