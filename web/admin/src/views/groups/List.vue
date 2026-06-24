<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { NDataTable, NButton, NTag, NSpace, NSpin, NCheckbox, NCheckboxGroup, useMessage, type DataTableColumns } from 'naive-ui'
import ListPage from '@/components/ListPage.vue'
import FormDrawer from '@/components/FormDrawer.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { NFormItem, NInput, NInputNumber } from 'naive-ui'
import { groupApi, type Group, type GroupInput } from '@/api/group'
import { channelApi, type Channel } from '@/api/channel'

const message = useMessage()
const data = ref<Group[]>([])
const loading = ref(false)
const drawerShow = ref(false)
const drawerMode = ref<'create'|'edit'>('create')
const drawerLoading = ref(false)
const editingId = ref(0)
const form = ref<GroupInput>({ name: '', display_name: '', ratio: 1 })
const confirmState = ref({ show: false, type: '', id: 0, title: '', content: '' })

const allChannels = ref<Channel[]>([])
const channelLoading = ref(false)
const selectedChannelIds = ref<number[]>([])
const originalChannelIds = ref<number[]>([])
const channelCountByGroup = ref<Record<number, number>>({})

async function load() {
  loading.value = true
  try {
    const [groupRes, chanRes] = await Promise.all([
      groupApi.list(),
      channelApi.list({ size: 100 }),
    ])
    data.value = groupRes.items
    const counts: Record<number, number> = {}
    for (const ch of chanRes.items) {
      if (ch.group_id) counts[ch.group_id] = (counts[ch.group_id] ?? 0) + 1
    }
    channelCountByGroup.value = counts
  } catch (_) { /* handled */ } finally { loading.value = false }
}

async function loadChannels(groupId: number) {
  channelLoading.value = true
  try {
    const res = await channelApi.list({ size: 100 })
    allChannels.value = res.items
    const ids = res.items.filter(c => c.group_id === groupId).map(c => c.id)
    selectedChannelIds.value = [...ids]
    originalChannelIds.value = [...ids]
  } catch (_) { /* handled */ } finally { channelLoading.value = false }
}

onMounted(load)

function openCreate() {
  drawerMode.value = 'create'
  form.value = { name: '', display_name: '', ratio: 1 }
  selectedChannelIds.value = []
  originalChannelIds.value = []
  allChannels.value = []
  drawerShow.value = true
}

function openEdit(row: Group) {
  drawerMode.value = 'edit'
  editingId.value = row.id
  form.value = { name: row.name, display_name: row.display_name, ratio: row.ratio }
  selectedChannelIds.value = []
  originalChannelIds.value = []
  drawerShow.value = true
  loadChannels(row.id)
}

async function handleSubmit() {
  drawerLoading.value = true
  try {
    if (drawerMode.value === 'create') {
      await groupApi.create(form.value)
      message.success('分组已创建')
    } else {
      await groupApi.patch(editingId.value, form.value)
      const added = selectedChannelIds.value.filter(id => !originalChannelIds.value.includes(id))
      const removed = originalChannelIds.value.filter(id => !selectedChannelIds.value.includes(id))
      await Promise.all([
        ...added.map(id => channelApi.patch(id, { group_id: editingId.value })),
        ...removed.map(id => channelApi.patch(id, { group_id: 0 })),
      ])
      message.success('分组已更新')
    }
    drawerShow.value = false
    load()
  } catch (_) { /* handled */ } finally { drawerLoading.value = false }
}

async function handleConfirm() {
  const { type, id } = confirmState.value
  try {
    if (type === 'disable') { await groupApi.patch(id, { status: 1 }); message.success('已禁用') }
    else if (type === 'enable') { await groupApi.patch(id, { status: 0 }); message.success('已启用') }
    else if (type === 'delete') { await groupApi.remove(id); message.success('已删除') }
    load()
  } catch (_) { /* handled */ }
}

const columns: DataTableColumns<Group> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '名称', key: 'name', width: 120 },
  { title: '显示名', key: 'display_name', width: 140 },
  { title: '倍率', key: 'ratio', width: 80 },
  { title: '渠道数', key: 'channels', width: 70, render: (row) => channelCountByGroup.value[row.id] ?? 0 },
  { title: '状态', key: 'status', width: 90, render: (row) => h(NTag, { type: row.status===0?'success':'error', size: 'small' }, { default: () => row.status===0?'正常':'禁用' }) },
  { title: '操作', key: 'actions', width: 200, render: (row) => h(NSpace, { size: 'small' }, { default: () => [
    h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => '编辑' }),
    row.status === 0
      ? h(NButton, { size: 'small', type: 'warning', onClick: () => { confirmState.value = { show:true, type:'disable', id:row.id, title:'禁用分组', content:`确认禁用分组 "${row.display_name}"？` } } }, { default: () => '禁用' })
      : h(NButton, { size: 'small', type: 'primary', onClick: () => { confirmState.value = { show:true, type:'enable', id:row.id, title:'启用分组', content:`确认启用分组 "${row.display_name}"？` } } }, { default: () => '启用' }),
    h(NButton, { size: 'small', type: 'error', onClick: () => { confirmState.value = { show:true, type:'delete', id:row.id, title:'删除分组', content:`确认删除分组 "${row.display_name}"？` } } }, { default: () => '删除' }),
  ] }) },
]
</script>

<template>
  <ListPage title="分组管理">
    <template #actions>
      <NButton type="primary" @click="openCreate">新建分组</NButton>
    </template>
    <NDataTable :columns="columns" :data="data" :loading="loading" size="small" />
  </ListPage>

  <FormDrawer :show="drawerShow" :mode="drawerMode" :title="drawerMode==='create'?'新建分组':'编辑分组'" :loading="drawerLoading" @update:show="drawerShow=$event" @submit="handleSubmit">
    <NFormItem label="名称（唯一标识）"><NInput v-model:value="form.name" :disabled="drawerMode==='edit'" /></NFormItem>
    <NFormItem label="显示名"><NInput v-model:value="form.display_name" /></NFormItem>
    <NFormItem label="倍率"><NInputNumber v-model:value="form.ratio" :min="0" :step="0.01" style="width:100%" /></NFormItem>
    <NFormItem v-if="drawerMode === 'edit'" label="绑定渠道">
      <NSpin v-if="channelLoading" size="small" />
      <NCheckboxGroup v-else v-model:value="selectedChannelIds">
        <NSpace vertical :size="8">
          <NCheckbox
            v-for="ch in allChannels"
            :key="ch.id"
            :value="ch.id"
          >
            <NSpace align="center" :size="6">
              <span>{{ ch.name }}</span>
              <NTag size="small" :bordered="false">{{ ch.provider }}</NTag>
              <span
                v-if="ch.group_id && ch.group_id !== editingId"
                style="font-size:11px;color:var(--n-color-warning)"
              >已绑定其他分组</span>
            </NSpace>
          </NCheckbox>
        </NSpace>
      </NCheckboxGroup>
    </NFormItem>
  </FormDrawer>

  <ConfirmDialog v-model:show="confirmState.show" :title="confirmState.title" :content="confirmState.content" :type="confirmState.type==='delete'?'error':'warning'" @confirm="handleConfirm" />
</template>
