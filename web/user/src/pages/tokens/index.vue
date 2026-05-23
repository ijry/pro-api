<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { tokenApi, type TokenView, type CreateTokenParams } from '@/api/token'
import { useToast } from '@/composables/useToast'
import { useForm } from '@/composables/useForm'
import TokenCard from '@/components/biz/TokenCard.vue'
import Skeleton from '@/components/ui/Skeleton.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import Drawer from '@/components/ui/Drawer.vue'
import Dialog from '@/components/ui/Dialog.vue'
import Input from '@/components/ui/Input.vue'
import Pagination from '@/components/ui/Pagination.vue'
import ClipboardButton from '@/components/ui/ClipboardButton.vue'
import Button from '@/components/ui/Button.vue'

const { t } = useI18n()
const toast = useToast()

const items = ref<TokenView[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(true)
const drawerOpen = ref(false)
const editingToken = ref<TokenView | null>(null)
const plaintextOpen = ref(false)
const plaintextValue = ref('')
const plaintextAck = ref(false)
const confirmOpen = ref(false)
const confirmTitle = ref('')
const confirmSubtitle = ref('')
const confirmAction = ref<() => Promise<void>>(async () => {})
const submitting = ref(false)
const advancedOpen = ref(false)

const form = useForm({
  initial: {
    name: '',
    quota_limit: null as number | null,
    quota_unlimited: true,
    allowed_models: [] as string[],
    allowed_ips: [] as string[],
    rpm_limit: 0,
    tpm_limit: 0,
    expires_at: null as string | null,
    never_expire: true,
    model_input: '',
    ip_input: '',
  },
  rules: {
    name: (v) => !v ? t('errors.required', { field: t('tokens.form.name') }) : v.length > 64 ? '名称过长' : '',
  },
})

async function loadTokens() {
  loading.value = true
  try {
    const r = await tokenApi.list(page.value, 20)
    items.value = r.items
    total.value = r.total
  } catch {
    toast.error('加载令牌失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadTokens)

function openCreate() {
  editingToken.value = null
  form.reset()
  advancedOpen.value = false
  drawerOpen.value = true
}

function openEdit(token: TokenView) {
  editingToken.value = token
  form.setValues({
    name: token.name,
    quota_limit: token.quota_limit,
    quota_unlimited: token.quota_limit === null,
    allowed_models: [...token.allowed_models],
    allowed_ips: [...token.allowed_ips],
    rpm_limit: token.rpm_limit,
    tpm_limit: token.tpm_limit,
    expires_at: token.expires_at,
    never_expire: token.expires_at === null,
    model_input: '',
    ip_input: '',
  })
  advancedOpen.value = false
  drawerOpen.value = true
}

function addChip(field: 'allowed_models' | 'allowed_ips', inputField: 'model_input' | 'ip_input') {
  const val = (form.values[inputField] as string).trim()
  if (!val) return
  ;(form.values[field] as string[]).push(val)
  ;(form.values as Record<string, unknown>)[inputField] = ''
}

function removeChip(field: 'allowed_models' | 'allowed_ips', idx: number) {
  ;(form.values[field] as string[]).splice(idx, 1)
}

async function onSubmit() {
  if (!form.validate()) return
  submitting.value = true
  try {
    const params: CreateTokenParams = {
      name: form.values.name,
      quota_limit: form.values.quota_unlimited ? null : form.values.quota_limit,
      allowed_models: form.values.allowed_models,
      allowed_ips: form.values.allowed_ips,
      rpm_limit: form.values.rpm_limit,
      tpm_limit: form.values.tpm_limit,
      expires_at: form.values.never_expire ? null : form.values.expires_at,
    }
    if (editingToken.value) {
      await tokenApi.update(editingToken.value.id, params)
      toast.success(t('tokens.toast.updated'))
      drawerOpen.value = false
      await loadTokens()
    } else {
      const res = await tokenApi.create(params)
      drawerOpen.value = false
      plaintextValue.value = res.plaintext_key
      plaintextAck.value = false
      plaintextOpen.value = true
      await loadTokens()
    }
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { message?: string } } })?.response?.data?.message || '操作失败'
    toast.error(msg)
  } finally {
    submitting.value = false
  }
}

function showConfirm(title: string, subtitle: string, action: () => Promise<void>) {
  confirmTitle.value = title
  confirmSubtitle.value = subtitle
  confirmAction.value = action
  confirmOpen.value = true
}

async function onRevoke(token: TokenView) {
  showConfirm(
    t('tokens.revoke.confirm.title'),
    t('tokens.revoke.confirm.subtitle'),
    async () => {
      await tokenApi.revoke(token.id)
      toast.success(t('tokens.toast.revoked'))
      await loadTokens()
    },
  )
}

async function onRegen(token: TokenView) {
  showConfirm(
    t('tokens.regenerate.confirm.title'),
    t('tokens.regenerate.confirm.subtitle'),
    async () => {
      const res = await tokenApi.regenerate(token.id)
      plaintextValue.value = res.plaintext_key
      plaintextAck.value = false
      plaintextOpen.value = true
      toast.success(t('tokens.toast.regenerated'))
      await loadTokens()
    },
  )
}

async function runConfirm() {
  submitting.value = true
  try {
    await confirmAction.value()
    confirmOpen.value = false
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { message?: string } } })?.response?.data?.message || '操作失败'
    toast.error(msg)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-start justify-between">
      <div>
        <h1 class="text-2xl font-bold text-fg">{{ t('tokens.title') }}</h1>
        <p class="text-sm text-fg-muted mt-1">{{ t('tokens.subtitle') }}</p>
        <p class="text-xs text-fg-muted mt-0.5">
          {{ t('tokens.base_url') }}: <span class="font-mono text-primary">https://api.proapi.io/v1</span>
        </p>
      </div>
      <Button @click="openCreate">
        <span class="i-lucide-plus w-4 h-4 mr-1" />{{ t('tokens.create') }}
      </Button>
    </div>

    <!-- List -->
    <div v-if="loading" class="space-y-3">
      <Skeleton v-for="i in 3" :key="i" class="h-44" />
    </div>
    <EmptyState
      v-else-if="!items.length"
      icon="i-lucide-key-round"
      :title="t('tokens.empty.title')"
      :subtitle="t('tokens.empty.subtitle')"
      :cta="t('tokens.empty.cta')"
      @cta="openCreate"
    />
    <div v-else class="space-y-3">
      <TokenCard
        v-for="token in items"
        :key="token.id"
        :token="token"
        @edit="openEdit"
        @regenerate="onRegen"
        @revoke="onRevoke"
      />
    </div>

    <Pagination v-if="total > 20" v-model="page" :total="total" :size="20" />
  </div>

  <!-- Create / Edit Drawer -->
  <Drawer v-model:open="drawerOpen" :title="editingToken ? t('tokens.action.edit') : t('tokens.create')">
    <form @submit.prevent="onSubmit" class="space-y-4">
      <!-- Name -->
      <div>
        <label class="block text-sm font-medium text-fg mb-1">{{ t('tokens.form.name') }} <span class="text-rose-400">*</span></label>
        <Input v-model="form.values.name" :error="form.errors.name" @blur="form.touch('name')" />
      </div>

      <!-- Advanced toggle -->
      <button type="button" @click="advancedOpen = !advancedOpen"
        class="flex items-center gap-1 text-sm text-fg-muted hover:text-fg transition-colors">
        <span :class="advancedOpen ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'" class="w-4 h-4" />
        {{ t('tokens.form.advanced') }}
      </button>

      <template v-if="advancedOpen">
        <!-- Quota limit -->
        <div>
          <label class="block text-sm font-medium text-fg mb-1">{{ t('tokens.form.quota_limit') }}</label>
          <div class="flex items-center gap-2">
            <Input v-model="(form.values.quota_limit as string | number | undefined)" type="number" :disabled="form.values.quota_unlimited" class="flex-1" />
            <label class="flex items-center gap-1 text-sm text-fg-muted cursor-pointer">
              <input type="checkbox" v-model="form.values.quota_unlimited" class="accent-primary" />
              {{ t('tokens.form.unlimited') }}
            </label>
          </div>
        </div>

        <!-- Allowed models chip input -->
        <div>
          <label class="block text-sm font-medium text-fg mb-1">{{ t('tokens.form.allowed_models') }}</label>
          <div class="flex flex-wrap gap-1 mb-1">
            <span v-for="(m, i) in form.values.allowed_models" :key="i"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-primary/10 text-primary">
              {{ m }}
              <button type="button" @click="removeChip('allowed_models', i)">
                <span class="i-lucide-x w-3 h-3" />
              </button>
            </span>
          </div>
          <div class="flex gap-2">
            <Input v-model="form.values.model_input" placeholder="gpt-4*" size="sm" @keydown.enter.prevent="addChip('allowed_models', 'model_input')" />
            <button type="button" @click="addChip('allowed_models', 'model_input')"
              class="px-3 h-8 rounded-md border border-border text-sm text-fg hover:bg-bg-elevated">+</button>
          </div>
        </div>

        <!-- Allowed IPs chip input -->
        <div>
          <label class="block text-sm font-medium text-fg mb-1">{{ t('tokens.form.allowed_ips') }}</label>
          <div class="flex flex-wrap gap-1 mb-1">
            <span v-for="(ip, i) in form.values.allowed_ips" :key="i"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-border/40 text-fg-muted">
              {{ ip }}
              <button type="button" @click="removeChip('allowed_ips', i)">
                <span class="i-lucide-x w-3 h-3" />
              </button>
            </span>
          </div>
          <div class="flex gap-2">
            <Input v-model="form.values.ip_input" placeholder="10.0.0.0/8" size="sm" @keydown.enter.prevent="addChip('allowed_ips', 'ip_input')" />
            <button type="button" @click="addChip('allowed_ips', 'ip_input')"
              class="px-3 h-8 rounded-md border border-border text-sm text-fg hover:bg-bg-elevated">+</button>
          </div>
        </div>

        <!-- RPM / TPM -->
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-fg mb-1">{{ t('tokens.form.rpm_limit') }}</label>
            <Input v-model="form.values.rpm_limit" type="number" size="sm" />
          </div>
          <div>
            <label class="block text-sm font-medium text-fg mb-1">{{ t('tokens.form.tpm_limit') }}</label>
            <Input v-model="form.values.tpm_limit" type="number" size="sm" />
          </div>
        </div>

        <!-- Expires at -->
        <div>
          <label class="block text-sm font-medium text-fg mb-1">{{ t('tokens.form.expires_at') }}</label>
          <div class="flex items-center gap-2">
            <Input v-model="(form.values.expires_at as string | undefined)" type="text" placeholder="2027-12-31T23:59:59Z" :disabled="form.values.never_expire" class="flex-1" />
            <label class="flex items-center gap-1 text-sm text-fg-muted cursor-pointer">
              <input type="checkbox" v-model="form.values.never_expire" class="accent-primary" />
              {{ t('tokens.form.never_expire') }}
            </label>
          </div>
        </div>
      </template>

      <!-- Actions -->
      <div class="flex gap-3 pt-2">
        <button type="button" @click="drawerOpen = false"
          class="flex-1 h-10 rounded-md border border-border text-sm text-fg hover:bg-bg-elevated transition-colors">
          {{ t('tokens.form.cancel') }}
        </button>
        <button type="submit" :disabled="submitting"
          class="flex-1 h-10 rounded-md bg-primary text-white text-sm font-medium hover:bg-primary-hover disabled:opacity-60 transition-colors">
          {{ editingToken ? t('tokens.form.submit_edit') : t('tokens.form.submit') }}
        </button>
      </div>
    </form>
  </Drawer>

  <!-- Plaintext reveal dialog -->
  <Dialog v-model:open="plaintextOpen" :title="t('tokens.plaintext.title')">
    <div class="space-y-4">
      <p class="text-sm text-fg-muted">{{ t('tokens.plaintext.hint') }}</p>
      <div class="flex items-center gap-2 bg-bg rounded-md px-3 py-2 border border-border">
        <span class="flex-1 font-mono text-sm text-primary break-all">{{ plaintextValue }}</span>
        <ClipboardButton :text="plaintextValue" />
      </div>
      <label class="flex items-center gap-2 text-sm cursor-pointer">
        <input type="checkbox" v-model="plaintextAck" class="accent-primary" />
        <span class="text-fg">{{ t('tokens.plaintext.ack') }}</span>
      </label>
      <button
        :disabled="!plaintextAck"
        @click="plaintextOpen = false"
        class="w-full h-10 rounded-md bg-primary text-white text-sm font-medium hover:bg-primary-hover disabled:opacity-40 transition-colors">
        {{ t('tokens.plaintext.confirm') }}
      </button>
    </div>
  </Dialog>

  <!-- Confirm dialog -->
  <Dialog v-model:open="confirmOpen" :title="confirmTitle" size="sm">
    <div class="space-y-4">
      <p class="text-sm text-fg-muted">{{ confirmSubtitle }}</p>
      <div class="flex gap-3">
        <button @click="confirmOpen = false"
          class="flex-1 h-9 rounded-md border border-border text-sm text-fg hover:bg-bg-elevated transition-colors">
          {{ t('common.cancel') }}
        </button>
        <button @click="runConfirm" :disabled="submitting"
          class="flex-1 h-9 rounded-md bg-rose-500 text-white text-sm font-medium hover:bg-rose-600 disabled:opacity-60 transition-colors">
          确认
        </button>
      </div>
    </div>
  </Dialog>
</template>
