<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { NTimeline, NTimelineItem, NPagination, NSpin, NEmpty, NTag } from 'naive-ui'
import { accountApi, type AccountEvent } from '@/api/account'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ accountId: number }>()
const { t } = useI18n()

const loading = ref(false)
const items = ref<AccountEvent[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(20)

async function load() {
  loading.value = true
  try {
    const res = await accountApi.events(props.accountId, page.value, size.value)
    items.value = res.items
    total.value = res.total
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

const eventTypeColor: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
  enabled: 'success', disabled: 'warning',
  test_ok: 'success', test_fail: 'error',
  refreshed: 'info', refresh_failed: 'error',
  cooldown_started: 'warning', cooldown_cleared: 'info',
  imported: 'info', deleted: 'error',
}

function lineFor(ev: AccountEvent) {
  return eventTypeColor[ev.event_type] ?? 'default'
}

function fmt(s: string) {
  try { return new Date(s).toLocaleString() } catch { return s }
}

const formattedPayload = computed(() => (ev: AccountEvent) => {
  if (ev.payload == null) return ''
  if (typeof ev.payload === 'string') return ev.payload
  try { return JSON.stringify(ev.payload) } catch { return String(ev.payload) }
})
</script>

<template>
  <div>
    <NSpin :show="loading">
      <NEmpty v-if="!loading && items.length === 0" :description="t('accounts.detail.no_events')" />
      <NTimeline v-else>
        <NTimelineItem
          v-for="ev in items"
          :key="ev.id"
          :type="lineFor(ev)"
          :title="ev.event_type"
          :time="fmt(ev.created_at)"
        >
          <NTag size="tiny" :type="lineFor(ev)">#{{ ev.id }}</NTag>
          <span class="ml-2 text-xs text-gray-500">{{ formattedPayload(ev) }}</span>
        </NTimelineItem>
      </NTimeline>
    </NSpin>
    <div v-if="total > size" class="mt-3 text-right">
      <NPagination
        v-model:page="page"
        :page-size="size"
        :item-count="total"
        size="small"
        @update:page="load"
      />
    </div>
  </div>
</template>
