import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { walletApi, type WalletInfo } from '@/api/wallet'

export const useWalletStore = defineStore('wallet', () => {
  const info = ref<WalletInfo | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const balanceUsd = computed(() => info.value?.balance_usd ?? 0)
  const balanceCny = computed(() => info.value?.balance_cny ?? 0)

  async function refresh() {
    loading.value = true
    error.value = null
    try {
      info.value = await walletApi.get()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'error'
    } finally {
      loading.value = false
    }
  }

  return { info, loading, error, balanceUsd, balanceCny, refresh }
})
