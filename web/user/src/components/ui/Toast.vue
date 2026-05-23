<script setup lang="ts">
import { useToast } from '@/composables/useToast'
const { toasts, dismiss } = useToast()

const colorMap = {
  success: 'bg-emerald-500',
  error: 'bg-rose-500',
  warn: 'bg-amber-500',
  info: 'bg-sky-500',
}
const iconMap = {
  success: 'i-lucide-check-circle',
  error: 'i-lucide-x-circle',
  warn: 'i-lucide-alert-triangle',
  info: 'i-lucide-info',
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none">
      <TransitionGroup name="toast">
        <div
          v-for="t in toasts"
          :key="t.id"
          class="flex items-center gap-3 px-4 py-3 rounded-lg shadow-xl text-white text-sm pointer-events-auto min-w-60 max-w-sm"
          :class="colorMap[t.type]"
        >
          <span :class="['w-4 h-4 shrink-0', iconMap[t.type]]" />
          <span class="flex-1">{{ t.message }}</span>
          <button @click="dismiss(t.id)" class="shrink-0 opacity-70 hover:opacity-100">
            <span class="i-lucide-x w-4 h-4" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-enter-active, .toast-leave-active { transition: all 0.25s ease; }
.toast-enter-from { opacity: 0; transform: translateX(24px); }
.toast-leave-to { opacity: 0; transform: translateX(24px); }
</style>
