<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import {
  NGrid, NGridItem, NCard, NStatistic, NSkeleton, NEmpty, NSelect, NDatePicker,
} from 'naive-ui'
import { statsApi, type TimeseriesPoint, type ByModelRow, type ByChannelRow } from '@/api/stats'
import dayjs from 'dayjs'

type Granularity = 'hour' | 'day'

const PRESETS = [
  { label: '今日', value: 'today' },
  { label: '7天', value: '7d' },
  { label: '30天', value: '30d' },
  { label: '自定义', value: 'custom' },
]

const preset = ref('today')
const customRange = ref<[number, number] | null>(null)
const granularity = ref<Granularity>('hour')

const loading = ref({ timeseries: false, byModel: false, byChannel: false })
const timeseries = ref<TimeseriesPoint[]>([])
const byModel = ref<ByModelRow[]>([])
const byChannel = ref<ByChannelRow[]>([])

const dateParams = computed(() => {
  if (preset.value === 'today') {
    return { from: dayjs().startOf('day').toISOString(), to: dayjs().endOf('day').toISOString(), granularity: 'hour' as Granularity }
  } else if (preset.value === '7d') {
    return { from: dayjs().subtract(7, 'day').startOf('day').toISOString(), to: dayjs().toISOString(), granularity: 'day' as Granularity }
  } else if (preset.value === '30d') {
    return { from: dayjs().subtract(30, 'day').startOf('day').toISOString(), to: dayjs().toISOString(), granularity: 'day' as Granularity }
  } else if (customRange.value) {
    return { from: new Date(customRange.value[0]).toISOString(), to: new Date(customRange.value[1]).toISOString(), granularity: granularity.value }
  }
  return {}
})

const totalRequests = computed(() => timeseries.value.reduce((s, p) => s + p.requests, 0))
const totalErrors = computed(() => timeseries.value.reduce((s, p) => s + p.errors, 0))
const totalQuota = computed(() => timeseries.value.reduce((s, p) => s + p.quota, 0))
const errorRate = computed(() => totalRequests.value > 0 ? ((totalErrors.value / totalRequests.value) * 100).toFixed(2) + '%' : '--')

async function loadAll() {
  const p = dateParams.value
  loading.value.timeseries = true; loading.value.byModel = true; loading.value.byChannel = true
  try {
    const [ts, models, channels] = await Promise.all([
      statsApi.timeseries({ from: p.from, to: p.to, granularity: p.granularity }),
      statsApi.byModel({ from: p.from, to: p.to, order_by: 'requests', limit: 10 }),
      statsApi.byChannel({ from: p.from, to: p.to, order_by: 'requests', limit: 10 }),
    ])
    timeseries.value = ts.points
    byModel.value = models.rows
    byChannel.value = channels.rows
  } catch (_) { /* handled */ } finally {
    loading.value.timeseries = false; loading.value.byModel = false; loading.value.byChannel = false
  }
}

onMounted(loadAll)
</script>

