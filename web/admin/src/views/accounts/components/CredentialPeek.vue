<script setup lang="ts">
import { ref } from 'vue'
import { NModal, NButton, NInput, NCode, NSpace, NAlert, useMessage } from 'naive-ui'
import { accountApi } from '@/api/account'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ accountId: number; accountName: string }>()
const { t } = useI18n()
const message = useMessage()

const show = ref(false)
const confirmInput = ref('')
const loading = ref(false)
const credentials = ref<Record<string, unknown> | null>(null)

function open() {
  show.value = true
  confirmInput.value = ''
  credentials.value = null
}

async function doPeek() {
  if (confirmInput.value !== props.accountName) {
    message.warning(t('accounts.detail.peek_confirm'))
    return
  }
  loading.value = true
  try {
    const r = await accountApi.peek(props.accountId)
    credentials.value = r.credentials
  } catch (_) { /* handled */ } finally { loading.value = false }
}

function close() {
  show.value = false
  credentials.value = null
  confirmInput.value = ''
}

const credString = (c: Record<string, unknown> | null) => c ? JSON.stringify(c, null, 2) : ''
</script>

<template>
  <NButton size="small" type="warning" ghost @click="open">{{ t('accounts.detail.peek') }}</NButton>
  <NModal v-model:show="show" preset="card" :title="t('accounts.detail.peek')" style="width:560px" :on-after-leave="() => { credentials = null; confirmInput = '' }">
    <NSpace vertical>
      <NAlert type="warning" :show-icon="true">{{ t('accounts.detail.peek_warning') }}</NAlert>
      <div v-if="!credentials">
        <div class="text-sm text-gray-500 mb-2">{{ t('accounts.detail.peek_confirm') }}</div>
        <NInput v-model:value="confirmInput" :placeholder="t('accounts.detail.peek_confirm_placeholder')" />
      </div>
      <div v-else>
        <div class="text-sm text-gray-500 mb-2">{{ t('accounts.detail.peek_credentials') }}</div>
        <NCode :code="credString(credentials)" language="json" word-wrap />
      </div>
    </NSpace>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="close">{{ t('accounts.add_dialog.cancel') }}</NButton>
        <NButton v-if="!credentials" type="primary" :loading="loading" @click="doPeek">{{ t('accounts.detail.peek') }}</NButton>
      </NSpace>
    </template>
  </NModal>
</template>
