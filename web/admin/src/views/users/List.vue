<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import {
  NDataTable, NInput, NSelect, NButton, NTag, NSpace,
  NDropdown, NModal, NForm, NFormItem, NInputNumber,
  useMessage, type DataTableColumns,
} from 'naive-ui'
import { useRouter } from 'vue-router'
import ListPage from '@/components/ListPage.vue'
import TimeDisplay from '@/components/TimeDisplay.vue'
import MoneyDisplay from '@/components/MoneyDisplay.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { userApi } from '@/api/user'
import type { AdminUser } from '@/api/auth'

const message = useMessage()
const router = useRouter()

const data = ref<(AdminUser & { wallet?: { balance: number } })[]>([])
const total = ref(0)
const loading = ref(false)

const filter = ref({
  keyword: '',
  role: null as number | null,
  status: null as number | null,
  page: 1,
  size: 20,
})

const confirmState = ref({ show: false, type: '' as string, userId: 0, title: '', content: '' })
const quotaModal = ref({ show: false, userId: 0, delta: 0, reason: '' })
const resetPassModal = ref({ show: false, tempPassword: '' })

async function load() {
  loading.value = true
  try {
    const res = await userApi.list({
      page: filter.value.page,
      size: filter.value.size,
      keyword: filter.value.keyword || undefined,
      role: filter.value.role !== null ? filter.value.role as 0|1|2|3 : undefined,
      status: filter.value.status !== null ? filter.value.status as 0|1|2 : undefined,
    })
    data.value = res.items
    total.value = res.total
  } catch (_) { /* handled */ }
  finally { loading.value = false }
}

onMounted(() => {
  load()
})

watch(() => filter.value.page, load)

const roleLabels: Record<number, { label: string; type: 'default' | 'info' | 'warning' | 'error' | 'success' | 'primary' }> = {
  0: { label: 'User', type: 'default' },
  1: { label: 'Dept', type: 'info' },
  2: { label: 'Tenant', type: 'warning' },
  3: { label: 'SuperAdmin', type: 'error' },
}

const statusLabels: Record<number, { label: string; type: 'default' | 'info' | 'warning' | 'error' | 'success' | 'primary' }> = {
  0: { label: '正常', type: 'success' },
  1: { label: '禁用', type: 'error' },
  2: { label: '待验证', type: 'warning' },
}

const columns: DataTableColumns<AdminUser & { wallet?: { balance: number } }> = [
  { title: 'ID', key: 'id', width: 100, render: (row) => h('span', { class: 'font-mono text-xs' }, String(row.id)) },
  {
    title: '用户名', key: 'username', width: 160,
    render: (row) => h('a', { class: 'text-blue-500 hover:underline cursor-pointer', onClick: () => router.push(`/users/${row.id}`) }, row.username),
  },
  { title: '邮箱', key: 'email', width: 200, render: (row) => row.email || '--' },
  {
    title: 'Role', key: 'role', width: 100,
    render: (row) => h(NTag, { type: roleLabels[row.role]?.type ?? 'default', size: 'small' }, { default: () => roleLabels[row.role]?.label ?? row.role }),
  },
  {
    title: '状态', key: 'status', width: 100,
    render: (row) => h(NTag, { type: statusLabels[row.status]?.type ?? 'default', size: 'small' }, { default: () => statusLabels[row.status]?.label ?? row.status }),
  },
  {
    title: '余额', key: 'wallet', width: 140,
    render: (row) => h(MoneyDisplay, { quota: row.wallet?.balance ?? 0 }),
  },
  {
    title: '最后登录', key: 'last_login_at', width: 160,
    render: (row) => h(TimeDisplay, { value: row.last_login_at, relative: true }),
  },
  {
    title: '操作', key: 'actions', width: 180, fixed: 'right',
    render: (row) => h(NSpace, { size: 'small' }, {
      default: () => [
        h(NButton, { size: 'small', onClick: () => router.push(`/users/${row.id}`) }, { default: () => '详情' }),
        h(NDropdown, {
          options: [
            { label: row.status === 1 ? '启用' : '禁用', key: row.status === 1 ? 'enable' : 'disable' },
            { label: '调额度', key: 'quota' },
            { label: '重置密码', key: 'reset_password' },
            { type: 'divider', key: 'd1' },
            { label: '删除', key: 'delete' },
          ],
          onSelect: (key: string) => handleAction(key, row),
        }, {
          default: () => h(NButton, { size: 'small' }, { default: () => '更多 ▾' }),
        }),
      ],
    }),
  },
]

