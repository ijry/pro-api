<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NGrid, NGridItem, NCard, NStatistic, NSkeleton, NEmpty, NList, NListItem, NText } from 'naive-ui'
import { useRouter } from 'vue-router'
import { statsApi, type Overview, type TimeseriesPoint, type ByModelRow, type ByChannelRow, type ByUserRow } from '@/api/stats'
import { noticeApi, type Notice } from '@/api/notice'
import TimeDisplay from '@/components/TimeDisplay.vue'

const router = useRouter()

const overview = ref<Overview | null>(null)
const timeseries = ref<TimeseriesPoint[]>([])
const byModel = ref<ByModelRow[]>([])
const byChannel = ref<ByChannelRow[]>([])
const byUser = ref<ByUserRow[]>([])
const notices = ref<Notice[]>([])

const loading = ref({
  overview: true,
  charts: true,
  notices: true,
})

async function loadOverview() {
  loading.value.overview = true
  try { overview.value = await statsApi.overview() }
  catch (_) { /* handled */ }
  finally { loading.value.overview = false }
}

async function loadCharts() {
  loading.value.charts = true
  try {
    const [ts, models, channels, users] = await Promise.all([
      statsApi.timeseries({ granularity: 'hour' }),
      statsApi.byModel({ order_by: 'quota', limit: 10 }),
      statsApi.byChannel({ order_by: 'requests', limit: 10 }),
      statsApi.byUser({ order_by: 'quota', limit: 10 }),
    ])
    timeseries.value = ts.points
    byModel.value = models.rows
    byChannel.value = channels.rows
    byUser.value = users.rows
  } catch (_) { /* handled */ }
  finally { loading.value.charts = false }
}

async function loadNotices() {
  loading.value.notices = true
  try {
    const res = await noticeApi.list({ status: 1, page: 1, size: 5 })
    notices.value = res.items
  } catch (_) { /* handled */ }
  finally { loading.value.notices = false }
}

onMounted(() => {
  loadOverview()
  loadCharts()
  loadNotices()
})

const trendClass = (v: number) => v >= 0 ? 'text-green-500' : 'text-red-500'
const trendSign = (v: number) => v >= 0 ? '+' : ''
</script>

