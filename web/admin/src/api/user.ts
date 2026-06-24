import { get, patch, post, del, type Page } from './http'
import type { AdminUser } from './auth'

export interface UserListParams {
  page?: number; size?: number; keyword?: string
  role?: 0 | 1 | 2 | 3
  status?: 0 | 1 | 2
  order?: string
}

export interface UserDetailResponse {
  user: AdminUser
  wallet: { id: number; quota_balance: number; quota_total_recharged: number; quota_total_consumed: number; currency: string } | null
  bindings: Array<{ provider: string; uid: string; email: string; bound_at: string }>
}

export interface UserPatch {
  display_name?: string
  role?: 0 | 1 | 2 | 3
  status?: 0 | 1 | 2
}

export interface ResetPasswordResp { ok: true; temp_password?: string }
export interface QuotaAdjustResp { ok: true; balance_after: number }

export const userApi = {
  list: (p: UserListParams) => get<Page<AdminUser & { wallet?: { balance: number } }>>('/api/admin/users', p as Record<string, unknown>),
  detail: (id: number) => get<UserDetailResponse>(`/api/admin/users/${id}`),
  patch: (id: number, p: UserPatch) => patch<AdminUser>(`/api/admin/users/${id}`, p),
  resetPassword: (id: number, newPassword?: string) =>
    post<ResetPasswordResp>(`/api/admin/users/${id}/reset_password`, { new_password: newPassword }),
  adjustQuota: (id: number, body: { delta_quota: number; reason: string }) =>
    post<QuotaAdjustResp>(`/api/admin/users/${id}/quota`, body),
  remove: (id: number) => del<{ ok: true }>(`/api/admin/users/${id}`),
}
