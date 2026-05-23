<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { useToast } from '@/composables/useToast'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const toast = useToast()

const email = ref('')
const password = ref('')
const loading = ref(false)
const errors = ref({ email: '', password: '' })

function validate() {
  errors.value = { email: '', password: '' }
  if (!email.value) { errors.value.email = '请输入邮箱'; return false }
  if (!password.value) { errors.value.password = '请输入密码'; return false }
  return true
}

async function onSubmit() {
  if (!validate()) return
  loading.value = true
  try {
    await authApi.login({ email: email.value, password: password.value })
    await auth.refresh()
    const redirect = route.query.redirect as string | undefined
    router.push(redirect || '/')
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { message?: string } } })?.response?.data?.message || '登录失败，请检查邮箱和密码'
    toast.error(msg)
  } finally {
    loading.value = false
  }
}

async function oauthLogin() {
  try {
    const { redirect_url } = await authApi.oauthStart('github')
    window.location.href = redirect_url
  } catch (_) {
    toast.error('GitHub 登录启动失败')
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-bg px-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-fg">ProAPI</h1>
        <p class="text-fg-muted mt-2">登录到您的账号</p>
      </div>

      <div class="bg-bg-elevated border border-border rounded-xl p-8 shadow-sm">
        <form @submit.prevent="onSubmit" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-fg mb-1">邮箱</label>
            <Input v-model="email" type="email" placeholder="you@example.com" :error="errors.email" @blur="() => { if (!email) errors.email='请输入邮箱' }" />
          </div>
          <div>
            <label class="block text-sm font-medium text-fg mb-1">密码</label>
            <Input v-model="password" type="password" placeholder="••••••••" :error="errors.password" />
          </div>

          <div class="flex justify-end">
            <router-link to="/forgot" class="text-sm text-primary hover:underline">忘记密码？</router-link>
          </div>

          <Button type="submit" :disabled="loading" class="w-full">
            <span v-if="loading" class="i-lucide-loader-circle w-4 h-4 mr-1 animate-spin" />
            登录
          </Button>
        </form>

        <div class="relative my-5">
          <div class="absolute inset-0 flex items-center"><div class="w-full border-t border-border" /></div>
          <div class="relative flex justify-center"><span class="bg-bg-elevated px-2 text-xs text-fg-muted">或通过第三方登录</span></div>
        </div>

        <button
          type="button"
          @click="oauthLogin"
          class="w-full h-10 flex items-center justify-center gap-2 rounded-md border border-border text-sm text-fg hover:bg-bg transition-colors"
        >
          <span class="i-lucide-github w-4 h-4" />
          GitHub 登录
        </button>

        <p class="mt-5 text-center text-sm text-fg-muted">
          没有账号？
          <router-link to="/register" class="text-primary hover:underline">注册</router-link>
        </p>
      </div>
    </div>
  </div>
</template>
