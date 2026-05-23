<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { NDataTable, NInput, NButton, type DataTableColumns } from 'naive-ui'
import ListPage from '@/components/ListPage.vue'
import TimeDisplay from '@/components/TimeDisplay.vue'
import { logApi, type ErrorLog } from '@/api/log'

const data = ref<ErrorLog[]>([])
const total = ref(0)
const loading = ref(false)
const userIdInput = ref('')
const filter = ref({ page: 1, size: 20, user_id: undefined as number|undefined })

async function load() {
  loading.value = true
  try {
    const res = await logApi.errors({ page: filter.value.page, size: filter.value.size, user_id: filter.value.user_id })
    data.value = res.items; total.value = res.total
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

function onUserChange(v: string) {
  filter.value.user_id = v ? Number(v) : undefined
  filter.value.page = 1
  load()
}

const columns: DataTableColumns<ErrorLog> = [
  { title: '时间', key: 'created_at', width: 160, render: (row) => h(TimeDisplay, { value: row.created_at }) },
  { title: '用户ID', key: 'user_id', width: 80, render: (row) => row.user_id ?? '--' },
  { title: '错误码', key: 'error_code', width: 100 },
  { title: '错误类型', key: 'error_type', width: 160 },
  { title: 'Trace ID', key: 'trace_id', width: 220, render: (row) => h('code', { class: 'text-xs opacity-60' }, row.trace_id) },
]
</script>

<template>
  <ListPage title="错误日志">
    <template #actions>
      <NButton @click="load">刷新</NButton>
    </template>
    <template #filters>
      <NInput v-model:value="userIdInput" placeholder="用户ID" clearable style="width:100px" @update:value="onUserChange" />
    </template>
    <NDataTable :columns="columns" :data="data" :loading="loading" :pagination="{ page: filter.page, pageSize: filter.size, itemCount: total, onChange: (p:number)=>{ filter.page=p; load() } }" remote scroll-x="800" size="small" />
  </ListPage>
</template>
