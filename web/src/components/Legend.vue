<script setup lang="ts">
import { useGraphStore } from '../stores/graph'
import { CATEGORY_META } from '../config/categories'

const graphStore = useGraphStore()
</script>

<template>
  <div class="legend">
    <h4>Legend</h4>
    <div class="legend-items">
      <div
        v-for="(meta, key) in CATEGORY_META"
        :key="key"
        class="legend-item"
        :class="{ active: graphStore.categoryFilter === key }"
        @click="graphStore.setCategoryFilter(graphStore.categoryFilter === key ? null : key)"
      >
        <span class="legend-color" :style="{ backgroundColor: meta.color }" />
        <span class="legend-label">{{ meta.label }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.legend {
  padding: 12px 16px;
  background: white;
  border-top: 1px solid #e0e0e0;
}
.legend h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
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
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  background: #f5f5f5;
}
.legend-item.active {
  background: #e3f2fd;
}
.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}
.legend-label {
  white-space: nowrap;
}
</style>
