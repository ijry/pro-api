import { createI18n } from 'vue-i18n'
import sharedZh from '@proapi/shared/i18n/zh'
import sharedEn from '@proapi/shared/i18n/en'
import zh from './zh.json'
import en from './en.json'

export const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('proapi.locale') || 'zh',
  fallbackLocale: 'en',
  messages: {
    zh: { ...sharedZh, ...zh },
    en: { ...sharedEn, ...en },
  },
})
