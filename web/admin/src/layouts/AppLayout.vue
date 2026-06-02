<script setup lang="ts">
import {
  NLayout, NLayoutSider, NLayoutHeader, NLayoutContent,
  NMenu, NButton, NIcon, NDropdown, NBreadcrumb, NBreadcrumbItem,
  NTag, NSpace, NText,
} from 'naive-ui'
import { computed, h } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useUserStore } from '@/stores/user'
import { useI18n } from 'vue-i18n'
import type { MenuOption } from 'naive-ui'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const app = useAppStore()
const userStore = useUserStore()

const menuOptions: MenuOption[] = [
  { label: () => h(RouterLink, { to: '/' }, { default: () => t('nav.dashboard') }), key: 'dashboard' },
  { label: () => h(RouterLink, { to: '/users' }, { default: () => t('nav.users') }), key: 'users' },
  { label: () => h(RouterLink, { to: '/groups' }, { default: () => t('nav.groups') }), key: 'groups' },
  { label: () => h(RouterLink, { to: '/channels' }, { default: () => t('nav.channels') }), key: 'channels' },
  { label: () => h(RouterLink, { to: '/accounts' }, { default: () => t('nav.accounts') }), key: 'accounts' },
  { label: () => h(RouterLink, { to: '/models' }, { default: () => t('nav.models') }), key: 'models' },
  { label: () => h(RouterLink, { to: '/pricing' }, { default: () => t('nav.pricing') }), key: 'pricing' },
  { label: () => h(RouterLink, { to: '/tokens' }, { default: () => t('nav.tokens') }), key: 'tokens' },
  {
    label: t('nav.logs'),
    key: 'logs',
    children: [
      { label: () => h(RouterLink, { to: '/logs/requests' }, { default: () => t('nav.logs_requests') }), key: 'logs-requests' },
      { label: () => h(RouterLink, { to: '/logs/errors' }, { default: () => t('nav.logs_errors') }), key: 'logs-errors' },
      { label: () => h(RouterLink, { to: '/logs/audit' }, { default: () => t('nav.logs_audit') }), key: 'logs-audit' },
    ],
  },
  { label: () => h(RouterLink, { to: '/stats' }, { default: () => t('nav.stats') }), key: 'stats' },
  { label: () => h(RouterLink, { to: '/notices' }, { default: () => t('nav.notices') }), key: 'notices' },
  {
    label: t('nav.payments'),
    key: 'payments',
    children: [
      { label: () => h(RouterLink, { to: '/payments/recharges' }, { default: () => t('nav.payments_recharges') }), key: 'payments-recharges' },
      { label: () => h(RouterLink, { to: '/payments/redeem' }, { default: () => t('nav.payments_redeem') }), key: 'payments-redeem' },
    ],
  },
  { label: () => h(RouterLink, { to: '/settings' }, { default: () => t('nav.settings') }), key: 'settings' },
]

const activeMenu = computed(() => String(route.name ?? 'dashboard'))

const userDropdownOptions = [
  { label: t('nav.profile'), key: 'profile' },
  { label: t('nav.logout'), key: 'logout' },
]

async function handleUserDropdown(key: string) {
  if (key === 'logout') {
    await userStore.logout()
    router.push({ name: 'login' })
  } else if (key === 'profile') {
    router.push({ name: 'profile' })
  }
}

const breadcrumbs = computed(() => {
  const bc = route.meta?.breadcrumb as string[] | undefined
  if (!bc) return []
  return bc.map((b) => ({
    label: b.startsWith(':') ? route.params[b.slice(1)] as string : t(`nav.${b}`),
    key: b,
  }))
})
</script>

<template>
  <NLayout has-sider style="height: 100vh">
    <NLayoutSider
      :collapsed="app.sidebarCollapsed"
      collapse-mode="width"
      :collapsed-width="64"
      :width="220"
      bordered
      show-trigger="bar"
      @update:collapsed="app.toggleSidebar"
    >
      <div class="h-14 flex items-center justify-center px-4 border-b border-border">
        <NText v-if="!app.sidebarCollapsed" strong class="text-lg">proapi</NText>
        <NText v-else strong>P</NText>
      </div>
      <NMenu
        :collapsed="app.sidebarCollapsed"
        :collapsed-width="64"
        :collapsed-icon-size="20"
        :options="menuOptions"
        :value="activeMenu"
        :default-expanded-keys="['logs', 'payments']"
      />
    </NLayoutSider>

    <NLayout>
      <NLayoutHeader bordered class="px-4 h-14 flex items-center gap-3">
        <NButton text @click="app.toggleSidebar">
          <NIcon size="20">
            <span class="i-lucide-menu" />
          </NIcon>
        </NButton>

        <NBreadcrumb v-if="breadcrumbs.length" class="flex-1">
          <NBreadcrumbItem v-for="bc in breadcrumbs" :key="bc.key">
            {{ bc.label }}
          </NBreadcrumbItem>
        </NBreadcrumb>
        <div v-else class="flex-1" />

        <NSpace align="center" :size="12">
          <NButton text @click="app.toggleTheme">
            <NIcon size="20">
              <span :class="app.resolvedTheme === 'dark' ? 'i-lucide-sun' : 'i-lucide-moon'" />
            </NIcon>
          </NButton>

          <NButton text @click="app.setLocale(app.locale === 'zh' ? 'en' : 'zh')" size="small">
            {{ app.locale === 'zh' ? 'EN' : '中' }}
          </NButton>

          <NDropdown
            :options="userDropdownOptions"
            @select="handleUserDropdown"
          >
            <NButton text>
              <NTag type="primary" size="small">
                {{ userStore.user?.username ?? '--' }}
              </NTag>
            </NButton>
          </NDropdown>
        </NSpace>
      </NLayoutHeader>

      <NLayoutContent class="p-4 overflow-auto" style="height: calc(100vh - 56px)">
        <RouterView v-slot="{ Component }">
          <Transition name="fade" mode="out-in">
            <component :is="Component" />
          </Transition>
        </RouterView>
      </NLayoutContent>
    </NLayout>
  </NLayout>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
