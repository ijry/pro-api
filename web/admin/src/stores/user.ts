import { defineStore } from 'pinia'
import { authApi, type AdminUser } from '@/api/auth'

interface State {
  user: AdminUser | null
  fetched: boolean
  fetching: boolean
}

export const useUserStore = defineStore('user', {
  state: (): State => ({ user: null, fetched: false, fetching: false }),
  getters: {
    isLoggedIn: (s) => !!s.user,
    role: (s) => s.user?.role ?? 0,
    isSuperAdmin: (s) => s.user?.role === 3,
  },
  actions: {
    async fetchMe() {
      if (this.fetching) return
      this.fetching = true
      try {
        const u = await authApi.me()
        this.user = u
      } catch (e) {
        this.user = null
        throw e
      } finally {
        this.fetched = true
        this.fetching = false
      }
    },
    async login(identity: string, password: string) {
      const r = await authApi.login({ identity, password })
      this.user = r.user
      this.fetched = true
    },
    async logout() {
      try { await authApi.logout() } catch (_) { /* ignore */ }
      this.clear()
    },
    clear() {
      this.user = null
      this.fetched = true
    },
    hasRole(min: 0 | 1 | 2 | 3) { return (this.user?.role ?? 0) >= min },
  },
})
