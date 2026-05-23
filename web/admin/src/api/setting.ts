import { get, patch, post } from './http'

export interface SettingItem {
  key: string; value: unknown
  description: string
  is_sensitive: boolean
}
export interface SettingGroup { name: string; items: SettingItem[] }

export const settingApi = {
  all: () => get<{ groups: SettingGroup[] }>('/api/admin/settings'),
  one: (key: string) => get<SettingItem>(`/api/admin/settings/${key}`),
  patch: (key: string, value: unknown) => patch<SettingItem>(`/api/admin/settings/${key}`, { value }),
  testSMTP: (to: string) => post<{ ok: boolean; stubbed?: boolean; error?: string }>('/api/admin/settings/test_smtp', { to }),
}
