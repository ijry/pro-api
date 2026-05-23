<script setup lang="ts">
import { NTooltip } from 'naive-ui'
import CopyButton from './CopyButton.vue'
import { computed } from 'vue'

const props = defineProps<{
  value: string | null | undefined
  show?: boolean
}>()

const masked = computed(() => {
  if (!props.value) return '--'
  if (props.value.length <= 12) return props.value.slice(0, 4) + '****'
  return props.value.slice(0, 8) + '****' + props.value.slice(-4)
})
</script>

<template>
  <span class="inline-flex items-center gap-1 font-mono text-sm">
    <NTooltip v-if="props.value">
      <template #trigger>{{ masked }}</template>
      {{ props.value }}
    </NTooltip>
    <span v-else>--</span>
    <CopyButton v-if="props.value" :value="props.value" size="tiny" />
  </span>
</template>
