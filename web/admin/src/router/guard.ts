import type { Router } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { i18n } from '@/i18n'

const MOCK = import.meta.env.VITE_DEMO_MOCK === 'true'

export function installGuards(router: Router) {
  router.beforeEach(async (to) => {
    if (MOCK) {
      const us = useUserStore()
      if (!us.fetched) {
        try { await us.fetchMe() } catch (_) { /* mock 不会失败 */ }
      }
      if (to.name === 'login') {
        return { name: 'dashboard' }
      }
      return true
    }

    const publicRoutes = ['login', 'forbidden', 'not-found']
    if (typeof to.name === 'string' && publicRoutes.includes(to.name)) return true

    const us = useUserStore()
    if (!us.fetched) {
      try { await us.fetchMe() } catch (_) { /* 401 handled by http interceptor */ }
    }

    if (!us.user) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }

    const roles = (to.meta?.roles as number[] | undefined) ?? []
    if (roles.length > 0 && !roles.includes(us.user.role)) {
      return { name: 'forbidden' }
    }

    return true
  })

  router.afterEach((to) => {
    const titleKey = to.meta?.title as string | undefined
    if (typeof document !== 'undefined') {
      document.title = titleKey
        ? `${i18n.global.t(titleKey)} · pro-api admin`
        : 'pro-api admin'
    }
  })
}
