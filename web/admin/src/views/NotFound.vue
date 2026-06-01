<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NResult, NButton } from 'naive-ui'

const router = useRouter()
const { t } = useI18n()
</script>

<template>
  <NResult status="404" :title="t('notfound.title')" :description="t('notfound.desc')">
    <template #footer>
      <NButton type="primary" @click="router.push('/')">{{ t('common.back_home') }}</NButton>
    </template>
  </NResult>
</template>
