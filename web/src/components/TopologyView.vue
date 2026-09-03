<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import cytoscape, { type Core, type EventObject } from 'cytoscape'
import { useGraphStore } from '../stores/graph'
import { getCategoryColor } from '../config/categories'
import { tierOf, type LayoutMode } from '../stores/graph'

const graphStore = useGraphStore()
const containerRef = ref<HTMLDivElement | null>(null)
const regionFilterRef = ref<HTMLElement | null>(null)
const typeFilterRef = ref<HTMLElement | null>(null)
const tagFilterRef = ref<HTMLElement | null>(null)
const advancedFilterRef = ref<HTMLElement | null>(null)
const valueInputRef = ref<HTMLElement | null>(null)
const valueDropdownPos = ref({ top: 0, left: 0, width: 0 })
const regionDropdownOpen = ref(false)
const typeDropdownOpen = ref(false)
const tagDropdownOpen = ref(false)
const advancedDropdownOpen = ref(false)
const showAddRule = ref(false)
const valueDropdownOpen = ref(false)
const newRule = ref({ field: '', value: '' })
const showAddTag = ref(false)
const tagValueInputRef = ref<HTMLElement | null>(null)
const tagValueDropdownPos = ref({ top: 0, left: 0, width: 0 })
const tagValueDropdownOpen = ref(false)
const newTag = ref({ key: '', value: '' })

function openTagValueDropdown() {
  if (!tagValueInputRef.value) return
  const rect = tagValueInputRef.value.getBoundingClientRect()
  tagValueDropdownPos.value = {
    top: rect.bottom + 2,
    left: rect.left,
    width: rect.width,
  }
  tagValueDropdownOpen.value = true
}

function addTagRule() {
  if (newTag.value.key && newTag.value.value) {
    const key = newTag.value.key
    const value = newTag.value.value
    const token = `${key}=${value}`
    if (!graphStore.tagFilter.includes(token)) {
      graphStore.setTagFilter([...graphStore.tagFilter, token])
    }
    newTag.value = { key: '', value: '' }
    tagValueDropdownOpen.value = false
  }
}

function removeTagRule(token: string) {
  graphStore.setTagFilter(graphStore.tagFilter.filter(t => t !== token))
}

function handleClickOutside(e: MouseEvent) {
  const target = e.target as Node
  if (regionFilterRef.value && !regionFilterRef.value.contains(target)) {
    regionDropdownOpen.value = false
  }
  if (typeFilterRef.value && !typeFilterRef.value.contains(target)) {
    typeDropdownOpen.value = false
  }
  if (tagFilterRef.value && !tagFilterRef.value.contains(target)) {
    if ((target as HTMLElement).closest?.('.tag-value-dropdown')) {
      return
    }
    tagDropdownOpen.value = false
    tagValueDropdownOpen.value = false
  }
  if (valueInputRef.value && !valueInputRef.value.contains(target)) {
    if (!(target as HTMLElement).closest?.('.value-dropdown')) {
      valueDropdownOpen.value = false
    }
  }
  if (advancedFilterRef.value && !advancedFilterRef.value.contains(target)) {
    if ((target as HTMLElement).closest?.('.value-dropdown')) {
      return
    }
    if (advancedDropdownOpen.value) {
      advancedDropdownOpen.value = false
      valueDropdownOpen.value = false
    }
  }
}

onMounted(() => {
  setTimeout(() => {
    document.addEventListener('mousedown', handleClickOutside)
  }, 0)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', handleClickOutside)
})

function addRule() {
  if (newRule.value.field && newRule.value.value) {
    graphStore.addFilterRule(newRule.value.field, newRule.value.value)
    newRule.value = { field: '', value: '' }
    valueDropdownOpen.value = false
    valueInputRef.value?.querySelector('input')?.blur()
  }
}

function openValueDropdown() {
  if (!valueInputRef.value) return
  const rect = valueInputRef.value.getBoundingClientRect()
  valueDropdownPos.value = {
    top: rect.bottom + 2,
    left: rect.left,
    width: rect.width,
  }
  valueDropdownOpen.value = true
}

