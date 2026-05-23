<script setup lang="ts">
import { watch, nextTick } from 'vue'

interface Props {
  title?: string
  size?: 'sm' | 'md' | 'lg'
}
withDefaults(defineProps<Props>(), { size: 'md' })

const open = defineModel<boolean>({ default: false })

function close() { open.value = false }

watch(open, (v) => {
  if (v) {
    nextTick(() => document.querySelector('[data-dialog-close]')?.addEventListener('click', close))
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})
</script>

<template>
  <Teleport to="body">
    <Transition name="dialog">
      <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="close" />
        <div
          :class="[
            'relative bg-bg-elevated border border-border rounded-xl shadow-2xl w-full',
            size === 'sm' ? 'max-w-sm' : size === 'lg' ? 'max-w-2xl' : 'max-w-md',
          ]"
        >
          <div v-if="title" class="flex items-center justify-between px-6 pt-5 pb-4 border-b border-border">
            <h3 class="text-lg font-semibold text-fg">{{ title }}</h3>
            <button @click="close" class="text-fg-muted hover:text-fg transition-colors">
              <span class="i-lucide-x w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5">
            <slot />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.dialog-enter-active, .dialog-leave-active { transition: opacity 0.2s; }
.dialog-enter-from, .dialog-leave-to { opacity: 0; }
</style>
