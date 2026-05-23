<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import WalletCard from '@/components/biz/WalletCard.vue'
import UsageStatCard from '@/components/biz/UsageStatCard.vue'
import RecentLogsCard from '@/components/biz/RecentLogsCard.vue'
import QuickStartCode from '@/components/biz/QuickStartCode.vue'

const { t } = useI18n()
const auth = useAuthStore()
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-fg">
        {{ t('home.greeting', { name: auth.user?.display_name || auth.user?.email || '' }) }}
      </h1>
    </div>

    <!-- Stat cards row -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <WalletCard variant="dashboard" />
      <UsageStatCard scope="today" />
      <UsageStatCard scope="month" />
    </div>

    <!-- Bottom row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <RecentLogsCard :limit="5" />
      <QuickStartCode />
    </div>
  </div>
</template>
