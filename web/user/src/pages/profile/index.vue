<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { profileApi } from '@/api/profile'
import { useAuthStore } from '@/stores/auth'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { useToast } from '@/composables/useToast'

const auth = useAuthStore()
const toast = useToast()
const loading = ref(false)
const saving = ref(false)

const form = ref({ display_name: '', avatar_url: '' })

onMounted(async () => {
  loading.value = true
  try {
    const profile = await profileApi.get()
    form.value = { display_name: profile.display_name || '', avatar_url: profile.avatar_url || '' }
  } catch { toast.error('加载个人信息失败') } finally { loading.value = false }
})

async function save() {
  saving.value = true
  try {
    await profileApi.update({ display_name: form.value.display_name, avatar_url: form.value.avatar_url })
    await auth.refresh()
    toast.success('个人信息已更新')
  } catch (e: unknown) {
    toast.error((e as { response?: { data?: { message?: string } } })?.response?.data?.message || '保存失败')
  } finally { saving.value = false }
}
</script>

<template>
  <div class="max-w-lg space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-fg">个人信息</h1>
      <p class="text-sm text-fg-muted mt-1">管理您的基本信息</p>
    </div>

    <div class="bg-bg-elevated border border-border rounded-xl p-6 space-y-4">
      <div class="flex items-center gap-4 pb-4 border-b border-border">
        <div class="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center overflow-hidden">
          <img v-if="form.avatar_url" :src="form.avatar_url" alt="avatar" class="w-full h-full object-cover" />
          <span v-else class="text-2xl font-bold text-primary">
            {{ (auth.user?.display_name || auth.user?.email || 'U').slice(0,1).toUpperCase() }}
          </span>
        </div>
        <div>
          <p class="font-semibold text-fg">{{ auth.user?.display_name || '未设置' }}</p>
          <p class="text-sm text-fg-muted">{{ auth.user?.email }}</p>
        </div>
      </div>

      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-fg mb-1">显示名称</label>
          <Input v-model="form.display_name" placeholder="您的显示名称" />
        </div>
        <div>
          <label class="block text-sm font-medium text-fg mb-1">头像 URL</label>
          <Input v-model="form.avatar_url" placeholder="https://..." />
        </div>
        <div>
          <label class="block text-sm font-medium text-fg mb-1">邮箱</label>
          <Input :model-value="auth.user?.email || ''" disabled />
        </div>
      </div>

      <Button @click="save" :disabled="saving" class="w-full mt-2">
        <span v-if="saving" class="i-lucide-loader-circle w-4 h-4 mr-1 animate-spin" />
        保存
      </Button>
    </div>

    <div class="flex gap-3 text-sm">
      <router-link to="/profile/security" class="text-primary hover:underline">修改密码</router-link>
      <span class="text-border">|</span>
      <router-link to="/profile/oauth" class="text-primary hover:underline">绑定账号</router-link>
    </div>
  </div>
</template>
