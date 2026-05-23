<script setup lang="ts">
import { type ModelInfo } from '@/api/model'
import Badge from '@/components/ui/Badge.vue'

defineProps<{ model: ModelInfo }>()
</script>

<template>
  <div class="rounded-xl border border-border bg-bg-elevated p-4 hover:border-primary/30 transition-colors">
    <div class="flex items-start justify-between gap-2">
      <div>
        <h4 class="font-medium text-fg text-sm">{{ model.name || model.id }}</h4>
        <p class="text-xs text-fg-muted mt-0.5">{{ model.provider }}</p>
      </div>
      <Badge variant="info" size="sm">{{ model.provider }}</Badge>
    </div>
    <p v-if="model.description" class="text-xs text-fg-muted mt-2 line-clamp-2">{{ model.description }}</p>
    <div v-if="model.input_price_per_1k !== undefined" class="mt-3 flex gap-3 text-xs text-fg-muted">
      <span>输入: ${{ model.input_price_per_1k }}/1k</span>
      <span>输出: ${{ model.output_price_per_1k }}/1k</span>
    </div>
    <div v-if="model.context_length" class="mt-1 text-xs text-fg-muted">
      上下文: {{ model.context_length.toLocaleString() }} tokens
    </div>
  </div>
</template>
