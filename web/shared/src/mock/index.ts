import { routeMock, type MockMethod } from './routes'

export interface MockResult<T> {
  matched: boolean
  data: T | null
}

export async function matchMock<T = unknown>(
  method: MockMethod,
  url: string,
  params?: unknown,
): Promise<MockResult<T>> {
  const delay = 100 + Math.random() * 200
  await new Promise((r) => setTimeout(r, delay))
  const { matched, data } = routeMock(method, url, params)
  return { matched, data: data as T | null }
}

export type { MockMethod } from './routes'
export type { Page } from './helpers'
