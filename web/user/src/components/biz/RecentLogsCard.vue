<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { logApi, type LogEntry } from '@/api/log'
import Skeleton from '@/components/ui/Skeleton.vue'
import EmptyState from '@/components/ui/EmptyState.vue'

interface Props { limit?: number }
const props = withDefaults(defineProps<Props>(), { limit: 5 })
const { t } = useI18n()

const items = ref<LogEntry[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const r = await logApi.list({ page_size: props.limit })
    items.value = r.items
  } finally {
    loading.value = false
  }
})

function statusColor(status: number) {
  if (status >= 200 && status < 300) return 'text-emerald-400'
  return 'text-rose-400'
}
function statusIcon(status: number) {
  return status >= 200 && status < 300 ? 'i-lucide-check-circle' : 'i-lucide-x-circle'
}
function formatLatency(ms: number) {
  return ms >= 1000 ? `${(ms / 1000).toFixed(1)}s` : `${ms}ms`
}
</script>

<template>
  <div class="rounded-xl border border-border bg-bg-elevated backdrop-blur-md p-5 shadow-sm">
    <div class="flex items-center justify-between mb-4">
      <h3 class="font-semibold text-fg">{{ t('home.logs.title') }}</h3>
      <router-link to="/logs" class="text-xs text-primary hover:underline">{{ t('home.logs.view_all') }}</router-link>
    </div>
    <div v-if="loading" class="space-y-2">
      <Skeleton v-for="i in 5" :key="i" class="h-10" />
    </div>
    <EmptyState
      v-else-if="!items.length"
      icon="i-lucide-inbox"
      :title="t('home.logs.empty.title')"
      :subtitle="t('home.logs.empty.subtitle')"
      :cta="t('home.logs.empty.cta')"
      cta-to="/tokens"
    />
    <div v-else class="space-y-1">
      <div
        v-for="item in items"
        :key="item.id"
        class="flex items-center gap-3 py-2 px-2 rounded-md hover:bg-bg transition-colors text-sm"
      >
        <span :class="[statusIcon(item.status), statusColor(item.status), 'w-4 h-4 shrink-0']" />
        <span class="text-fg flex-1 truncate">{{ item.model }}</span>
        <span class="text-fg-muted text-xs">$ {{ item.cost_usd.toFixed(4) }}</span>
        <span class="text-fg-muted text-xs">{{ formatLatency(item.latency_ms) }}</span>
      </div>
    </div>
  </div>
</template>
