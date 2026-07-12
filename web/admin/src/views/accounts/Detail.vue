<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  NCard, NDescriptions, NDescriptionsItem, NSpace, NButton, NTag, NInput, NInputNumber, NSelect, NPageHeader, NSpin, NGrid, NGi,
  useMessage, type SelectOption,
} from 'naive-ui'
import { accountApi, type AccountDetail, type QuotaMode } from '@/api/account'
import QuotaRing from './components/QuotaRing.vue'
import EventTimeline from './components/EventTimeline.vue'
import CredentialPeek from './components/CredentialPeek.vue'
import { useAccountActions } from './composables/useAccountActions'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const message = useMessage()

const id = computed(() => Number(route.params.id))
const data = ref<AccountDetail | null>(null)
const loading = ref(false)

async function load() {
  loading.value = true
  try { data.value = await accountApi.get(id.value) } catch (_) { /* handled */ } finally { loading.value = false }
}
onMounted(load)

// tag 行内编辑
const editingTag = ref(false)
const tagDraft = ref('')
const savingTag = ref(false)
function startEditTag() {
  tagDraft.value = data.value?.tag ?? ''
  editingTag.value = true
}
async function saveTag() {
  if (!data.value) return
  savingTag.value = true
  try {
    const updated = await accountApi.patch(data.value.id, { tag: tagDraft.value })
    data.value = updated
    editingTag.value = false
    message.success(t('accounts.detail.tag_saved'))
  } catch (_) { /* handled */ } finally { savingTag.value = false }
}

// quota_mode / 手动额度 行内编辑
const editingQuota = ref(false)
const savingQuota = ref(false)
const quotaDraft = ref({
  mode: 'auto' as QuotaMode,
  total: null as number | null,
  remaining: null as number | null,
})
function startEditQuota() {
  if (!data.value) return
  quotaDraft.value = {
    mode: data.value.quota_mode ?? 'auto',
    total: data.value.quota_5h?.total ?? null,
    remaining: data.value.quota_5h?.remaining ?? null,
  }
  editingQuota.value = true
}
async function saveQuota() {
  if (!data.value) return
  savingQuota.value = true
  try {
    // 仅 manual 模式提交手填额度;auto/none 只改模式(后端对非 manual 会忽略额度)。
    const payload = {
      quota_mode: quotaDraft.value.mode,
      ...(quotaDraft.value.mode === 'manual'
        ? { quota_total: quotaDraft.value.total, quota_remaining: quotaDraft.value.remaining }
        : {}),
    }
    const updated = await accountApi.patch(data.value.id, payload)
    data.value = updated
    editingQuota.value = false
    message.success(t('accounts.detail.quota_saved'))
  } catch (_) { /* handled */ } finally { savingQuota.value = false }
}

const actions = useAccountActions(load)

const quotaModeOptions = computed<SelectOption[]>(() => [
  { label: t('accounts.quota_mode.auto'), value: 'auto' },
  { label: t('accounts.quota_mode.manual'), value: 'manual' },
  { label: t('accounts.quota_mode.none'), value: 'none' },
])

function quotaModeTagType(m?: QuotaMode): 'success' | 'warning' | 'default' {
  if (m === 'manual') return 'warning'
  if (m === 'none') return 'default'
  return 'success'
}

function statusTagType(s: number): 'success' | 'warning' | 'error' | 'default' {
  if (s === 0) return 'success'
  if (s === 1) return 'default'
  if (s === 2) return 'warning'
  return 'error'
}

function fmtTime(s?: string | null) {
  if (!s) return '--'
  try { return new Date(s).toLocaleString() } catch { return s }
}

function refreshValidLabel(v?: 0 | 1 | 2) {
  if (v === 1) return t('accounts.detail.refresh_valid_ok')
  if (v === 2) return t('accounts.detail.refresh_valid_invalid')
  return t('accounts.detail.refresh_valid_unknown')
}
function refreshValidType(v?: 0 | 1 | 2): 'success' | 'error' | 'default' {
  if (v === 1) return 'success'
  if (v === 2) return 'error'
  return 'default'
}
</script>

