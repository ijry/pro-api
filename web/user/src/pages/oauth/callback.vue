<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { get } from '@/api/http'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const toast = useToast()

onMounted(async () => {
  const provider = route.params.provider as string
  const code = route.query.code as string
  const state = route.query.state as string

  if (!code) {
    toast.error('OAuth 回调缺少 code 参数')
    router.push('/login')
    return
  }

  try {
    await get(`/api/auth/oauth/${provider}/callback`, { params: { code, state } })
    await auth.refresh()
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { message?: string } } })?.response?.data?.message || 'OAuth 登录失败'
    toast.error(msg)
    router.push('/login')
  }
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-bg">
    <div class="text-center space-y-3">
      <span class="i-lucide-loader-circle w-10 h-10 animate-spin text-primary block mx-auto" />
      <p class="text-fg-muted">正在处理登录，请稍候...</p>
    </div>
  </div>
</template>
