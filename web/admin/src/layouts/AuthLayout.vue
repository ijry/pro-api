<script setup lang="ts">
// Auth layout: centered card
</script>

<template>
  <div class="min-h-screen flex flex-col items-center justify-center bg-gray-50 dark:bg-gray-900 p-4">
    <slot />
    <div class="mt-6 text-sm text-gray-400">
      proapi admin &middot; {{ new Date().getFullYear() }}
    </div>
  </div>
</template>
