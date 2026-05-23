<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import {
  NDataTable, NButton, NTag, NSelect, NSpace, NModal, NForm, NFormItem, NInput, NInputNumber,
  useMessage, type DataTableColumns,
} from 'naive-ui'
import ListPage from '@/components/ListPage.vue'
import TimeDisplay from '@/components/TimeDisplay.vue'
import { rechargeApi, type ManualRecharge } from '@/api/recharge'

const message = useMessage()
const data = ref<ManualRecharge[]>([])
const total = ref(0)
const loading = ref(false)
const filter = ref({ status: null as 0|1|2|3|null, page: 1, size: 20 })

const approveModal = ref({ show: false, id: 0, amount_quota: 0, review_note: '' })
const rejectModal = ref({ show: false, id: 0, review_note: '' })

const statusMap: Record<number, { label: string; type: 'default'|'info'|'success'|'error'|'warning' }> = {
  0: { label: '待审核', type: 'warning' }, 1: { label: '已通过', type: 'success' },
  2: { label: '已拒绝', type: 'error' }, 3: { label: '已取消', type: 'default' },
}

async function load() {
  loading.value = true
  try {
    const res = await rechargeApi.list({ status: filter.value.status ?? undefined, page: filter.value.page, size: filter.value.size })
    data.value = res.items; total.value = res.total
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

async function doApprove() {
  try {
    await rechargeApi.approve(approveModal.value.id, approveModal.value.review_note)
    message.success('已批准'); approveModal.value.show = false; load()
  } catch (_) { /* handled */ }
}

async function doReject() {
  try {
    await rechargeApi.reject(rejectModal.value.id, rejectModal.value.review_note)
    message.success('已拒绝'); rejectModal.value.show = false; load()
  } catch (_) { /* handled */ }
}

const columns: DataTableColumns<ManualRecharge> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '用户', key: 'username', width: 120 },
  { title: '用户ID', key: 'user_id', width: 80 },
  { title: '申请金额', key: 'amount_money', width: 110, render: (row) => `${row.amount_money} ${row.currency}` },
  { title: '状态', key: 'status', width: 90, render: (row) => h(NTag, { type: statusMap[row.status]?.type, size: 'small' }, { default: () => statusMap[row.status]?.label ?? row.status }) },
  { title: '申请备注', key: 'applicant_note', ellipsis: { tooltip: true }, width: 160 },
  { title: '申请时间', key: 'created_at', width: 160, render: (row) => h(TimeDisplay, { value: row.created_at }) },
  { title: '操作', key: 'actions', width: 160, fixed: 'right', render: (row) => row.status === 0 ? h(NSpace, { size: 'small' }, { default: () => [
    h(NButton, { size: 'small', type: 'primary', onClick: () => { approveModal.value = { show:true, id:row.id, amount_quota: row.amount_money * 7, review_note:'' } } }, { default: () => '批准' }),
    h(NButton, { size: 'small', type: 'error', onClick: () => { rejectModal.value = { show:true, id:row.id, review_note:'' } } }, { default: () => '拒绝' }),
  ] }) : h('span', '--') },
]
</script>

<template>
  <ListPage title="充值申请管理">
    <template #actions>
      <NButton @click="load">刷新</NButton>
    </template>
    <template #filters>
      <NSelect v-model:value="filter.status" placeholder="状态" :options="[{label:'待审核',value:0},{label:'已通过',value:1},{label:'已拒绝',value:2},{label:'已取消',value:3}]" clearable style="width:110px" @update:value="() => { filter.page=1; load() }" />
    </template>
    <NDataTable :columns="columns" :data="data" :loading="loading" :pagination="{ page: filter.page, pageSize: filter.size, itemCount: total, onChange: (p:number)=>{ filter.page=p; load() } }" remote scroll-x="1000" size="small" />
  </ListPage>

  <NModal v-model:show="approveModal.show" preset="card" title="批准充值申请" style="width:420px">
    <NForm label-placement="top">
      <NFormItem label="发放额度（quota）">
        <NInputNumber v-model:value="approveModal.amount_quota" :min="0" style="width:100%" />
      </NFormItem>
      <NFormItem label="审核备注">
        <NInput v-model:value="approveModal.review_note" type="textarea" :rows="3" />
      </NFormItem>
    </NForm>
    <NSpace justify="end">
      <NButton @click="approveModal.show=false">取消</NButton>
      <NButton type="primary" @click="doApprove">确认批准</NButton>
    </NSpace>
  </NModal>

  <NModal v-model:show="rejectModal.show" preset="card" title="拒绝充值申请" style="width:400px">
    <NForm label-placement="top">
      <NFormItem label="拒绝原因">
        <NInput v-model:value="rejectModal.review_note" type="textarea" :rows="3" placeholder="请填写拒绝原因..." />
      </NFormItem>
    </NForm>
    <NSpace justify="end">
      <NButton @click="rejectModal.show=false">取消</NButton>
      <NButton type="error" @click="doReject">确认拒绝</NButton>
    </NSpace>
  </NModal>
</template>
