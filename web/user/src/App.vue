<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/layout/AppLayout.vue'
import AuthLayout from '@/layout/AuthLayout.vue'
import { useAppStore } from '@/stores/app'
import Toast from '@/components/ui/Toast.vue'
import { setHttpToast } from '@/api/http'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const app = useAppStore()
const toast = useToast()

// Connect http interceptor to toast
setHttpToast((msg, type) => {
  if (type === 'warn') toast.warn(msg)
  else if (type === 'info') toast.info(msg)
  else toast.error(msg)
})

// Apply theme on load
app.applyTheme()

const layoutMap = { app: AppLayout, auth: AuthLayout }
const Layout = computed(() => layoutMap[(route.meta.layout as 'app' | 'auth') || 'app'])
</script>

<template>
  <component :is="Layout">
    <RouterView />
  </component>
  <Toast />
</template>
