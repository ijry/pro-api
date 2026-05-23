import { get, patch, post, del } from './http'

export interface UserProfile {
  id: string
  email: string
  display_name: string
  avatar_url?: string
  created_at: string
}

export interface OAuthBinding {
  provider: string
  provider_user_id: string
  provider_login: string
  bound_at: string
}

export const profileApi = {
  get: () => get<UserProfile>('/api/user/profile'),
  update: (data: Partial<UserProfile>) => patch<UserProfile>('/api/user/profile', data),
  changePassword: (old_password: string, new_password: string) =>
    post('/api/user/password', { old_password, new_password }),
  listOAuthBindings: () => get<OAuthBinding[]>('/api/user/oauth/bindings'),
  bindGitHub: () => get<{ redirect_url: string }>('/api/auth/oauth/github/start'),
  unbindGitHub: () => del('/api/user/oauth/bindings/github'),
}
