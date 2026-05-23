import axios, { type AxiosInstance, type AxiosError, type AxiosRequestConfig } from 'axios'
import qs from 'qs'

declare global {
  interface Window {
    $message?: { error: (msg: string) => void; warning: (msg: string) => void; success: (msg: string) => void }
    $loadingBar?: { start: () => void; finish: () => void; error: () => void }
  }
}

const csrfTokenFromCookie = () =>
  document.cookie
    .split('; ')
    .find((c) => c.startsWith('proapi_csrf='))
    ?.split('=')[1] ?? ''

const isUnsafeMethod = (m?: string) =>
  ['post', 'put', 'patch', 'delete'].includes((m || 'get').toLowerCase())

export const http: AxiosInstance = axios.create({
  baseURL: '/',
  timeout: 30_000,
  withCredentials: true,
  paramsSerializer: { serialize: (p) => qs.stringify(p, { arrayFormat: 'brackets', skipNulls: true }) },
})

http.interceptors.request.use((cfg) => {
  if (isUnsafeMethod(cfg.method)) {
    const t = csrfTokenFromCookie()
    if (t) cfg.headers.set('X-CSRF-Token', t)
  }
  window.$loadingBar?.start()
  return cfg
})

http.interceptors.response.use(
  (resp) => {
    window.$loadingBar?.finish()
    return resp
  },
  (err: AxiosError<{ code: number; message: string; details?: unknown }>) => {
    window.$loadingBar?.error()
    const status = err.response?.status
    const message = err.response?.data?.message || err.message

    if (status === 401) {
      // Lazy import to avoid circular dependency
      import('@/stores/user').then(({ useUserStore }) => {
        const us = useUserStore()
        us.clear()
      })
      import('@/router').then(({ router }) => {
        if (router.currentRoute.value.name !== 'login') {
          router.push({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } })
        }
      })
      return Promise.reject(err)
    }
    if (status === 403) {
      window.$message?.error(message || '权限不足')
      return Promise.reject(err)
    }
    if (status === 429) {
      const retry = err.response?.headers?.['retry-after']
      window.$message?.warning(retry ? `请求过快，${String(retry)}s 后再试` : '请求过快，请稍后再试')
      return Promise.reject(err)
    }
    if (status && status >= 500) {
      window.$message?.error('服务器错误，请稍后重试')
      return Promise.reject(err)
    }
    window.$message?.error(message || '未知错误')
    return Promise.reject(err)
  }
)

export type Page<T> = { items: T[]; total: number; page: number; size: number }
export type Ok = { ok: true }

export async function get<T>(url: string, params?: Record<string, unknown>, cfg?: AxiosRequestConfig) {
  return http.get<T>(url, { params, ...cfg }).then((r) => r.data)
}
export async function post<T>(url: string, body?: unknown, cfg?: AxiosRequestConfig) {
  return http.post<T>(url, body, cfg).then((r) => r.data)
}
export async function patch<T>(url: string, body?: unknown, cfg?: AxiosRequestConfig) {
  return http.patch<T>(url, body, cfg).then((r) => r.data)
}
export async function del<T>(url: string, params?: Record<string, unknown>, cfg?: AxiosRequestConfig) {
  return http.delete<T>(url, { params, ...cfg }).then((r) => r.data)
}