// Need h for render functions
import { h } from 'vue'

function handleAction(key: string, row: AdminUser) {
  if (key === 'disable') {
    confirmState.value = { show: true, type: 'disable', userId: row.id, title: '禁用用户', content: `确认禁用用户 ${row.username}？` }
  } else if (key === 'enable') {
    confirmState.value = { show: true, type: 'enable', userId: row.id, title: '启用用户', content: `确认启用用户 ${row.username}？` }
  } else if (key === 'delete') {
    confirmState.value = { show: true, type: 'delete', userId: row.id, title: '删除用户', content: `确认删除用户 ${row.username}？此操作不可撤销。` }
  } else if (key === 'quota') {
    quotaModal.value = { show: true, userId: row.id, delta: 0, reason: '' }
  } else if (key === 'reset_password') {
    confirmState.value = { show: true, type: 'reset_password', userId: row.id, title: '重置密码', content: `确认重置用户 ${row.username} 的密码？` }
  }
}

async function handleConfirm() {
  const { type, userId } = confirmState.value
  try {
    if (type === 'disable') {
      await userApi.patch(userId, { status: 1 })
      message.success('已禁用')
    } else if (type === 'enable') {
      await userApi.patch(userId, { status: 0 })
      message.success('已启用')
    } else if (type === 'delete') {
      await userApi.remove(userId)
      message.success('已删除')
    } else if (type === 'reset_password') {
      const res = await userApi.resetPassword(userId)
      if (res.temp_password) {
        resetPassModal.value = { show: true, tempPassword: res.temp_password }
      }
    }
    load()
  } catch (_) { /* handled by interceptor */ }
}

async function handleQuotaSubmit() {
  try {
    await userApi.adjustQuota(quotaModal.value.userId, { delta_quota: quotaModal.value.delta, reason: quotaModal.value.reason })
    message.success('额度已调整')
    quotaModal.value.show = false
    load()
  } catch (_) { /* handled */ }
}
</script>

<template>
  <ListPage title="用户管理">
    <template #actions>
      <NButton @click="load">刷新</NButton>
    </template>
    <template #filters>
      <NInput v-model:value="filter.keyword" placeholder="搜索用户名/邮箱" clearable style="width: 200px" @update:value="() => { filter.page = 1; load() }" />
      <NSelect v-model:value="filter.role" placeholder="角色" :options="[{label:'User',value:0},{label:'Dept',value:1},{label:'Tenant',value:2},{label:'SuperAdmin',value:3}]" clearable style="width: 130px" @update:value="() => { filter.page = 1; load() }" />
      <NSelect v-model:value="filter.status" placeholder="状态" :options="[{label:'正常',value:0},{label:'禁用',value:1},{label:'待验证',value:2}]" clearable style="width: 100px" @update:value="() => { filter.page = 1; load() }" />
    </template>

    <NDataTable
      :columns="columns"
      :data="data"
      :loading="loading"
      :pagination="{
        page: filter.page,
        pageSize: filter.size,
        itemCount: total,
        onChange: (p: number) => { filter.page = p },
      }"
      remote
      scroll-x="1200"
      size="small"
    />
  </ListPage>

  <ConfirmDialog
    v-model:show="confirmState.show"
    :title="confirmState.title"
    :content="confirmState.content"
    :type="confirmState.type === 'delete' ? 'error' : 'warning'"
    @confirm="handleConfirm"
  />

  <NModal v-model:show="quotaModal.show" preset="card" title="调整额度" style="width: 400px">
    <NForm label-placement="top">
      <NFormItem label="额度变更（正数增加，负数扣减）">
        <NInputNumber v-model:value="quotaModal.delta" style="width: 100%" />
      </NFormItem>
      <NFormItem label="原因">
        <NInput v-model:value="quotaModal.reason" type="textarea" :rows="2" />
      </NFormItem>
    </NForm>
    <NSpace justify="end">
      <NButton @click="quotaModal.show = false">取消</NButton>
      <NButton type="primary" @click="handleQuotaSubmit">确认</NButton>
    </NSpace>
  </NModal>

  <NModal v-model:show="resetPassModal.show" preset="card" title="密码已重置" style="width: 400px">
    <p class="mb-2 text-sm text-gray-500">临时密码（一次性，请及时修改）：</p>
    <div class="flex items-center gap-2">
      <code class="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-sm font-mono">{{ resetPassModal.tempPassword }}</code>
    </div>
    <template #footer>
      <NButton type="primary" @click="resetPassModal.show = false">关闭</NButton>
    </template>
  </NModal>
</template>
