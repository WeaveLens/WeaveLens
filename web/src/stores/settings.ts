import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { THEMES, THEME_LIST, getCategoryColor, getSystemColor, type ThemeId, type ThemePreset } from '../config/categories'

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
  if (themeId === 'default') {
    root.style.removeProperty('--text')
    root.style.removeProperty('--text-h')
    root.style.removeProperty('--text-muted')
    root.style.removeProperty('--bg')
    root.style.removeProperty('--surface')
    root.style.removeProperty('--surface-alt')
    root.style.removeProperty('--border')
    root.style.removeProperty('--code-bg')
    root.style.removeProperty('--accent')
    root.style.removeProperty('--accent-bg')
    root.style.removeProperty('--accent-border')
    root.style.removeProperty('--shadow')
    root.style.removeProperty('--app-header-bg')
    root.style.removeProperty('--app-header-text')
    root.style.removeProperty('--app-panel-bg')
    root.style.removeProperty('--app-panel-border')
    root.style.removeProperty('--app-graph-bg')
  } else {
    root.style.setProperty('--text', getSystemColor('text', themeId))
    root.style.setProperty('--text-h', getSystemColor('textHeading', themeId))
    root.style.setProperty('--text-muted', getSystemColor('textMuted', themeId))
    root.style.setProperty('--bg', getSystemColor('background', themeId))
    root.style.setProperty('--surface', getSystemColor('surface', themeId))
    root.style.setProperty('--surface-alt', getSystemColor('surfaceAlt', themeId))
    root.style.setProperty('--border', getSystemColor('border', themeId))
    root.style.setProperty('--code-bg', getSystemColor('codeBg', themeId))
    root.style.setProperty('--accent', getSystemColor('accent', themeId))
    root.style.setProperty('--accent-bg', getSystemColor('accentBg', themeId))
    root.style.setProperty('--accent-border', getSystemColor('accentBorder', themeId))
    root.style.setProperty('--shadow', getSystemColor('shadow', themeId))
    root.style.setProperty('--app-header-bg', getSystemColor('headerBg', themeId))
    root.style.setProperty('--app-header-text', getSystemColor('headerText', themeId))
    root.style.setProperty('--app-panel-bg', getSystemColor('panelBg', themeId))
    root.style.setProperty('--app-panel-border', getSystemColor('panelBorder', themeId))
    root.style.setProperty('--app-graph-bg', getSystemColor('graphBg', themeId))
  }
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
    currentTheme,
  }
})
