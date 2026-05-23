<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { noticeApi, type Notice } from '@/api/notice'
import { useToast } from '@/composables/useToast'
import Skeleton from '@/components/ui/Skeleton.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const notice = ref<Notice | null>(null)
const loading = ref(true)

async function load() {
  loading.value = true
  try {
    const id = route.params.id as string
    notice.value = await noticeApi.get(id)
    noticeApi.markRead(id).catch(() => {})
  } catch { toast.error('加载公告失败'); router.push('/notices') } finally { loading.value = false }
}

onMounted(load)

function formatDate(d: string) { return new Date(d).toLocaleString('zh-CN') }

// Simple markdown to HTML for basic formatting (paragraphs + line breaks)
function renderContent(content: string): string {
  return content
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/`(.+?)`/g, '<code class="bg-bg-elevated px-1 rounded text-primary text-sm">$1</code>')
    .replace(/\n\n/g, '</p><p class="mb-3">')
    .replace(/\n/g, '<br/>')
}
</script>

<template>
  <div class="max-w-2xl space-y-5">
    <button @click="router.push('/notices')" class="flex items-center gap-1 text-sm text-fg-muted hover:text-fg transition-colors">
      <span class="i-lucide-arrow-left w-4 h-4" />
      返回公告列表
    </button>

    <div v-if="loading" class="space-y-3">
      <Skeleton class="h-8 w-2/3" />
      <Skeleton class="h-4 w-1/3" />
      <Skeleton class="h-32" />
    </div>

    <div v-else-if="notice">
      <div class="mb-2 flex items-center gap-2">
        <span v-if="notice.is_pinned" class="text-yellow-500">⭐</span>
        <h1 class="text-2xl font-bold text-fg">{{ notice.title }}</h1>
      </div>
      <p class="text-xs text-fg-muted mb-6">发布于 {{ formatDate(notice.created_at) }}</p>

      <div class="prose prose-sm max-w-none text-fg leading-relaxed">
        <p class="mb-3" v-html="renderContent(notice.content)" />
      </div>
    </div>
  </div>
</template>
