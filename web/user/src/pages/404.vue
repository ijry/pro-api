<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'

const router = useRouter()
const { t } = useI18n()

const goHome = () => {
  router.push('/')
}
</script>

<template>
  <Card class="w-full max-w-md space-y-6 text-center">
    <div class="text-6xl font-bold text-primary">404</div>
    <h2 class="text-fg">{{ t('notfound.title') }}</h2>
    <p class="text-fg-muted">{{ t('notfound.desc') }}</p>
    <Button @click="goHome">{{ t('common.back_home') }}</Button>
  </Card>
</template>
