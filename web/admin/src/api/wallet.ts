import { get, post } from './http'

export interface Wallet {
  id: number; owner_type: 'user'; owner_id: number
  quota_balance: number; quota_total_recharged: number; quota_total_consumed: number
  currency: 'USD' | 'CNY'
}

export const walletApi = {
  mine: () => get<Wallet>('/api/user/wallet'),
  adminAdjust: (userId: number, body: { delta_quota: number; reason: string }) =>
    post<{ ok: true; balance_after: number }>(`/api/admin/users/${userId}/quota`, body),
}
