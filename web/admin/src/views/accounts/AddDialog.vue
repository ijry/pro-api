<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import {
  NModal, NTabs, NTabPane, NForm, NFormItem, NInput, NSelect,
  NButton, NSpace, NAlert, NDataTable, NTag, useMessage,
  type DataTableColumns, type SelectOption,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { accountApi, type Account } from '@/api/account'
import { channelApi, type Channel } from '@/api/channel'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{
  'update:show': [val: boolean]
  imported: []
}>()

const { t } = useI18n()
const message = useMessage()

const tab = ref<'oauth' | 'token_json' | 'apikey' | 'import'>('token_json')

const channelOptions = ref<SelectOption[]>([])
async function loadChannels() {
  try {
    const r = await channelApi.list({ size: 200 })
    channelOptions.value = r.items.map((c: Channel) => ({ label: `${c.name} (#${c.id})`, value: c.id }))
  } catch (_) { /* handled */ }
}
onMounted(loadChannels)

const common = ref({
  channel_id: null as number | null,
  name: '',
})

const tokenForm = ref({ text: '', format: 'auto' })
const apikeyForm = ref({ text: '' })
const importForm = ref({ text: '', format: 'auto' })
const oauthProvider = ref<'anthropic' | 'openai'>('anthropic')

const formatOptions: SelectOption[] = [
  { label: t('accounts.add_dialog.format_auto'), value: 'auto' },
  { label: 'access_token', value: 'access_token' },
  { label: 'refresh_token', value: 'refresh_token' },
  { label: 'token_pair', value: 'token_pair' },
  { label: 'JSON', value: 'json' },
]

const previewData = ref<Account[]>([])
const previewErrors = ref<{ index: number; reason: string }[]>([])
const submitting = ref(false)
const previewing = ref(false)
const oauthStarting = ref(false)
let oauthMessageHandler: ((event: MessageEvent) => void) | null = null

watch(() => props.show, (v) => {
  if (v) {
    tab.value = 'token_json'
    common.value = { channel_id: null, name: '' }
    tokenForm.value = { text: '', format: 'auto' }
    apikeyForm.value = { text: '' }
    importForm.value = { text: '', format: 'auto' }
    oauthProvider.value = 'anthropic'
    previewData.value = []
    previewErrors.value = []
  }
})

function clearOAuthMessageHandler() {
  if (!oauthMessageHandler) return
  window.removeEventListener('message', oauthMessageHandler)
  oauthMessageHandler = null
}

onBeforeUnmount(clearOAuthMessageHandler)

const providerOptions: SelectOption[] = [
  { label: 'Anthropic', value: 'anthropic' },
  { label: 'OpenAI', value: 'openai' },
]

const canSubmit = computed(() => {
  if (!common.value.channel_id) return false
  if (tab.value === 'token_json') return tokenForm.value.text.trim().length > 0
  if (tab.value === 'apikey') return apikeyForm.value.text.trim().length > 0
  if (tab.value === 'import') return importForm.value.text.trim().length > 0
  return false
})

const canOAuthStart = computed(() => Boolean(common.value.channel_id && oauthProvider.value))

async function doPreview() {
  if (!common.value.channel_id) {
    message.warning(t('accounts.add_dialog.channel_placeholder'))
    return
  }
  if (tab.value !== 'import') return
  previewing.value = true
  previewData.value = []
  previewErrors.value = []
  try {
    const r = await accountApi.import({
      channel_id: common.value.channel_id,
      text: importForm.value.text,
      format: importForm.value.format || undefined,
      dry_run: true,
    })
    previewData.value = r.preview ?? []
    previewErrors.value = r.errors ?? []
    message.info(t('accounts.add_dialog.preview_total', { n: r.total }))
    if (previewErrors.value.length) {
      message.warning(t('accounts.add_dialog.preview_errors', { n: previewErrors.value.length }))
    }
  } catch (_) { /* handled */ } finally { previewing.value = false }
}

async function doSubmit() {
  if (!canSubmit.value || !common.value.channel_id) return
  submitting.value = true
  try {
    if (tab.value === 'token_json') {
      await accountApi.create({
        channel_id: common.value.channel_id,
        name: common.value.name || undefined,
        text: tokenForm.value.text,
        format: tokenForm.value.format,
      })
      message.success(t('accounts.add_dialog.imported', { n: 1 }))
    } else if (tab.value === 'apikey') {
      await accountApi.create({
        channel_id: common.value.channel_id,
        name: common.value.name || undefined,
        text: apikeyForm.value.text,
        format: 'apikey',
      })
      message.success(t('accounts.add_dialog.imported', { n: 1 }))
    } else if (tab.value === 'import') {
      const r = await accountApi.import({
        channel_id: common.value.channel_id,
        text: importForm.value.text,
        format: importForm.value.format || undefined,
      })
      message.success(t('accounts.add_dialog.imported', { n: r.imported ?? 0 }))
    }
    emit('imported')
    emit('update:show', false)
  } catch (_) { /* handled */ } finally { submitting.value = false }
}

async function doOAuthStart() {
  if (!common.value.channel_id) {
    message.warning(t('accounts.add_dialog.channel_placeholder'))
    return
  }
  oauthStarting.value = true
  try {
    const r = await accountApi.oauthStart({
      channel_id: common.value.channel_id,
      provider: oauthProvider.value,
    })
    clearOAuthMessageHandler()
    oauthMessageHandler = (event: MessageEvent) => {
      if (event.origin !== window.location.origin) return
      if ((event.data as { type?: string })?.type !== 'account_oauth_done') return
      clearOAuthMessageHandler()
      message.success(t('accounts.add_dialog.oauth_done'))
      emit('imported')
      emit('update:show', false)
    }
    window.addEventListener('message', oauthMessageHandler)
    const popup = window.open(r.auth_url, 'proapi-account-oauth', 'width=560,height=720')
    if (!popup) {
      clearOAuthMessageHandler()
      message.warning(t('accounts.add_dialog.oauth_popup_blocked'))
      return
    }
    popup.focus()
    message.info(t('accounts.add_dialog.oauth_started'))
  } catch (_) { /* handled */ } finally { oauthStarting.value = false }
}

const previewColumns: DataTableColumns<Account> = [
  { title: '#', key: 'index', width: 50, render: (_row, idx) => idx + 1 },
  { title: t('accounts.columns.name'), key: 'name', width: 180 },
  { title: t('accounts.columns.cred_type'), key: 'cred_type', width: 110, render: (row) => row.cred_type },
  { title: t('accounts.columns.provider'), key: 'provider', width: 100 },
  { title: t('accounts.columns.tier'), key: 'tier', width: 100 },
  { title: 'email', key: 'email' },
]
</script>

<template>
  <NModal :show="props.show" preset="card" :title="t('accounts.add_dialog.title')" style="width:760px"
    @update:show="emit('update:show', $event)">
    <NTabs v-model:value="tab" type="line" animated>
      <NTabPane name="oauth" :tab="t('accounts.add_dialog.tabs.oauth')">
        <NAlert type="info" class="mb-3">{{ t('accounts.add_dialog.oauth_hint') }}</NAlert>
        <NForm label-placement="top" size="small">
          <NFormItem :label="t('accounts.add_dialog.channel_label')" required>
            <NSelect v-model:value="common.channel_id" :options="channelOptions" :placeholder="t('accounts.add_dialog.channel_placeholder')" clearable />
          </NFormItem>
          <NFormItem :label="t('accounts.add_dialog.provider_label')" required>
            <NSelect v-model:value="oauthProvider" :options="providerOptions" />
          </NFormItem>
        </NForm>
      </NTabPane>

      <NTabPane name="token_json" :tab="t('accounts.add_dialog.tabs.token_json')">
        <NForm label-placement="top" size="small">
          <NFormItem :label="t('accounts.add_dialog.channel_label')" required>
            <NSelect v-model:value="common.channel_id" :options="channelOptions" :placeholder="t('accounts.add_dialog.channel_placeholder')" clearable />
          </NFormItem>
          <NFormItem :label="t('accounts.add_dialog.name_label')">
            <NInput v-model:value="common.name" :placeholder="t('accounts.add_dialog.name_placeholder')" />
          </NFormItem>
          <NFormItem :label="t('accounts.add_dialog.format_label')">
            <NSelect v-model:value="tokenForm.format" :options="formatOptions" />
          </NFormItem>
          <NFormItem :label="t('accounts.add_dialog.token_text_label')" required>
            <NInput v-model:value="tokenForm.text" type="textarea" :rows="6" :placeholder="t('accounts.add_dialog.token_text_placeholder')" />
          </NFormItem>
        </NForm>
      </NTabPane>

      <NTabPane name="apikey" :tab="t('accounts.add_dialog.tabs.apikey')">
        <NForm label-placement="top" size="small">
          <NFormItem :label="t('accounts.add_dialog.channel_label')" required>
            <NSelect v-model:value="common.channel_id" :options="channelOptions" :placeholder="t('accounts.add_dialog.channel_placeholder')" clearable />
          </NFormItem>
          <NFormItem :label="t('accounts.add_dialog.name_label')">
            <NInput v-model:value="common.name" :placeholder="t('accounts.add_dialog.name_placeholder')" />
          </NFormItem>
          <NFormItem :label="t('accounts.add_dialog.apikey_label')" required>
            <NInput v-model:value="apikeyForm.text" type="password" show-password-on="click" :placeholder="t('accounts.add_dialog.apikey_placeholder')" />
          </NFormItem>
        </NForm>
      </NTabPane>

      <NTabPane name="import" :tab="t('accounts.add_dialog.tabs.import')">
        <NForm label-placement="top" size="small">
          <NFormItem :label="t('accounts.add_dialog.channel_label')" required>
            <NSelect v-model:value="common.channel_id" :options="channelOptions" :placeholder="t('accounts.add_dialog.channel_placeholder')" clearable />
          </NFormItem>
          <NFormItem :label="t('accounts.add_dialog.format_label')">
            <NSelect v-model:value="importForm.format" :options="formatOptions" />
          </NFormItem>
          <NFormItem :label="t('accounts.add_dialog.import_text_label')" required>
            <NInput v-model:value="importForm.text" type="textarea" :rows="8" :placeholder="t('accounts.add_dialog.import_text_placeholder')" />
          </NFormItem>
        </NForm>
        <div v-if="previewData.length" class="mt-2">
          <NDataTable :columns="previewColumns" :data="previewData" size="small" :max-height="240" />
        </div>
        <div v-if="previewErrors.length" class="mt-2">
          <NTag type="error" size="small" class="mr-2">
            {{ t('accounts.add_dialog.preview_errors', { n: previewErrors.length }) }}
          </NTag>
          <div v-for="e in previewErrors" :key="e.index" class="text-xs text-red-500">#{{ e.index }}: {{ e.reason }}</div>
        </div>
      </NTabPane>
    </NTabs>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="emit('update:show', false)">{{ t('accounts.add_dialog.cancel') }}</NButton>
        <NButton v-if="tab === 'import'" :loading="previewing" :disabled="!canSubmit" @click="doPreview">
          {{ t('accounts.add_dialog.preview') }}
        </NButton>
        <NButton v-if="tab === 'oauth'" type="primary" :loading="oauthStarting" :disabled="!canOAuthStart" @click="doOAuthStart">
          {{ t('accounts.add_dialog.oauth_start') }}
        </NButton>
        <NButton v-else type="primary" :loading="submitting" :disabled="!canSubmit" @click="doSubmit">
          {{ t('accounts.add_dialog.submit') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>
