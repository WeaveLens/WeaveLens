<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import cytoscape, { type Core, type EventObject } from 'cytoscape'
import { useGraphStore } from '../stores/graph'
import { getCategoryColor } from '../config/categories'

const graphStore = useGraphStore()
const containerRef = ref<HTMLDivElement | null>(null)
let cy: Core | null = null

const elements = computed(() => {
  const nodes = graphStore.filteredNodes.map(node => ({
    data: {
      id: node.id,
      label: node.name,
      category: node.category,
      color: getCategoryColor(node.category),
      type: node.type,
      region: node.region,
    },
  }))
  const edges = graphStore.filteredEdges.map(edge => ({
    data: {
      id: edge.id,
      source: edge.sourceId,
      target: edge.targetId,
      label: edge.type,
    },
  }))
  return [...nodes, ...edges]
})

function initCy() {
  if (!containerRef.value) return
  cy = cytoscape({
    container: containerRef.value,
    elements: elements.value,
    style: [
      {
        selector: 'node',
        style: {
          label: 'data(label)',
          'background-color': 'data(color)',
          width: 44,
          height: 44,
          'font-size': 11,
          'text-valign': 'bottom',
          'text-margin-y': 8,
          color: '#333',
          'border-width': 2,
          'border-color': '#fff',
        },
      },
      {
        selector: 'edge',
        style: {
          label: 'data(label)',
          width: 2,
          'line-color': '#999',
          'target-arrow-color': '#999',
          'target-arrow-shape': 'triangle',
          'curve-style': 'bezier',
          'font-size': 9,
          color: '#666',
          'text-rotation': 'autorotate',
          'text-margin-y': -12,
        },
      },
      {
        selector: 'node:selected',
        style: {
          'border-width': 4,
          'border-color': '#1976d2',
        },
      },
      {
        selector: 'node.highlighted',
        style: {
          'border-width': 3,
          'border-color': '#4CAF50',
        },
      },
    ],
    layout: {
      name: 'cose',
      animate: false,
      padding: 40,
      nodeRepulsion: () => 8000,
      idealEdgeLength: () => 120,
    },
    minZoom: 0.1,
    maxZoom: 4,
    wheelSensitivity: 0.3,
  })

  cy.on('tap', 'node', (e: EventObject) => {
    const node = e.target
    const resourceId = node.data('id')
    const nodeData = graphStore.nodes.find(n => n.id === resourceId)
    if (nodeData) {
      graphStore.selectResource(nodeData)
    }
  })

  cy.on('tap', (e: EventObject) => {
    if (e.target === cy) {
      graphStore.clearSelection()
    }
  })
}

function fitGraph() {
  if (cy) {
    cy.fit(undefined, 40)
  }
}

watch(elements, () => {
  if (cy) {
    cy.json({ elements: elements.value })
    cy.layout({
      name: 'cose',
      animate: false,
      padding: 40,
      nodeRepulsion: () => 8000,
      idealEdgeLength: () => 120,
    }).run()
  }
})

let resizeObserver: ResizeObserver | null = null

onMounted(() => {
  initCy()
  if (containerRef.value) {
    resizeObserver = new ResizeObserver(() => {
      if (cy) {
        cy.resize()
        cy.fit(undefined, 40)
      }
    })
    resizeObserver.observe(containerRef.value)
  }
})

onUnmounted(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
  if (cy) {
    cy.destroy()
  }
})

defineExpose({ fitGraph })
</script>

<template>
  <div class="topology-container">
    <div class="graph-controls">
      <button @click="fitGraph" class="control-btn" title="Fit to screen">
        Fit
      </button>
      <span class="node-count" v-if="graphStore.nodes.length">
        {{ graphStore.nodes.length }} nodes
      </span>
    </div>
    <div ref="containerRef" class="topology-view" />
    <div v-if="!graphStore.nodes.length" class="empty-state">
      <div class="empty-icon">📊</div>
      <h3>No Data</h3>
      <p>Connect an AWS account and start a scan to visualize your infrastructure.</p>
    </div>
  </div>
</template>

<style scoped>
.topology-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
  min-height: 0;
  background: #fafafa;
}

.graph-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: white;
  border-bottom: 1px solid #e0e0e0;
}

.control-btn {
  padding: 6px 12px;
  background: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.control-btn:hover {
  background: #e0e0e0;
}

.node-count {
  font-size: 12px;
  color: #666;
}

.topology-view {
  flex: 1;
  min-height: 400px;
  background: #fafafa;
}

.empty-state {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  color: #666;
  pointer-events: none;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-state h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
  color: #333;
}

.empty-state p {
  margin: 0;
  font-size: 14px;
  max-width: 300px;
}
</style>
