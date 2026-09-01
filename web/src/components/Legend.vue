<script setup lang="ts">
import { ref } from 'vue'
import { useGraphStore } from '../stores/graph'
import { CATEGORY_META } from '../config/categories'

const graphStore = useGraphStore()
const collapsed = ref(false)
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
          v-for="(meta, key) in CATEGORY_META"
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
  background: white;
  border-top: 1px solid #e0e0e0;
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
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.collapse-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 10px;
  color: #666;
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
  font-size: 12px;
  background: #f5f5f5;
  border: 2px solid transparent;
  transition: all 0.2s;
}

.legend-item:hover {
  background: #e8e8e8;
}

.legend-item.active {
  background: #e3f2fd;
  border-color: #1976d2;
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
  font-size: 11px;
  color: #666;
}

.clear-btn:hover {
  background: #f5f5f5;
}
</style>
