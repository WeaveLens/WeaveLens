<script setup lang="ts">
import { ref, computed } from 'vue'
import { useGraphStore } from '../stores/graph'
import { useSettingsStore } from '../stores/settings'

const graphStore = useGraphStore()
const settingsStore = useSettingsStore()
const collapsed = ref(false)
const categories = computed(() => settingsStore.currentTheme().categories)
</script>

<template>
  <div class="legend" :class="{ collapsed }">
    <div class="legend-header" @click="collapsed = !collapsed">
      <h4>Legend</h4>
      <button class="collapse-btn">
        {{ collapsed ? '▶' : '▼' }}
      </button>
    </div>
    <div v-if="!collapsed" class="legend-content">
      <div class="legend-items">
        <div
          v-for="(meta, key) in categories"
          :key="key"
          class="legend-item"
          :class="{ active: graphStore.categoryFilter === key }"
          @click="graphStore.setCategoryFilter(graphStore.categoryFilter === key ? null : key)"
          :title="meta.label"
        >
          <span class="legend-color" :style="{ backgroundColor: meta.color }" />
          <span class="legend-label">{{ meta.label }}</span>
        </div>
      </div>
      <div v-if="graphStore.categoryFilter" class="legend-actions">
        <button @click="graphStore.setCategoryFilter(null)" class="clear-btn">
          Clear filter
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.legend {
  background: var(--surface, #fff);
  border-top: 1px solid var(--app-panel-border, #e0e0e0);
  max-height: 200px;
  overflow: hidden;
  transition: max-height 0.2s ease;
}

.legend.collapsed {
  max-height: 40px;
}

.legend-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  cursor: pointer;
  user-select: none;
}

.legend-header h4 {
  margin: 0;
  font-size: calc(13px * var(--app-font-scale));
  font-weight: 600;
  color: var(--text-h, #333);
}

.collapse-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: calc(10px * var(--app-font-scale));
  color: var(--text-muted, #666);
  padding: 4px;
}

.legend-content {
  padding: 0 16px 12px;
}

.legend-items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  font-size: calc(12px * var(--app-font-scale));
  background: var(--surface-alt, #f5f5f5);
  border: 2px solid transparent;
  transition: all 0.2s;
}

.legend-item:hover {
  background: var(--surface-alt, #e8e8e8);
}

.legend-item.active {
  background: var(--accent-bg, #e3f2fd);
  border-color: var(--accent, #1976d2);
}

.legend-color {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  flex-shrink: 0;
}

.legend-label {
  white-space: nowrap;
}

.legend-actions {
  margin-top: 8px;
  text-align: center;
}

.clear-btn {
  padding: 4px 12px;
  background: none;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(11px * var(--app-font-scale));
  color: var(--text-muted, #666);
}

.clear-btn:hover {
  background: var(--surface-alt, #f5f5f5);
}
</style>
