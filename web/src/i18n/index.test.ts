import { describe, it, expect } from 'vitest'
import i18n, { supportedLanguages, type SupportedLanguage } from './index'

describe('i18n', () => {
  it('exports supportedLanguages with en and ro', () => {
    expect(supportedLanguages).toContain('en')
    expect(supportedLanguages).toContain('ro')
  })

  it('initializes i18n with en as fallback', () => {
    expect(i18n.options.fallbackLng).toEqual(['en'])
  })

  it('has translation resources for en', () => {
    expect(i18n.hasResourceBundle('en', 'translation')).toBe(true)
  })

  it('has translation resources for ro', () => {
    expect(i18n.hasResourceBundle('ro', 'translation')).toBe(true)
  })

  it('SupportedLanguage type matches array values', () => {
    const lang: SupportedLanguage = 'en'
    expect(supportedLanguages).toContain(lang)
  })
})
