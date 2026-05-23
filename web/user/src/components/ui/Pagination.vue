<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  total: number
  size?: number
}
const props = withDefaults(defineProps<Props>(), { size: 20 })
const page = defineModel<number>({ default: 1 })

const totalPages = computed(() => Math.ceil(props.total / props.size))
const pages = computed(() => {
  const p = page.value
  const t = totalPages.value
  if (t <= 7) return Array.from({ length: t }, (_, i) => i + 1)
  if (p <= 4) return [1, 2, 3, 4, 5, '...', t]
  if (p >= t - 3) return [1, '...', t - 4, t - 3, t - 2, t - 1, t]
  return [1, '...', p - 1, p, p + 1, '...', t]
})
</script>

<template>
  <div v-if="totalPages > 1" class="flex items-center gap-1 mt-4">
    <button
      class="px-2 h-8 rounded border border-border text-fg-muted hover:text-fg hover:bg-bg-elevated disabled:opacity-40 text-sm"
      :disabled="page <= 1"
      @click="page--"
    >
      <span class="i-lucide-chevron-left w-4 h-4" />
    </button>
    <button
      v-for="p in pages"
      :key="p"
      class="min-w-8 h-8 rounded border text-sm transition-colors"
      :class="p === page
        ? 'border-primary bg-primary text-white'
        : p === '...'
          ? 'border-transparent text-fg-muted cursor-default'
          : 'border-border text-fg hover:bg-bg-elevated'"
      :disabled="p === '...'"
      @click="typeof p === 'number' && (page = p)"
    >{{ p }}</button>
    <button
      class="px-2 h-8 rounded border border-border text-fg-muted hover:text-fg hover:bg-bg-elevated disabled:opacity-40 text-sm"
      :disabled="page >= totalPages"
      @click="page++"
    >
      <span class="i-lucide-chevron-right w-4 h-4" />
    </button>
  </div>
</template>
