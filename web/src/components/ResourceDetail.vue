<script setup lang="ts">
import { useGraphStore } from '../stores/graph'
import { getCategoryColor } from '../config/categories'
import type { Relationship } from '../types'

const graphStore = useGraphStore()

const relatedNode = (relationship: Relationship) => {
  if (graphStore.selectedResource?.id === relationship.sourceId) {
    return graphStore.nodes.find(n => n.id === relationship.targetId)
  }
  return graphStore.nodes.find(n => n.id === relationship.sourceId)
}
</script>

<template>
  <aside class="sidebar" :class="{ open: graphStore.selectedResource }">
    <div v-if="graphStore.selectedResource" class="resource-detail">
      <h3>Resource Detail</h3>
      <div class="detail-card">
        <div class="detail-row">
          <span class="label">ID:</span>
          <span class="value">{{ graphStore.selectedResource.id }}</span>
        </div>
        <div class="detail-row">
          <span class="label">Name:</span>
          <span class="value">{{ graphStore.selectedResource.name }}</span>
        </div>
        <div class="detail-row">
          <span class="label">Type:</span>
          <span class="value">{{ graphStore.selectedResource.type }}</span>
        </div>
        <div class="detail-row">
          <span class="label">Category:</span>
          <span class="value" :style="{ color: getCategoryColor(graphStore.selectedResource.category) }">
            {{ graphStore.selectedResource.category }}
          </span>
        </div>
        <div class="detail-row">
          <span class="label">ARN:</span>
          <span class="value mono">{{ graphStore.selectedResource.arn }}</span>
        </div>
        <div class="detail-row">
          <span class="label">Region:</span>
          <span class="value">{{ graphStore.selectedResource.region }}</span>
        </div>
      </div>

      <div v-if="graphStore.selectedRelationships.length" class="relationships">
        <h4>Relationships</h4>
        <ul>
          <li v-for="rel in graphStore.selectedRelationships" :key="rel.id">
            <span class="rel-type">{{ rel.type }}</span>
            <span class="rel-node">
              {{ relatedNode(rel)?.name ?? rel.sourceId }}
            </span>
          </li>
        </ul>
      </div>
    </div>
    <div v-else class="empty-state">
      Select a resource to view details
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 320px;
  background: white;
  border-left: 1px solid #e0e0e0;
  overflow-y: auto;
  padding: 16px;
}
.resource-detail h3 {
  margin-top: 0;
}
.detail-card {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  margin-bottom: 16px;
}
.detail-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
  font-size: 14px;
}
.detail-row .label {
  font-weight: 500;
}
.detail-row .value {
  text-align: right;
  max-width: 180px;
  word-break: break-all;
}
.detail-row .mono {
  font-family: monospace;
  font-size: 12px;
}
.relationships h4 {
  margin-top: 0;
}
.relationships ul {
  list-style: none;
  padding: 0;
  margin: 0;
}
.relationships li {
  padding: 6px 0;
  border-bottom: 1px solid #eee;
  font-size: 14px;
}
.rel-type {
  font-weight: 500;
  color: #666;
}
.rel-node {
  float: right;
}
.empty-state {
  color: #999;
  text-align: center;
  margin-top: 40px;
  font-size: 14px;
}
</style>
