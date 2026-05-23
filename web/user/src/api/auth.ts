import { get, post } from './http'

export interface LoginParams { email: string; password: string }
export interface RegisterParams { email: string; password: string; code: string }
export interface ForgotParams { email: string }
export interface ResetPasswordParams { token: string; password: string }

export const authApi = {
  login: (p: LoginParams) => post('/api/auth/login', p),
  register: (p: RegisterParams) => post('/api/auth/register', p),
  logout: () => post('/api/auth/logout'),
  sendEmailCode: (email: string) => post('/api/auth/email/send_code', { email }),
  emailLogin: (email: string, code: string) => post('/api/auth/email/login', { email, code }),
  forgotPassword: (p: ForgotParams) => post('/api/auth/password/forgot', p),
  resetPassword: (p: ResetPasswordParams) => post('/api/auth/password/reset', p),
  oauthStart: (provider: string) => get<{ redirect_url: string }>(`/api/auth/oauth/${provider}/start`),
}
