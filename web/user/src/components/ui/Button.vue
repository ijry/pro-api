<script setup lang="ts">
interface Props {
  variant?: 'primary' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
}
const props = withDefaults(defineProps<Props>(), { variant: 'primary', size: 'md' })
const sizeCls = { sm: 'h-8 px-3 text-sm', md: 'h-10 px-4', lg: 'h-12 px-6 text-lg' }[props.size]
const variantCls = props.variant === 'primary'
  ? 'bg-primary text-white hover:bg-primary-hover'
  : 'bg-transparent text-fg border border-border hover:bg-bg-elevated'
</script>

<template>
  <button :class="['inline-flex items-center justify-center rounded-md font-medium transition-colors active:scale-[0.98]', sizeCls, variantCls]">
    <slot />
  </button>
</template>
