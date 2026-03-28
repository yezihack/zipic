import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import zh from './locales/zh.json'
import ja from './locales/ja.json'

const STORAGE_KEY = 'image_compressor_lang'

export type LocaleType = 'en' | 'zh' | 'ja'

function getDefaultLocale(): LocaleType {
  // Check localStorage first
  const stored = localStorage.getItem(STORAGE_KEY) as LocaleType | null
  if (stored && ['en', 'zh', 'ja'].includes(stored)) {
    return stored
  }

  // Check browser language
  const browserLang = navigator.language.toLowerCase()
  if (browserLang.startsWith('zh')) return 'zh'
  if (browserLang.startsWith('ja')) return 'ja'
  return 'en'
}

export const i18n = createI18n({
  legacy: false,
  locale: getDefaultLocale(),
  fallbackLocale: 'en',
  messages: {
    en,
    zh,
    ja
  }
})

export function setLocale(locale: LocaleType): void {
  i18n.global.locale.value = locale
  localStorage.setItem(STORAGE_KEY, locale)
  document.documentElement.lang = locale
}

export function getLocale(): LocaleType {
  return i18n.global.locale.value as LocaleType
}