<script setup lang="ts">
import { h, ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NCard, NDescriptions, NDescriptionsItem, NDataTable, NButton, NSpace, NFormItem,
  NInput, NInputNumber, NTag, NSpin, useMessage, type DataTableColumns,
} from 'naive-ui'
import { channelApi, type Channel, type Mapping, type ChannelInput } from '@/api/channel'
import FormDrawer from '@/components/FormDrawer.vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const isNew = computed(() => route.meta.mode === 'new' || route.path === '/channels/new')
const channelId = computed(() => isNew.value ? 0 : Number(route.params.id))

const channel = ref<Channel | null>(null)
const loadingChannel = ref(false)
const mappings = ref<Mapping[]>([])
const loadingMappings = ref(false)
const savingMappings = ref(false)

const editDrawer = ref(false)
const editLoading = ref(false)
const editForm = ref<Partial<ChannelInput>>({ name: '', provider: '', base_url: '', priority: 0, weight: 1 })

async function loadChannel() {
  if (isNew.value) return
  loadingChannel.value = true
  try {
    channel.value = await channelApi.detail(channelId.value)
    editForm.value = { name: channel.value.name, provider: channel.value.provider, base_url: channel.value.base_url, priority: channel.value.priority, weight: channel.value.weight }
  } catch (_) { /* handled */ } finally { loadingChannel.value = false }
}

async function loadMappings() {
  if (isNew.value) return
  loadingMappings.value = true
  try {
    const res = await channelApi.listMappings(channelId.value)
    mappings.value = res.items.map(m => ({ ...m }))
  } catch (_) { /* handled */ } finally { loadingMappings.value = false }
}

onMounted(() => { loadChannel(); loadMappings() })

async function saveEdit() {
  editLoading.value = true
  try {
    await channelApi.patch(channelId.value, editForm.value)
    message.success('已保存'); editDrawer.value = false; loadChannel()
  } catch (_) { /* handled */ } finally { editLoading.value = false }
}

function addMapping() { mappings.value.push({ client_model: '', upstream_model: '', input_ratio: null, output_ratio: null }) }
function removeMapping(idx: number) { mappings.value.splice(idx, 1) }

async function saveMappings() {
  savingMappings.value = true
  try {
    await channelApi.putMappings(channelId.value, mappings.value)
    message.success('映射已保存'); loadMappings()
  } catch (_) { /* handled */ } finally { savingMappings.value = false }
}

const mappingColumns: DataTableColumns<Mapping & { _idx: number }> = [
  { title: '客户端模型', key: 'client_model', render: (row) => h(NInput, { value: row.client_model, size: 'small', onUpdateValue: (v) => { mappings.value[row._idx].client_model = v } }) },
  { title: '上游模型', key: 'upstream_model', render: (row) => h(NInput, { value: row.upstream_model, size: 'small', onUpdateValue: (v) => { mappings.value[row._idx].upstream_model = v } }) },
  { title: '输入倍率', key: 'input_ratio', width: 100, render: (row) => h(NInputNumber, { value: row.input_ratio, size: 'small', min: 0, step: 0.01, onUpdateValue: (v) => { mappings.value[row._idx].input_ratio = v } }) },
  { title: '输出倍率', key: 'output_ratio', width: 100, render: (row) => h(NInputNumber, { value: row.output_ratio, size: 'small', min: 0, step: 0.01, onUpdateValue: (v) => { mappings.value[row._idx].output_ratio = v } }) },
  { title: '操作', key: 'actions', width: 80, render: (row) => h(NButton, { size: 'small', type: 'error', onClick: () => removeMapping(row._idx) }, { default: () => '删除' }) },
]
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center gap-2 mb-2">
      <NButton text @click="router.push('/channels')">← 返回渠道列表</NButton>
    </div>

    <NCard title="渠道基本信息" size="small">
      <NSpin :show="loadingChannel">
        <NDescriptions v-if="channel" bordered label-placement="left" :column="2">
          <NDescriptionsItem label="ID">{{ channel.id }}</NDescriptionsItem>
          <NDescriptionsItem label="名称">{{ channel.name }}</NDescriptionsItem>
          <NDescriptionsItem label="供应商">{{ channel.provider }}</NDescriptionsItem>
          <NDescriptionsItem label="Base URL">{{ channel.base_url }}</NDescriptionsItem>
          <NDescriptionsItem label="优先级">{{ channel.priority }}</NDescriptionsItem>
          <NDescriptionsItem label="权重">{{ channel.weight }}</NDescriptionsItem>
          <NDescriptionsItem label="状态">
            <NTag :type="channel.status===0?'success':channel.status===1?'error':'warning'" size="small">
              {{ ['启用','禁用','熔断'][channel.status] }}
            </NTag>
          </NDescriptionsItem>
          <NDescriptionsItem label="健康状态">{{ channel.health?.state ?? '--' }}</NDescriptionsItem>
        </NDescriptions>
        <div v-else-if="!loadingChannel" class="text-gray-400">暂无数据</div>
      </NSpin>
      <NSpace class="mt-3">
        <NButton @click="editDrawer=true" :disabled="!channel">编辑基本信息</NButton>
        <NButton @click="router.push(`/channels/${channelId}/mappings`)">管理模型映射</NButton>
      </NSpace>
    </NCard>

    <NCard title="模型映射" size="small">
      <NSpin :show="loadingMappings">
        <NDataTable
          :columns="mappingColumns"
          :data="mappings.map((m,i) => ({...m, _idx: i}))"
          size="small"
          :bordered="false"
        />
      </NSpin>
      <NSpace class="mt-3">
        <NButton @click="addMapping">添加映射</NButton>
        <NButton type="primary" :loading="savingMappings" @click="saveMappings">保存映射</NButton>
      </NSpace>
    </NCard>
  </div>

  <FormDrawer :show="editDrawer" mode="edit" title="编辑渠道" :loading="editLoading" @update:show="editDrawer=$event" @submit="saveEdit">
    <NFormItem label="名称"><NInput v-model:value="editForm.name" /></NFormItem>
    <NFormItem label="Base URL"><NInput v-model:value="editForm.base_url" /></NFormItem>
    <NFormItem label="优先级"><NInputNumber v-model:value="editForm.priority" style="width:100%" /></NFormItem>
    <NFormItem label="权重"><NInputNumber v-model:value="editForm.weight" :min="1" style="width:100%" /></NFormItem>
  </FormDrawer>
</template>
