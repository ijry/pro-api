import { get, post, patch, del, type Page } from './http'

export interface Notice {
  id: number; title: string; content: string
  level: 'info' | 'warning' | 'danger' | 'success'
  target: 'all' | 'user' | 'admin'
  status: 0 | 1 | 2
  publish_at: string | null; expires_at: string | null
  pinned: boolean
  created_by: number; created_at: string; updated_at: string
}

export interface NoticeInput {
  title: string; content: string
  level: Notice['level']; target: Notice['target']
  publish_at?: string | null; expires_at?: string | null
  pinned?: boolean
}

export const noticeApi = {
  list: (p: { status?: 0 | 1 | 2; target?: Notice['target']; page?: number; size?: number }) =>
    get<Page<Notice>>('/api/admin/notices', p as Record<string, unknown>),
  detail: (id: number) => get<Notice>(`/api/admin/notices/${id}`),
  create: (b: NoticeInput) => post<Notice>('/api/admin/notices', b),
  patch: (id: number, b: Partial<NoticeInput>) => patch<Notice>(`/api/admin/notices/${id}`, b),
  remove: (id: number) => del<{ ok: true }>(`/api/admin/notices/${id}`),
  publish: (id: number) => post<Notice>(`/api/admin/notices/${id}/publish`, {}),
  unpublish: (id: number) => post<Notice>(`/api/admin/notices/${id}/unpublish`, {}),
}
