import { get, type Page } from './http'

export interface LedgerEntry {
  id: number; wallet_id: number
  direction: 'credit' | 'debit'
  amount_quota: number; amount_money: number; currency: 'USD' | 'CNY'
  ref_type: 'usage' | 'manual' | 'redeem' | 'refund'
  ref_id: number | null
  balance_after: number
  description: string
  created_at: string
}

export const ledgerApi = {
  mine: (p: { page?: number; size?: number; ref_type?: LedgerEntry['ref_type'] }) =>
    get<Page<LedgerEntry>>('/api/user/wallet/ledger', p as Record<string, unknown>),
}
