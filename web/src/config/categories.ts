import { theme } from './theme'

export interface CategoryMeta {
  label: string
  color: string
  icon: string
}

export interface SystemColors {
  text: string
  textHeading: string
  textMuted: string
  background: string
  surface: string
  surfaceAlt: string
  border: string
  codeBg: string
  accent: string
  accentBg: string
  accentBorder: string
  shadow: string
  headerBg: string
  headerText: string
  panelBg: string
  panelBorder: string
  graphBg: string
  status: {
    success: string
    info: string
    warning: string
    error: string
    neutral: string
  }
}

export type ThemeId = 'default' | 'pastel' | 'monochrome' | 'high-contrast'

export interface ThemePreset {
  id: ThemeId
  label: string
  description: string
  system: SystemColors
  categories: Record<string, CategoryMeta>
}

const ICON_CIRCLE =
  'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z'

function withIcons(cats: Record<string, Omit<CategoryMeta, 'icon'>>): Record<string, CategoryMeta> {
  const out: Record<string, CategoryMeta> = {}
  for (const [k, v] of Object.entries(cats)) {
    out[k] = { ...v, icon: ICON_CIRCLE }
  }
  return out
}

export const CATEGORY_META: Record<string, CategoryMeta> = withIcons({
  compute: { label: 'Compute', color: theme.category.compute },
  network: { label: 'Network', color: theme.category.network },
  database: { label: 'Database', color: theme.category.database },
  storage: { label: 'Storage', color: theme.category.storage },
  security: { label: 'Security', color: theme.category.security },
  integration: { label: 'Integration', color: theme.category.integration },
  other: { label: 'Other', color: theme.category.other },
})

const light: SystemColors = {
  text: '#6b6375',
  textHeading: '#08060d',
  textMuted: '#666',
  background: '#fff',
  surface: '#fff',
  surfaceAlt: '#f5f5f5',
  border: '#e5e4e7',
  codeBg: '#f4f3ec',
  accent: '#1976d2',
  accentBg: 'rgba(25, 118, 210, 0.1)',
  accentBorder: 'rgba(25, 118, 210, 0.5)',
  shadow:
    'rgba(0, 0, 0, 0.1) 0 10px 15px -3px, rgba(0, 0, 0, 0.05) 0 4px 6px -2px',
  headerBg: 'linear-gradient(135deg, #1976d2 0%, #1565c0 100%)',
  headerText: '#fff',
  panelBg: '#fafafa',
  panelBorder: '#e0e0e0',
  graphBg: '#fafafa',
  status: { success: '#4CAF50', info: '#2196F3', warning: '#FF9800', error: '#F44336', neutral: '#9E9E9E' },
}

