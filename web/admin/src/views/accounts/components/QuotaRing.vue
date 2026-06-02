<script setup lang="ts">
import { computed } from 'vue'
import { NProgress, NTooltip } from 'naive-ui'
import type { QuotaWindow } from '@/api/account'

const props = defineProps<{
  label: string
  quota: QuotaWindow
  size?: number
}>()

const pct = computed(() => {
  if (typeof props.quota?.pct === 'number') return Math.round(props.quota.pct * 100)
  const total = props.quota?.total ?? 0
  const remaining = props.quota?.remaining ?? 0
  if (!total) return null
  return Math.round((remaining / total) * 100)
})

const status = computed<'default' | 'success' | 'warning' | 'error'>(() => {
  if (pct.value == null) return 'default'
  if (pct.value <= 10) return 'error'
  if (pct.value <= 30) return 'warning'
  return 'success'
})

const resetCountdown = computed(() => {
  if (!props.quota?.reset_at) return ''
  const t = new Date(props.quota.reset_at).getTime() - Date.now()
  if (t <= 0) return ''
  const h = Math.floor(t / 3600_000)
  const m = Math.floor((t % 3600_000) / 60_000)
  return h > 0 ? `${h}h${m}m` : `${m}m`
})

const tooltip = computed(() => {
  const parts: string[] = []
  if (props.quota?.remaining != null && props.quota?.total != null) {
    parts.push(`${props.quota.remaining} / ${props.quota.total}`)
  }
  if (props.quota?.reset_at) {
    parts.push(`reset: ${new Date(props.quota.reset_at).toLocaleString()}`)
  }
  return parts.join(' · ') || '--'
})
</script>

<template>
  <NTooltip placement="top">
    <template #trigger>
      <div class="quota-ring" :style="{ width: (size ?? 96) + 'px' }">
        <NProgress
          type="circle"
          :percentage="pct ?? 0"
          :status="status"
          :stroke-width="10"
          :style="{ width: (size ?? 96) + 'px' }"
        >
          <span v-if="pct != null" class="text-sm font-semibold">{{ pct }}%</span>
          <span v-else class="text-xs text-gray-400">--</span>
        </NProgress>
        <div class="text-xs text-center mt-1 text-gray-500">{{ label }}</div>
        <div v-if="resetCountdown" class="text-xs text-center text-gray-400">reset {{ resetCountdown }}</div>
      </div>
    </template>
    {{ tooltip }}
  </NTooltip>
</template>

<style scoped>
.quota-ring {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
}
</style>
