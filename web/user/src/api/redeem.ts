import { post } from './http'

export interface RedeemResult {
  granted_usd: number
  code: string
}

export const redeemApi = {
  redeem: (code: string) => post<RedeemResult>('/api/user/payment/redeem', { code }),
}
