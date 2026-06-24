import { createRouter, createWebHashHistory, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { i18n } from '@/i18n'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/pages/index.vue'),
    meta: { auth: true, title: 'nav.home', layout: 'app' },
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/pages/login.vue'),
    meta: { auth: false, title: 'auth.login.title', layout: 'auth' },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/pages/register.vue'),
    meta: { auth: false, title: 'auth.register.title', layout: 'auth' },
  },
  {
    path: '/forgot',
    name: 'forgot',
    component: () => import('@/pages/forgot.vue'),
    meta: { auth: false, title: 'auth.forgot.title', layout: 'auth' },
  },
  {
    path: '/oauth/:provider/callback',
    name: 'oauth-callback',
    component: () => import('@/pages/oauth/callback.vue'),
    meta: { auth: false, layout: 'auth' },
  },
  {
    path: '/apikeys',
    name: 'apikeys',
    component: () => import('@/pages/tokens/index.vue'),
    meta: { auth: true, title: 'nav.apikeys', layout: 'app' },
  },
  {
    path: '/recharge',
    name: 'recharge',
    component: () => import('@/pages/recharge/index.vue'),
    meta: { auth: true, title: 'nav.recharge', layout: 'app' },
  },
  {
    path: '/recharge/:orderId',
    name: 'recharge-detail',
    component: () => import('@/pages/recharge/detail.vue'),
    meta: { auth: true, title: 'recharge.detail.title', layout: 'app' },
  },
  {
    path: '/redeem',
    name: 'redeem',
    component: () => import('@/pages/redeem.vue'),
    meta: { auth: true, title: 'nav.redeem', layout: 'app' },
  },
  {
    path: '/logs',
    name: 'logs',
    component: () => import('@/pages/logs/index.vue'),
    meta: { auth: true, title: 'nav.logs', layout: 'app' },
  },
  {
    path: '/models',
    name: 'models',
    component: () => import('@/pages/models.vue'),
    meta: { auth: true, title: 'nav.models', layout: 'app' },
  },
  {
    path: '/notices',
    name: 'notices',
    component: () => import('@/pages/notices/index.vue'),
    meta: { auth: true, title: 'nav.notices', layout: 'app' },
  },
  {
    path: '/notices/:id',
    name: 'notice-detail',
    component: () => import('@/pages/notices/detail.vue'),
    meta: { auth: true, title: 'notices.detail.title', layout: 'app' },
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('@/pages/profile/index.vue'),
    meta: { auth: true, title: 'nav.profile', layout: 'app' },
  },
  {
    path: '/profile/security',
    name: 'profile-security',
    component: () => import('@/pages/profile/security.vue'),
    meta: { auth: true, title: 'profile.security.title', layout: 'app' },
  },
  {
    path: '/profile/oauth',
    name: 'profile-oauth',
    component: () => import('@/pages/profile/oauth.vue'),
    meta: { auth: true, title: 'profile.oauth.title', layout: 'app' },
  },
  {
    path: '/playground',
    name: 'playground',
    component: () => import('@/pages/playground.vue'),
    meta: { auth: true, title: 'nav.playground', layout: 'app' },
  },
  {
    path: '/invites',
    name: 'invites',
    component: () => import('@/pages/invites.vue'),
    meta: { auth: true, title: 'nav.invites', layout: 'app' },
  },
  {
    path: '/:all(.*)*',
    name: '404',
    component: () => import('@/pages/404.vue'),
    meta: { auth: false, layout: 'auth' },
  },
]

const MOCK = import.meta.env.VITE_DEMO_MOCK === 'true'
const baseUrl = MOCK ? import.meta.env.BASE_URL : '/'

export const router = createRouter({
  history: MOCK
    ? createWebHashHistory(baseUrl)
    : createWebHistory(baseUrl),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (MOCK) {
    if (!auth.user) {
      try { await auth.refresh() } catch { /* mock 不会失败 */ }
    }
    if (['login', 'register', 'forgot'].includes(to.name as string)) {
      return { name: 'home' }
    }
    return
  }

  if (to.meta.auth) {
    if (!auth.user) {
      try { await auth.refresh() } catch { /* fall through */ }
    }
    if (!auth.user) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }
  }
  if (!to.meta.auth && auth.user && ['login', 'register', 'forgot'].includes(to.name as string)) {
    return { name: 'home' }
  }
})

router.afterEach((to) => {
  const titleKey = to.meta.title as string | undefined
  document.title = titleKey ? `${i18n.global.t(titleKey)} · pro-api` : 'pro-api'
})
