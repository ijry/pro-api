export interface PageParams {
  page?: number | string
  size?: number | string
  [k: string]: unknown
}

export interface Page<T> {
  items: T[]
  total: number
  page: number
  size: number
}

export function clone<T>(v: T): T {
  return JSON.parse(JSON.stringify(v))
}

export function paginate<T>(items: T[], params?: PageParams): Page<T> {
  const page = Math.max(1, Number(params?.page ?? 1))
  const size = Math.max(1, Number(params?.size ?? 20))
  const start = (page - 1) * size
  return {
    items: clone(items.slice(start, start + size)),
    total: items.length,
    page,
    size,
  }
}

export function ok(extra?: Record<string, unknown>) {
  return { ok: true, ...(extra ?? {}) }
}
