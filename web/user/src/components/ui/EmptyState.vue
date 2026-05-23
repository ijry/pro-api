<script setup lang="ts">
interface Props {
  icon?: string
  title: string
  subtitle?: string
  cta?: string
  ctaTo?: string
}
defineProps<Props>()
const emit = defineEmits<{ (e: 'cta'): void }>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-16 text-center gap-3">
    <div v-if="icon" class="text-fg-muted mb-2">
      <span :class="[icon, 'w-12 h-12']" />
    </div>
    <h3 class="text-lg font-semibold text-fg">{{ title }}</h3>
    <p v-if="subtitle" class="text-fg-muted text-sm max-w-xs">{{ subtitle }}</p>
    <router-link
      v-if="cta && ctaTo"
      :to="ctaTo"
      class="mt-2 btn-primary"
    >{{ cta }}</router-link>
    <button
      v-else-if="cta"
      class="mt-2 btn-primary"
      @click="emit('cta')"
    >{{ cta }}</button>
  </div>
</template>
