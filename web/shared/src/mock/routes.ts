import type { PageParams } from './helpers'

export type MockMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE'
export type MockHandler = (method: MockMethod, url: string, params?: unknown) => unknown

export interface MockRoute {
  pattern: RegExp
  handler: MockHandler
  methods?: MockMethod[]
}

export const routes: MockRoute[] = []

export function routeMock(method: MockMethod, url: string, params?: unknown):
  { matched: boolean; data: unknown } {
  const path = url.split('?')[0]
  for (const r of routes) {
    if (r.methods && !r.methods.includes(method)) continue
    if (r.pattern.test(path)) {
      return { matched: true, data: r.handler(method, url, params) }
    }
  }
  return { matched: false, data: null }
}

export type { PageParams }
