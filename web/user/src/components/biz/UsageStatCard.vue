<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { usageApi, type UsageStat } from '@/api/usage'
import Skeleton from '@/components/ui/Skeleton.vue'

interface Props { scope: 'today' | 'month' }
const props = defineProps<Props>()
const { t } = useI18n()

const stat = ref<UsageStat | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    stat.value = await usageApi.get(props.scope)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="rounded-xl border border-border bg-bg-elevated backdrop-blur-md p-5 shadow-sm">
    <template v-if="loading">
      <Skeleton class="h-6 w-24 mb-2" />
      <Skeleton class="h-8 w-20" />
    </template>
    <template v-else>
      <p class="text-xs text-fg-muted uppercase tracking-wide">
        {{ scope === 'today' ? t('home.usage.today') : t('home.usage.month') }}
      </p>
      <p class="text-2xl font-semibold text-fg mt-1">
        $ {{ (stat?.cost_usd ?? 0).toFixed(4) }}
      </p>
      <p class="text-xs text-fg-muted mt-1">
        {{ t('home.usage.requests', { n: (stat?.request_count ?? 0).toLocaleString() }) }}
      </p>
    </template>
  </div>
</template>
