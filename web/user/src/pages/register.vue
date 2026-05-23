<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '@/api/auth'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { useToast } from '@/composables/useToast'

const router = useRouter()
const toast = useToast()

const email = ref('')
const password = ref('')
const code = ref('')
const sendingCode = ref(false)
const loading = ref(false)
const errors = ref({ email: '', password: '', code: '' })
const codeSent = ref(false)
const countdown = ref(0)

let timer: ReturnType<typeof setInterval> | null = null

async function sendCode() {
  if (!email.value) { errors.value.email = '请输入邮箱'; return }
  sendingCode.value = true
  try {
    await authApi.sendEmailCode(email.value)
    codeSent.value = true
    countdown.value = 60
    timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0 && timer) { clearInterval(timer); timer = null }
    }, 1000)
    toast.success('验证码已发送，请查收邮件')
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { message?: string } } })?.response?.data?.message || '发送失败'
    toast.error(msg)
  } finally {
    sendingCode.value = false
  }
}

function validate() {
  errors.value = { email: '', password: '', code: '' }
  if (!email.value) { errors.value.email = '请输入邮箱'; return false }
  if (!password.value || password.value.length < 8) { errors.value.password = '密码至少 8 位'; return false }
  if (!code.value) { errors.value.code = '请输入验证码'; return false }
  return true
}

async function onSubmit() {
  if (!validate()) return
  loading.value = true
  try {
    await authApi.register({ email: email.value, password: password.value, code: code.value })
    toast.success('注册成功，请登录')
    router.push('/login')
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { message?: string } } })?.response?.data?.message || '注册失败'
    toast.error(msg)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-bg px-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-fg">ProAPI</h1>
        <p class="text-fg-muted mt-2">创建新账号</p>
      </div>

      <div class="bg-bg-elevated border border-border rounded-xl p-8 shadow-sm">
        <form @submit.prevent="onSubmit" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-fg mb-1">邮箱</label>
            <Input v-model="email" type="email" placeholder="you@example.com" :error="errors.email" />
          </div>
          <div>
            <label class="block text-sm font-medium text-fg mb-1">密码</label>
            <Input v-model="password" type="password" placeholder="至少 8 位" :error="errors.password" />
          </div>
          <div>
            <label class="block text-sm font-medium text-fg mb-1">验证码</label>
            <div class="flex gap-2">
              <Input v-model="code" placeholder="6位验证码" :error="errors.code" class="flex-1" />
              <button
                type="button"
                :disabled="sendingCode || countdown > 0"
                @click="sendCode"
                class="px-3 h-10 rounded-md border border-border text-sm text-fg hover:bg-bg transition-colors whitespace-nowrap disabled:opacity-50"
              >
                {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
              </button>
            </div>
          </div>

          <Button type="submit" :disabled="loading" class="w-full">
            <span v-if="loading" class="i-lucide-loader-circle w-4 h-4 mr-1 animate-spin" />
            注册
          </Button>
        </form>

        <p class="mt-5 text-center text-sm text-fg-muted">
          已有账号？
          <router-link to="/login" class="text-primary hover:underline">登录</router-link>
        </p>
      </div>
    </div>
  </div>
</template>
