<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWalletStore } from '@/stores/wallet'
import { useNoticeStore } from '@/stores/notice'
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const wallet = useWalletStore()
const notice = useNoticeStore()
const app = useAppStore()
const { t, locale } = useI18n()

const avatarMenuOpen = ref(false)
const langMenuOpen = ref(false)

const navLinks = [
  { to: '/', label: 'nav.home', icon: 'i-lucide-layout-dashboard', exact: true },
  { to: '/tokens', label: 'nav.tokens', icon: 'i-lucide-key-round' },
  { to: '/recharge', label: 'nav.recharge', icon: 'i-lucide-wallet' },
  { to: '/redeem', label: 'nav.redeem', icon: 'i-lucide-ticket' },
  { to: '/logs', label: 'nav.logs', icon: 'i-lucide-list' },
  { to: '/models', label: 'nav.models', icon: 'i-lucide-bot' },
  { to: '/notices', label: 'nav.notices', icon: 'i-lucide-megaphone' },
]

const balanceDisplay = computed(() => {
  const b = wallet.balanceUsd
  if (app.locale === 'zh') {
    return `¥ ${wallet.balanceCny.toFixed(2)}`
  }
  return `$ ${b.toFixed(3)}`
})

function isActive(to: string, exact?: boolean) {
  if (exact) return route.path === to
  return route.path.startsWith(to)
}

async function logout() {
  avatarMenuOpen.value = false
  await auth.logout()
  router.push('/login')
}

function switchLocale(l: 'zh' | 'en') {
  app.setLocale(l)
  locale.value = l
  langMenuOpen.value = false
}

function toggleAvatar() {
  avatarMenuOpen.value = !avatarMenuOpen.value
  langMenuOpen.value = false
}
</script>

<template>
  <header class="sticky top-0 z-40 border-b border-border bg-bg/80 backdrop-blur-md">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-14 flex items-center gap-4">
      <!-- Logo -->
      <router-link to="/" class="flex items-center gap-2 shrink-0">
        <span class="text-primary font-bold text-xl tracking-tight">proapi</span>
      </router-link>

      <!-- Nav links (desktop) -->
      <nav class="hidden md:flex items-center gap-1 flex-1">
        <router-link
          v-for="link in navLinks"
          :key="link.to"
          :to="link.to"
          class="flex items-center gap-1.5 px-3 h-8 rounded-md text-sm transition-colors"
          :class="isActive(link.to, link.exact)
            ? 'bg-primary/10 text-primary font-medium'
            : 'text-fg-muted hover:text-fg hover:bg-bg-elevated'"
        >
          <span :class="[link.icon, 'w-4 h-4']" />
          {{ t(link.label) }}
          <span v-if="link.to === '/notices' && notice.unreadCount > 0"
            class="ml-0.5 bg-rose-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center">
            {{ notice.unreadCount > 9 ? '9+' : notice.unreadCount }}
          </span>
        </router-link>
      </nav>

      <!-- Right cluster -->
      <div class="flex items-center gap-2 ml-auto">
        <!-- Balance pill -->
        <router-link
          to="/recharge"
          class="hidden sm:flex items-center gap-1.5 px-3 h-8 rounded-md border border-border text-sm text-fg hover:bg-bg-elevated transition-colors"
        >
          <span class="i-lucide-wallet w-3.5 h-3.5 text-primary" />
          <span v-if="wallet.loading" class="text-fg-muted">--</span>
          <span v-else>{{ balanceDisplay }}</span>
        </router-link>

        <!-- Theme toggle -->
        <button @click="app.toggleTheme"
          class="w-8 h-8 flex items-center justify-center rounded-md text-fg-muted hover:text-fg hover:bg-bg-elevated transition-colors">
          <span v-if="app.theme === 'dark'" class="i-lucide-sun w-4 h-4" />
          <span v-else class="i-lucide-moon w-4 h-4" />
        </button>

        <!-- Language -->
        <div class="relative">
          <button @click="langMenuOpen = !langMenuOpen; avatarMenuOpen = false"
            class="w-8 h-8 flex items-center justify-center rounded-md text-fg-muted hover:text-fg hover:bg-bg-elevated transition-colors">
            <span class="i-lucide-globe w-4 h-4" />
          </button>
          <div v-if="langMenuOpen"
            class="absolute right-0 top-10 w-32 bg-bg-elevated border border-border rounded-lg shadow-xl overflow-hidden z-50">
            <button v-for="l in ['zh', 'en']" :key="l"
              @click="switchLocale(l as 'zh' | 'en')"
              class="w-full px-4 py-2 text-left text-sm hover:bg-border/30 transition-colors"
              :class="app.locale === l ? 'text-primary' : 'text-fg'">
              {{ l === 'zh' ? '中文' : 'English' }}
            </button>
          </div>
        </div>

        <!-- Avatar -->
        <div class="relative">
          <button @click="toggleAvatar"
            class="w-8 h-8 flex items-center justify-center rounded-full bg-primary/20 text-primary text-sm font-medium hover:bg-primary/30 transition-colors">
            {{ auth.user?.display_name?.[0]?.toUpperCase() || '?' }}
          </button>
          <div v-if="avatarMenuOpen"
            class="absolute right-0 top-10 w-52 bg-bg-elevated border border-border rounded-lg shadow-xl overflow-hidden z-50">
            <div class="px-4 py-3 border-b border-border">
              <p class="text-sm font-medium text-fg">{{ auth.user?.display_name }}</p>
              <p class="text-xs text-fg-muted truncate">{{ auth.user?.email }}</p>
            </div>
            <router-link to="/profile" @click="avatarMenuOpen = false"
              class="flex items-center gap-2 px-4 py-2 text-sm text-fg hover:bg-border/30 transition-colors">
              <span class="i-lucide-user w-4 h-4" /> {{ t('nav.profile') }}
            </router-link>
            <router-link to="/profile/security" @click="avatarMenuOpen = false"
              class="flex items-center gap-2 px-4 py-2 text-sm text-fg hover:bg-border/30 transition-colors">
              <span class="i-lucide-shield w-4 h-4" /> {{ t('profile.security.title') }}
            </router-link>
            <router-link to="/profile/oauth" @click="avatarMenuOpen = false"
              class="flex items-center gap-2 px-4 py-2 text-sm text-fg hover:bg-border/30 transition-colors">
              <span class="i-lucide-link w-4 h-4" /> {{ t('profile.oauth.title') }}
            </router-link>
            <div class="border-t border-border" />
            <button @click="logout"
              class="flex items-center gap-2 w-full px-4 py-2 text-sm text-rose-400 hover:bg-border/30 transition-colors">
              <span class="i-lucide-log-out w-4 h-4" /> {{ t('auth.logout') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>
