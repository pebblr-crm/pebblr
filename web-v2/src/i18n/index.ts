import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import en from './en'
import ro from './ro'

export const supportedLanguages = ['en', 'ro'] as const
export type SupportedLanguage = (typeof supportedLanguages)[number]

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: { en, ro },
    fallbackLng: 'en',
    interpolation: { escapeValue: false },
    detection: {
      order: ['localStorage', 'navigator'],
      lookupLocalStorage: 'pebblr-language',
      caches: ['localStorage'],
    },
  })

export default i18n
