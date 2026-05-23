import { get } from './http'

export interface LedgerEntry {
  id: string
  type: string
  amount_usd: number
  balance_after_usd: number
  description: string
  created_at: string
}

export interface LedgerResponse {
  items: LedgerEntry[]
  total: number
}

export const ledgerApi = {
  list: (page = 1, size = 20) =>
    get<LedgerResponse>(`/api/user/wallet/ledger?page=${page}&page_size=${size}`),
}
