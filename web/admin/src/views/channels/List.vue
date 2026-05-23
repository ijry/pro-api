<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import {
  NDataTable, NInput, NSelect, NButton, NTag, NSpace, NModal, NFormItem,
  NInputNumber, useMessage, type DataTableColumns,
} from 'naive-ui'
import ListPage from '@/components/ListPage.vue'
import FormDrawer from '@/components/FormDrawer.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { channelApi, type Channel, type ChannelInput } from '@/api/channel'
import { useRouter } from 'vue-router'

const message = useMessage()
const router = useRouter()

const data = ref<Channel[]>([])
const total = ref(0)
const loading = ref(false)
const filter = ref({ keyword: '', provider: null as string | null, status: null as 0|1|2|null, page: 1, size: 20 })

const drawerShow = ref(false)
const drawerMode = ref<'create'|'edit'>('create')
const drawerLoading = ref(false)
const editingId = ref<number>(0)
const form = ref<ChannelInput>({ name: '', provider: 'openai', base_url: '', credentials: { api_key: '' }, priority: 0, weight: 1, status: 0, tags: [], extra: {} })

const testModal = ref({ show: false, channelId: 0, result: null as null|{ ok: boolean; latency_ms?: number; error?: string } })
const confirmState = ref({ show: false, type: '', id: 0, title: '', content: '' })

const providerOptions = [
  { label: 'OpenAI', value: 'openai' }, { label: 'Azure', value: 'azure' },
  { label: 'Anthropic', value: 'anthropic' }, { label: 'Google', value: 'google' },
  { label: 'Groq', value: 'groq' }, { label: 'Custom', value: 'custom' },
]

const statusMap: Record<number, { label: string; type: 'success'|'error'|'warning' }> = {
  0: { label: '启用', type: 'success' },
  1: { label: '禁用', type: 'error' },
  2: { label: '熔断', type: 'warning' },
}

async function load() {
  loading.value = true
  try {
    const res = await channelApi.list({ page: filter.value.page, size: filter.value.size, keyword: filter.value.keyword || undefined, provider: filter.value.provider ?? undefined, status: filter.value.status ?? undefined })
    data.value = res.items; total.value = res.total
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

function openCreate() {
  drawerMode.value = 'create'
  form.value = { name: '', provider: 'openai', base_url: '', credentials: { api_key: '' }, priority: 0, weight: 1, status: 0, tags: [], extra: {} }
  drawerShow.value = true
}

function openEdit(row: Channel) {
  drawerMode.value = 'edit'
  editingId.value = row.id
  form.value = { name: row.name, provider: row.provider, base_url: row.base_url, credentials: { api_key: '' }, priority: row.priority, weight: row.weight, status: row.status as 0|1, tags: [...row.tags], extra: { ...row.extra } }
  drawerShow.value = true
}

async function handleSubmit() {
  drawerLoading.value = true
  try {
    if (drawerMode.value === 'create') { await channelApi.create(form.value); message.success('渠道已创建') }
    else { await channelApi.patch(editingId.value, form.value); message.success('渠道已更新') }
    drawerShow.value = false; load()
  } catch (_) { /* handled */ } finally { drawerLoading.value = false }
}

async function doTest(id: number) {
  testModal.value = { show: true, channelId: id, result: null }
  try {
    const r = await channelApi.test(id)
    testModal.value.result = r
  } catch (_) { testModal.value.result = { ok: false, error: '测试失败' } }
}

function confirmAction(type: string, row: Channel) {
  const labels: Record<string, string> = { enable: '启用', disable: '禁用', delete: '删除' }
  confirmState.value = { show: true, type, id: row.id, title: `${labels[type]}渠道`, content: `确认${labels[type]}渠道 "${row.name}"？` }
}

async function handleConfirm() {
  const { type, id } = confirmState.value
  try {
    if (type === 'enable') { await channelApi.enable(id); message.success('已启用') }
    else if (type === 'disable') { await channelApi.disable(id); message.success('已禁用') }
    else if (type === 'delete') { await channelApi.remove(id); message.success('已删除') }
    load()
  } catch (_) { /* handled */ }
}

const columns: DataTableColumns<Channel> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '名称', key: 'name', width: 160, render: (row) => h('a', { class: 'text-blue-500 hover:underline cursor-pointer', onClick: () => router.push(`/channels/${row.id}`) }, row.name) },
  { title: '供应商', key: 'provider', width: 100 },
  { title: '优先级', key: 'priority', width: 80 },
  { title: '权重', key: 'weight', width: 80 },
  { title: '状态', key: 'status', width: 90, render: (row) => h(NTag, { type: statusMap[row.status]?.type, size: 'small' }, { default: () => statusMap[row.status]?.label ?? row.status }) },
  { title: '健康', key: 'health', width: 100, render: (row) => h('span', { class: row.health?.state === 'closed' ? 'text-green-500' : 'text-red-500' }, row.health?.state ?? '--') },
  { title: '最后测试', key: 'last_test', width: 150, render: (row) => row.last_test ? h('span', { class: row.last_test.ok ? 'text-green-600' : 'text-red-500' }, row.last_test.ok ? `${row.last_test.latency_ms}ms` : row.last_test.error ?? '失败') : h('span', '--') },
  { title: '操作', key: 'actions', width: 240, fixed: 'right', render: (row) => h(NSpace, { size: 'small' }, { default: () => [
    h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => '编辑' }),
    h(NButton, { size: 'small', onClick: () => doTest(row.id) }, { default: () => '测试' }),
    row.status === 0 ? h(NButton, { size: 'small', type: 'warning', onClick: () => confirmAction('disable', row) }, { default: () => '禁用' }) : h(NButton, { size: 'small', type: 'primary', onClick: () => confirmAction('enable', row) }, { default: () => '启用' }),
    h(NButton, { size: 'small', type: 'error', onClick: () => confirmAction('delete', row) }, { default: () => '删除' }),
  ] }) },
]
</script>

