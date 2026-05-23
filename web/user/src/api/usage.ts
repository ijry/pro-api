import { get } from './http'

export interface UsageStat {
  cost_usd: number
  request_count: number
  token_count: number
  range: string
}

export const usageApi = {
  get: (range: 'today' | 'month' | 'total') =>
    get<UsageStat>(`/api/user/usage?range=${range}`),
}
