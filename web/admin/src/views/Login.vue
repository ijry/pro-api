<script setup lang="ts">
import { ref } from 'vue'
import { NCard, NForm, NFormItem, NInput, NButton } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const form = ref({ email: '', password: '' })

function submit() {
  // M1 起接入 /api/admin/auth/login
  // eslint-disable-next-line no-console
  console.log('login submitted (M0 stub)', form.value)
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-bg p-4">
    <NCard class="w-full max-w-md" :title="t('login.title')">
      <template #header-extra>
        <span class="text-fg-muted text-sm">{{ t('login.subtitle') }}</span>
      </template>
      <NForm label-placement="top">
        <NFormItem :label="t('login.email')">
          <NInput v-model:value="form.email" placeholder="admin@example.com" />
        </NFormItem>
        <NFormItem :label="t('login.password')">
          <NInput v-model:value="form.password" type="password" show-password-on="click" />
        </NFormItem>
        <NButton type="primary" block @click="submit">{{ t('login.submit') }}</NButton>
      </NForm>
    </NCard>
  </div>
</template>