<template>
  <NSpin :show="loading">
    <NPageHeader :title="data ? data.name : '--'" @back="router.push('/accounts')">
      <template #extra>
        <NSpace>
          <NButton size="small" @click="actions.doTest(data!)" :disabled="!data">{{ t('accounts.actions.test') }}</NButton>
          <NButton size="small" @click="actions.doRefresh(data!)" :disabled="!data || (data!.cred_type !== 'oauth' && data!.cred_type !== 'token_pasted')">{{ t('accounts.actions.refresh') }}</NButton>
          <NButton size="small" :disabled="!data || data!.status !== 2" @click="actions.doClearCooldown(data!)">{{ t('accounts.actions.clear_cd') }}</NButton>
          <CredentialPeek v-if="data" :account-id="data.id" :account-name="data.name" />
        </NSpace>
      </template>
    </NPageHeader>

    <div v-if="data" class="mt-3">
      <NGrid :cols="2" :x-gap="12">
        <NGi>
          <NCard size="small" :title="t('accounts.detail.basic')">
            <NDescriptions :column="2" size="small" bordered>
              <NDescriptionsItem :label="t('accounts.columns.id')">{{ data.id }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.name')">{{ data.name }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.tag')">
                <NSpace v-if="!editingTag" size="small" align="center">
                  <NTag v-if="data.tag" size="small" type="info">{{ data.tag }}</NTag>
                  <span v-else>--</span>
                  <NButton text size="tiny" @click="startEditTag">{{ t('accounts.actions.edit') }}</NButton>
                </NSpace>
                <NSpace v-else size="small" align="center">
                  <NInput v-model:value="tagDraft" size="tiny" style="width:140px" :placeholder="t('accounts.add_dialog.tag_placeholder')" />
                  <NButton text size="tiny" type="primary" :loading="savingTag" @click="saveTag">{{ t('accounts.add_dialog.submit') }}</NButton>
                  <NButton text size="tiny" @click="editingTag = false">{{ t('accounts.add_dialog.cancel') }}</NButton>
                </NSpace>
              </NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.channel')">#{{ data.channel_id }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.provider')">{{ data.provider }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.tier')">{{ data.tier }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.cred_type')">{{ t(`accounts.cred_type.${data.cred_type}`) }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.quota_mode')">
                <NSpace v-if="!editingQuota" size="small" align="center">
                  <NTag size="small" :type="quotaModeTagType(data.quota_mode)">
                    {{ t(`accounts.quota_mode.${data.quota_mode || 'auto'}`) }}
                  </NTag>
                  <NButton text size="tiny" @click="startEditQuota">{{ t('accounts.actions.edit') }}</NButton>
                </NSpace>
                <NSpace v-else vertical size="small">
                  <NSelect v-model:value="quotaDraft.mode" :options="quotaModeOptions" size="tiny" style="width:160px" />
                  <template v-if="quotaDraft.mode === 'manual'">
                    <NInputNumber v-model:value="quotaDraft.total" :min="0" size="tiny" style="width:160px" :placeholder="t('accounts.add_dialog.quota_total_label')" />
                    <NInputNumber v-model:value="quotaDraft.remaining" :min="0" size="tiny" style="width:160px" :placeholder="t('accounts.add_dialog.quota_remaining_label')" />
                  </template>
                  <NSpace size="small" align="center">
                    <NButton text size="tiny" type="primary" :loading="savingQuota" @click="saveQuota">{{ t('accounts.add_dialog.submit') }}</NButton>
                    <NButton text size="tiny" @click="editingQuota = false">{{ t('accounts.add_dialog.cancel') }}</NButton>
                  </NSpace>
                </NSpace>
              </NDescriptionsItem>
              <NDescriptionsItem label="email">{{ data.email || '--' }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.status')">
                <NTag :type="statusTagType(data.status)" size="small">{{ t(`accounts.status.${data.status}`) }}</NTag>
              </NDescriptionsItem>
              <NDescriptionsItem label="priority/weight">{{ data.priority }} / {{ data.weight }}</NDescriptionsItem>
              <NDescriptionsItem label="import_source">{{ data.import_source || '--' }}</NDescriptionsItem>
              <NDescriptionsItem label="external_account_id">{{ data.external_account_id || '--' }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.last_used_at')">{{ fmtTime(data.last_used_at) }}</NDescriptionsItem>
              <NDescriptionsItem label="last_success_at">{{ fmtTime(data.last_success_at) }}</NDescriptionsItem>
              <NDescriptionsItem label="last_failure_at">{{ fmtTime(data.last_failure_at) }}</NDescriptionsItem>
              <NDescriptionsItem label="cooldown_until">{{ fmtTime(data.cooldown_until) }}</NDescriptionsItem>
              <NDescriptionsItem label="created_at">{{ fmtTime(data.created_at) }}</NDescriptionsItem>
              <NDescriptionsItem label="updated_at">{{ fmtTime(data.updated_at) }}</NDescriptionsItem>
            </NDescriptions>
          </NCard>
        </NGi>

        <NGi>
          <NCard size="small" :title="t('accounts.detail.token_health')">
            <NDescriptions :column="1" size="small" bordered>
              <NDescriptionsItem :label="t('accounts.detail.token_expires_at')">{{ fmtTime(data.access_token_expires_at) }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.detail.refresh_token_valid')">
                <NTag :type="refreshValidType(data.refresh_token_valid)" size="small">{{ refreshValidLabel(data.refresh_token_valid) }}</NTag>
              </NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.detail.consec_failures')">{{ data.consec_failures }}</NDescriptionsItem>
            </NDescriptions>
            <div class="mt-3">
              <NSpace>
                <QuotaRing :label="t('accounts.detail.quota_5h')" :quota="data.quota_5h" />
                <QuotaRing :label="t('accounts.detail.quota_week')" :quota="data.quota_week" />
              </NSpace>
            </div>
          </NCard>
        </NGi>
      </NGrid>

      <NCard size="small" :title="t('accounts.detail.events')" class="mt-3">
        <EventTimeline :account-id="data.id" />
      </NCard>
    </div>
  </NSpin>
</template>
