<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { type TokenView } from '@/api/token'
import Badge from '@/components/ui/Badge.vue'
import ClipboardButton from '@/components/ui/ClipboardButton.vue'

interface Props { token: TokenView }
const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'edit', token: TokenView): void
  (e: 'regenerate', token: TokenView): void
  (e: 'revoke', token: TokenView): void
}>()
const { t } = useI18n()

const usagePct = computed(() => {
  if (!props.token.quota_limit) return 0
  return Math.min(100, Math.round(props.token.quota_used / props.token.quota_limit * 100))
})

const barColor = computed(() => {
  if (usagePct.value >= 100) return 'bg-rose-500'
  if (usagePct.value >= 80) return 'bg-amber-500'
  return 'bg-primary'
})

function formatWhen(dateStr: string | null) {
  if (!dateStr) return t('tokens.last_used.never')
  const diff = Date.now() - new Date(dateStr).getTime()
  const min = Math.floor(diff / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr} 小时前`
  return `${Math.floor(hr / 24)} 天前`
}
</script>

<template>
  <div class="rounded-xl border border-border bg-bg-elevated p-5 space-y-3 hover:border-primary/30 transition-colors">
    <!-- Header row -->
    <div class="flex items-start justify-between gap-2">
      <div class="flex items-center gap-2">
        <span :class="[
          'w-2 h-2 rounded-full shrink-0',
          token.status === 'enabled' ? 'bg-emerald-400' : 'bg-zinc-500',
        ]" />
        <h4 class="font-medium text-fg">{{ token.name }}</h4>
      </div>
      <Badge :variant="token.status === 'enabled' ? 'success' : 'muted'" size="sm">
        {{ token.status === 'enabled' ? t('tokens.status.enabled') : t('tokens.status.disabled') }}
      </Badge>
    </div>

    <!-- Key prefix -->
    <div class="flex items-center gap-2 font-mono text-sm text-fg-muted bg-bg rounded-md px-3 py-2">
      <span class="flex-1 truncate">{{ token.prefix }}****</span>
      <ClipboardButton :text="token.prefix" :success-msg="t('tokens.toast.copied')" size="sm" />
    </div>

    <!-- Usage bar -->
    <div>
      <div class="flex justify-between text-xs text-fg-muted mb-1">
        <span v-if="token.quota_limit">
          {{ t('tokens.usage_label', { used: token.quota_used.toLocaleString(), limit: token.quota_limit.toLocaleString() }) }}
        </span>
        <span v-else>{{ t('tokens.usage_unlimited', { used: token.quota_used.toLocaleString() }) }}</span>
        <span v-if="token.quota_limit">{{ usagePct }}%</span>
      </div>
      <div v-if="token.quota_limit" class="h-1.5 rounded-full bg-bg overflow-hidden">
        <div :class="[barColor, 'h-full rounded-full transition-all']" :style="{ width: `${usagePct}%` }" />
      </div>
    </div>

    <!-- Meta info -->
    <div class="text-xs text-fg-muted space-y-0.5">
      <div v-if="token.allowed_models.length" class="truncate">
        模型: {{ token.allowed_models.join(', ') || '全部' }}
      </div>
      <div>RPM: {{ token.rpm_limit || '不限' }} · TPM: {{ token.tpm_limit || '不限' }}</div>
      <div class="flex gap-3">
        <span>{{ t('tokens.last_used') }}: {{ formatWhen(token.last_used_at) }}</span>
        <span>{{ t('tokens.created_at') }}: {{ new Date(token.created_at).toLocaleDateString() }}</span>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex gap-2 pt-1 border-t border-border/50">
      <button @click="emit('edit', token)"
        class="inline-flex items-center gap-1 px-2 h-7 rounded text-xs text-fg-muted hover:text-fg hover:bg-bg transition-colors">
        <span class="i-lucide-pencil w-3.5 h-3.5" />{{ t('tokens.action.edit') }}
      </button>
      <button @click="emit('regenerate', token)"
        class="inline-flex items-center gap-1 px-2 h-7 rounded text-xs text-fg-muted hover:text-fg hover:bg-bg transition-colors">
        <span class="i-lucide-refresh-cw w-3.5 h-3.5" />{{ t('tokens.action.regenerate') }}
      </button>
      <button @click="emit('revoke', token)"
        class="inline-flex items-center gap-1 px-2 h-7 rounded text-xs text-rose-400 hover:text-rose-300 hover:bg-rose-500/10 transition-colors">
        <span class="i-lucide-trash-2 w-3.5 h-3.5" />{{ t('tokens.action.revoke') }}
      </button>
    </div>
  </div>
</template>
