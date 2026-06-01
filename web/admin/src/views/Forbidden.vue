<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NResult, NButton, NSpace } from 'naive-ui'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const { t } = useI18n()
const userStore = useUserStore()

async function onRelogin() {
  await userStore.logout()
  router.push('/login')
}
</script>

<template>
  <NResult status="403" :title="t('forbidden.title')" :description="t('forbidden.desc')">
    <template #footer>
      <NSpace>
        <NButton @click="router.push('/')">{{ t('common.back_home') }}</NButton>
        <NButton type="primary" @click="onRelogin">{{ t('auth.relogin') }}</NButton>
      </NSpace>
    </template>
  </NResult>
</template>
