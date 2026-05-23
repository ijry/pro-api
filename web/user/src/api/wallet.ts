import { get } from './http'

export interface WalletInfo {
  balance_usd: number
  balance_cny: number
  total_recharged_usd: number
  total_consumed_usd: number
  currency: string
}

export const walletApi = {
  get: () => get<WalletInfo>('/api/user/wallet'),
}
