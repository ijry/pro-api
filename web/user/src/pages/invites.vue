<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { inviteApi, type InviteSummary, type Invitee, type InviteRecord } from '@/api/invite'
import { useToast } from '@/composables/useToast'
import Card from '@/components/ui/Card.vue'
import ClipboardButton from '@/components/ui/ClipboardButton.vue'
import Pagination from '@/components/ui/Pagination.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import Skeleton from '@/components/ui/Skeleton.vue'

const { t, locale } = useI18n()
const toast = useToast()

const summary = ref<InviteSummary | null>(null)
const summaryLoading = ref(true)
const summaryError = ref(false)

const activeTab = ref<'invitees' | 'records'>('invitees')

const invitees = ref<Invitee[]>([])
const inviteesTotal = ref(0)
const inviteesPage = ref(1)
const inviteesLoading = ref(false)
const inviteesError = ref(false)

const records = ref<InviteRecord[]>([])
const recordsTotal = ref(0)
const recordsPage = ref(1)
const recordsLoading = ref(false)
const recordsError = ref(false)

const PAGE_SIZE = 10

const ratioPercent = computed(() =>
  summary.value ? Math.round(summary.value.rebate_ratio * 100) : 0
)

function formatRebate(credits: number) {
  return `¥${(credits / 100).toFixed(2)}`
}

function formatDate(iso: string) {
  const d = new Date(iso)
  return locale.value === 'zh'
    ? d.toLocaleString('zh-CN', { hour12: false })
    : d.toLocaleString('en-US')
}

async function loadSummary() {
  summaryLoading.value = true
  summaryError.value = false
  try {
    summary.value = await inviteApi.getSummary()
  } catch {
    summaryError.value = true
    toast.error(t('invites.load_failed'))
  } finally {
    summaryLoading.value = false
  }
}

async function loadInvitees() {
  inviteesLoading.value = true
  inviteesError.value = false
  try {
    const r = await inviteApi.listInvitees(inviteesPage.value, PAGE_SIZE)
    invitees.value = r.items
    inviteesTotal.value = r.total
  } catch {
    inviteesError.value = true
    toast.error(t('invites.load_failed'))
  } finally {
    inviteesLoading.value = false
  }
}

async function loadRecords() {
  recordsLoading.value = true
  recordsError.value = false
  try {
    const r = await inviteApi.listRecords(recordsPage.value, PAGE_SIZE)
    records.value = r.items
    recordsTotal.value = r.total
  } catch {
    recordsError.value = true
    toast.error(t('invites.load_failed'))
  } finally {
    recordsLoading.value = false
  }
}

onMounted(() => {
  loadSummary()
  loadInvitees()
})

watch(activeTab, (tab) => {
  if (tab === 'records' && records.value.length === 0 && !recordsLoading.value) {
    loadRecords()
  }
})

watch(inviteesPage, loadInvitees)
watch(recordsPage, loadRecords)
</script>

