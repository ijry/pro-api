<script setup lang="ts">
import {
  NConfigProvider, NMessageProvider, NDialogProvider,
  NNotificationProvider, NLoadingBarProvider,
  darkTheme, zhCN, dateZhCN, enUS, dateEnUS,
  createDiscreteApi,
} from 'naive-ui'
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

onMounted(() => {
  app.init()
  // Inject discrete API for use in http interceptors
  const { message, loadingBar } = createDiscreteApi(['message', 'loadingBar'])
  window.$message = message
  window.$loadingBar = loadingBar
})

const theme = computed(() => (app.resolvedTheme === 'dark' ? darkTheme : null))
const locale = computed(() => (app.locale === 'en' ? enUS : zhCN))
const dateLocale = computed(() => (app.locale === 'en' ? dateEnUS : dateZhCN))
</script>

<template>
  <NConfigProvider :theme="theme" :locale="locale" :date-locale="dateLocale">
    <NLoadingBarProvider>
      <NMessageProvider>
        <NDialogProvider>
          <NNotificationProvider>
            <RouterView />
          </NNotificationProvider>
        </NDialogProvider>
      </NMessageProvider>
    </NLoadingBarProvider>
  </NConfigProvider>
</template>
