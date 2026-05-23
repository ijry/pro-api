<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  quota: number
  currency?: 'CNY' | 'USD'
  showBoth?: boolean
}>()

// 1 quota = 0.000002 CNY (based on spec: 1 CNY = 500000 quota)
const QUOTA_PER_CNY = 500_000

const moneyValue = computed(() => props.quota / QUOTA_PER_CNY)
const symbol = computed(() => props.currency === 'USD' ? '$' : '¥')

const formatted = computed(() => {
  if (props.showBoth) {
    return `${symbol.value}${moneyValue.value.toFixed(4)} (${props.quota.toLocaleString()} quota)`
  }
  return `${props.quota.toLocaleString()} quota`
})
</script>

<template>
  <span>{{ formatted }}</span>
</template>
