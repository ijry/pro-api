<script setup lang="ts">
interface Props {
  type?: 'text' | 'password' | 'email' | 'number'
  placeholder?: string
  disabled?: boolean
  error?: string
  size?: 'sm' | 'md'
}
withDefaults(defineProps<Props>(), { type: 'text', size: 'md' })
const model = defineModel<string | number>()
</script>

<template>
  <div class="flex flex-col gap-1">
    <input
      v-model="model"
      :type="type"
      :placeholder="placeholder"
      :disabled="disabled"
      :class="[
        'w-full rounded-md border bg-bg-elevated px-3 text-fg placeholder:text-fg-muted outline-none transition-colors',
        size === 'sm' ? 'h-8 text-sm' : 'h-10',
        error ? 'border-rose-500 focus:ring-1 focus:ring-rose-500' : 'border-border focus:border-primary focus:ring-1 focus:ring-primary/30',
        disabled ? 'opacity-50 cursor-not-allowed' : '',
      ]"
    />
    <span v-if="error" class="text-rose-400 text-xs">{{ error }}</span>
  </div>
</template>