let cy: Core | null = null

function toggleRegion(region: string) {
  const current = graphStore.regionFilter
  if (current.includes(region)) {
    graphStore.setRegionFilter(current.filter(r => r !== region))
  } else {
    graphStore.setRegionFilter([...current, region])
  }
}

function clearRegions() {
  graphStore.setRegionFilter([])
}

function toggleType(t: string) {
  const current = graphStore.typeFilter
  if (current.includes(t)) {
    graphStore.setTypeFilter(current.filter(x => x !== t))
  } else {
    graphStore.setTypeFilter([...current, t])
  }
}

function clearTypes() {
  graphStore.setTypeFilter([])
}

function clearTags() {
  graphStore.setTagFilter([])
}

const elements = computed(() => {
  const nodes = graphStore.filteredNodes.map(node => ({
    data: {
      id: node.id,
      label: node.name,
      category: node.category,
      color: getCategoryColor(node.category),
      type: node.type,
      region: node.region,
      tier: tierOf(node.category),
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
    layout: buildLayoutConfig(graphStore.layoutMode) as cytoscape.LayoutOptions,
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

  cy.on('dragfree', 'node', (e: EventObject) => {
    const node = e.target
    const pos = node.position()
    graphStore.pinPosition(node.id(), pos.x, pos.y)
  })
}

function fitGraph() {
  if (cy) {
    cy.fit(undefined, 40)
  }
}

function buildLayoutConfig(mode: LayoutMode) {
  if (mode === 'none') {
    return { name: 'preset', fit: true, padding: 40 }
  }
  if (mode === 'tiers') {
    return {
      name: 'breadthfirst',
      directed: false,
      fit: true,
      padding: 40,
      spacingFactor: 1.4,
      avoidOverlap: true,
      transform: (_node: unknown, pos: { x: number; y: number }) => ({ x: pos.y, y: pos.x }),
    }
  }
  if (mode === 'concentric') {
    return {
      name: 'concentric',
      fit: true,
      padding: 40,
      minNodeSpacing: 40,
      avoidOverlap: true,
      concentric: (node: cytoscape.NodeSingular) => -tierOf(String(node.data('category'))),
      levelWidth: () => 1,
    }
  }
  return {
    name: 'cose',
    animate: false,
    padding: 40,
    nodeRepulsion: () => 8000,
    idealEdgeLength: () => 120,
    randomize: false,
    fit: true,
  }
}

function applyPinnedPositions() {
  const instance = cy
  if (!instance) return
  const pins = graphStore.pinnedPositions
  if (!pins || Object.keys(pins).length === 0) return
  instance.batch(() => {
    instance.nodes().forEach(n => {
      const p = pins[n.id()]
      if (p) {
        n.position(p)
        n.lock()
      } else {
        n.unlock()
      }
    })
  })
}

function runLayout(mode: LayoutMode, opts: { lock: boolean } = { lock: false }) {
  if (!cy) return
  if (mode === 'none' || opts.lock) {
    cy.layout({ name: 'preset', fit: true, padding: 40 }).run()
  } else {
    cy.layout(buildLayoutConfig(mode) as cytoscape.LayoutOptions).run()
  }
  if (opts.lock) applyPinnedPositions()
}

watch(elements, () => {
  if (!cy) return
  cy.json({ elements: elements.value })
  if (graphStore.layoutLocked) {
    applyPinnedPositions()
    return
  }
  runLayout(graphStore.layoutMode, { lock: false })
})

watch(
  () => [graphStore.layoutMode, graphStore.layoutLocked],
  ([mode, locked]) => {
    if (!cy) return
    runLayout(mode as LayoutMode, { lock: locked as boolean })
  }
)

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
      <div class="graph-controls-group">
        <label class="layout-label" for="layout-mode-select">Layout</label>
        <select
          id="layout-mode-select"
          class="control-select"
          :value="graphStore.layoutMode"
          @change="(e) => graphStore.setLayoutMode((e.target as HTMLSelectElement).value as LayoutMode)"
        >
          <option value="tiers">Tiers (by category)</option>
          <option value="concentric">Concentric (by category)</option>
          <option value="force">Force (cose)</option>
          <option value="none">None (keep positions)</option>
        </select>
        <button
          class="control-btn"
          :class="{ active: graphStore.layoutLocked }"
          :title="graphStore.layoutLocked ? 'Unlock layout (re-run auto layout on updates)' : 'Lock layout (keep positions on updates)'"
          @click="graphStore.setLayoutLocked(!graphStore.layoutLocked)"
        >
          {{ graphStore.layoutLocked ? '🔒 Locked' : '🔓 Unlock' }}
        </button>
      </div>
      <div class="graph-controls-group">
        <button @click="fitGraph" class="control-btn" title="Fit to screen">
          Fit
        </button>
        <button
          class="control-btn"
          title="Rerun current layout"
          @click="runLayout(graphStore.layoutMode, { lock: graphStore.layoutLocked }); graphStore.setLayoutLocked(false)"
        >
          Relayout
        </button>
        <span class="node-count" v-if="graphStore.nodes.length">
          {{ graphStore.nodes.length }} {{ (graphStore.nodes.length === 1 ? 'service' : 'services') }}
        </span>
      </div>
    </div>
    <div ref="containerRef" class="topology-view" />
    <div class="graph-overlay" v-if="graphStore.regions.length > 0 || graphStore.types.length > 0 || graphStore.availableTags.length > 0">
      <div class="filter-group" v-if="graphStore.regions.length > 0" ref="regionFilterRef">
        <button class="overlay-btn" @click="regionDropdownOpen = !regionDropdownOpen">
          🌍 Region ({{ graphStore.regionFilter.length || 'All' }})
        </button>
        <div v-if="regionDropdownOpen" class="filter-dropdown">
          <label v-for="region in graphStore.regions" :key="region" class="filter-option">
            <input
              type="checkbox"
              :checked="graphStore.regionFilter.includes(region)"
              @change="toggleRegion(region)"
            />
            {{ region }}
          </label>
          <button class="clear-btn" @click="clearRegions" v-if="graphStore.regionFilter.length">
            Clear
          </button>
        </div>
      </div>
      <div class="filter-group" v-if="graphStore.types.length > 0" ref="typeFilterRef">
        <button class="overlay-btn" @click="typeDropdownOpen = !typeDropdownOpen">
          📦 Type ({{ graphStore.typeFilter.length || 'All' }})
        </button>
        <div v-if="typeDropdownOpen" class="filter-dropdown">
          <label v-for="t in graphStore.types" :key="t" class="filter-option">
            <input
              type="checkbox"
              :checked="graphStore.typeFilter.includes(t)"
              @change="toggleType(t)"
            />
            {{ t }}
          </label>
          <button class="clear-btn" @click="clearTypes" v-if="graphStore.typeFilter.length">
            Clear
          </button>
        </div>
      </div>
      <div class="filter-group" v-if="graphStore.availableTags.length > 0" ref="tagFilterRef">
        <button class="overlay-btn" @click="tagDropdownOpen = !tagDropdownOpen">
          🏷️ Tags ({{ graphStore.tagFilter.length }})
        </button>
        <div v-if="tagDropdownOpen" class="filter-dropdown advanced-dropdown">
          <div class="filter-rule" v-for="token in graphStore.tagFilter" :key="token">
            <span class="rule-field">{{ token.split('=')[0] }}</span>
            <span class="rule-value">{{ token.split('=').slice(1).join('=') }}</span>
            <button class="rule-remove" @click="removeTagRule(token)">×</button>
          </div>
          <div class="add-rule" v-if="!showAddTag">
            <button class="add-btn" @click="showAddTag = true">+ Add Tag</button>
          </div>
          <div class="add-rule-form" v-else>
            <select v-model="newTag.key" class="rule-select" @change="newTag.value = ''; tagValueDropdownOpen = false">
              <option value="" disabled>Select tag key</option>
              <option v-for="k in graphStore.availableTagKeys" :key="k" :value="k">
                {{ k }}
              </option>
            </select>
            <div class="value-input-wrapper" ref="tagValueInputRef">
              <input
                v-model="newTag.value"
                class="rule-input"
                :placeholder="newTag.key ? 'Value (click for suggestions)' : 'Select key first'"
                :disabled="!newTag.key"
                @focus="openTagValueDropdown()"
                @click="openTagValueDropdown()"
                @input="tagValueDropdownOpen = false"
              />
            </div>
            <Teleport to="body">
              <div
                v-if="tagValueDropdownOpen && newTag.key && graphStore.getTagValues(newTag.key).length"
                class="value-dropdown tag-value-dropdown"
                :style="{
                  top: tagValueDropdownPos.top + 'px',
                  left: tagValueDropdownPos.left + 'px',
                  width: tagValueDropdownPos.width + 'px',
                }"
              >
                <div
                  v-for="val in graphStore.getTagValues(newTag.key)"
                  :key="val"
                  class="value-option"
                  @mousedown.prevent="newTag.value = val; tagValueDropdownOpen = false"
                >
                  {{ val }}
                </div>
              </div>
            </Teleport>
            <div class="add-rule-actions">
              <button class="add-btn" @click="addTagRule" :disabled="!newTag.key || !newTag.value">Add</button>
              <button class="cancel-btn" @click="showAddTag = false; newTag = { key: '', value: '' }; tagValueDropdownOpen = false">Cancel</button>
            </div>
          </div>
          <button class="clear-btn" @click="clearTags" v-if="graphStore.tagFilter.length">
            Clear All
          </button>
        </div>
      </div>
      <div class="filter-group" ref="advancedFilterRef">
        <button class="overlay-btn" @click="advancedDropdownOpen = !advancedDropdownOpen">
          🔍 Advanced ({{ graphStore.filterRules.length }})
        </button>
        <div v-if="advancedDropdownOpen" class="filter-dropdown advanced-dropdown">
          <div class="filter-rule" v-for="rule in graphStore.filterRules" :key="rule.id">
            <span class="rule-field">{{ rule.field }}</span>
            <span class="rule-value">{{ rule.value }}</span>
            <button class="rule-remove" @click="graphStore.removeFilterRule(rule.id)">×</button>
          </div>
          <div class="add-rule" v-if="!showAddRule">
            <button class="add-btn" @click="showAddRule = true">+ Add Filter</button>
          </div>
          <div class="add-rule-form" v-else>
            <select v-model="newRule.field" class="rule-select" @change="newRule.value = ''; valueDropdownOpen = false">
              <option value="" disabled>Select field</option>
              <option v-for="field in graphStore.availableFields" :key="field" :value="field">
                {{ field }}
              </option>
            </select>
            <div class="value-input-wrapper" ref="valueInputRef">
              <input
                v-model="newRule.value"
                class="rule-input"
                :placeholder="newRule.field ? 'Value (click for suggestions)' : 'Select field first'"
                :disabled="!newRule.field"
                @focus="openValueDropdown()"
                @click="openValueDropdown()"
                @input="valueDropdownOpen = false"
              />
            </div>
            <Teleport to="body">
              <div
                v-if="valueDropdownOpen && graphStore.getAvailableValues(newRule.field).length"
                class="value-dropdown"
                :style="{
                  top: valueDropdownPos.top + 'px',
                  left: valueDropdownPos.left + 'px',
                  width: valueDropdownPos.width + 'px',
                }"
              >
                <div
                  v-for="val in graphStore.getAvailableValues(newRule.field)"
                  :key="val"
                  class="value-option"
                  @mousedown.prevent="newRule.value = val; valueDropdownOpen = false"
                >
                  {{ val }}
                </div>
              </div>
            </Teleport>
            <div class="add-rule-actions">
              <button class="add-btn" @click="addRule" :disabled="!newRule.field || !newRule.value">Add</button>
              <button class="cancel-btn" @click="showAddRule = false; newRule = { field: '', value: '' }; valueDropdownOpen = false">Cancel</button>
            </div>
          </div>
          <button class="clear-btn" @click="graphStore.clearFilterRules()" v-if="graphStore.filterRules.length">
            Clear All
          </button>
        </div>
      </div>
    </div>
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
  flex-shrink: 0;
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

.control-btn.active {
  background: #1976d2;
  color: white;
  border-color: #1976d2;
}

.control-btn.active:hover {
  background: #1565c0;
}

.graph-controls-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.layout-label {
  font-size: 11px;
  color: #666;
  font-weight: 600;
}

.control-select {
  padding: 5px 8px;
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}

.control-select:hover {
  border-color: #1976d2;
}

.node-count {
  font-size: 12px;
  color: #666;
}

.topology-view {
  flex: 1;
  min-height: 0;
  width: 100%;
  background: #fafafa;
}

.graph-overlay {
  position: absolute;
  top: 60px;
  left: 16px;
  z-index: 10;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-group {
  position: relative;
}

.overlay-btn {
  padding: 6px 12px;
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.overlay-btn:hover {
  background: #f5f5f5;
}

.filter-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  min-width: 150px;
  max-height: 200px;
  overflow-y: auto;
  z-index: 100;
  padding: 8px;
  margin-top: 4px;
}

.filter-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  cursor: pointer;
  font-size: 12px;
}

.filter-option:hover {
  background: #f5f5f5;
}

.match-toggle {
  display: flex;
  gap: 4px;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid #eee;
}

.toggle-btn {
  flex: 1;
  padding: 4px 8px;
  background: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
  font-weight: 600;
}

.toggle-btn.active {
  background: #1976d2;
  color: white;
  border-color: #1976d2;
}

.toggle-btn:hover:not(.active) {
  background: #e0e0e0;
}

.advanced-dropdown {
  min-width: 220px;
}

.filter-rule {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  background: #f5f5f5;
  border-radius: 4px;
  margin-bottom: 4px;
  font-size: 11px;
}

.rule-field {
  font-weight: 600;
  color: #1976d2;
}

.rule-value {
  flex: 1;
  color: #333;
}

.rule-remove {
  width: 20px;
  height: 20px;
  background: #ff5252;
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
}

.rule-remove:hover {
  background: #d32f2f;
}

.add-rule {
  margin-top: 8px;
}

.add-rule-form {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 8px;
}

.add-rule-actions {
  display: flex;
  gap: 6px;
}

.rule-select,
.rule-input {
  padding: 6px 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 11px;
}

.add-btn {
  padding: 6px 12px;
  background: #1976d2;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
}

.add-btn:hover {
  background: #1565c0;
}

.cancel-btn {
  padding: 6px 12px;
  background: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
}

.cancel-btn:hover {
  background: #e0e0e0;
}

.rule-input:disabled {
  background: #f5f5f5;
  color: #999;
  cursor: not-allowed;
}

.add-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.value-input-wrapper {
  position: relative;
}

.value-dropdown {
  position: fixed;
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  max-height: 150px;
  overflow-y: auto;
  z-index: 9999;
}

.value-option {
  padding: 6px 8px;
  cursor: pointer;
  font-size: 11px;
}

.value-option:hover {
  background: #f0f0f0;
}

.tag-section {
  margin-bottom: 8px;
}

.tag-key {
  font-size: 10px;
  font-weight: 600;
  color: #1976d2;
  padding: 4px 8px;
  background: #f5f5f5;
  border-radius: 4px;
  margin-bottom: 4px;
}

.filter-help {
  font-size: 10px;
  color: #666;
  padding: 4px 8px;
  background: #f5f5f5;
  border-radius: 4px;
  margin-top: 4px;
}

.clear-btn {
  width: 100%;
  margin-top: 8px;
  padding: 4px;
  background: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
}

.clear-btn:hover {
  background: #e0e0e0;
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
