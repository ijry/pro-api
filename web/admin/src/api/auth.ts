import { post, get } from './http'

export interface AdminUser {
  id: number
  username: string
  email: string | null
  display_name: string | null
  avatar: string | null
  role: 0 | 1 | 2 | 3
  status: 0 | 1 | 2
  group_id: number | null
  group_name: string
  email_verified_at: string | null
  last_login_at: string | null
  created_at: string
}

export interface LoginInput { identity: string; password: string }
export interface LoginOutput { user: AdminUser; session: { id: string; expires_at: string }; email_verify_required?: boolean }

export const authApi = {
  login: (input: LoginInput) => post<LoginOutput>('/api/admin/auth/login', input),
  logout: () => post<{ ok: true }>('/api/admin/auth/logout', {}),
  me: () => get<AdminUser>('/api/admin/auth/me'),
}
