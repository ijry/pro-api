<script setup lang="ts">
import { NButton, NTooltip, NIcon } from 'naive-ui'
import { useMessage } from 'naive-ui'
import { ref } from 'vue'

const props = defineProps<{
  value: string
  size?: 'small' | 'tiny'
  tooltip?: string
}>()

const message = useMessage()
const copied = ref(false)

async function copy() {
  try {
    await navigator.clipboard.writeText(props.value)
    copied.value = true
    message.success('已复制')
    setTimeout(() => { copied.value = false }, 2000)
  } catch (_) {
    // fallback
    const ta = document.createElement('textarea')
    ta.value = props.value
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    message.success('已复制')
  }
}
</script>

<template>
  <NTooltip>
    <template #trigger>
      <NButton text :size="props.size ?? 'small'" @click.stop="copy">
        <NIcon>
          <span :class="copied ? 'i-lucide-check' : 'i-lucide-copy'" />
        </NIcon>
      </NButton>
    </template>
    {{ props.tooltip ?? '复制' }}
  </NTooltip>
</template>
