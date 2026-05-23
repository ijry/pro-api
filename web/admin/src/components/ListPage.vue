<script setup lang="ts">
import { NPageHeader, NCard, NSpace } from 'naive-ui'

defineProps<{
  title: string
}>()
</script>

<template>
  <div>
    <NPageHeader :title="title" class="mb-3">
      <template #extra>
        <slot name="actions" />
      </template>
    </NPageHeader>
    <NCard size="small" :bordered="false">
      <NSpace v-if="$slots.filters" align="end" wrap class="mb-3">
        <slot name="filters" />
      </NSpace>
      <slot />
    </NCard>
  </div>
</template>
