import { get, post, patch, del, type Page } from './http'

export interface Channel {
  id: number; name: string; provider: string
  base_url: string
  priority: number; weight: number
  status: 0 | 1 | 2
  group_id?: number
  tags: string[]
  extra: Record<string, unknown>
  credentials_masked?: { api_key?: string }
  created_at: string; updated_at: string
  health?: { state: 'closed' | 'half_open' | 'open'; consec_fail: number; opened_at: string | null }
  last_test?: { at: string; ok: boolean; latency_ms?: number; error?: string }
}

export interface ChannelListParams {
  page?: number; size?: number; keyword?: string
  provider?: string; status?: 0 | 1 | 2; tags?: string[]; group_id?: number
}

export interface ChannelInput {
  name: string; provider: string; base_url?: string
  credentials?: { secret?: string; region?: string; extra?: Record<string, string> }
  priority: number; weight: number
  status: 0 | 1
  group_id?: number
  tags: string[]; extra: Record<string, unknown>
}

export interface Mapping {
  id?: number; channel_id?: number
  client_model: string; upstream_model: string
  input_ratio?: number | null; output_ratio?: number | null
  cached_ratio?: number | null; reasoning_ratio?: number | null
}

export interface TestResult { ok: boolean; latency_ms?: number; error?: string; channel_id?: number; model?: string }

export const channelApi = {
  list: (p: ChannelListParams) => get<Page<Channel>>('/api/admin/channels', p as Record<string, unknown>),
  create: (b: ChannelInput) => post<Channel>('/api/admin/channels', b),
  detail: (id: number) => get<Channel>(`/api/admin/channels/${id}`),
  patch: (id: number, b: Partial<ChannelInput>) => patch<Channel>(`/api/admin/channels/${id}`, b),
  remove: (id: number) => del<{ ok: true }>(`/api/admin/channels/${id}`),
  enable: (id: number) => post<{ ok: true; status: 0 }>(`/api/admin/channels/${id}/enable`, {}),
  disable: (id: number) => post<{ ok: true; status: 1 }>(`/api/admin/channels/${id}/disable`, {}),
  test: (id: number, model?: string) => post<TestResult>(`/api/admin/channels/${id}/test`, { model }),
  batchTest: (ids: number[], model?: string) => post<{ results: TestResult[] }>('/api/admin/channels/batch_test', { ids, model }),
  health: (id: number) => get<Channel['health']>(`/api/admin/channels/${id}/health`),
  listMappings: (id: number) => get<{ items: Mapping[] }>(`/api/admin/channels/${id}/model_mappings`),
  putMappings: (id: number, mappings: Mapping[]) => post<{ items: Mapping[] }>(`/api/admin/channels/${id}/model_mappings`, { mappings }),
}
