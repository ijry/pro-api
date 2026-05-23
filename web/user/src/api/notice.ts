import { get, post } from './http'

export interface Notice {
  id: string
  title: string
  content: string
  is_pinned: boolean
  is_published: boolean
  created_at: string
  read?: boolean
}

export interface NoticeListResponse {
  items: Notice[]
  total: number
}

export const noticeApi = {
  list: (page = 1, size = 20) =>
    get<NoticeListResponse>(`/api/user/notices?page=${page}&page_size=${size}`),
  get: (id: string) => get<Notice>(`/api/user/notices/${id}`),
  markRead: (id: string) => post(`/api/user/notices/${id}/read`),
  unreadCount: () => get<{ count: number }>('/api/user/notices/unread_count'),
  publicList: (page = 1, size = 20) =>
    get<NoticeListResponse>(`/api/public/notices?page=${page}&page_size=${size}`),
}
