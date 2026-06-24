<script setup lang="ts">
import { h, ref, onMounted, computed } from 'vue'
import {
  NDataTable, NInput, NSelect, NButton, NTag, NSpace, NCard, NGrid, NGi,
  NStatistic, NDropdown, type DataTableColumns, type SelectOption,
} from 'naive-ui'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import ListPage from '@/components/ListPage.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import AddDialog from './AddDialog.vue'
import { accountApi, type Account, type StatsResp } from '@/api/account'
import { channelApi, type Channel } from '@/api/channel'
import { useAccountActions } from './composables/useAccountActions'

const router = useRouter()
const { t } = useI18n()

const data = ref<Account[]>([])
const total = ref(0)
const loading = ref(false)
const stats = ref<StatsResp | null>(null)

const filter = ref({
  channel_id: null as number | null,
  status: null as number | null,
  keyword: '',
  provider: null as string | null,
  tier: null as string | null,
})

const channelOptions = ref<SelectOption[]>([])
async function loadChannels() {
  try {
    const r = await channelApi.list({ size: 200 })
    channelOptions.value = r.items.map((c: Channel) => ({ label: c.name, value: c.id }))
    if (!filter.value.channel_id && channelOptions.value.length > 0) {
      filter.value.channel_id = channelOptions.value[0].value as number
    }
  } catch (_) { /* handled */ }
}

const providerOptions: SelectOption[] = [
  { label: 'Anthropic', value: 'anthropic' },
  { label: 'OpenAI', value: 'openai' },
]
const tierOptions: SelectOption[] = [
  { label: 'free', value: 'free' },
  { label: 'plus', value: 'plus' },
  { label: 'pro', value: 'pro' },
  { label: 'team', value: 'team' },
  { label: 'enterprise', value: 'enterprise' },
]
const statusOptions = computed<SelectOption[]>(() => [
  { label: t('accounts.status.0'), value: 0 },
  { label: t('accounts.status.1'), value: 1 },
  { label: t('accounts.status.2'), value: 2 },
  { label: t('accounts.status.3'), value: 3 },
  { label: t('accounts.status.4'), value: 4 },
])

async function load() {
  if (!filter.value.channel_id) {
    data.value = []
    total.value = 0
    return
  }
  loading.value = true
  try {
    const res = await accountApi.list({
      channel_id: filter.value.channel_id ?? undefined,
      status: filter.value.status ?? undefined,
    })
    let items = res.items
    if (filter.value.provider) items = items.filter(a => a.provider === filter.value.provider)
    if (filter.value.tier) items = items.filter(a => a.tier === filter.value.tier)
    if (filter.value.keyword) {
      const kw = filter.value.keyword.toLowerCase()
      items = items.filter(a => a.name.toLowerCase().includes(kw) || a.email.toLowerCase().includes(kw))
    }
    data.value = items
    total.value = items.length
  } catch (_) { /* handled */ } finally { loading.value = false }
}

async function loadStats() {
  try { stats.value = await accountApi.stats() } catch (_) { /* handled */ }
}

onMounted(async () => { await loadChannels(); await load(); loadStats() })

function reset() {
  filter.value = { channel_id: channelOptions.value[0]?.value as number ?? null, status: null, keyword: '', provider: null, tier: null }
  load()
}

const showAdd = ref(false)
const confirmState = ref({ show: false, type: '', row: null as Account | null, title: '', content: '', requireInput: '' })

function askConfirm(type: 'delete' | 'disable' | 'clear_cd', row: Account) {
  const map: Record<string, { titleKey: string; contentKey: string; requireInput?: string }> = {
    delete: { titleKey: 'accounts.confirm.delete_title', contentKey: 'accounts.confirm.delete_content', requireInput: row.name },
    disable: { titleKey: 'accounts.confirm.disable_title', contentKey: 'accounts.confirm.disable_content' },
    clear_cd: { titleKey: 'accounts.confirm.clear_cd_title', contentKey: 'accounts.confirm.clear_cd_content' },
  }
  const m = map[type]
  confirmState.value = {
    show: true, type, row,
    title: t(m.titleKey),
    content: t(m.contentKey, { name: row.name }),
    requireInput: m.requireInput ?? '',
  }
}

const actions = useAccountActions(() => { load(); loadStats() })

async function handleConfirm() {
  const { type, row } = confirmState.value
  if (!row) return
  if (type === 'delete') await actions.doDelete(row)
  else if (type === 'disable') await actions.doDisable(row)
  else if (type === 'clear_cd') await actions.doClearCooldown(row)
}

function statusTagType(s: number): 'success' | 'warning' | 'error' | 'default' {
  if (s === 0) return 'success'
  if (s === 1) return 'default'
  if (s === 2) return 'warning'
  return 'error'
}

function fmtPct(q: Account['quota_5h']) {
  if (typeof q?.pct === 'number') return `${Math.round(q.pct * 100)}%`
  if (q?.total && q?.remaining != null) return `${Math.round((q.remaining / q.total) * 100)}%`
  return '--'
}

function fmtTime(s?: string | null) {
  if (!s) return '--'
  try { return new Date(s).toLocaleString() } catch { return s }
}

