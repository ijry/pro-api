<script setup lang="ts">
import { ref } from 'vue'
import { useToast } from '@/composables/useToast'

interface Props {
  text: string
  successMsg?: string
  size?: 'sm' | 'md'
}
const props = withDefaults(defineProps<Props>(), { successMsg: '已复制', size: 'md' })
const toast = useToast()
const copied = ref(false)

async function copy() {
  try {
    await navigator.clipboard.writeText(props.text)
    copied.value = true
    toast.success(props.successMsg)
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    toast.error('复制失败，请手动复制')
  }
}
</script>

<template>
  <button
    @click="copy"
    :class="[
      'inline-flex items-center gap-1 transition-colors',
      size === 'sm' ? 'p-1' : 'px-2 py-1',
      'rounded text-fg-muted hover:text-fg hover:bg-bg-elevated',
    ]"
    :title="copied ? successMsg : '复制'"
  >
    <span :class="[copied ? 'i-lucide-check' : 'i-lucide-clipboard', 'w-4 h-4']" />
    <span v-if="$slots.default" class="text-sm"><slot /></span>
  </button>
</template>