<template>
  <div>
    <div class="flex items-center gap-3 mb-4 flex-wrap">
      <h2 class="text-2xl font-semibold">统计分析</h2>
      <NSelect v-model:value="preset" :options="PRESETS" style="width:120px" @update:value="loadAll" />
      <NDatePicker v-if="preset==='custom'" v-model:value="customRange" type="datetimerange" clearable style="width:360px" @update:value="loadAll" />
    </div>

    <!-- Summary stats -->
    <NGrid :cols="4" :x-gap="16" :y-gap="16" responsive="screen" :item-responsive="true" class="mb-4">
      <NGridItem span="4 m:2 l:1">
        <NCard size="small">
          <NSkeleton v-if="loading.timeseries" text :repeat="2" />
          <NStatistic v-else label="总请求数" :value="totalRequests.toLocaleString()" />
        </NCard>
      </NGridItem>
      <NGridItem span="4 m:2 l:1">
        <NCard size="small">
          <NSkeleton v-if="loading.timeseries" text :repeat="2" />
          <NStatistic v-else label="总错误数" :value="totalErrors.toLocaleString()" />
        </NCard>
      </NGridItem>
      <NGridItem span="4 m:2 l:1">
        <NCard size="small">
          <NSkeleton v-if="loading.timeseries" text :repeat="2" />
          <NStatistic v-else label="总消耗额度" :value="totalQuota.toLocaleString()" />
        </NCard>
      </NGridItem>
      <NGridItem span="4 m:2 l:1">
        <NCard size="small">
          <NSkeleton v-if="loading.timeseries" text :repeat="2" />
          <NStatistic v-else label="错误率" :value="errorRate" />
        </NCard>
      </NGridItem>
    </NGrid>

    <!-- Timeseries table -->
    <NGrid :cols="1" :x-gap="16" :y-gap="16" responsive="screen" :item-responsive="true" class="mb-4">
      <NGridItem>
        <NCard title="请求量 & 消耗（时序）" size="small">
          <NSkeleton v-if="loading.timeseries" text :repeat="8" />
          <NEmpty v-else-if="!timeseries.length" description="暂无数据" />
          <div v-else class="overflow-auto max-h-64">
            <table class="w-full text-sm">
              <thead><tr><th class="text-left py-1">时间</th><th class="text-right py-1">请求数</th><th class="text-right py-1">错误数</th><th class="text-right py-1">消耗Quota</th></tr></thead>
              <tbody>
                <tr v-for="p in timeseries" :key="p.ts" class="border-t border-gray-100 dark:border-gray-800">
                  <td class="py-1 font-mono text-xs">{{ p.ts }}</td>
                  <td class="py-1 text-right">{{ p.requests.toLocaleString() }}</td>
                  <td class="py-1 text-right" :class="p.errors > 0 ? 'text-red-500' : ''">{{ p.errors.toLocaleString() }}</td>
                  <td class="py-1 text-right">{{ p.quota.toLocaleString() }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </NCard>
      </NGridItem>
    </NGrid>

    <!-- By model & channel -->
    <NGrid :cols="2" :x-gap="16" :y-gap="16" responsive="screen" :item-responsive="true">
      <NGridItem span="2 l:1">
        <NCard title="Top 10 模型" size="small">
          <NSkeleton v-if="loading.byModel" text :repeat="5" />
          <NEmpty v-else-if="!byModel.length" description="暂无数据" />
          <table v-else class="w-full text-sm">
            <thead><tr><th class="text-left py-1">模型</th><th class="text-right py-1">请求数</th><th class="text-right py-1">Tokens入</th><th class="text-right py-1">Tokens出</th><th class="text-right py-1">Quota</th></tr></thead>
            <tbody>
              <tr v-for="r in byModel" :key="r.model" class="border-t border-gray-100 dark:border-gray-800">
                <td class="py-1 font-mono text-xs">{{ r.model }}</td>
                <td class="py-1 text-right">{{ r.requests.toLocaleString() }}</td>
                <td class="py-1 text-right">{{ r.tokens_in.toLocaleString() }}</td>
                <td class="py-1 text-right">{{ r.tokens_out.toLocaleString() }}</td>
                <td class="py-1 text-right">{{ r.quota.toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </NCard>
      </NGridItem>
      <NGridItem span="2 l:1">
        <NCard title="Top 10 渠道" size="small">
          <NSkeleton v-if="loading.byChannel" text :repeat="5" />
          <NEmpty v-else-if="!byChannel.length" description="暂无数据" />
          <table v-else class="w-full text-sm">
            <thead><tr><th class="text-left py-1">渠道</th><th class="text-right py-1">供应商</th><th class="text-right py-1">请求数</th><th class="text-right py-1">Quota</th></tr></thead>
            <tbody>
              <tr v-for="r in byChannel" :key="r.channel_id" class="border-t border-gray-100 dark:border-gray-800">
                <td class="py-1">{{ r.channel_name }}</td>
                <td class="py-1">{{ r.provider }}</td>
                <td class="py-1 text-right">{{ r.requests.toLocaleString() }}</td>
                <td class="py-1 text-right">{{ r.quota.toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </NCard>
      </NGridItem>
    </NGrid>
  </div>
</template>