<template>
  <div>
    <h2 class="text-2xl font-semibold mb-4">仪表盘</h2>

    <!-- Overview cards -->
    <NGrid :cols="4" :x-gap="16" :y-gap="16" responsive="screen" :item-responsive="true" class="mb-4">
      <NGridItem span="4 600:2 900:1">
        <NCard size="small" hoverable @click="router.push('/logs/requests')">
          <NSkeleton v-if="loading.overview" text :repeat="2" />
          <NStatistic v-else label="今日请求" :value="overview?.requests_today ?? 0">
            <template #suffix>
              <span :class="trendClass(overview?.delta?.requests ?? 0)">
                {{ trendSign(overview?.delta?.requests ?? 0) }}{{ ((overview?.delta?.requests ?? 0) * 100).toFixed(1) }}%
              </span>
            </template>
          </NStatistic>
        </NCard>
      </NGridItem>
      <NGridItem span="4 600:2 900:1">
        <NCard size="small" hoverable @click="router.push('/stats')">
          <NSkeleton v-if="loading.overview" text :repeat="2" />
          <NStatistic v-else label="今日收入(quota)" :value="(overview?.revenue_today ?? 0).toLocaleString()">
            <template #suffix>
              <span :class="trendClass(overview?.delta?.revenue ?? 0)">
                {{ trendSign(overview?.delta?.revenue ?? 0) }}{{ ((overview?.delta?.revenue ?? 0) * 100).toFixed(1) }}%
              </span>
            </template>
          </NStatistic>
        </NCard>
      </NGridItem>
      <NGridItem span="4 600:2 900:1">
        <NCard size="small" hoverable @click="router.push('/users')">
          <NSkeleton v-if="loading.overview" text :repeat="2" />
          <NStatistic v-else label="活跃用户" :value="overview?.active_users ?? 0">
            <template #suffix>
              <span :class="trendClass(overview?.delta?.users ?? 0)">
                {{ trendSign(overview?.delta?.users ?? 0) }}{{ ((overview?.delta?.users ?? 0) * 100).toFixed(1) }}%
              </span>
            </template>
          </NStatistic>
        </NCard>
      </NGridItem>
      <NGridItem span="4 600:2 900:1">
        <NCard size="small" hoverable @click="router.push('/logs/errors')">
          <NSkeleton v-if="loading.overview" text :repeat="2" />
          <NStatistic v-else label="错误率" :value="`${((overview?.error_rate ?? 0) * 100).toFixed(2)}%`">
            <template #suffix>
              <span :class="trendClass(-(overview?.delta?.error_rate ?? 0))">
                {{ trendSign(-(overview?.delta?.error_rate ?? 0)) }}{{ ((overview?.delta?.error_rate ?? 0) * 100).toFixed(2) }}%
              </span>
            </template>
          </NStatistic>
        </NCard>
      </NGridItem>
    </NGrid>

    <!-- Top stats tables -->
    <NGrid :cols="2" :x-gap="16" :y-gap="16" responsive="screen" :item-responsive="true" class="mb-4">
      <NGridItem span="2 900:1">
        <NCard title="Top 模型 (by quota)" size="small">
          <NSkeleton v-if="loading.charts" text :repeat="5" />
          <NEmpty v-else-if="!byModel.length" description="暂无数据" />
          <table v-else class="w-full text-sm">
            <thead><tr><th class="text-left py-1">模型</th><th class="text-right py-1">请求数</th><th class="text-right py-1">Quota</th></tr></thead>
            <tbody>
              <tr v-for="r in byModel" :key="r.model" class="border-t border-gray-100 dark:border-gray-800">
                <td class="py-1 font-mono text-xs">{{ r.model }}</td>
                <td class="py-1 text-right">{{ r.requests.toLocaleString() }}</td>
                <td class="py-1 text-right">{{ r.quota.toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </NCard>
      </NGridItem>
      <NGridItem span="2 900:1">
        <NCard title="Top 渠道 (by requests)" size="small">
          <NSkeleton v-if="loading.charts" text :repeat="5" />
          <NEmpty v-else-if="!byChannel.length" description="暂无数据" />
          <table v-else class="w-full text-sm">
            <thead><tr><th class="text-left py-1">渠道</th><th class="text-right py-1">请求数</th><th class="text-right py-1">Quota</th></tr></thead>
            <tbody>
              <tr v-for="r in byChannel" :key="r.channel_id" class="border-t border-gray-100 dark:border-gray-800">
                <td class="py-1">{{ r.channel_name }}</td>
                <td class="py-1 text-right">{{ r.requests.toLocaleString() }}</td>
                <td class="py-1 text-right">{{ r.quota.toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </NCard>
      </NGridItem>
    </NGrid>

    <!-- Notices + top users -->
    <NGrid :cols="2" :x-gap="16" :y-gap="16" responsive="screen" :item-responsive="true">
      <NGridItem span="2 900:1">
        <NCard title="Top 用户 (by quota)" size="small">
          <NSkeleton v-if="loading.charts" text :repeat="5" />
          <NEmpty v-else-if="!byUser.length" description="暂无数据" />
          <table v-else class="w-full text-sm">
            <thead><tr><th class="text-left py-1">用户</th><th class="text-right py-1">请求数</th><th class="text-right py-1">Quota</th></tr></thead>
            <tbody>
              <tr v-for="r in byUser" :key="r.user_id" class="border-t border-gray-100 dark:border-gray-800">
                <td class="py-1">{{ r.username }}</td>
                <td class="py-1 text-right">{{ r.requests.toLocaleString() }}</td>
                <td class="py-1 text-right">{{ r.quota.toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </NCard>
      </NGridItem>
      <NGridItem span="2 900:1">
        <NCard title="最新公告" size="small">
          <NSkeleton v-if="loading.notices" text :repeat="5" />
          <NEmpty v-else-if="!notices.length" description="暂无公告" />
          <NList v-else>
            <NListItem v-for="n in notices" :key="n.id">
              <div class="flex items-center justify-between">
                <NText>{{ n.title }}</NText>
                <TimeDisplay :value="n.publish_at" relative class="text-xs opacity-50" />
              </div>
            </NListItem>
          </NList>
        </NCard>
      </NGridItem>
    </NGrid>
  </div>
</template>