<template>
  <ListPage title="渠道管理">
    <template #actions>
      <NButton type="primary" @click="openCreate">新建渠道</NButton>
    </template>
    <template #filters>
      <NInput v-model:value="filter.keyword" placeholder="搜索名称" clearable style="width:180px" @update:value="() => { filter.page=1; load() }" />
      <NSelect v-model:value="filter.provider" placeholder="供应商" :options="providerOptions" clearable style="width:120px" @update:value="() => { filter.page=1; load() }" />
      <NSelect v-model:value="filter.status" placeholder="状态" :options="[{label:'启用',value:0},{label:'禁用',value:1},{label:'熔断',value:2}]" clearable style="width:100px" @update:value="() => { filter.page=1; load() }" />
    </template>
    <NDataTable :columns="columns" :data="data" :loading="loading" :pagination="{ page: filter.page, pageSize: filter.size, itemCount: total, onChange: (p:number)=>{ filter.page=p; load() } }" remote scroll-x="1100" size="small" />
  </ListPage>

  <FormDrawer :show="drawerShow" :mode="drawerMode" :title="drawerMode === 'create' ? '新建渠道' : '编辑渠道'" :loading="drawerLoading" @update:show="drawerShow=$event" @submit="handleSubmit">
    <NFormItem label="名称"><NInput v-model:value="form.name" /></NFormItem>
    <NFormItem label="供应商"><NSelect v-model:value="form.provider" :options="providerOptions" /></NFormItem>
    <NFormItem label="Base URL"><NInput v-model:value="form.base_url" placeholder="https://api.openai.com/v1" /></NFormItem>
    <NFormItem label="API Key"><NInput v-model:value="form.credentials.api_key" type="password" show-password-on="click" placeholder="sk-..." /></NFormItem>
    <NFormItem label="优先级"><NInputNumber v-model:value="form.priority" style="width:100%" /></NFormItem>
    <NFormItem label="权重"><NInputNumber v-model:value="form.weight" :min="1" style="width:100%" /></NFormItem>
  </FormDrawer>

  <ConfirmDialog v-model:show="confirmState.show" :title="confirmState.title" :content="confirmState.content" :type="confirmState.type==='delete'?'error':'warning'" @confirm="handleConfirm" />

  <NModal v-model:show="testModal.show" preset="card" title="渠道测试结果" style="width:400px">
    <div v-if="!testModal.result" class="text-center py-4 text-gray-400">测试中...</div>
    <div v-else>
      <div v-if="testModal.result.ok" class="text-green-500 font-semibold">测试通过 ✓ &nbsp; {{ testModal.result.latency_ms }}ms</div>
      <div v-else class="text-red-500">测试失败: {{ testModal.result.error }}</div>
    </div>
    <template #footer>
      <NButton @click="testModal.show=false">关闭</NButton>
    </template>
  </NModal>
</template>
