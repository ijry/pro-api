<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { logApi, type LogEntry } from '@/api/log'
import { useToast } from '@/composables/useToast'
import Skeleton from '@/components/ui/Skeleton.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import Input from '@/components/ui/Input.vue'

const toast = useToast()
const items = ref<LogEntry[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(true)
const filterModel = ref('')
const filterFrom = ref('')
const filterTo = ref('')

async function load() {
  loading.value = true
  try {
    const r = await logApi.list({ page: page.value, page_size: 20, model: filterModel.value || undefined, from: filterFrom.value || undefined, to: filterTo.value || undefined })
    items.value = r.items; total.value = r.total
  } catch { toast.error('加载日志失败') } finally { loading.value = false }
}

onMounted(load)

function doSearch() { page.value = 1; load() }

function statusColor(status: number) {
  if (status < 300) return 'text-green-500'
  if (status < 500) return 'text-yellow-500'
  return 'text-rose-500'
}

function formatQuota(v: number) { return v.toLocaleString(undefined, { maximumFractionDigits: 4 }) }
</script>

<template>
  <div class="space-y-5">
    <div>
      <h1 class="text-2xl font-bold text-fg">消费日志</h1>
      <p class="text-sm text-fg-muted mt-1">查看 API 请求记录和消耗情况</p>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 items-end">
      <div>
        <label class="block text-xs text-fg-muted mb-1">模型</label>
        <Input v-model="filterModel" placeholder="gpt-4" size="sm" class="w-36" @keydown.enter="doSearch" />
      </div>
      <div>
        <label class="block text-xs text-fg-muted mb-1">开始时间</label>
        <Input v-model="filterFrom" type="text" placeholder="2024-01-01" size="sm" class="w-36" @keydown.enter="doSearch" />
      </div>
      <div>
        <label class="block text-xs text-fg-muted mb-1">结束时间</label>
        <Input v-model="filterTo" type="text" placeholder="2024-12-31" size="sm" class="w-36" @keydown.enter="doSearch" />
      </div>
      <button @click="doSearch" class="h-8 px-3 rounded-md bg-primary text-white text-sm hover:bg-primary-hover transition-colors">搜索</button>
    </div>

    <!-- Table -->
    <div v-if="loading" class="space-y-2">
      <Skeleton v-for="i in 5" :key="i" class="h-10" />
    </div>
    <EmptyState v-else-if="!items.length" icon="i-lucide-scroll-text" title="暂无日志" subtitle="API 调用记录将显示在这里" />
    <div v-else class="overflow-x-auto rounded-lg border border-border">
      <table class="w-full text-sm min-w-[800px]">
        <thead class="bg-bg-elevated border-b border-border">
          <tr>
            <th class="text-left px-3 py-2 text-fg-muted font-medium">时间</th>
            <th class="text-left px-3 py-2 text-fg-muted font-medium">模型</th>
            <th class="text-right px-3 py-2 text-fg-muted font-medium">状态码</th>
            <th class="text-right px-3 py-2 text-fg-muted font-medium">延迟(ms)</th>
            <th class="text-right px-3 py-2 text-fg-muted font-medium">输入tokens</th>
            <th class="text-right px-3 py-2 text-fg-muted font-medium">输出tokens</th>
            <th class="text-right px-3 py-2 text-fg-muted font-medium">消耗</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in items" :key="log.id" class="border-b border-border last:border-0 hover:bg-bg-elevated/50 transition-colors">
            <td class="px-3 py-2 text-fg-muted text-xs font-mono">{{ new Date(log.created_at).toLocaleString('zh-CN') }}</td>
            <td class="px-3 py-2 font-mono text-xs text-fg">{{ log.model }}</td>
            <td class="px-3 py-2 text-right font-mono" :class="statusColor(log.status)">{{ log.status }}</td>
            <td class="px-3 py-2 text-right text-fg-muted">{{ log.latency_ms }}</td>
            <td class="px-3 py-2 text-right text-fg-muted">{{ log.prompt_tokens.toLocaleString() }}</td>
            <td class="px-3 py-2 text-right text-fg-muted">{{ log.completion_tokens.toLocaleString() }}</td>
            <td class="px-3 py-2 text-right text-fg font-mono text-xs">{{ formatQuota(log.cost_usd) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination v-if="total > 20" v-model="page" :total="total" :size="20" @update:model-value="load" />
  </div>
</template>
