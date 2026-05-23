<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { NDataTable, NInput, NSelect, NButton, NTag, NSpace, useMessage, type DataTableColumns } from 'naive-ui'
import ListPage from '@/components/ListPage.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import TimeDisplay from '@/components/TimeDisplay.vue'
import { tokenApi, type Token } from '@/api/token'

const message = useMessage()
const data = ref<Token[]>([])
const total = ref(0)
const loading = ref(false)
const filter = ref({ keyword: '', user_id: null as number | null, status: null as 0|1|2|null, page: 1, size: 20 })
const confirmState = ref({ show: false, type: '', id: 0, title: '', content: '' })

const statusMap: Record<number, { label: string; type: 'success'|'error'|'warning' }> = {
  0: { label: '正常', type: 'success' },
  1: { label: '禁用', type: 'error' },
  2: { label: '已撤销', type: 'warning' },
}

const userIdInput = ref('')

async function load() {
  loading.value = true
  try {
    const res = await tokenApi.list({ page: filter.value.page, size: filter.value.size, keyword: filter.value.keyword || undefined, user_id: filter.value.user_id ?? undefined, status: filter.value.status ?? undefined })
    data.value = res.items; total.value = res.total
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

function doConfirm(type: string, row: Token) {
  const labels: Record<string, string> = { disable: '禁用', revoke: '撤销' }
  confirmState.value = { show: true, type, id: row.id, title: `${labels[type]}令牌`, content: `确认${labels[type]}令牌 "${row.name}"？` }
}

async function handleConfirm() {
  const { type, id } = confirmState.value
  try {
    if (type === 'disable') { await tokenApi.patch(id, { status: 1 }); message.success('已禁用') }
    else if (type === 'revoke') { await tokenApi.remove(id); message.success('已撤销') }
    load()
  } catch (_) { /* handled */ }
}

const columns: DataTableColumns<Token> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '名称', key: 'name', width: 160 },
  { title: '前缀', key: 'key_prefix', width: 120, render: (row) => h('code', { class: 'text-xs bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded' }, row.key_prefix) },
  { title: '用户ID', key: 'user_id', width: 80 },
  { title: '状态', key: 'status', width: 90, render: (row) => h(NTag, { type: statusMap[row.status]?.type, size: 'small' }, { default: () => statusMap[row.status]?.label ?? row.status }) },
  { title: '已用额度', key: 'quota_used', width: 100 },
  { title: '到期时间', key: 'expires_at', width: 160, render: (row) => h(TimeDisplay, { value: row.expires_at }) },
  { title: '最后使用', key: 'last_used_at', width: 160, render: (row) => h(TimeDisplay, { value: row.last_used_at, relative: true }) },
  { title: '操作', key: 'actions', width: 150, fixed: 'right', render: (row) => h(NSpace, { size: 'small' }, { default: () => [
    row.status !== 1 ? h(NButton, { size: 'small', type: 'warning', onClick: () => doConfirm('disable', row) }, { default: () => '禁用' }) : null,
    row.status !== 2 ? h(NButton, { size: 'small', type: 'error', onClick: () => doConfirm('revoke', row) }, { default: () => '撤销' }) : null,
  ] }) },
]
</script>

<template>
  <ListPage title="令牌管理（全局）">
    <template #actions>
      <NButton @click="load">刷新</NButton>
    </template>
    <template #filters>
      <NInput v-model:value="filter.keyword" placeholder="搜索名称" clearable style="width:180px" @update:value="() => { filter.page=1; load() }" />
      <NInput v-model:value="userIdInput" placeholder="用户ID" clearable style="width:100px" @update:value="(v) => { filter.user_id = v ? Number(v) : null; filter.page=1; load() }" />
      <NSelect v-model:value="filter.status" placeholder="状态" :options="[{label:'正常',value:0},{label:'禁用',value:1},{label:'已撤销',value:2}]" clearable style="width:100px" @update:value="() => { filter.page=1; load() }" />
    </template>
    <NDataTable :columns="columns" :data="data" :loading="loading" :pagination="{ page: filter.page, pageSize: filter.size, itemCount: total, onChange: (p:number)=>{ filter.page=p; load() } }" remote scroll-x="1100" size="small" />
  </ListPage>

  <ConfirmDialog v-model:show="confirmState.show" :title="confirmState.title" :content="confirmState.content" :type="confirmState.type==='revoke'?'error':'warning'" @confirm="handleConfirm" />
</template>
