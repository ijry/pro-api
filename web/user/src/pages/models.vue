<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { modelApi, type ModelInfo } from '@/api/model'
import { useToast } from '@/composables/useToast'
import ModelCard from '@/components/biz/ModelCard.vue'
import Skeleton from '@/components/ui/Skeleton.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import Input from '@/components/ui/Input.vue'

const toast = useToast()
const models = ref<ModelInfo[]>([])
const loading = ref(true)
const search = ref('')

async function load() {
  loading.value = true
  try {
    const r = await modelApi.list()
    models.value = r.models
  } catch { toast.error('加载模型失败') } finally { loading.value = false }
}

onMounted(load)

const filtered = computed(() => {
  if (!search.value) return models.value
  const q = search.value.toLowerCase()
  return models.value.filter(m => (m.name || m.id).toLowerCase().includes(q) || (m.provider || '').toLowerCase().includes(q))
})

const grouped = computed(() => {
  const groups: Record<string, ModelInfo[]> = {}
  for (const m of filtered.value) {
    const key = m.provider || 'Other'
    if (!groups[key]) groups[key] = []
    groups[key].push(m)
  }
  return Object.entries(groups)
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-start justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold text-fg">模型广场</h1>
        <p class="text-sm text-fg-muted mt-1">查看可用的 AI 模型及定价</p>
      </div>
      <Input v-model="search" placeholder="搜索模型..." size="sm" class="w-48" />
    </div>

    <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <Skeleton v-for="i in 9" :key="i" class="h-28" />
    </div>

    <EmptyState v-else-if="!filtered.length" icon="i-lucide-cpu" title="没有匹配的模型" subtitle="尝试调整搜索关键词" />

    <div v-else class="space-y-6">
      <div v-for="[provider, list] in grouped" :key="provider">
        <h2 class="text-sm font-semibold text-fg-muted uppercase tracking-wider mb-3">{{ provider }}</h2>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <ModelCard v-for="model in list" :key="model.id" :model="model" />
        </div>
      </div>
    </div>
  </div>
</template>