<template>
  <div class="space-y-5">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-fg">{{ t('invites.title') }}</h1>
      <p class="text-sm text-fg-muted mt-1">
        {{ t('invites.subtitle', { ratio: ratioPercent }) }}
      </p>
    </div>

    <!-- Code card + stats -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- Invite code / share link -->
      <Card>
        <div v-if="summaryLoading" class="space-y-3">
          <Skeleton class="h-6 w-32" />
          <Skeleton class="h-10" />
          <Skeleton class="h-10" />
        </div>
        <div v-else-if="summaryError" class="py-6 text-center">
          <p class="text-fg-muted text-sm mb-3">{{ t('invites.load_failed') }}</p>
          <button class="btn-primary" @click="loadSummary">{{ t('invites.retry') }}</button>
        </div>
        <div v-else-if="summary" class="space-y-4">
          <div>
            <label class="block text-xs text-fg-muted mb-1.5">{{ t('invites.my_code') }}</label>
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 rounded-md bg-bg-elevated border border-border font-mono text-fg text-lg tracking-wider">
                {{ summary.invite_code }}
              </code>
              <ClipboardButton :text="summary.invite_code" />
            </div>
          </div>
          <div>
            <label class="block text-xs text-fg-muted mb-1.5">{{ t('invites.share_url') }}</label>
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 rounded-md bg-bg-elevated border border-border font-mono text-fg text-xs truncate">
                {{ summary.share_url }}
              </code>
              <ClipboardButton :text="summary.share_url" />
            </div>
          </div>
        </div>
      </Card>

      <!-- Stats -->
      <Card>
        <div v-if="summaryLoading" class="grid grid-cols-3 gap-4">
          <Skeleton v-for="i in 3" :key="i" class="h-16" />
        </div>
        <div v-else-if="summary" class="grid grid-cols-3 gap-4">
          <div class="text-center">
            <div class="text-xs text-fg-muted mb-1">{{ t('invites.stats.invitee_count') }}</div>
            <div class="text-2xl font-bold text-fg">{{ summary.stats.invitee_count }}</div>
          </div>
          <div class="text-center border-l border-r border-border">
            <div class="text-xs text-fg-muted mb-1">{{ t('invites.stats.rebate_total') }}</div>
            <div class="text-2xl font-bold text-primary">{{ formatRebate(summary.stats.rebate_credits_total) }}</div>
          </div>
          <div class="text-center">
            <div class="text-xs text-fg-muted mb-1">{{ t('invites.stats.rebate_month') }}</div>
            <div class="text-2xl font-bold text-fg">{{ formatRebate(summary.stats.rebate_credits_month) }}</div>
          </div>
        </div>
      </Card>
    </div>

    <!-- Tabs -->
    <div role="tablist" class="flex items-center gap-1 border-b border-border">
      <button
        type="button"
        role="tab"
        :aria-selected="activeTab === 'invitees'"
        class="px-4 h-9 text-sm border-b-2 -mb-px transition-colors"
        :class="activeTab === 'invitees'
          ? 'border-primary text-primary font-medium'
          : 'border-transparent text-fg-muted hover:text-fg'"
        @click="activeTab = 'invitees'"
      >
        {{ t('invites.tab.invitees') }}
        <span v-if="summary" class="ml-1 text-fg-muted">({{ summary.stats.invitee_count }})</span>
      </button>
      <button
        type="button"
        role="tab"
        :aria-selected="activeTab === 'records'"
        class="px-4 h-9 text-sm border-b-2 -mb-px transition-colors"
        :class="activeTab === 'records'
          ? 'border-primary text-primary font-medium'
          : 'border-transparent text-fg-muted hover:text-fg'"
        @click="activeTab = 'records'"
      >
        {{ t('invites.tab.records') }}
        <span v-if="recordsTotal > 0" class="ml-1 text-fg-muted">({{ recordsTotal }})</span>
      </button>
    </div>

    <!-- Invitees tab -->
    <div v-if="activeTab === 'invitees'" role="tabpanel">
      <div v-if="inviteesLoading" class="space-y-2">
        <Skeleton v-for="i in 5" :key="i" class="h-10" />
      </div>
      <EmptyState
        v-else-if="inviteesError"
        icon="i-lucide-alert-circle"
        :title="t('invites.load_failed')"
        :cta="t('invites.retry')"
        @cta="loadInvitees"
      />
      <EmptyState
        v-else-if="!invitees.length"
        icon="i-lucide-user-plus"
        :title="t('invites.empty.invitees')"
      />
      <div v-else class="overflow-x-auto rounded-lg border border-border">
        <table class="w-full text-sm min-w-[600px]">
          <thead class="bg-bg-elevated border-b border-border">
            <tr>
              <th class="text-left px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.user') }}</th>
              <th class="text-left px-3 py-2 text-fg-muted font-medium hidden sm:table-cell">{{ t('invites.table.email') }}</th>
              <th class="text-left px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.registered_at') }}</th>
              <th class="text-right px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.total_rebate') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in invitees" :key="u.user_id" class="border-b border-border last:border-0 hover:bg-bg-elevated/50 transition-colors">
              <td class="px-3 py-2 text-fg">{{ u.display_name }}</td>
              <td class="px-3 py-2 text-fg-muted text-xs font-mono hidden sm:table-cell">{{ u.email_masked }}</td>
              <td class="px-3 py-2 text-fg-muted text-xs">{{ formatDate(u.registered_at) }}</td>
              <td class="px-3 py-2 text-right font-mono text-primary">{{ formatRebate(u.total_rebate_credits) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <Pagination v-if="inviteesTotal > PAGE_SIZE" v-model="inviteesPage" :total="inviteesTotal" :size="PAGE_SIZE" />
    </div>

    <!-- Records tab -->
    <div v-else role="tabpanel">
      <div v-if="recordsLoading" class="space-y-2">
        <Skeleton v-for="i in 5" :key="i" class="h-10" />
      </div>
      <EmptyState
        v-else-if="recordsError"
        icon="i-lucide-alert-circle"
        :title="t('invites.load_failed')"
        :cta="t('invites.retry')"
        @cta="loadRecords"
      />
      <EmptyState
        v-else-if="!records.length"
        icon="i-lucide-coins"
        :title="t('invites.empty.records')"
      />
      <div v-else class="overflow-x-auto rounded-lg border border-border">
        <table class="w-full text-sm min-w-[600px]">
          <thead class="bg-bg-elevated border-b border-border">
            <tr>
              <th class="text-left px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.user') }}</th>
              <th class="text-left px-3 py-2 text-fg-muted font-medium hidden sm:table-cell">{{ t('invites.table.order_id') }}</th>
              <th class="text-right px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.rebate') }}</th>
              <th class="text-left px-3 py-2 text-fg-muted font-medium">{{ t('invites.table.created_at') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in records" :key="r.id" class="border-b border-border last:border-0 hover:bg-bg-elevated/50 transition-colors">
              <td class="px-3 py-2 text-fg">{{ r.invitee_display_name }}</td>
              <td class="px-3 py-2 text-fg-muted text-xs font-mono hidden sm:table-cell">#{{ r.order_id }}</td>
              <td class="px-3 py-2 text-right font-mono text-primary">{{ formatRebate(r.rebate_credits) }}</td>
              <td class="px-3 py-2 text-fg-muted text-xs">{{ formatDate(r.created_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <Pagination v-if="recordsTotal > PAGE_SIZE" v-model="recordsPage" :total="recordsTotal" :size="PAGE_SIZE" />
    </div>

    <!-- Rules -->
    <Card>
      <h2 class="text-base font-semibold text-fg mb-3 flex items-center gap-2">
        <span class="i-lucide-scroll-text w-4 h-4 text-primary" />
        {{ t('invites.rules.title') }}
      </h2>
      <ul class="space-y-2 text-sm text-fg-muted">
        <li class="flex gap-2"><span class="text-primary">•</span>{{ t('invites.rules.item1', { ratio: ratioPercent }) }}</li>
        <li class="flex gap-2"><span class="text-primary">•</span>{{ t('invites.rules.item2') }}</li>
        <li class="flex gap-2"><span class="text-primary">•</span>{{ t('invites.rules.item3') }}</li>
      </ul>
    </Card>
  </div>
</template>
