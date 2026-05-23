<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { NDataTable, NInput, NButton, type DataTableColumns } from 'naive-ui'
import ListPage from '@/components/ListPage.vue'
import TimeDisplay from '@/components/TimeDisplay.vue'
import { logApi, type AuditLog } from '@/api/log'

const data = ref<AuditLog[]>([])
const total = ref(0)
const loading = ref(false)
const actorInput = ref('')
const filter = ref({ page: 1, size: 20, actor_id: undefined as number|undefined, action: undefined as string|undefined, target_type: undefined as string|undefined })

async function load() {
  loading.value = true
  try {
    const res = await logApi.audit({ page: filter.value.page, size: filter.value.size, actor_id: filter.value.actor_id, action: filter.value.action, target_type: filter.value.target_type })
    data.value = res.items; total.value = res.total
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

function onActorChange(v: string) {
  filter.value.actor_id = v ? Number(v) : undefined
  filter.value.page = 1
  load()
}

const columns: DataTableColumns<AuditLog> = [
  { title: '时间', key: 'created_at', width: 160, render: (row) => h(TimeDisplay, { value: row.created_at }) },
  { title: '操作者ID', key: 'actor_id', width: 90, render: (row) => row.actor_id ?? '--' },
  { title: '动作', key: 'action', width: 160, render: (row) => h('code', { class: 'text-xs' }, row.action) },
  { title: '目标类型', key: 'target_type', width: 120 },
  { title: '目标ID', key: 'target_id', width: 80, render: (row) => row.target_id ?? '--' },
  { title: 'IP', key: 'ip', width: 140 },
]
</script>

<template>
  <ListPage title="审计日志">
    <template #actions>
      <NButton @click="load">刷新</NButton>
    </template>
    <template #filters>
      <NInput v-model:value="actorInput" placeholder="操作者ID" clearable style="width:110px" @update:value="onActorChange" />
      <NInput v-model:value="filter.action" placeholder="动作" clearable style="width:140px" @update:value="() => { filter.page=1; load() }" />
      <NInput v-model:value="filter.target_type" placeholder="目标类型" clearable style="width:120px" @update:value="() => { filter.page=1; load() }" />
    </template>
    <NDataTable :columns="columns" :data="data" :loading="loading" :pagination="{ page: filter.page, pageSize: filter.size, itemCount: total, onChange: (p:number)=>{ filter.page=p; load() } }" remote scroll-x="900" size="small" />
  </ListPage>
</template>
