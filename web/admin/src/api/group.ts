import { get, post, patch, del } from './http'

export interface Group {
  id: number; name: string; display_name: string
  ratio: number; priority: number; status: 0 | 1
  created_at: string; updated_at: string
}

export interface GroupInput { name: string; display_name: string; ratio: number; priority: number }

export const groupApi = {
  list: () => get<{ items: Group[] }>('/api/admin/groups'),
  create: (b: GroupInput) => post<Group>('/api/admin/groups', b),
  patch: (id: number, b: Partial<GroupInput> & { status?: 0 | 1 }) => patch<Group>(`/api/admin/groups/${id}`, b),
  remove: (id: number) => del<{ ok: true }>(`/api/admin/groups/${id}`),
}
