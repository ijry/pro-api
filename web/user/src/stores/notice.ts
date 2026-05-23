import { defineStore } from 'pinia'
import { ref } from 'vue'
import { noticeApi, type Notice } from '@/api/notice'

export const useNoticeStore = defineStore('notice', () => {
  const unreadCount = ref(0)
  const topBanner = ref<Notice | null>(null)
  const dismissedIds = ref<Set<string>>(new Set())

  async function refreshUnreadCount() {
    try {
      const r = await noticeApi.unreadCount()
      unreadCount.value = r.count
    } catch {
      // ignore
    }
  }

  async function refreshTopBanner() {
    try {
      const r = await noticeApi.publicList(1, 1)
      const pinned = r.items.find(n => n.is_pinned && !dismissedIds.value.has(n.id))
      topBanner.value = pinned ?? null
    } catch {
      // ignore
    }
  }

  function dismissBanner(id: string) {
    dismissedIds.value.add(id)
    topBanner.value = null
  }

  return { unreadCount, topBanner, refreshUnreadCount, refreshTopBanner, dismissBanner }
})
