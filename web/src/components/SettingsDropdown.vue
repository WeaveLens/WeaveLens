<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useSettingsStore, type FontSize } from '../stores/settings'

const settingsStore = useSettingsStore()
const rootRef = ref<HTMLElement | null>(null)

function onDocumentClick(e: MouseEvent) {
  if (!settingsStore.open) return
  if (rootRef.value && !rootRef.value.contains(e.target as Node)) {
    settingsStore.close()
  }
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && settingsStore.open) {
    settingsStore.close()
  }
}

onMounted(() => {
  document.addEventListener('mousedown', onDocumentClick)
  document.addEventListener('keydown', onKey)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', onDocumentClick)
  document.removeEventListener('keydown', onKey)
})

const fontOptions: { value: FontSize; label: string }[] = [
  { value: 'small', label: 'Small' },
  { value: 'medium', label: 'Medium' },
  { value: 'large', label: 'Large' },
  { value: 'very-large', label: 'XLarge' },
]

const shortcuts: { keys: string; desc: string }[] = [
  { keys: 'L', desc: 'Toggle layout lock' },
  { keys: 'P', desc: 'Pin/unpin current scan' },
  { keys: 'F', desc: 'Fit the entire graph into the canvas' },
  { keys: 'R', desc: 'Recalculate node positions and fit the graph' },
  { keys: '+ / =', desc: 'Zoom in around the canvas center' },
  { keys: '-', desc: 'Zoom out around the canvas center' },
  { keys: '0', desc: 'Reset to 100% zoom and center the graph' },
]
</script>

<template>
  <div ref="rootRef" class="settings-root">
    <button
      class="header-btn"
      :class="{ active: settingsStore.open }"
      title="Settings"
      @click="settingsStore.toggle"
    >
      ⚙️ Settings
    </button>

    <div v-if="settingsStore.open" class="settings-dropdown">
      <section class="settings-section">
        <h4 class="settings-title">Font size</h4>
        <div class="font-size-group">
          <label
            v-for="opt in fontOptions"
            :key="opt.value"
            class="font-size-option"
            :class="{ active: settingsStore.fontSize === opt.value }"
          >
            <input
              type="radio"
              name="font-size"
              :value="opt.value"
              :checked="settingsStore.fontSize === opt.value"
              @change="settingsStore.setFontSize(opt.value)"
            />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </section>

      <section class="settings-section">
        <h4 class="settings-title">Theme</h4>
        <div class="theme-list">
          <label
            v-for="theme in settingsStore.themes"
            :key="theme.id"
            class="theme-option"
            :class="{ active: settingsStore.themeId === theme.id }"
          >
            <input
              type="radio"
              name="theme"
              :value="theme.id"
              :checked="settingsStore.themeId === theme.id"
              @change="settingsStore.setTheme(theme.id)"
            />
            <div class="theme-info">
              <div class="theme-header">
                <span class="theme-name">{{ theme.label }}</span>
                <span class="theme-desc">{{ theme.description }}</span>
              </div>
              <div class="theme-swatches">
                <span
                  v-for="(meta, key) in theme.categories"
                  :key="key"
                  class="theme-swatch"
                  :style="{ background: meta.color }"
                  :title="meta.label"
                />
              </div>
            </div>
          </label>
        </div>
      </section>

      <section class="settings-section">
        <h4 class="settings-title">Keyboard shortcuts</h4>
        <table class="shortcuts-table">
          <tbody>
            <tr v-for="s in shortcuts" :key="s.keys">
              <td class="shortcut-keys">
                <kbd>{{ s.keys }}</kbd>
              </td>
              <td class="shortcut-desc">{{ s.desc }}</td>
            </tr>
          </tbody>
        </table>
      </section>
    </div>
  </div>
</template>

<style scoped>
.settings-root {
  position: relative;
  display: inline-block;
}

.header-btn {
  width: 120px;
  flex: 0 0 120px;
  min-height: 36px;
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 16px;
  margin: 0;
  line-height: 18px;
  font-family: inherit;
  appearance: none;
  background: var(--color-overlay-white-15);
  border: 1px solid var(--color-overlay-white-20);
  color: var(--app-header-text);
  border-radius: 6px;
  cursor: pointer;
  font-size: calc(13px * var(--app-font-scale));
  transition: background 0.2s;
}

.header-btn:hover {
  background: var(--color-overlay-white-25);
}

.header-btn.active {
  background: var(--color-overlay-white-30);
}

.settings-dropdown {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  width: 320px;
  max-height: 70vh;
  overflow-y: auto;
  background: var(--color-bg-card);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  box-shadow: 0 6px 20px var(--color-shadow-strong);
  z-index: 1000;
  padding: 12px;
}

.settings-section {
  padding: 8px 0;
  border-bottom: 1px solid var(--color-border-lighter);
}

.settings-section:last-child {
  border-bottom: none;
}

.settings-title {
  margin: 0 0 8px 0;
  font-size: calc(11px * var(--app-font-scale));
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-text-secondary);
}

.font-size-group {
  display: flex;
  gap: 6px;
}

.font-size-option {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 6px 8px;
  border: 1px solid var(--border, #ddd);
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(12px * var(--app-font-scale));
  user-select: none;
}

.font-size-option:hover {
  background: var(--surface-alt, #f5f5f5);
}

.font-size-option.active {
  background: var(--color-primary);
  color: var(--app-header-text);
  border-color: var(--color-primary);
}

.font-size-option input {
  display: none;
}

.theme-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.theme-option {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  user-select: none;
}

.theme-option:hover {
  background: var(--color-bg-soft);
}

.theme-option.active {
  border-color: var(--color-primary);
  background: var(--color-primary-bg);
}

.theme-option input {
  margin-top: 4px;
}

.theme-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.theme-header {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.theme-name {
  font-size: calc(12px * var(--app-font-scale));
  font-weight: 600;
  color: var(--color-text-primary);
}

.theme-desc {
  font-size: calc(10px * var(--app-font-scale));
  color: var(--color-text-secondary);
}

.theme-swatches {
  display: flex;
  gap: 3px;
}

.theme-swatch {
  width: 18px;
  height: 18px;
  border-radius: 3px;
  border: 1px solid var(--color-border);
}

.shortcuts-table {
  width: 100%;
  border-collapse: collapse;
  font-size: calc(12px * var(--app-font-scale));
}

.shortcuts-table td {
  padding: 4px 0;
}

.shortcut-keys {
  width: 60px;
}

kbd {
  display: inline-block;
  padding: 2px 6px;
  background: var(--surface-alt, #f5f5f5);
  border: 1px solid var(--border, #ddd);
  border-radius: 3px;
  font-family: var(--mono);
  font-size: calc(11px * var(--app-font-scale));
  font-weight: 600;
  color: var(--color-text-primary);
}

.shortcut-desc {
  color: var(--color-text-secondary);
}
</style>
