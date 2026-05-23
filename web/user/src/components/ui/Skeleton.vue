<script setup lang="ts">
interface Props {
  class?: string
}
defineProps<Props>()
</script>

<template>
  <div
    :class="['animate-pulse rounded-md bg-border/40', $props.class]"
  />
</template>
