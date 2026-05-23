<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NDataTable, NButton, NTag, NSelect, NInput, useMessage, type DataTableColumns } from 'naive-ui'
import ListPage from '@/components/ListPage.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import TimeDisplay from '@/components/TimeDisplay.vue'
import { redeemApi, type RedeemCode } from '@/api/redeem'

const message = useMessage()
const router = useRouter()
const data = ref<RedeemCode[]>([])
const total = ref(0)
const loading = ref(false)
const filter = ref({ batch_no: '', status: null as 0|1|2|null, page: 1, size: 20 })
const confirmState = ref({ show: false, id: 0, title: '', content: '' })

const statusMap: Record<number, { label: string; type: 'success'|'error'|'default' }> = {
  0: { label: '未使用', type: 'success' }, 1: { label: '已禁用', type: 'error' }, 2: { label: '已使用', type: 'default' },
}

async function load() {
  loading.value = true
  try {
    const res = await redeemApi.list({ batch_no: filter.value.batch_no || undefined, status: filter.value.status ?? undefined, page: filter.value.page, size: filter.value.size })
    data.value = res.items; total.value = res.total
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

async function handleConfirm() {
  try {
    await redeemApi.disableOne(confirmState.value.id)
    message.success('已禁用'); load()
  } catch (_) { /* handled */ }
}

const columns: DataTableColumns<RedeemCode> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '兑换码前缀', key: 'code_prefix', width: 150, render: (row) => h('code', { class: 'text-xs' }, row.code_prefix) },
  { title: '批次号', key: 'batch_no', width: 140 },
  { title: '金额(quota)', key: 'amount_quota', width: 100 },
  { title: '状态', key: 'status', width: 90, render: (row) => h(NTag, { type: statusMap[row.status]?.type, size: 'small' }, { default: () => statusMap[row.status]?.label ?? row.status }) },
  { title: '使用者ID', key: 'used_by', width: 90, render: (row) => row.used_by ?? '--' },
  { title: '到期时间', key: 'expires_at', width: 160, render: (row) => h(TimeDisplay, { value: row.expires_at }) },
  { title: '操作', key: 'actions', width: 100, fixed: 'right', render: (row) => row.status === 0 ? h(NButton, { size: 'small', type: 'error', onClick: () => { confirmState.value = { show:true, id:row.id, title:'禁用兑换码', content:'确认禁用该兑换码？' } } }, { default: () => '禁用' }) : h('span', '--') },
]
</script>

<template>
  <ListPage title="兑换码管理">
    <template #actions>
      <NButton type="primary" @click="router.push('/payments/redeem/batch_new')">批量生成</NButton>
    </template>
    <template #filters>
      <NInput v-model:value="filter.batch_no" placeholder="批次号" clearable style="width:150px" @update:value="() => { filter.page=1; load() }" />
      <NSelect v-model:value="filter.status" placeholder="状态" :options="[{label:'未使用',value:0},{label:'已禁用',value:1},{label:'已使用',value:2}]" clearable style="width:110px" @update:value="() => { filter.page=1; load() }" />
    </template>
    <NDataTable :columns="columns" :data="data" :loading="loading" :pagination="{ page: filter.page, pageSize: filter.size, itemCount: total, onChange: (p:number)=>{ filter.page=p; load() } }" remote scroll-x="1000" size="small" />
  </ListPage>

  <ConfirmDialog v-model:show="confirmState.show" :title="confirmState.title" :content="confirmState.content" type="error" @confirm="handleConfirm" />
</template>
