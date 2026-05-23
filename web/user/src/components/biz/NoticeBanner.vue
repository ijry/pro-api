<script setup lang="ts">
import { type Notice } from '@/api/notice'

interface Props { notice: Notice }
defineProps<Props>()
const emit = defineEmits<{ (e: 'dismiss'): void }>()
</script>

<template>
  <div class="bg-primary/10 border-b border-primary/20 px-4 py-2.5">
    <div class="max-w-7xl mx-auto flex items-center gap-3">
      <span class="i-lucide-megaphone w-4 h-4 text-primary shrink-0" />
      <p class="flex-1 text-sm text-fg">
        <span class="font-medium text-primary mr-2">{{ notice.title }}</span>
        <router-link :to="`/notices/${notice.id}`" class="hover:underline">查看详情 →</router-link>
      </p>
      <button @click="emit('dismiss')"
        class="text-fg-muted hover:text-fg transition-colors">
        <span class="i-lucide-x w-4 h-4" />
      </button>
    </div>
  </div>
</template>
