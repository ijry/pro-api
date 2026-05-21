<script setup lang="ts">
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<template>
  <main class="min-h-screen flex items-center justify-center p-6">
    <Card class="max-w-xl w-full text-center">
      <h1 class="text-3xl font-bold mb-2">{{ t('brand.name') }}</h1>
      <p class="text-fg-muted mb-6">{{ t('brand.slogan') }}</p>
      <div class="flex gap-3 justify-center">
        <Button>{{ t('home.cta.start') }}</Button>
        <Button variant="ghost">{{ t('home.cta.docs') }}</Button>
      </div>
    </Card>
  </main>
</template>
