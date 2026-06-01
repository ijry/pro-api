import { get } from './http'

export interface InviteSummary {
  invite_code: string
  share_url: string
  rebate_ratio: number
  stats: {
    invitee_count: number
    rebate_credits_total: number
    rebate_credits_month: number
  }
}

export interface Invitee {
  user_id: number
  display_name: string
  email_masked: string
  registered_at: string
  total_rebate_credits: number
}

export interface InviteRecord {
  id: number
  invitee_id: number
  invitee_display_name: string
  order_id: number
  rebate_cents: number
  rebate_credits: number
  created_at: string
}

export interface PageResp<T> {
  items: T[]
  total: number
  page: number
  size: number
}

export const inviteApi = {
  getSummary: () => get<InviteSummary>('/api/user/invites/me'),
  listInvitees: (page = 1, size = 10) =>
    get<PageResp<Invitee>>(`/api/user/invites/invitees?page=${page}&size=${size}`),
  listRecords: (page = 1, size = 10) =>
    get<PageResp<InviteRecord>>(`/api/user/invites/records?page=${page}&size=${size}`),
}
