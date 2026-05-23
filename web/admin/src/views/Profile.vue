<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NForm, NFormItem, NInput, NButton, NAvatar, useMessage } from 'naive-ui'
import { useUserStore } from '@/stores/user'
import { patch } from '@/api/http'

const message = useMessage()
const userStore = useUserStore()

const profileForm = ref({ display_name: '', email: '', avatar: '' })
const passwordForm = ref({ old_password: '', new_password: '', confirm_password: '' })
const savingProfile = ref(false)
const savingPassword = ref(false)

onMounted(() => {
  if (userStore.user) {
    profileForm.value = {
      display_name: userStore.user.display_name ?? '',
      email: userStore.user.email ?? '',
      avatar: userStore.user.avatar ?? '',
    }
  }
})

async function saveProfile() {
  savingProfile.value = true
  try {
    await patch('/api/admin/auth/me', { display_name: profileForm.value.display_name, avatar: profileForm.value.avatar })
    message.success('个人信息已保存')
    await userStore.fetchMe()
  } catch (_) { /* handled */ } finally { savingProfile.value = false }
}

async function changePassword() {
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    message.error('两次输入的密码不一致'); return
  }
  if (passwordForm.value.new_password.length < 8) {
    message.error('密码长度至少 8 位'); return
  }
  savingPassword.value = true
  try {
    await patch('/api/admin/auth/password', { old_password: passwordForm.value.old_password, new_password: passwordForm.value.new_password })
    message.success('密码已修改')
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
  } catch (_) { /* handled */ } finally { savingPassword.value = false }
}
</script>

<template>
  <div class="max-w-lg space-y-4">
    <h2 class="text-2xl font-semibold">个人资料</h2>

    <NCard title="基本信息" size="small">
      <div class="flex items-center gap-4 mb-4">
        <NAvatar :src="profileForm.avatar || undefined" :size="64" round>
          {{ (userStore.user?.username || 'A').slice(0, 1).toUpperCase() }}
        </NAvatar>
        <div>
          <p class="font-semibold">{{ userStore.user?.username }}</p>
          <p class="text-sm text-gray-400">{{ userStore.user?.email }}</p>
        </div>
      </div>
      <NForm label-placement="top">
        <NFormItem label="显示名称">
          <NInput v-model:value="profileForm.display_name" placeholder="显示名称" />
        </NFormItem>
        <NFormItem label="头像 URL">
          <NInput v-model:value="profileForm.avatar" placeholder="https://..." />
        </NFormItem>
        <NFormItem label="邮箱">
          <NInput v-model:value="profileForm.email" disabled />
        </NFormItem>
        <NButton type="primary" :loading="savingProfile" @click="saveProfile">保存</NButton>
      </NForm>
    </NCard>

    <NCard title="修改密码" size="small">
      <NForm label-placement="top">
        <NFormItem label="当前密码">
          <NInput v-model:value="passwordForm.old_password" type="password" show-password-on="click" />
        </NFormItem>
        <NFormItem label="新密码">
          <NInput v-model:value="passwordForm.new_password" type="password" show-password-on="click" />
        </NFormItem>
        <NFormItem label="确认新密码">
          <NInput v-model:value="passwordForm.confirm_password" type="password" show-password-on="click" />
        </NFormItem>
        <NButton type="primary" :loading="savingPassword" @click="changePassword">修改密码</NButton>
      </NForm>
    </NCard>
  </div>
</template>
