<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  NCard, NDescriptions, NDescriptionsItem, NSpace, NButton, NTag, NPageHeader, NSpin, NGrid, NGi,
} from 'naive-ui'
import { accountApi, type AccountDetail } from '@/api/account'
import QuotaRing from './components/QuotaRing.vue'
import EventTimeline from './components/EventTimeline.vue'
import CredentialPeek from './components/CredentialPeek.vue'
import { useAccountActions } from './composables/useAccountActions'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const id = computed(() => Number(route.params.id))
const data = ref<AccountDetail | null>(null)
const loading = ref(false)

async function load() {
  loading.value = true
  try { data.value = await accountApi.get(id.value) } catch (_) { /* handled */ } finally { loading.value = false }
}
onMounted(load)

const actions = useAccountActions(load)

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
              <NDescriptionsItem :label="t('accounts.columns.channel')">#{{ data.channel_id }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.share_tag')">{{ data.share_tag || '--' }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.provider')">{{ data.provider }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.tier')">{{ data.tier }}</NDescriptionsItem>
              <NDescriptionsItem :label="t('accounts.columns.cred_type')">{{ t(`accounts.cred_type.${data.cred_type}`) }}</NDescriptionsItem>
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
