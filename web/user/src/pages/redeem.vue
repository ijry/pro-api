<script setup lang="ts">
import { ref } from 'vue'
import { redeemApi } from '@/api/redeem'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'

const toast = useToast()
const auth = useAuthStore()

const code = ref('')
const loading = ref(false)
const result = ref<{ granted_usd: number; code: string } | null>(null)
const error = ref('')

async function redeem() {
  const trimmed = code.value.trim()
  if (!trimmed) { error.value = '请输入兑换码'; return }
  error.value = ''
  loading.value = true
  try {
    result.value = await redeemApi.redeem(trimmed)
    toast.success(`兑换成功！已获得 $${result.value.granted_usd} 额度`)
    code.value = ''
    await auth.refresh()
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { message?: string } } })?.response?.data?.message || '兑换失败，请检查兑换码'
    error.value = msg
  } finally { loading.value = false }
}
</script>

<template>
  <div class="max-w-md mx-auto space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-fg">兑换码</h1>
      <p class="text-sm text-fg-muted mt-1">输入兑换码以获得账户额度</p>
    </div>

    <div class="bg-bg-elevated border border-border rounded-xl p-6 space-y-4">
      <div>
        <label class="block text-sm font-medium text-fg mb-1">兑换码</label>
        <Input
          v-model="code"
          placeholder="PROA-XXXX-XXXX-XXXX"
          :error="error"
          @keydown.enter="redeem"
          class="font-mono uppercase"
        />
      </div>

      <Button @click="redeem" :disabled="loading" class="w-full">
        <span v-if="loading" class="i-lucide-loader-circle w-4 h-4 mr-1 animate-spin" />
        <span class="i-lucide-gift w-4 h-4 mr-1" />
        兑换
      </Button>
    </div>

    <!-- Success result -->
    <div v-if="result" class="bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800 rounded-xl p-5 text-center">
      <span class="i-lucide-party-popper w-8 h-8 text-green-500 mx-auto block mb-2" />
      <p class="font-semibold text-green-700 dark:text-green-400">兑换成功！</p>
      <p class="text-sm text-green-600 dark:text-green-500 mt-1">已获得 <span class="font-bold">${{ result.granted_usd }}</span> 额度</p>
    </div>
  </div>
</template>
