<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
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
          width: 40,
          height: 40,
          'font-size': '10px',
          'text-valign': 'bottom',
          'text-margin-y': 8,
          color: '#333',
        },
      },
      {
        selector: 'edge',
        style: {
          label: 'data(label)',
          width: 2,
          'line-color': '#666',
          'target-arrow-color': '#666',
          'target-arrow-shape': 'triangle',
          'curve-style': 'bezier',
          'font-size': '8px',
          color: '#666',
          'text-rotation': 'autorotate',
          'text-margin-y': -10,
        },
      },
      {
        selector: 'node:selected',
        style: {
          'border-width': 3,
          'border-color': '#000',
        },
      },
    ],
    layout: {
      name: 'cose',
      animate: false,
      padding: 40,
    },
    minZoom: 0.2,
    maxZoom: 3,
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

watch(elements, () => {
  if (cy) {
    cy.json({ elements: elements.value })
    cy.layout({ name: 'cose', animate: false, padding: 40 }).run()
  }
})

onMounted(() => {
  initCy()
})
</script>

<template>
  <div ref="containerRef" class="topology-view" />
</template>

<style scoped>
.topology-view {
  width: 100%;
  height: 100%;
  min-height: 500px;
  background: #fafafa;
  border-radius: 8px;
}
</style>
