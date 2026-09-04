export const theme = {
  colorWhite: '#ffffff',
  colorBlack: '#000000',

  text: {
    primary: '#333333',
    secondary: '#666666',
    muted: '#999999',
    mutedAlt: '#888888',
    inverse: '#ffffff',
  },

  bg: {
    default: '#ffffff',
    subtle: '#fafafa',
    soft: '#f5f5f5',
    softAlt: '#f0f0f0',
    card: '#ffffff',
  },

  border: {
    default: '#e0e0e0',
    light: '#e8e8e8',
    lighter: '#eeeeee',
    lightest: '#e0e0e0',
    input: '#cccccc',
    inputAlt: '#bbbbbb',
  },

  primary: {
    DEFAULT: '#1976d2',
    hover: '#1565c0',
    light: '#90caf9',
    bg: '#e3f2fd',
    contrast: '#ffffff',
  },

  accent: {
    DEFAULT: '#aa3bff',
    bg: 'rgba(170, 59, 255, 0.1)',
    border: 'rgba(170, 59, 255, 0.5)',
    contrast: '#ffffff',
  },

  overlay: {
    white15: 'rgba(255, 255, 255, 0.15)',
    white20: 'rgba(255, 255, 255, 0.2)',
    white25: 'rgba(255, 255, 255, 0.25)',
    white30: 'rgba(255, 255, 255, 0.3)',
  },

  success: '#4CAF50',

  info: '#2196F3',

  warning: {
    DEFAULT: '#FF9800',
    dark: '#e65100',
    bg: '#fff3e0',
    contrast: '#000000',
  },

  amber: {
    DEFAULT: '#ffc107',
    bg: '#fff8e1',
  },

  error: {
    DEFAULT: '#F44336',
    dark: '#c62828',
    darkAlt: '#d32f2f',
    light: '#ff5252',
    bg: '#ffebee',
    contrast: '#ffffff',
  },

  gray: {
    DEFAULT: '#9E9E9E',
    50: '#fafafa',
    100: '#f5f5f5',
    200: '#e8e8e8',
    300: '#e0e0e0',
    400: '#cccccc',
    500: '#999999',
    600: '#666666',
    700: '#333333',
  },

  category: {
    compute: '#4CAF50',
    network: '#2196F3',
    database: '#9C27B0',
    storage: '#FF9800',
    security: '#F44336',
    integration: '#00BCD4',
    other: '#9E9E9E',
  },

  terminal: {
    bg: '#263238',
    panel: '#263238',
    btn: '#37474f',
    border: '#455a64',
    text: '#e0e0e0',
  },
} as const

export type Theme = typeof theme
export type StatusType = 'success' | 'info' | 'warning' | 'error' | 'neutral'
