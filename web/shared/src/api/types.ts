/**
 * 与后端 /api/* 接口对齐的通用类型。
 * 代理 API(/v1/*)不走此类型,直接透传上游协议。
 */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
  details?: Record<string, unknown>
}

export interface Pagination<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface ApiError {
  code: number
  message: string
  details?: Record<string, unknown>
}
