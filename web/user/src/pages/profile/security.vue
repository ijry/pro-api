<script setup lang="ts">
import { ref } from 'vue'
import { profileApi } from '@/api/profile'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const saving = ref(false)
const form = ref({ old_password: '', new_password: '', confirm_password: '' })
const errors = ref({ old_password: '', new_password: '', confirm_password: '' })

function validate() {
  errors.value = { old_password: '', new_password: '', confirm_password: '' }
  if (!form.value.old_password) { errors.value.old_password = '请输入当前密码'; return false }
  if (form.value.new_password.length < 8) { errors.value.new_password = '新密码至少 8 位'; return false }
  if (form.value.new_password !== form.value.confirm_password) { errors.value.confirm_password = '两次密码不一致'; return false }
  return true
}

async function save() {
  if (!validate()) return
  saving.value = true
  try {
    await profileApi.changePassword(form.value.old_password, form.value.new_password)
    toast.success('密码已修改')
    form.value = { old_password: '', new_password: '', confirm_password: '' }
  } catch (e: unknown) {
    toast.error((e as { response?: { data?: { message?: string } } })?.response?.data?.message || '修改失败')
  } finally { saving.value = false }
}
</script>

<template>
  <div class="max-w-lg space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-fg">修改密码</h1>
      <p class="text-sm text-fg-muted mt-1">更新您的登录密码</p>
    </div>

    <div class="bg-bg-elevated border border-border rounded-xl p-6 space-y-4">
      <div>
        <label class="block text-sm font-medium text-fg mb-1">当前密码</label>
        <Input v-model="form.old_password" type="password" placeholder="••••••••" :error="errors.old_password" />
      </div>
      <div>
        <label class="block text-sm font-medium text-fg mb-1">新密码</label>
        <Input v-model="form.new_password" type="password" placeholder="至少 8 位" :error="errors.new_password" />
      </div>
      <div>
        <label class="block text-sm font-medium text-fg mb-1">确认新密码</label>
        <Input v-model="form.confirm_password" type="password" placeholder="再次输入新密码" :error="errors.confirm_password" />
      </div>

      <Button @click="save" :disabled="saving" class="w-full mt-2">
        <span v-if="saving" class="i-lucide-loader-circle w-4 h-4 mr-1 animate-spin" />
        修改密码
      </Button>
    </div>

    <router-link to="/profile" class="text-sm text-fg-muted hover:text-fg">← 返回个人信息</router-link>
  </div>
</template>
