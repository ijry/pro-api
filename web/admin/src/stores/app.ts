import { defineStore } from 'pinia'

type Theme = 'light' | 'dark' | 'system'
type Locale = 'zh' | 'en'

interface State {
  theme: Theme
  resolvedTheme: 'light' | 'dark'
  locale: Locale
  sidebarCollapsed: boolean
}

const KS = {
  theme: 'proapi_admin_theme',
  locale: 'proapi.locale',
  collapsed: 'proapi_admin_sidebar_collapsed',
}

function resolveTheme(t: Theme): 'light' | 'dark' {
  if (t === 'system') return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  return t
}

export const useAppStore = defineStore('app', {
  state: (): State => ({
    theme: (localStorage.getItem(KS.theme) as Theme) || 'system',
    resolvedTheme: 'light',
    locale: (localStorage.getItem(KS.locale) as Locale) || 'zh',
    sidebarCollapsed: localStorage.getItem(KS.collapsed) === '1',
  }),
  actions: {
    init() {
      this.resolvedTheme = resolveTheme(this.theme)
      this.applyTheme()
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        if (this.theme === 'system') { this.resolvedTheme = resolveTheme('system'); this.applyTheme() }
      })
    },
    setTheme(t: Theme) {
      this.theme = t
      localStorage.setItem(KS.theme, t)
      this.resolvedTheme = resolveTheme(t)
      this.applyTheme()
    },
    toggleTheme() {
      const next = this.resolvedTheme === 'dark' ? 'light' : 'dark'
      this.setTheme(next)
    },
    applyTheme() {
      document.documentElement.classList.toggle('dark', this.resolvedTheme === 'dark')
    },
    setLocale(l: Locale) {
      this.locale = l
      localStorage.setItem(KS.locale, l)
    },
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      localStorage.setItem(KS.collapsed, this.sidebarCollapsed ? '1' : '0')
    },
  },
})
