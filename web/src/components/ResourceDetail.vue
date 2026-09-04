<script setup lang="ts">
import { computed } from 'vue'
import { useGraphStore } from '../stores/graph'
import { getCategoryColor } from '../config/categories'
import { theme } from '../config/theme'
import Icon from './Icon.vue'
import type { Relationship } from '../types'

const graphStore = useGraphStore()

const selectedResource = computed(() => graphStore.selectedResource)
const relationships = computed(() => graphStore.selectedRelationships)

const relatedNode = (relationship: Relationship) => {
  if (selectedResource.value?.id === relationship.sourceId) {
    return graphStore.nodes.find(n => n.id === relationship.targetId)
  }
  return graphStore.nodes.find(n => n.id === relationship.sourceId)
}

const categoryColor = computed(() => {
  if (!selectedResource.value) return theme.text.muted
  return getCategoryColor(selectedResource.value.category)
})
</script>

<template>
  <div class="resource-detail">
    <div v-if="selectedResource" class="detail-content">
      <div class="detail-header">
        <h3>Resource Detail</h3>
        <button @click="graphStore.clearSelection" class="close-btn"><Icon name="icon-close" size="16" /></button>
      </div>
      <div class="detail-card">
        <div class="resource-type-badge" :style="{ backgroundColor: categoryColor }">
          {{ selectedResource.type }}
        </div>
        <div class="detail-rows">
          <div class="detail-row">
            <span class="label">ID</span>
            <span class="value mono">{{ selectedResource.id }}</span>
          </div>
          <div class="detail-row">
            <span class="label">Name</span>
            <span class="value">{{ selectedResource.name }}</span>
          </div>
          <div class="detail-row">
            <span class="label">Type</span>
            <span class="value">{{ selectedResource.type }}</span>
          </div>
          <div class="detail-row">
            <span class="label">Category</span>
            <span class="value" :style="{ color: categoryColor }">
              {{ selectedResource.category }}
            </span>
          </div>
          <div class="detail-row">
            <span class="label">ARN</span>
            <span class="value mono" :title="selectedResource.arn">
              {{ selectedResource.arn || '—' }}
            </span>
          </div>
          <div class="detail-row">
            <span class="label">Region</span>
            <span class="value">{{ selectedResource.region || '—' }}</span>
          </div>
        </div>
      </div>

      <div v-if="Object.keys(selectedResource.tags).length" class="tags-section">
        <h4>Tags</h4>
        <div class="detail-card">
          <div class="detail-row" v-for="(value, key) in selectedResource.tags" :key="key">
            <span class="label">{{ key }}</span>
            <span class="value mono">{{ value }}</span>
          </div>
        </div>
      </div>

      <div v-if="Object.keys(selectedResource.metadata).length" class="metadata-section">
        <h4>Metadata</h4>
        <div class="detail-card">
          <div class="detail-row" v-for="(value, key) in selectedResource.metadata" :key="key">
            <span class="label">{{ key }}</span>
            <span class="value mono">{{ value }}</span>
          </div>
        </div>
      </div>

      <div v-if="relationships.length" class="relationships-section">
        <h4>Relationships ({{ relationships.length }})</h4>
        <ul class="relationships-list">
          <li v-for="rel in relationships" :key="rel.id" class="relationship-item">
            <span class="rel-type">{{ rel.type }}</span>
            <span class="rel-arrow">→</span>
            <span class="rel-node" :title="relatedNode(rel)?.id">
              {{ relatedNode(rel)?.name || 'Unknown' }}
            </span>
          </li>
        </ul>
      </div>
    </div>

    <div v-else class="empty-state">
      <div class="empty-icon"><Icon name="icon-search" size="48" /></div>
      <h3>No Resource Selected</h3>
      <p>Click on a resource in the graph to view its details.</p>
    </div>
  </div>
</template>

<style scoped>
.resource-detail {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.detail-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 16px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.detail-header h3 {
  margin: 0;
  font-size: calc(16px * var(--app-font-scale));
  font-weight: 600;
}

.close-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: var(--surface-alt, #f5f5f5);
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(14px * var(--app-font-scale));
  color: #666;
  background: var(--color-bg-soft);
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
}

.close-btn:hover {
  background: var(--color-border);
}

.detail-card {
  background: var(--color-white);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
}

.resource-type-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  color: white;
  font-size: calc(11px * var(--app-font-scale));
  color: var(--color-white);
  font-size: 11px;
  font-weight: 600;
  margin-bottom: 12px;
}

.detail-rows {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  font-size: calc(12px * var(--app-font-scale));
  gap: 8px;
}

.label {
  font-weight: 500;
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.value {
  color: var(--color-text-primary);
  text-align: right;
  word-break: break-word;
  max-width: 60%;
}

.mono {
  font-family: monospace;
  font-size: calc(11px * var(--app-font-scale));
}

.metadata-section h4,
.tags-section h4,
.relationships-section h4 {
  margin: 0 0 8px 0;
  font-size: calc(13px * var(--app-font-scale));
  font-weight: 600;
  color: var(--color-text-primary);
}

.relationships-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.relationship-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--color-white);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  margin-bottom: 6px;
  font-size: calc(12px * var(--app-font-scale));
}

.rel-type {
  font-weight: 500;
  color: var(--color-primary);
  background: var(--color-primary-bg);
  padding: 2px 8px;
  border-radius: 4px;
}

.rel-arrow {
  color: var(--color-text-muted);
}

.rel-node {
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px;
  text-align: center;
  color: var(--color-text-secondary);
}

.empty-icon {
  font-size: calc(48px * var(--app-font-scale));
  margin-bottom: 16px;
  display: flex;
  justify-content: center;
}

.empty-state h3 {
  margin: 0 0 8px 0;
  font-size: calc(16px * var(--app-font-scale));
  color: #333;
  font-size: 16px;
  color: var(--color-text-primary);
}

.empty-state p {
  margin: 0;
  font-size: calc(13px * var(--app-font-scale));
  max-width: 200px;
}
</style>
