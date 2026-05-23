<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { noticeApi, type Notice } from '@/api/notice'
import { useToast } from '@/composables/useToast'
import Skeleton from '@/components/ui/Skeleton.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import Pagination from '@/components/ui/Pagination.vue'

const toast = useToast()
const router = useRouter()
const items = ref<Notice[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(true)

async function load() {
  loading.value = true
  try {
    const r = await noticeApi.list(page.value, 20)
    items.value = r.items; total.value = r.total
  } catch { toast.error('加载公告失败') } finally { loading.value = false }
}

onMounted(load)

async function openNotice(notice: Notice) {
  if (!notice.read) {
    noticeApi.markRead(notice.id).catch(() => {})
  }
  router.push(`/notices/${notice.id}`)
}

function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime()
  const days = Math.floor(diff / 86400000)
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 30) return `${days}天前`
  return new Date(dateStr).toLocaleDateString('zh-CN')
}
</script>

<template>
  <div class="space-y-5">
    <div>
      <h1 class="text-2xl font-bold text-fg">系统公告</h1>
      <p class="text-sm text-fg-muted mt-1">查看最新的系统通知和公告</p>
    </div>

    <div v-if="loading" class="space-y-3">
      <Skeleton v-for="i in 5" :key="i" class="h-20" />
    </div>

    <EmptyState v-else-if="!items.length" icon="i-lucide-bell" title="暂无公告" subtitle="系统公告将显示在这里" />

    <div v-else class="space-y-3">
      <div
        v-for="notice in items"
        :key="notice.id"
        @click="openNotice(notice)"
        class="p-4 rounded-lg border border-border bg-bg-elevated hover:border-primary/30 cursor-pointer transition-colors"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="flex items-start gap-2 flex-1 min-w-0">
            <span v-if="notice.is_pinned" class="text-yellow-500 shrink-0 mt-0.5">⭐</span>
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <h3 class="font-medium text-fg truncate">{{ notice.title }}</h3>
                <span v-if="!notice.read" class="w-2 h-2 rounded-full bg-primary shrink-0" />
              </div>
              <p class="text-xs text-fg-muted mt-0.5 line-clamp-2">{{ notice.content }}</p>
            </div>
          </div>
          <span class="text-xs text-fg-muted whitespace-nowrap shrink-0">{{ timeAgo(notice.created_at) }}</span>
        </div>
      </div>
    </div>

    <Pagination v-if="total > 20" v-model="page" :total="total" :size="20" @update:model-value="load" />
  </div>
</template>
