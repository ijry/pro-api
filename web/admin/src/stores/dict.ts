import { defineStore } from 'pinia'
import { modelApi, type ModelCatalog } from '@/api/model'
import { channelApi, type Channel } from '@/api/channel'
import { groupApi, type Group } from '@/api/group'

interface CachedList<T> { data: T[]; fetchedAt: number; loading: boolean; promise: Promise<void> | null }
const empty = <T>(): CachedList<T> => ({ data: [], fetchedAt: 0, loading: false, promise: null })

const TTL = 5 * 60 * 1000

export const useDictStore = defineStore('dict', {
  state: () => ({
    models: empty<ModelCatalog>(),
    channels: empty<Channel>(),
    groups: empty<Group>(),
  }),
  getters: {
    modelOptions: (s) => s.models.data.map((m) => ({ label: m.name, value: m.name })),
    groupOptions: (s) => s.groups.data.map((g) => ({ label: g.display_name || g.name, value: g.id })),
    channelOptions: (s) => s.channels.data.map((c) => ({ label: c.name, value: c.id })),
  },
  actions: {
    async ensureModels() { await this.ensure('models', async () => (await modelApi.list({ page: 1, size: 1000 })).items) },
    async ensureChannels() { await this.ensure('channels', async () => (await channelApi.list({ page: 1, size: 1000 })).items) },
    async ensureGroups() { await this.ensure('groups', async () => { const r = await groupApi.list(); return r.items }) },
    async refreshAll() {
      this.models.fetchedAt = 0; this.channels.fetchedAt = 0; this.groups.fetchedAt = 0
      await Promise.all([this.ensureModels(), this.ensureChannels(), this.ensureGroups()])
    },
    async ensure<K extends 'models' | 'channels' | 'groups'>(key: K, fetcher: () => Promise<unknown[]>) {
      const c: CachedList<unknown> = (this as unknown as Record<string, CachedList<unknown>>)[key]
      const now = Date.now()
      if (now - c.fetchedAt < TTL && c.data.length) return
      if (c.promise) { await c.promise; return }
      c.loading = true
      c.promise = (async () => {
        try { c.data = await fetcher(); c.fetchedAt = now }
        finally { c.loading = false; c.promise = null }
      })()
      await c.promise
    },
  },
})
