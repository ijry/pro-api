<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Card from '@/components/ui/Card.vue'
const { t } = useI18n()
</script>

<template>
  <div class="min-h-screen flex flex-col items-center justify-center px-4 bg-gradient-to-br from-[#0f172a] via-[#0c1226] to-[#1e1b4b]">
    <div class="max-w-md w-full">
      <div class="text-center mb-6">
        <div class="text-3xl font-bold text-primary mb-1">proapi</div>
        <p class="text-fg-muted text-sm">{{ t('brand.slogan') }}</p>
      </div>
      <Card class="bg-bg-elevated/80 backdrop-blur-xl border-border/60">
        <slot />
      </Card>
      <p class="text-center text-xs text-fg-muted mt-6">© 2026 proapi · MIT</p>
    </div>
  </div>
</template>
