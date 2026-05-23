import { defineStore } from 'pinia'
import { ref } from 'vue'
import { profileApi, type UserProfile } from '@/api/profile'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserProfile | null>(null)
  const loading = ref(false)

  async function refresh() {
    loading.value = true
    try {
      user.value = await profileApi.get()
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    const { authApi } = await import('@/api/auth')
    await authApi.logout().catch(() => {})
    user.value = null
  }

  function clear() {
    user.value = null
  }

  return { user, loading, refresh, logout, clear }
})
