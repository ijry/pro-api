<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NDataTable, NButton, NSpace, NInput, NInputNumber, NSpin, useMessage, type DataTableColumns } from 'naive-ui'
import { channelApi, type Mapping } from '@/api/channel'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const channelId = Number(route.params.id)

const mappings = ref<Mapping[]>([])
const loading = ref(false)
const saving = ref(false)

async function load() {
  loading.value = true
  try {
    const res = await channelApi.listMappings(channelId)
    mappings.value = res.items.map(m => ({ ...m }))
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

function addRow() { mappings.value.push({ client_model: '', upstream_model: '', input_ratio: null, output_ratio: null }) }
function removeRow(idx: number) { mappings.value.splice(idx, 1) }

async function save() {
  saving.value = true
  try {
    await channelApi.putMappings(channelId, mappings.value)
    message.success('映射已保存')
    load()
  } catch (_) { /* handled */ } finally { saving.value = false }
}

const columns: DataTableColumns<Mapping & { _idx: number }> = [
  { title: '客户端模型', key: 'client_model', render: (row) => h(NInput, { value: row.client_model, size: 'small', placeholder: 'gpt-4', onUpdateValue: (v) => { mappings.value[row._idx].client_model = v } }) },
  { title: '上游模型', key: 'upstream_model', render: (row) => h(NInput, { value: row.upstream_model, size: 'small', placeholder: 'gpt-4-0613', onUpdateValue: (v) => { mappings.value[row._idx].upstream_model = v } }) },
  { title: '输入倍率', key: 'input_ratio', width: 110, render: (row) => h(NInputNumber, { value: row.input_ratio, size: 'small', min: 0, step: 0.001, placeholder: '1.0', onUpdateValue: (v) => { mappings.value[row._idx].input_ratio = v } }) },
  { title: '输出倍率', key: 'output_ratio', width: 110, render: (row) => h(NInputNumber, { value: row.output_ratio, size: 'small', min: 0, step: 0.001, placeholder: '1.0', onUpdateValue: (v) => { mappings.value[row._idx].output_ratio = v } }) },
  { title: '操作', key: 'actions', width: 80, render: (row) => h(NButton, { size: 'small', type: 'error', onClick: () => removeRow(row._idx) }, { default: () => '删除' }) },
]
</script>

<template>
  <div>
    <div class="flex items-center gap-2 mb-4">
      <NButton text @click="router.push(`/channels/${channelId}`)">← 返回渠道详情</NButton>
      <span class="text-gray-400">|</span>
      <span class="font-semibold">模型映射管理（渠道 #{{ channelId }}）</span>
    </div>
    <NCard size="small">
      <NSpin :show="loading">
        <NDataTable :columns="columns" :data="mappings.map((m,i) => ({...m, _idx: i}))" size="small" :bordered="false" />
      </NSpin>
      <NSpace class="mt-3">
        <NButton @click="addRow">添加一行</NButton>
        <NButton type="primary" :loading="saving" @click="save">保存全部</NButton>
        <NButton @click="load">重置</NButton>
      </NSpace>
    </NCard>
  </div>
</template>
