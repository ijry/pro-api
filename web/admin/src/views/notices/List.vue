<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NDataTable, NButton, NTag, NSelect, NSpace, useMessage, type DataTableColumns } from 'naive-ui'
import ListPage from '@/components/ListPage.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import TimeDisplay from '@/components/TimeDisplay.vue'
import { noticeApi, type Notice } from '@/api/notice'

const message = useMessage()
const router = useRouter()
const data = ref<Notice[]>([])
const total = ref(0)
const loading = ref(false)
const filter = ref({ status: null as 0|1|2|null, page: 1, size: 20 })
const confirmState = ref({ show: false, type: '', id: 0, title: '', content: '' })

const levelMap: Record<string, { label: string; type: 'default'|'info'|'warning'|'error'|'success' }> = {
  info: { label: '信息', type: 'info' }, warning: { label: '警告', type: 'warning' },
  danger: { label: '危险', type: 'error' }, success: { label: '成功', type: 'success' },
}
const statusMap: Record<number, { label: string; type: 'default'|'info'|'warning'|'error'|'success' }> = {
  0: { label: '草稿', type: 'default' }, 1: { label: '已发布', type: 'success' }, 2: { label: '已撤回', type: 'warning' },
}

async function load() {
  loading.value = true
  try {
    const res = await noticeApi.list({ status: filter.value.status ?? undefined, page: filter.value.page, size: filter.value.size })
    data.value = res.items; total.value = res.total
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

async function handlePublish(row: Notice) {
  try {
    await noticeApi.publish(row.id)
    message.success('已发布'); load()
  } catch (_) { /* handled */ }
}

async function handleUnpublish(row: Notice) {
  try {
    await noticeApi.unpublish(row.id)
    message.success('已撤回'); load()
  } catch (_) { /* handled */ }
}

async function handleConfirm() {
  try { await noticeApi.remove(confirmState.value.id); message.success('已删除'); load() } catch (_) { /* handled */ }
}

const columns: DataTableColumns<Notice> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '标题', key: 'title', width: 200, render: (row) => h('a', { class: 'text-blue-500 hover:underline cursor-pointer', onClick: () => router.push(`/notices/${row.id}`) }, row.title) },
  { title: '级别', key: 'level', width: 90, render: (row) => h(NTag, { type: levelMap[row.level]?.type ?? 'default', size: 'small' }, { default: () => levelMap[row.level]?.label ?? row.level }) },
  { title: '状态', key: 'status', width: 90, render: (row) => h(NTag, { type: statusMap[row.status]?.type ?? 'default', size: 'small' }, { default: () => statusMap[row.status]?.label ?? row.status }) },
  { title: '置顶', key: 'pinned', width: 70, render: (row) => row.pinned ? h('span', '⭐') : h('span', '--') },
  { title: '发布时间', key: 'publish_at', width: 160, render: (row) => h(TimeDisplay, { value: row.publish_at }) },
  { title: '操作', key: 'actions', width: 220, fixed: 'right', render: (row) => h(NSpace, { size: 'small' }, { default: () => [
    h(NButton, { size: 'small', onClick: () => router.push(`/notices/${row.id}`) }, { default: () => '编辑' }),
    row.status !== 1 ? h(NButton, { size: 'small', type: 'primary', onClick: () => handlePublish(row) }, { default: () => '发布' }) : null,
    row.status === 1 ? h(NButton, { size: 'small', type: 'warning', onClick: () => handleUnpublish(row) }, { default: () => '撤回' }) : null,
    h(NButton, { size: 'small', type: 'error', onClick: () => { confirmState.value = { show:true, type:'delete', id:row.id, title:'删除公告', content:`确认删除公告 "${row.title}"？` } } }, { default: () => '删除' }),
  ] }) },
]
</script>

<template>
  <ListPage title="公告管理">
    <template #actions>
      <NButton type="primary" @click="router.push('/notices/new')">新建公告</NButton>
    </template>
    <template #filters>
      <NSelect v-model:value="filter.status" placeholder="状态" :options="[{label:'草稿',value:0},{label:'已发布',value:1},{label:'已撤回',value:2}]" clearable style="width:110px" @update:value="() => { filter.page=1; load() }" />
    </template>
    <NDataTable :columns="columns" :data="data" :loading="loading" :pagination="{ page: filter.page, pageSize: filter.size, itemCount: total, onChange: (p:number)=>{ filter.page=p; load() } }" remote scroll-x="900" size="small" />
  </ListPage>

  <ConfirmDialog v-model:show="confirmState.show" :title="confirmState.title" :content="confirmState.content" type="error" @confirm="handleConfirm" />
</template>
