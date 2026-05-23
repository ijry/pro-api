<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { NDataTable, NButton, NTag, NSpace, NSelect, useMessage, type DataTableColumns } from 'naive-ui'
import { NFormItem, NInput, NInputNumber } from 'naive-ui'
import ListPage from '@/components/ListPage.vue'
import FormDrawer from '@/components/FormDrawer.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { pricingApi, type PricingRule, type PricingInput, type PricingScope } from '@/api/pricing'
import { useDictStore } from '@/stores/dict'

const message = useMessage()
const dictStore = useDictStore()
const data = ref<PricingRule[]>([])
const total = ref(0)
const loading = ref(false)
const filter = ref({ scope: null as PricingScope|null, page: 1, size: 20 })
const drawerShow = ref(false)
const drawerMode = ref<'create'|'edit'>('create')
const drawerLoading = ref(false)
const editingId = ref(0)
const form = ref<PricingInput>({ scope: 'global', group_id: null, model: null, input_ratio: null, output_ratio: null, priority: 0, status: 0 })
const confirmState = ref({ show: false, id: 0, title: '', content: '' })

const scopeOptions = [
  { label: '全局', value: 'global' }, { label: '分组', value: 'group' },
  { label: '模型', value: 'model' }, { label: '分组+模型', value: 'group_model' },
]

async function load() {
  loading.value = true
  try {
    const res = await pricingApi.list({ scope: filter.value.scope ?? undefined, page: filter.value.page, size: filter.value.size })
    data.value = res.items; total.value = res.total
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(() => { dictStore.ensureGroups(); load() })

function openCreate() {
  drawerMode.value = 'create'
  form.value = { scope: 'global', group_id: null, model: null, input_ratio: null, output_ratio: null, priority: 0, status: 0 }
  drawerShow.value = true
}

function openEdit(row: PricingRule) {
  drawerMode.value = 'edit'; editingId.value = row.id
  form.value = { scope: row.scope, group_id: row.group_id, model: row.model, input_ratio: row.input_ratio, output_ratio: row.output_ratio, priority: row.priority, status: row.status }
  drawerShow.value = true
}

async function handleSubmit() {
  drawerLoading.value = true
  try {
    if (drawerMode.value === 'create') { await pricingApi.create(form.value); message.success('定价规则已创建') }
    else { await pricingApi.patch(editingId.value, form.value); message.success('定价规则已更新') }
    drawerShow.value = false; load()
  } catch (_) { /* handled */ } finally { drawerLoading.value = false }
}

async function handleConfirm() {
  try { await pricingApi.remove(confirmState.value.id); message.success('已删除'); load() } catch (_) { /* handled */ }
}

const scopeLabels: Record<string, string> = { global: '全局', group: '分组', model: '模型', group_model: '分组+模型' }

const columns: DataTableColumns<PricingRule> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '范围', key: 'scope', width: 100, render: (row) => h(NTag, { size: 'small', type: 'info' }, { default: () => scopeLabels[row.scope] }) },
  { title: '分组ID', key: 'group_id', width: 80, render: (row) => row.group_id ?? '--' },
  { title: '模型', key: 'model', width: 150, render: (row) => row.model ? h('code', { class: 'text-xs' }, row.model) : h('span', '--') },
  { title: '输入倍率', key: 'input_ratio', width: 90, render: (row) => row.input_ratio ?? '--' },
  { title: '输出倍率', key: 'output_ratio', width: 90, render: (row) => row.output_ratio ?? '--' },
  { title: '优先级', key: 'priority', width: 80 },
  { title: '状态', key: 'status', width: 80, render: (row) => h(NTag, { type: row.status===0?'success':'error', size: 'small' }, { default: () => row.status===0?'启用':'禁用' }) },
  { title: '操作', key: 'actions', width: 150, fixed: 'right', render: (row) => h(NSpace, { size: 'small' }, { default: () => [
    h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => '编辑' }),
    h(NButton, { size: 'small', type: 'error', onClick: () => { confirmState.value = { show:true, id:row.id, title:'删除定价规则', content:`确认删除该定价规则？` } } }, { default: () => '删除' }),
  ] }) },
]
</script>

<template>
  <ListPage title="定价规则">
    <template #actions>
      <NButton type="primary" @click="openCreate">新建规则</NButton>
    </template>
    <template #filters>
      <NSelect v-model:value="filter.scope" placeholder="范围" :options="scopeOptions" clearable style="width:130px" @update:value="() => { filter.page=1; load() }" />
    </template>
    <NDataTable :columns="columns" :data="data" :loading="loading" :pagination="{ page: filter.page, pageSize: filter.size, itemCount: total, onChange: (p:number)=>{ filter.page=p; load() } }" remote size="small" scroll-x="900" />
  </ListPage>

  <FormDrawer :show="drawerShow" :mode="drawerMode" :title="drawerMode==='create'?'新建定价规则':'编辑定价规则'" :loading="drawerLoading" @update:show="drawerShow=$event" @submit="handleSubmit">
    <NFormItem label="范围"><NSelect v-model:value="form.scope" :options="scopeOptions" /></NFormItem>
    <NFormItem v-if="['group','group_model'].includes(form.scope)" label="分组">
      <NSelect v-model:value="form.group_id" :options="dictStore.groupOptions" clearable />
    </NFormItem>
    <NFormItem v-if="['model','group_model'].includes(form.scope)" label="模型"><NInput v-model:value="(form.model as string)" /></NFormItem>
    <NFormItem label="输入倍率"><NInputNumber v-model:value="form.input_ratio" :min="0" :step="0.001" style="width:100%" clearable /></NFormItem>
    <NFormItem label="输出倍率"><NInputNumber v-model:value="form.output_ratio" :min="0" :step="0.001" style="width:100%" clearable /></NFormItem>
    <NFormItem label="优先级"><NInputNumber v-model:value="form.priority" style="width:100%" /></NFormItem>
    <NFormItem label="状态"><NSelect v-model:value="form.status" :options="[{label:'启用',value:0},{label:'禁用',value:1}]" /></NFormItem>
  </FormDrawer>

  <ConfirmDialog v-model:show="confirmState.show" :title="confirmState.title" :content="confirmState.content" type="error" @confirm="handleConfirm" />
</template>
