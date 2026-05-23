<script setup lang="ts">
import { computed } from 'vue'
import { useWalletStore } from '@/stores/wallet'
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import Skeleton from '@/components/ui/Skeleton.vue'

interface Props {
  variant?: 'dashboard' | 'recharge-hero' | 'compact'
}
withDefaults(defineProps<Props>(), { variant: 'dashboard' })

const wallet = useWalletStore()
const app = useAppStore()
const { t } = useI18n()

const balanceDisplay = computed(() => {
  if (!wallet.info) return '--'
  return app.locale === 'zh'
    ? `¥ ${wallet.info.balance_cny.toFixed(2)}`
    : `$ ${wallet.info.balance_usd.toFixed(4)}`
})
const balanceSub = computed(() => {
  if (!wallet.info) return ''
  return app.locale === 'zh'
    ? `$ ${wallet.info.balance_usd.toFixed(4)}`
    : `¥ ${wallet.info.balance_cny.toFixed(2)}`
})
</script>

<template>
  <div class="rounded-xl border border-border bg-bg-elevated backdrop-blur-md p-5 shadow-sm">
    <template v-if="wallet.loading">
      <Skeleton class="h-8 w-32 mb-2" />
      <Skeleton class="h-4 w-24" />
    </template>
    <template v-else-if="wallet.error">
      <p class="text-fg-muted text-sm">{{ t('home.wallet.title') }}</p>
      <p class="text-2xl font-semibold text-fg mt-1">--</p>
    </template>
    <template v-else>
      <p class="text-xs text-fg-muted uppercase tracking-wide">{{ t('home.wallet.title') }}</p>
      <p class="text-3xl font-semibold mt-1 bg-gradient-to-r from-primary to-[#6366f1] bg-clip-text text-transparent">
        {{ balanceDisplay }}
      </p>
      <p class="text-xs text-fg-muted mt-0.5">≈ {{ balanceSub }}</p>
      <div v-if="variant !== 'compact'" class="flex gap-2 mt-4">
        <router-link to="/recharge"
          class="inline-flex items-center gap-1 px-3 h-8 rounded-md bg-primary text-white text-sm font-medium hover:bg-primary-hover transition-colors">
          <span class="i-lucide-plus w-3.5 h-3.5" />{{ t('home.wallet.cta.recharge') }}
        </router-link>
        <router-link to="/redeem"
          class="inline-flex items-center gap-1 px-3 h-8 rounded-md border border-border text-sm text-fg hover:bg-bg transition-colors">
          <span class="i-lucide-ticket w-3.5 h-3.5" />{{ t('home.wallet.cta.redeem') }}
        </router-link>
      </div>
    </template>
  </div>
</template>
