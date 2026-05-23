<script setup lang="ts">
import { onMounted } from 'vue'
import { useWalletStore } from '@/stores/wallet'
import { useNoticeStore } from '@/stores/notice'
import AppHeader from '@/components/biz/AppHeader.vue'
import AppFooter from '@/components/biz/AppFooter.vue'
import NoticeBanner from '@/components/biz/NoticeBanner.vue'

const wallet = useWalletStore()
const notice = useNoticeStore()

onMounted(async () => {
  await Promise.allSettled([
    wallet.refresh(),
    notice.refreshUnreadCount(),
    notice.refreshTopBanner(),
  ])
})
</script>

<template>
  <div class="min-h-screen flex flex-col bg-bg text-fg">
    <AppHeader />
    <NoticeBanner
      v-if="notice.topBanner"
      :notice="notice.topBanner"
      @dismiss="notice.dismissBanner(notice.topBanner!.id)"
    />
    <main class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
      <slot />
    </main>
    <AppFooter />
  </div>
</template>
