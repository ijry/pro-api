<script setup lang="ts">
type BadgeVariant = 'default' | 'success' | 'warn' | 'error' | 'info' | 'muted'
withDefaults(defineProps<{ variant?: BadgeVariant; size?: 'sm' | 'md' }>(), {
  variant: 'default',
  size: 'md',
})

const variantCls: Record<string, string> = {
  default: 'bg-primary/10 text-primary',
  success: 'bg-emerald-500/10 text-emerald-400',
  warn: 'bg-amber-500/10 text-amber-400',
  error: 'bg-rose-500/10 text-rose-400',
  info: 'bg-sky-500/10 text-sky-400',
  muted: 'bg-border/40 text-fg-muted',
}
</script>

<template>
  <span :class="[
    'inline-flex items-center font-medium rounded-full',
    size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-0.5 text-sm',
    variantCls[variant]
  ]">
    <slot />
  </span>
</template>
