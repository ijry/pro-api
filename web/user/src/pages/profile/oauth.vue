<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { profileApi, type OAuthBinding } from '@/api/profile'
import Button from '@/components/ui/Button.vue'
import Skeleton from '@/components/ui/Skeleton.vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const bindings = ref<OAuthBinding[]>([])
const loading = ref(true)
const unbinding = ref(false)

async function load() {
  loading.value = true
  try {
    bindings.value = await profileApi.listOAuthBindings()
  } catch { /* not critical */ } finally { loading.value = false }
}

onMounted(load)

const githubBinding = computed(() => bindings.value.find(b => b.provider === 'github'))

async function bindGitHub() {
  try {
    const { redirect_url } = await profileApi.bindGitHub()
    window.location.href = redirect_url
  } catch { toast.error('启动 GitHub 绑定失败') }
}

async function unbindGitHub() {
  unbinding.value = true
  try {
    await profileApi.unbindGitHub()
    toast.success('GitHub 绑定已解除')
    await load()
  } catch (e: unknown) {
    toast.error((e as { response?: { data?: { message?: string } } })?.response?.data?.message || '解绑失败')
  } finally { unbinding.value = false }
}
</script>

<template>
  <div class="max-w-lg space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-fg">OAuth 绑定</h1>
      <p class="text-sm text-fg-muted mt-1">管理第三方账号绑定</p>
    </div>

    <div class="bg-bg-elevated border border-border rounded-xl p-6">
      <div v-if="loading">
        <Skeleton class="h-16" />
      </div>
      <div v-else class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-full bg-gray-900 flex items-center justify-center">
            <span class="i-lucide-github w-5 h-5 text-white" />
          </div>
          <div>
            <p class="font-medium text-fg">GitHub</p>
            <p v-if="githubBinding" class="text-xs text-fg-muted">已绑定 · {{ githubBinding.provider_login }}</p>
            <p v-else class="text-xs text-fg-muted">未绑定</p>
          </div>
        </div>

        <Button v-if="!githubBinding" @click="bindGitHub" size="sm">绑定</Button>
        <Button v-else variant="ghost" :disabled="unbinding" @click="unbindGitHub" size="sm">
          <span v-if="unbinding" class="i-lucide-loader-circle w-4 h-4 mr-1 animate-spin" />
          解绑
        </Button>
      </div>
    </div>

    <router-link to="/profile" class="text-sm text-fg-muted hover:text-fg">← 返回个人信息</router-link>
  </div>
</template>
