<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { NDataTable, NInput, NButton, NTag, NDatePicker, type DataTableColumns } from 'naive-ui'
import ListPage from '@/components/ListPage.vue'
import TimeDisplay from '@/components/TimeDisplay.vue'
import { logApi, type RequestLog, type LogFilter } from '@/api/log'

const data = ref<RequestLog[]>([])
const total = ref(0)
const loading = ref(false)
const userIdInput = ref('')
const statusInput = ref('')
const dateRange = ref<[number, number] | null>(null)
const filter = ref<LogFilter & { page: number; size: number }>({
  page: 1, size: 20, user_id: undefined, model: undefined, status_code: undefined,
})

async function load() {
  loading.value = true
  try {
    const params: LogFilter = {
      page: filter.value.page, size: filter.value.size,
      user_id: filter.value.user_id, model: filter.value.model,
      status_code: filter.value.status_code,
    }
    if (dateRange.value) {
      params.from = new Date(dateRange.value[0]).toISOString()
      params.to = new Date(dateRange.value[1]).toISOString()
    }
    const res = await logApi.requests(params)
    data.value = res.items; total.value = res.total
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

function onUserChange(v: string) { filter.value.user_id = v ? Number(v) : undefined; filter.value.page = 1; load() }
function onStatusChange(v: string) { filter.value.status_code = v ? Number(v) : undefined; filter.value.page = 1; load() }

const columns: DataTableColumns<RequestLog> = [
  { title: '时间', key: 'created_at', width: 160, render: (row) => h(TimeDisplay, { value: row.created_at }) },
  { title: '用户ID', key: 'user_id', width: 80 },
  { title: '模型', key: 'client_model', width: 150, render: (row) => h('code', { class: 'text-xs' }, row.client_model) },
  { title: '渠道ID', key: 'channel_id', width: 80, render: (row) => row.channel_id ?? '--' },
  { title: '状态码', key: 'status_code', width: 80, render: (row) => h(NTag, { type: row.status_code < 300 ? 'success' : 'error', size: 'small' }, { default: () => String(row.status_code) }) },
  { title: '延迟(ms)', key: 'latency_ms', width: 90 },
  { title: '输入tokens', key: 'input_tokens', width: 100 },
  { title: '输出tokens', key: 'output_tokens', width: 100 },
  { title: '消耗额度', key: 'total_quota', width: 100 },
  { title: 'Trace ID', key: 'trace_id', width: 200, render: (row) => h('code', { class: 'text-xs opacity-60' }, (row.trace_id?.slice(0, 16) ?? '') + '...') },
]
</script>

<template>
  <ListPage title="请求日志">
    <template #actions>
      <NButton @click="load">刷新</NButton>
    </template>
    <template #filters>
      <NDatePicker
        v-model:value="dateRange"
        type="datetimerange"
        clearable
        style="width: 360px"
        @update:value="() => { filter.page=1; load() }"
      />
      <NInput v-model:value="userIdInput" placeholder="用户ID" clearable style="width:90px" @update:value="onUserChange" />
      <NInput v-model:value="filter.model" placeholder="模型" clearable style="width:160px" @update:value="() => { filter.page=1; load() }" />
      <NInput v-model:value="statusInput" placeholder="状态码" clearable style="width:90px" @update:value="onStatusChange" />
    </template>
    <NDataTable
      :columns="columns"
      :data="data"
      :loading="loading"
      :pagination="{ page: filter.page, pageSize: filter.size, itemCount: total, onChange: (p:number)=>{ filter.page=p; load() } }"
      remote
      scroll-x="1200"
      size="small"
    />
  </ListPage>
</template>