const columns = computed<DataTableColumns<Account>>(() => [
  { title: t('accounts.columns.id'), key: 'id', width: 70 },
  {
    title: t('accounts.columns.name'), key: 'name', width: 200,
    render: (row) => h('a', {
      class: 'text-blue-500 hover:underline cursor-pointer',
      onClick: () => router.push(`/accounts/${row.id}`),
    }, row.name),
  },
  {
    title: t('accounts.columns.channel'), key: 'channel_id', width: 110,
    render: (row) => {
      const opt = channelOptions.value.find(c => c.value === row.channel_id)
      return (typeof opt?.label === 'string' ? opt.label : '') || `#${row.channel_id}`
    },
  },
  { title: t('accounts.columns.provider'), key: 'provider', width: 100 },
  { title: t('accounts.columns.tier'), key: 'tier', width: 90 },
  {
    title: t('accounts.columns.cred_type'), key: 'cred_type', width: 110,
    render: (row) => h(NTag, { size: 'small' }, { default: () => t(`accounts.cred_type.${row.cred_type}`) }),
  },
  {
    title: t('accounts.columns.status'), key: 'status', width: 90,
    render: (row) => h(NTag, { type: statusTagType(row.status), size: 'small' }, { default: () => t(`accounts.status.${row.status}`) }),
  },
  { title: t('accounts.columns.quota_5h'), key: 'quota_5h', width: 90, render: (row) => fmtPct(row.quota_5h) },
  { title: t('accounts.columns.quota_week'), key: 'quota_week', width: 90, render: (row) => fmtPct(row.quota_week) },
  { title: t('accounts.columns.consec_failures'), key: 'consec_failures', width: 90 },
  { title: t('accounts.columns.last_used_at'), key: 'last_used_at', width: 170, render: (row) => fmtTime(row.last_used_at) },
  {
    title: t('accounts.columns.actions'), key: 'actions', width: 280, fixed: 'right',
    render: (row) => {
      const moreOptions = [
        { label: t('accounts.actions.clear_cd'), key: 'clear_cd', disabled: row.status !== 2 },
        row.status === 1
          ? { label: t('accounts.actions.enable'), key: 'enable' }
          : { label: t('accounts.actions.disable'), key: 'disable' },
        { label: t('accounts.actions.delete'), key: 'delete' },
      ]
      const onSelect = (k: string) => {
        if (k === 'enable') actions.doEnable(row)
        else if (k === 'disable') askConfirm('disable', row)
        else if (k === 'clear_cd') askConfirm('clear_cd', row)
        else if (k === 'delete') askConfirm('delete', row)
      }
      return h(NSpace, { size: 'small' }, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => actions.doTest(row) }, { default: () => t('accounts.actions.test') }),
          h(NButton, { size: 'small', onClick: () => actions.doRefresh(row), disabled: row.cred_type !== 'oauth' && row.cred_type !== 'token_pasted' }, { default: () => t('accounts.actions.refresh') }),
          h(NButton, { size: 'small', onClick: () => router.push(`/accounts/${row.id}`) }, { default: () => t('accounts.actions.view') }),
          h(NDropdown, { options: moreOptions, onSelect, trigger: 'click' }, {
            default: () => h(NButton, { size: 'small' }, { default: () => '⋯' }),
          }),
        ],
      })
    },
  },
])

const kpi = computed(() => {
  const totals = stats.value?.totals ?? {}
  return [
    { label: t('accounts.kpi.total'), value: totals.total ?? 0 },
    { label: t('accounts.kpi.active'), value: totals.active ?? 0 },
    { label: t('accounts.kpi.cooldown'), value: totals.cooldown ?? 0, color: '#f0a020' },
    { label: t('accounts.kpi.expired'), value: totals.expired ?? 0, color: '#d03050' },
    { label: t('accounts.kpi.low_5h'), value: totals.low_5h ?? 0, color: '#f0a020' },
    { label: t('accounts.kpi.expiring'), value: totals.expiring ?? 0, color: '#f0a020' },
  ]
})
</script>

<template>
  <ListPage :title="t('accounts.title')">
    <template #actions>
      <NButton type="primary" @click="showAdd = true">{{ t('accounts.actions.add') }}</NButton>
    </template>

    <NGrid :cols="6" :x-gap="12" class="mb-3">
      <NGi v-for="k in kpi" :key="k.label">
        <NCard size="small">
          <NStatistic :label="k.label" :value="k.value" :value-style="k.color ? { color: k.color } : undefined" />
        </NCard>
      </NGi>
    </NGrid>

    <template #filters>
      <NSelect v-model:value="filter.channel_id" :placeholder="t('accounts.filter.channel')" :options="channelOptions" clearable style="width:160px" @update:value="load" />
      <NSelect v-model:value="filter.provider" :placeholder="t('accounts.filter.provider')" :options="providerOptions" clearable style="width:130px" @update:value="load" />
      <NSelect v-model:value="filter.tier" :placeholder="t('accounts.filter.tier')" :options="tierOptions" clearable style="width:120px" @update:value="load" />
      <NSelect v-model:value="filter.status" :placeholder="t('accounts.filter.status')" :options="statusOptions" clearable style="width:120px" @update:value="load" />
      <NInput v-model:value="filter.keyword" :placeholder="t('accounts.filter.keyword')" clearable style="width:180px" @update:value="load" />
      <NButton size="small" @click="reset">{{ t('accounts.filter.reset') }}</NButton>
    </template>

    <NDataTable
      :columns="columns"
      :data="data"
      :loading="loading"
      :scroll-x="1700"
      size="small"
      :row-key="(row: Account) => row.id"
      :pagination="false"
    />
  </ListPage>

  <AddDialog v-model:show="showAdd" @imported="() => { load(); loadStats() }" />

  <ConfirmDialog
    v-model:show="confirmState.show"
    :title="confirmState.title"
    :content="confirmState.content"
    :type="confirmState.type === 'delete' ? 'error' : 'warning'"
    :require-input="confirmState.requireInput || undefined"
    @confirm="handleConfirm"
  />
</template>
