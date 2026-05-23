<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NCard, NDescriptions, NDescriptionsItem, NTag, NButton, NSpin,
  NDataTable, NStatistic, NGrid, NGridItem, type DataTableColumns,
} from 'naive-ui'
import TimeDisplay from '@/components/TimeDisplay.vue'
import MoneyDisplay from '@/components/MoneyDisplay.vue'
import { userApi, type UserDetailResponse } from '@/api/user'
import { tokenApi, type Token } from '@/api/token'

const route = useRoute()
const router = useRouter()
const userId = Number(route.params.id)

const detail = ref<UserDetailResponse | null>(null)
const tokens = ref<Token[]>([])
const loading = ref(false)
const loadingTokens = ref(false)

async function load() {
  loading.value = true
  try {
    detail.value = await userApi.detail(userId)
  } catch (_) { /* handled */ } finally { loading.value = false }
}

async function loadTokens() {
  loadingTokens.value = true
  try {
    const res = await tokenApi.list({ user_id: userId, size: 10 })
    tokens.value = res.items
  } catch (_) { /* handled */ } finally { loadingTokens.value = false }
}

onMounted(() => { load(); loadTokens() })

const roleLabels: Record<number, string> = { 0: 'User', 1: 'Dept', 2: 'Tenant', 3: 'SuperAdmin' }
const statusLabels: Record<number, { label: string; type: 'success'|'error'|'warning' }> = {
  0: { label: '正常', type: 'success' }, 1: { label: '禁用', type: 'error' }, 2: { label: '待验证', type: 'warning' },
}

const tokenColumns: DataTableColumns<Token> = [
  { title: '名称', key: 'name', width: 160 },
  { title: '前缀', key: 'key_prefix', width: 120, render: (row) => h('code', { class: 'text-xs' }, row.key_prefix) },
  { title: '状态', key: 'status', width: 90, render: (row) => h(NTag, { type: row.status===0?'success':row.status===1?'error':'warning', size: 'small' }, { default: () => ['正常','禁用','已撤销'][row.status] }) },
  { title: '最后使用', key: 'last_used_at', width: 160, render: (row) => h(TimeDisplay, { value: row.last_used_at, relative: true }) },
]
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center gap-2 mb-2">
      <NButton text @click="router.push('/users')">← 返回用户列表</NButton>
    </div>

    <NSpin :show="loading">
      <div v-if="detail">
        <NCard title="用户基本信息" size="small" class="mb-4">
          <NDescriptions bordered label-placement="left" :column="2">
            <NDescriptionsItem label="ID">{{ detail.user.id }}</NDescriptionsItem>
            <NDescriptionsItem label="用户名">{{ detail.user.username }}</NDescriptionsItem>
            <NDescriptionsItem label="邮箱">{{ detail.user.email || '--' }}</NDescriptionsItem>
            <NDescriptionsItem label="显示名">{{ detail.user.display_name || '--' }}</NDescriptionsItem>
            <NDescriptionsItem label="角色">{{ roleLabels[detail.user.role] }}</NDescriptionsItem>
            <NDescriptionsItem label="状态">
              <NTag :type="statusLabels[detail.user.status]?.type" size="small">{{ statusLabels[detail.user.status]?.label }}</NTag>
            </NDescriptionsItem>
            <NDescriptionsItem label="分组">{{ detail.user.group_name || '--' }}</NDescriptionsItem>
            <NDescriptionsItem label="最后登录">
              <TimeDisplay :value="detail.user.last_login_at" relative />
            </NDescriptionsItem>
            <NDescriptionsItem label="创建时间">
              <TimeDisplay :value="detail.user.created_at" />
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>

        <NCard v-if="detail.wallet" title="钱包信息" size="small" class="mb-4">
          <NGrid :cols="3" :x-gap="16">
            <NGridItem>
              <NStatistic label="余额">
                <MoneyDisplay :quota="detail.wallet.quota_balance" />
              </NStatistic>
            </NGridItem>
            <NGridItem>
              <NStatistic label="累计充值">
                <MoneyDisplay :quota="detail.wallet.quota_total_recharged" />
              </NStatistic>
            </NGridItem>
            <NGridItem>
              <NStatistic label="累计消耗">
                <MoneyDisplay :quota="detail.wallet.quota_total_consumed" />
              </NStatistic>
            </NGridItem>
          </NGrid>
        </NCard>

        <NCard title="最近令牌" size="small">
          <NDataTable :columns="tokenColumns" :data="tokens" :loading="loadingTokens" size="small" />
        </NCard>
      </div>
    </NSpin>
  </div>
</template>
