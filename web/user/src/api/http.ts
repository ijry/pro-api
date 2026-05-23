import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'

let _toastFn: ((msg: string, type?: 'error' | 'warn' | 'info') => void) | null = null
export function setHttpToast(fn: (msg: string, type?: 'error' | 'warn' | 'info') => void) {
  _toastFn = fn
}

function toast(msg: string, type: 'error' | 'warn' | 'info' = 'error') {
  if (_toastFn) _toastFn(msg, type)
  else console.warn('[http]', type, msg)
}

function getCsrf(): string {
  const m = document.cookie.match(/(?:^|;\s*)proapi_csrf=([^;]+)/)
  return m ? decodeURIComponent(m[1]) : ''
}

export const http: AxiosInstance = axios.create({
  baseURL: '/',
  withCredentials: true,
  timeout: 30_000,
  headers: { 'Content-Type': 'application/json' },
})

http.interceptors.request.use((config) => {
  const csrf = getCsrf()
  if (csrf) config.headers['X-CSRF-Token'] = csrf
  return config
})

http.interceptors.response.use(
  (res) => res,
  (err) => {
    if (!err.response) {
      toast('网络错误，请检查连接')
      return Promise.reject(err)
    }
    const { status, data } = err.response
    if (status === 401) {
      const code = data?.code as string | undefined
      const redirect = encodeURIComponent(location.pathname + location.search)
      if (code === 'CodeSessionExpired') {
        location.href = `/login?reason=session_expired&redirect=${redirect}`
      } else {
        location.href = `/login?redirect=${redirect}`
      }
      return Promise.reject(err)
    }
    if (status === 429) {
      const retryAfter = err.response.headers['retry-after']
      const sec = retryAfter ? parseInt(retryAfter) : 60
      toast(`请求过于频繁，请 ${sec} 秒后重试`, 'warn')
      const e = err as any
      e.retryAfter = sec
      return Promise.reject(e)
    }
    if (status >= 500) {
      toast('服务异常，请稍后重试')
      return Promise.reject(err)
    }
    // 4xx — caller handles
    return Promise.reject(err)
  }
)

export function get<T = unknown>(url: string, config?: AxiosRequestConfig) {
  return http.get<T>(url, config).then(r => r.data)
}
export function post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) {
  return http.post<T>(url, data, config).then(r => r.data)
}
export function patch<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) {
  return http.patch<T>(url, data, config).then(r => r.data)
}
export function del<T = unknown>(url: string, config?: AxiosRequestConfig) {
  return http.delete<T>(url, config).then(r => r.data)
}
