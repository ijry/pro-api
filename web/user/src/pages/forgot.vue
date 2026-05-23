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
const code = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const sendingCode = ref(false)
const loading = ref(false)
const countdown = ref(0)
const errors = ref({ email: '', code: '', newPassword: '', confirmPassword: '' })

let timer: ReturnType<typeof setInterval> | null = null

async function sendCode() {
  if (!email.value) { errors.value.email = '请输入邮箱'; return }
  sendingCode.value = true
  try {
    await authApi.sendEmailCode(email.value)
    countdown.value = 60
    timer = setInterval(() => { countdown.value--; if (countdown.value <= 0 && timer) { clearInterval(timer); timer = null } }, 1000)
    toast.success('验证码已发送')
  } catch (e: unknown) {
    toast.error((e as { response?: { data?: { message?: string } } })?.response?.data?.message || '发送失败')
  } finally { sendingCode.value = false }
}

function validate() {
  errors.value = { email: '', code: '', newPassword: '', confirmPassword: '' }
  if (!email.value) { errors.value.email = '请输入邮箱'; return false }
  if (!code.value) { errors.value.code = '请输入验证码'; return false }
  if (!newPassword.value || newPassword.value.length < 8) { errors.value.newPassword = '密码至少 8 位'; return false }
  if (newPassword.value !== confirmPassword.value) { errors.value.confirmPassword = '两次密码不一致'; return false }
  return true
}

async function onSubmit() {
  if (!validate()) return
  loading.value = true
  try {
    await authApi.resetPassword({ token: code.value, password: newPassword.value })
    toast.success('密码已重置，请登录')
    router.push('/login')
  } catch (e: unknown) {
    toast.error((e as { response?: { data?: { message?: string } } })?.response?.data?.message || '重置失败')
  } finally { loading.value = false }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-bg px-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-fg">重置密码</h1>
        <p class="text-fg-muted mt-2">输入邮箱和验证码来重置密码</p>
      </div>

      <div class="bg-bg-elevated border border-border rounded-xl p-8 shadow-sm">
        <form @submit.prevent="onSubmit" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-fg mb-1">邮箱</label>
            <Input v-model="email" type="email" placeholder="you@example.com" :error="errors.email" />
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
              >{{ countdown > 0 ? `${countdown}s` : '发送验证码' }}</button>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-fg mb-1">新密码</label>
            <Input v-model="newPassword" type="password" placeholder="至少 8 位" :error="errors.newPassword" />
          </div>
          <div>
            <label class="block text-sm font-medium text-fg mb-1">确认新密码</label>
            <Input v-model="confirmPassword" type="password" placeholder="再次输入新密码" :error="errors.confirmPassword" />
          </div>

          <Button type="submit" :disabled="loading" class="w-full">
            <span v-if="loading" class="i-lucide-loader-circle w-4 h-4 mr-1 animate-spin" />
            重置密码
          </Button>
        </form>

        <p class="mt-5 text-center text-sm text-fg-muted">
          想起密码了？<router-link to="/login" class="text-primary hover:underline">返回登录</router-link>
        </p>
      </div>
    </div>
  </div>
</template>
