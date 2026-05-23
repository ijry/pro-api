<script setup lang="ts">
import { NTooltip } from 'naive-ui'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'
import { computed } from 'vue'

dayjs.extend(relativeTime)

const props = defineProps<{
  value: string | null | undefined
  relative?: boolean
}>()

const formatted = computed(() => {
  if (!props.value) return '--'
  const d = dayjs(props.value)
  if (!d.isValid()) return '--'
  return props.relative ? d.fromNow() : d.format('YYYY-MM-DD HH:mm:ss')
})

const absolute = computed(() => {
  if (!props.value) return ''
  return dayjs(props.value).format('YYYY-MM-DD HH:mm:ss')
})
</script>

<template>
  <NTooltip v-if="props.relative && props.value">
    <template #trigger>
      <span>{{ formatted }}</span>
    </template>
    {{ absolute }}
  </NTooltip>
  <span v-else>{{ formatted }}</span>
</template>
