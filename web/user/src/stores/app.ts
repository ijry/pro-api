import { defineStore } from 'pinia'
import { ref } from 'vue'

type Theme = 'dark' | 'light'
type Locale = 'zh' | 'en'

export const useAppStore = defineStore('app', () => {
  const theme = ref<Theme>((localStorage.getItem('proapi.theme') as Theme) || 'dark')
  const locale = ref<Locale>((localStorage.getItem('proapi.locale') as Locale) || 'zh')

  function applyTheme() {
    document.documentElement.className = theme.value === 'dark' ? 'dark' : ''
  }

  function toggleTheme() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
    localStorage.setItem('proapi.theme', theme.value)
    applyTheme()
  }

  function setLocale(l: Locale) {
    locale.value = l
    localStorage.setItem('proapi.locale', l)
  }

  return { theme, locale, applyTheme, toggleTheme, setLocale }
})
