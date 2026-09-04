import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { THEMES, THEME_LIST, getCategoryColor, getResourceColor, type ThemeId, type ThemePreset } from '../config/categories'

export type FontSize = 'small' | 'medium' | 'large' | 'very-large'

const STORAGE_KEY = 'weavelens.settings.v1'

interface PersistedSettings {
  fontSize: FontSize
  themeId: ThemeId
}

const FONT_SCALE: Record<FontSize, string> = {
  small: '0.85',
  medium: '1',
  large: '1.2',
  'very-large': '1.4',
}

function loadSettings(): PersistedSettings {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { fontSize: 'medium', themeId: 'default' }
    const parsed = JSON.parse(raw) as Partial<PersistedSettings>
    const fontSize: FontSize = parsed.fontSize && ['small', 'medium', 'large', 'very-large'].includes(parsed.fontSize)
      ? parsed.fontSize
      : 'medium'
    const themeId: ThemeId = parsed.themeId && parsed.themeId in THEMES
      ? parsed.themeId
      : 'default'
    return { fontSize, themeId }
  } catch {
    return { fontSize: 'medium', themeId: 'default' }
  }
}

function applyFontSize(size: FontSize) {
  document.documentElement.style.setProperty('--app-font-scale', FONT_SCALE[size])
}

function applyTheme(themeId: ThemeId) {
  const root = document.documentElement
  const system = THEMES[themeId]?.system ?? THEMES.default.system
  const properties: Record<string, string> = {
    '--text': system.text,
    '--text-h': system.textHeading,
    '--text-muted': system.textMuted,
    '--bg': system.background,
    '--surface': system.surface,
    '--surface-alt': system.surfaceAlt,
    '--border': system.border,
    '--code-bg': system.codeBg,
    '--accent': system.accent,
    '--accent-bg': system.accentBg,
    '--accent-border': system.accentBorder,
    '--shadow': system.shadow,
    '--app-header-bg': system.headerBg,
    '--app-header-text': system.headerText,
    '--app-panel-bg': system.panelBg,
    '--app-panel-border': system.panelBorder,
    '--app-graph-bg': system.graphBg,
    '--color-text-primary': system.textHeading,
    '--color-text-secondary': system.text,
    '--color-text-muted': system.textMuted,
    '--color-bg': system.background,
    '--color-bg-subtle': system.graphBg,
    '--color-bg-soft': system.surfaceAlt,
    '--color-bg-card': system.surface,
    '--color-border': system.border,
    '--color-border-light': system.border,
    '--color-border-lighter': system.border,
    '--color-primary': system.accent,
    '--color-primary-hover': system.accent,
    '--color-primary-bg': system.accentBg,
  }
  Object.entries(properties).forEach(([name, value]) => root.style.setProperty(name, value))
}

export const useSettingsStore = defineStore('settings', () => {
  const initial = loadSettings()
  const fontSize = ref<FontSize>(initial.fontSize)
  const themeId = ref<ThemeId>(initial.themeId)
  const open = ref(false)

  function setFontSize(size: FontSize) {
    fontSize.value = size
  }

  function setTheme(id: ThemeId) {
    themeId.value = id
  }

  function toggle() {
    open.value = !open.value
  }

  function close() {
    open.value = false
  }

  function getColor(category: string): string {
    return getCategoryColor(category, themeId.value)
  }

  function getResourceColorForType(type: string, category: string): string {
    return getResourceColor(type, category, themeId.value)
  }

  function currentTheme(): ThemePreset {
    return THEMES[themeId.value] ?? THEMES.default
  }

  watch([fontSize, themeId], ([fs, tid]) => {
    applyFontSize(fs)
    applyTheme(tid)
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ fontSize: fs, themeId: tid }))
    } catch {
      // ignore quota errors
    }
  }, { immediate: true })

  return {
    fontSize,
    themeId,
    open,
    themes: THEME_LIST,
    setFontSize,
    setTheme,
    toggle,
    close,
    getColor,
    getResourceColorForType,
    currentTheme,
  }
})
