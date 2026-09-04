import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useSettingsStore } from '../stores/settings'
import { THEMES, THEME_LIST } from '../config/categories'

describe('settings store', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  it('defaults to medium font and default theme', () => {
    const store = useSettingsStore()
    expect(store.fontSize).toBe('medium')
    expect(store.themeId).toBe('default')
  })

  it('updates font size', () => {
    const store = useSettingsStore()
    store.setFontSize('large')
    expect(store.fontSize).toBe('large')
  })

  it('updates theme', () => {
    const store = useSettingsStore()
    store.setTheme('pastel')
    expect(store.themeId).toBe('pastel')
  })

  it('exposes all themes', () => {
    expect(THEME_LIST.length).toBeGreaterThan(0)
    for (const theme of THEME_LIST) {
      expect(THEMES[theme.id]).toBeDefined()
    }
  })

  it('returns color from current theme', () => {
    const store = useSettingsStore()
    store.setTheme('pastel')
    const defaultColor = THEMES.default.categories.compute.color
    const pastelColor = THEMES.pastel.categories.compute.color
    expect(defaultColor).not.toBe(pastelColor)
    expect(store.getColor('compute')).toBe(pastelColor)
  })

  it('toggles dropdown open state', () => {
    const store = useSettingsStore()
    expect(store.open).toBe(false)
    store.toggle()
    expect(store.open).toBe(true)
    store.close()
    expect(store.open).toBe(false)
  })

  it('falls back to default for unknown theme id', () => {
    const raw = '{"fontSize":"medium","themeId":"nonexistent"}'
    localStorage.setItem('weavelens.settings.v1', raw)
    setActivePinia(createPinia())
    const store2 = useSettingsStore()
    expect(store2.themeId).toBe('default')
  })
})
