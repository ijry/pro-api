<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { NDataTable, NButton, NTag, NSelect, NInput, NSpace, useMessage, type DataTableColumns } from 'naive-ui'
import { NFormItem, NInputNumber, NCheckbox, NCheckboxGroup } from 'naive-ui'
import ListPage from '@/components/ListPage.vue'
import FormDrawer from '@/components/FormDrawer.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { modelApi, type ModelCatalog, type ModelInput } from '@/api/model'

const message = useMessage()
const data = ref<ModelCatalog[]>([])
const total = ref(0)
const loading = ref(false)
const filter = ref({ keyword: '', family: null as string|null, status: null as 0|1|null, page: 1, size: 20 })
const drawerShow = ref(false)
const drawerMode = ref<'create'|'edit'>('create')
const drawerLoading = ref(false)
const editingId = ref(0)
const form = ref<ModelInput>({ name: '', family: 'chat', capabilities: [], default_input_ratio: 1, default_output_ratio: 1, max_input_tokens: 128000, status: 0 })
const confirmState = ref({ show: false, id: 0, title: '', content: '' })

const familyOptions = [
  { label: 'Chat', value: 'chat' }, { label: 'Embed', value: 'embed' },
  { label: 'Image', value: 'image' }, { label: 'Audio', value: 'audio' }, { label: 'Rerank', value: 'rerank' },
]
const capOptions = ['stream','function_call','vision','json_mode','reasoning']

async function load() {
  loading.value = true
  try {
    const res = await modelApi.list({ keyword: filter.value.keyword||undefined, family: filter.value.family ?? undefined, status: filter.value.status ?? undefined, page: filter.value.page, size: filter.value.size })
    data.value = res.items; total.value = res.total
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

function openCreate() {
  drawerMode.value = 'create'
  form.value = { name: '', family: 'chat', capabilities: [], default_input_ratio: 1, default_output_ratio: 1, max_input_tokens: 128000, status: 0 }
  drawerShow.value = true
}

function openEdit(row: ModelCatalog) {
  drawerMode.value = 'edit'; editingId.value = row.id
  form.value = { name: row.name, family: row.family, capabilities: [...row.capabilities], default_input_ratio: row.default_input_ratio, default_output_ratio: row.default_output_ratio, max_input_tokens: row.max_input_tokens, status: row.status }
  drawerShow.value = true
}

async function handleSubmit() {
  drawerLoading.value = true
  try {
    if (drawerMode.value === 'create') { await modelApi.create(form.value); message.success('模型已创建') }
    else { await modelApi.patch(editingId.value, form.value); message.success('模型已更新') }
    drawerShow.value = false; load()
  } catch (_) { /* handled */ } finally { drawerLoading.value = false }
}

async function handleConfirm() {
  try { await modelApi.remove(confirmState.value.id); message.success('已删除'); load() } catch (_) { /* handled */ }
}

const columns: DataTableColumns<ModelCatalog> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '模型名', key: 'name', width: 200, render: (row) => h('code', { class: 'text-xs' }, row.name) },
  { title: '家族', key: 'family', width: 80 },
  { title: '能力', key: 'capabilities', width: 200, render: (row) => h(NSpace, { size: 4 }, { default: () => row.capabilities.map(c => h(NTag, { size: 'small', type: 'info' as const }, { default: () => c })) }) },
  { title: '输入倍率', key: 'default_input_ratio', width: 90 },
  { title: '输出倍率', key: 'default_output_ratio', width: 90 },
  { title: '最大tokens', key: 'max_input_tokens', width: 100 },
  { title: '状态', key: 'status', width: 80, render: (row) => h(NTag, { type: row.status===0?'success':'error', size: 'small' }, { default: () => row.status===0?'启用':'禁用' }) },
  { title: '操作', key: 'actions', width: 160, fixed: 'right', render: (row) => h(NSpace, { size: 'small' }, { default: () => [
    h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => '编辑' }),
    h(NButton, { size: 'small', type: 'error', onClick: () => { confirmState.value = { show:true, id:row.id, title:'删除模型', content:`确认删除模型 "${row.name}"？` } } }, { default: () => '删除' }),
  ] }) },
]
</script>

<template>
  <ListPage title="模型字典">
    <template #actions>
      <NButton type="primary" @click="openCreate">新建模型</NButton>
    </template>
    <template #filters>
      <NInput v-model:value="filter.keyword" placeholder="搜索名称" clearable style="width:180px" @update:value="() => { filter.page=1; load() }" />
      <NSelect v-model:value="filter.family" placeholder="家族" :options="familyOptions" clearable style="width:110px" @update:value="() => { filter.page=1; load() }" />
      <NSelect v-model:value="filter.status" placeholder="状态" :options="[{label:'启用',value:0},{label:'禁用',value:1}]" clearable style="width:100px" @update:value="() => { filter.page=1; load() }" />
    </template>
    <NDataTable :columns="columns" :data="data" :loading="loading" :pagination="{ page: filter.page, pageSize: filter.size, itemCount: total, onChange: (p:number)=>{ filter.page=p; load() } }" remote scroll-x="1100" size="small" />
  </ListPage>

  <FormDrawer :show="drawerShow" :mode="drawerMode" :title="drawerMode==='create'?'新建模型':'编辑模型'" :loading="drawerLoading" @update:show="drawerShow=$event" @submit="handleSubmit">
    <NFormItem label="模型名（唯一标识）">
      <NInput v-model:value="form.name" :disabled="drawerMode==='edit'" placeholder="gpt-4o" />
    </NFormItem>
    <NFormItem label="家族">
      <NSelect v-model:value="form.family" :options="familyOptions" />
    </NFormItem>
    <NFormItem label="能力">
      <NCheckboxGroup v-model:value="form.capabilities">
        <NCheckbox v-for="c in capOptions" :key="c" :value="c" :label="c" />
      </NCheckboxGroup>
    </NFormItem>
    <NFormItem label="输入倍率">
      <NInputNumber v-model:value="form.default_input_ratio" :min="0" :step="0.001" style="width:100%" />
    </NFormItem>
    <NFormItem label="输出倍率">
      <NInputNumber v-model:value="form.default_output_ratio" :min="0" :step="0.001" style="width:100%" />
    </NFormItem>
    <NFormItem label="最大输入tokens">
      <NInputNumber v-model:value="form.max_input_tokens" :min="1" style="width:100%" />
    </NFormItem>
    <NFormItem label="状态">
      <NSelect v-model:value="form.status" :options="[{label:'启用',value:0},{label:'禁用',value:1}]" />
    </NFormItem>
  </FormDrawer>

  <ConfirmDialog v-model:show="confirmState.show" :title="confirmState.title" :content="confirmState.content" type="error" @confirm="handleConfirm" />
</template>