export const THEMES: Record<ThemeId, ThemePreset> = {
  default: {
    id: 'default',
    label: 'Default',
    description: 'Material Design inspired colors',
    system: light,
    categories: CATEGORY_META,
  },
  pastel: {
    id: 'pastel',
    label: 'Pastel',
    description: 'Soft, muted colors for long sessions',
    system: {
      text: '#5a5a6e',
      textHeading: '#2d2d3a',
      textMuted: '#8a8a9e',
      background: '#f7f5f9',
      surface: '#ffffff',
      surfaceAlt: '#eeecf2',
      border: '#ddd9e6',
      codeBg: '#f0eef4',
      accent: '#6a8caf',
      accentBg: 'rgba(106, 140, 175, 0.12)',
      accentBorder: 'rgba(106, 140, 175, 0.35)',
      shadow:
        'rgba(60, 50, 80, 0.08) 0 10px 15px -3px, rgba(60, 50, 80, 0.04) 0 4px 6px -2px',
      headerBg: 'linear-gradient(135deg, #7a9cc6 0%, #5a7ea8 100%)',
      headerText: '#ffffff',
      panelBg: '#f3f0f7',
      panelBorder: '#ddd9e6',
      graphBg: '#f4f2f7',
      status: { success: '#6fba8b', info: '#6a8caf', warning: '#d49a68', error: '#c97979', neutral: '#8a8a9e' },
    },
    categories: withIcons({
      compute: { label: 'Compute', color: '#A8D5BA' },
      network: { label: 'Network', color: '#A8C8E8' },
      database: { label: 'Database', color: '#C8A8D5' },
      storage: { label: 'Storage', color: '#F4C8A8' },
      security: { label: 'Security', color: '#E8A8A8' },
      integration: { label: 'Integration', color: '#A8D5D5' },
      other: { label: 'Other', color: '#C8C8C8' },
    }),
  },
  monochrome: {
    id: 'monochrome',
    label: 'Monochrome',
    description: 'Grayscale palette for accessibility',
    system: {
      text: '#333333',
      textHeading: '#000000',
      textMuted: '#666666',
      background: '#ffffff',
      surface: '#ffffff',
      surfaceAlt: '#f5f5f5',
      border: '#cccccc',
      codeBg: '#f5f5f5',
      accent: '#333333',
      accentBg: 'rgba(0, 0, 0, 0.08)',
      accentBorder: 'rgba(0, 0, 0, 0.25)',
      shadow:
        'rgba(0, 0, 0, 0.15) 0 10px 15px -3px, rgba(0, 0, 0, 0.05) 0 4px 6px -2px',
      headerBg: 'linear-gradient(135deg, #555 0%, #333 100%)',
      headerText: '#ffffff',
      panelBg: '#fafafa',
      panelBorder: '#e0e0e0',
      graphBg: '#f5f5f5',
      status: { success: '#333333', info: '#595959', warning: '#737373', error: '#1a1a1a', neutral: '#8c8c8c' },
    },
    categories: withIcons({
      compute: { label: 'Compute', color: '#2C2C2C' },
      network: { label: 'Network', color: '#595959' },
      database: { label: 'Database', color: '#737373' },
      storage: { label: 'Storage', color: '#8C8C8C' },
      security: { label: 'Security', color: '#A6A6A6' },
      integration: { label: 'Integration', color: '#BFBFBF' },
      other: { label: 'Other', color: '#D9D9D9' },
    }),
  },
  'high-contrast': {
    id: 'high-contrast',
    label: 'High Contrast',
    description: 'Vivid colors that pop on any background',
    system: {
      text: '#000000',
      textHeading: '#000000',
      textMuted: '#1a1a1a',
      background: '#ffffff',
      surface: '#ffffff',
      surfaceAlt: '#f0f0f0',
      border: '#000000',
      codeBg: '#f5f5f5',
      accent: '#d500f9',
      accentBg: 'rgba(213, 0, 249, 0.12)',
      accentBorder: 'rgba(213, 0, 249, 0.6)',
      shadow:
        'rgba(0, 0, 0, 0.25) 0 10px 15px -3px, rgba(0, 0, 0, 0.15) 0 4px 6px -2px',
      headerBg: 'linear-gradient(135deg, #000 0%, #333 100%)',
      headerText: '#00e676',
      panelBg: '#ffffff',
      panelBorder: '#000000',
      graphBg: '#ffffff',
      status: { success: '#00c853', info: '#2962ff', warning: '#ff6d00', error: '#d50000', neutral: '#666666' },
    },
    categories: withIcons({
      compute: { label: 'Compute', color: '#00c853' },
      network: { label: 'Network', color: '#2962ff' },
      database: { label: 'Database', color: '#aa00ff' },
      storage: { label: 'Storage', color: '#ff6d00' },
      security: { label: 'Security', color: '#d50000' },
      integration: { label: 'Integration', color: '#00bfa5' },
      other: { label: 'Other', color: '#000000' },
    }),
  },
}

export const THEME_LIST: ThemePreset[] = Object.values(THEMES)

export function getCategoryColor(category: string, themeId: ThemeId = 'default'): string {
  if (themeId === 'default') {
    return CATEGORY_META[category]?.color ?? theme.category.other
  }
  const selectedTheme = THEMES[themeId] ?? THEMES.default
  return selectedTheme.categories[category]?.color ?? selectedTheme.categories.other.color
}

export function getSystemColor(name: Exclude<keyof SystemColors, 'status'>, themeId: ThemeId = 'default'): string {
  const theme = THEMES[themeId] ?? THEMES.default
  return theme.system[name] ?? THEMES.default.system[name]
}
