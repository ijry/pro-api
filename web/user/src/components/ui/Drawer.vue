<script setup lang="ts">
import { watch } from 'vue'

interface Props { title?: string; width?: string }
withDefaults(defineProps<Props>(), { width: '420px' })

const open = defineModel<boolean>({ default: false })

function close() { open.value = false }

watch(open, (v) => {
  document.body.style.overflow = v ? 'hidden' : ''
})
</script>

<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="open" class="fixed inset-0 z-50 flex">
        <div class="flex-1 bg-black/60 backdrop-blur-sm" @click="close" />
        <div
          class="bg-bg-elevated border-l border-border h-full flex flex-col shadow-2xl overflow-auto"
          :style="{ width }"
        >
          <div class="flex items-center justify-between px-6 py-4 border-b border-border shrink-0">
            <h3 class="text-lg font-semibold text-fg">{{ title }}</h3>
            <button @click="close" class="text-fg-muted hover:text-fg transition-colors">
              <span class="i-lucide-x w-5 h-5" />
            </button>
          </div>
          <div class="flex-1 px-6 py-5 overflow-auto">
            <slot />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.drawer-enter-active, .drawer-leave-active { transition: transform 0.25s ease; }
.drawer-enter-from .bg-bg-elevated, .drawer-leave-to .bg-bg-elevated {
  transform: translateX(100%);
}
</style>
