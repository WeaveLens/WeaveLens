<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import cytoscape, { type Core, type EventObject } from 'cytoscape'
import { useGraphStore } from '../stores/graph'
import { useScanStore } from '../stores/scan'
import { useSettingsStore } from '../stores/settings'
import { getRegions, setScanLayout } from '../api/client'
import { tierOf, type LayoutMode } from '../stores/graph'
import { theme } from '../config/theme'
import Icon from './Icon.vue'

const graphStore = useGraphStore()
const scanStore = useScanStore()
const settingsStore = useSettingsStore()
const containerRef = ref<HTMLDivElement | null>(null)
const regionFilterRef = ref<HTMLElement | null>(null)
const typeFilterRef = ref<HTMLElement | null>(null)
const tagFilterRef = ref<HTMLElement | null>(null)
const advancedFilterRef = ref<HTMLElement | null>(null)
const valueInputRef = ref<HTMLElement | null>(null)
const valueDropdownRef = ref<HTMLElement | null>(null)
const valueDropdownPos = ref({ top: 0, left: 0, width: 0 })
const regionDropdownOpen = ref(false)
const typeDropdownOpen = ref(false)
const tagDropdownOpen = ref(false)
const advancedDropdownOpen = ref(false)
const showAddRule = ref(false)
const showAddTag = ref(false)
const valueDropdownOpen = ref(false)
const newRule = ref({ field: '', value: '' })
const newTag = ref({ key: '', value: '' })
const regionLabels = ref<Record<string, string>>({})

function handleClickOutside(e: MouseEvent) {
  const target = e.target as Node
  const insideValueDropdown = valueDropdownRef.value?.contains(target) ?? false
  if (regionFilterRef.value && !regionFilterRef.value.contains(target)) {
    regionDropdownOpen.value = false
  }
  if (typeFilterRef.value && !typeFilterRef.value.contains(target)) {
    typeDropdownOpen.value = false
  }
  if (tagFilterRef.value && !tagFilterRef.value.contains(target)) {
    tagDropdownOpen.value = false
  }
  if (valueInputRef.value && !valueInputRef.value.contains(target) && !insideValueDropdown) {
    valueDropdownOpen.value = false
  }
  if (advancedFilterRef.value && !advancedFilterRef.value.contains(target) && !insideValueDropdown) {
    if (advancedDropdownOpen.value) {
      advancedDropdownOpen.value = false
      valueDropdownOpen.value = false
    }
  }
}

onMounted(async () => {
  try {
    const regions = await getRegions(true)
    regionLabels.value = Object.fromEntries(regions.map(region => [region.value, region.label]))
  } catch {
    // Region codes remain available from the graph when metadata cannot be loaded.
  }
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
    showAddRule.value = false
    valueDropdownOpen.value = false
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

function getRegionLabel(region: string): string {
  return regionLabels.value[region] || region
}

const regionFilterLabel = computed(() => {
  if (graphStore.regionFilter.length === 0) return 'All'
  if (graphStore.regionFilter.length > 1) return `${graphStore.regionFilter.length} selected`

  const region = graphStore.regionFilter[0]
  const label = getRegionLabel(region)
  return label === region ? region : `${label} (${region})`
})

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

function addTag() {
  if (!newTag.value.key || !newTag.value.value) return

  const tag = `${newTag.value.key}=${newTag.value.value}`
  if (!graphStore.tagFilter.includes(tag)) {
    graphStore.setTagFilter([...graphStore.tagFilter, tag])
  }
  newTag.value = { key: '', value: '' }
  showAddTag.value = false
}

function removeTag(tag: string) {
  graphStore.setTagFilter(graphStore.tagFilter.filter(item => item !== tag))
}

function cancelTag() {
  newTag.value = { key: '', value: '' }
  showAddTag.value = false
}

function tagKey(tag: string): string {
  const separator = tag.indexOf('=')
  return separator >= 0 ? tag.slice(0, separator) : tag
}

function tagValue(tag: string): string {
  const separator = tag.indexOf('=')
  return separator >= 0 ? tag.slice(separator + 1) : ''
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
      color: settingsStore.getResourceColorForType(node.type, node.category),
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
          color: theme.text.primary,
          'border-width': 2,
          'border-color': theme.colorWhite,
        },
      },
      {
        selector: 'edge',
        style: {
          label: 'data(label)',
          width: 2,
          'line-color': theme.text.secondary,
          'target-arrow-color': theme.text.secondary,
          'target-arrow-shape': 'triangle',
          'curve-style': 'bezier',
          'font-size': 9,
          color: theme.text.secondary,
          'text-rotation': 'autorotate',
          'text-margin-y': -12,
        },
      },
      {
        selector: 'node:selected',
        style: {
          'border-width': 4,
          'border-color': theme.primary.DEFAULT,
        },
      },
      {
        selector: 'node.highlighted',
        style: {
          'border-width': 3,
          'border-color': settingsStore.currentTheme().system.status.success,
        },
      },
      {
        selector: '.focus-node',
        style: {
          'border-width': 3,
          'border-color': settingsStore.currentTheme().system.accent,
          opacity: 1,
        },
      },
      {
        selector: '.focus-edge',
        style: {
          width: 4,
          'line-color': settingsStore.currentTheme().system.accent,
          'target-arrow-color': settingsStore.currentTheme().system.accent,
          opacity: 1,
        },
      },
      {
        selector: '.dimmed',
        style: {
          opacity: 0.18,
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
      highlightNeighborhood(node)
    }
  })

  cy.on('tap', (e: EventObject) => {
    if (e.target === cy) {
      graphStore.clearSelection()
      clearNeighborhoodHighlight()
    }
  })

  cy.on('dragfree', 'node', (e: EventObject) => {
    const node = e.target
    const pos = node.position()
    graphStore.pinPosition(node.id(), pos.x, pos.y)
    const scanId = graphStore.currentScanId
    if (scanId) graphStore.saveLayout(scanId)
  })

  cy.on('layoutstop', () => {
    const scanId = graphStore.currentScanId
    if (scanId && cy) {
      if (cy.nodes().length > 0) {
        const positions: Record<string, { x: number; y: number }> = {}
        cy.nodes().forEach(node => {
          const pos = node.position()
          positions[node.id()] = { x: pos.x, y: pos.y }
        })
        graphStore.setPinnedPositions(positions)
        graphStore.saveLayout(scanId)
        if (graphStore.layoutLocked) {
          lockAllNodes()
        }
      }
      graphStore.saveViewport(scanId, cy.zoom(), { x: cy.pan().x, y: cy.pan().y })
      graphStore.savePositionsToBackend(scanId).catch(() => {})
    }
  })

  cy.on('viewport', () => {
    const scanId = graphStore.currentScanId
    if (scanId && cy) {
      graphStore.saveViewport(scanId, cy.zoom(), { x: cy.pan().x, y: cy.pan().y })
    }
  })
}

function highlightNeighborhood(node: cytoscape.NodeSingular) {
  if (!cy) return
  cy.elements().removeClass('focus-node focus-edge dimmed')
  node.addClass('focus-node')
  node.neighborhood('node').addClass('focus-node')
  node.connectedEdges().addClass('focus-edge')
  cy.elements().not(node.union(node.neighborhood())).addClass('dimmed')
}

function clearNeighborhoodHighlight() {
  cy?.elements().removeClass('focus-node focus-edge dimmed')
}

function updateCyTheme() {
  if (!cy) return
  const colors = settingsStore.currentTheme().system
  cy.style()
    .selector('node')
    .style({ color: colors.textHeading, 'border-color': colors.headerText })
    .selector('edge')
    .style({
      color: colors.text,
      'line-color': colors.text,
      'target-arrow-color': colors.text,
    })
    .selector('node:selected')
    .style({ 'border-color': colors.accent })
    .selector('.focus-node')
    .style({ 'border-color': colors.accent })
    .selector('.focus-edge')
    .style({ 'line-color': colors.accent, 'target-arrow-color': colors.accent })
    .update()
}

function fitGraph() {
  if (cy) {
    cy.fit(undefined, 40)
  }
}

function zoomGraph(factor: number) {
  if (!cy) return
  const center = { x: cy.width() / 2, y: cy.height() / 2 }
  cy.zoom({ level: Math.max(0.1, Math.min(4, cy.zoom() * factor)), renderedPosition: center })
}

function resetZoom() {
  if (!cy) return
  cy.zoom(1)
  cy.center()
}

function onToggleLock() {
  const wasLocked = graphStore.layoutLocked
  const nextLocked = !wasLocked
  graphStore.setLayoutLocked(nextLocked)
  if (wasLocked) {
    releaseAllNodes()
  } else {
    lockAllNodes()
  }
  const scanId = scanStore.currentScan?.id
  if (scanId) {
    scanStore.toggleLocked(scanId, nextLocked).catch(() => {
      graphStore.setLayoutLocked(wasLocked)
      if (wasLocked) {
        lockAllNodes()
      } else {
        releaseAllNodes()
      }
    })
    graphStore.savePositionsToBackend(scanId).catch(() => {})
  }
}

function onRelayout() {
  if (graphStore.layoutLocked) {
    releaseAllNodes()
  } else {
    graphStore.setPinnedPositions({})
    runLayout(graphStore.layoutMode, { lock: false, fit: true })
  }
}

function buildLayoutConfig(mode: LayoutMode) {
  if (mode === 'none') {
    return { name: 'preset', padding: 40 }
  }
  if (mode === 'tiers' || mode === 'tiers-vertical') {
    return {
      name: 'breadthfirst',
      directed: false,
      padding: 40,
      spacingFactor: 1,
      idealEdgeLength: () => 100,
      avoidOverlap: true,
      ...(mode === 'tiers' ? {
        transform: (_node: unknown, pos: { x: number; y: number }) => ({ x: pos.y, y: pos.x }),
      } : {}),
    }
  }
  if (mode === 'concentric') {
    return {
      name: 'concentric',
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

function applySavedPositions() {
  const instance = cy
  if (!instance) return
  const positions = graphStore.pinnedPositions
  if (!positions || Object.keys(positions).length === 0) return
  instance.batch(() => {
    instance.nodes().forEach(node => {
      const position = positions[node.id()]
      if (!position) return
      node.position(position)
      if (graphStore.layoutLocked) {
        node.lock()
      } else {
        node.unlock()
      }
    })
  })
}

function releaseAllNodes() {
  if (!cy) return
  const instance = cy
  instance.batch(() => {
    instance.nodes().forEach((n) => {
      n.unlock()
      return undefined
    })
  })
}

function lockAllNodes() {
  if (!cy) return
  const instance = cy
  const positions: Record<string, { x: number; y: number }> = {}
  instance.batch(() => {
    instance.nodes().forEach((n) => {
      const p = n.position()
      positions[n.id()] = { x: p.x, y: p.y }
      n.lock()
    })
  })
  graphStore.setPinnedPositions(positions)
}

function runLayout(mode: LayoutMode, opts: { lock: boolean; fit?: boolean } = { lock: false }) {
  if (!cy) return
  if (mode === 'none' || opts.lock) {
    cy.layout({ name: 'preset', padding: 40 }).run()
  } else {
    cy.layout(buildLayoutConfig(mode) as cytoscape.LayoutOptions).run()
  }
  if (opts.fit) {
    cy.fit(undefined, 40)
  }
  if (opts.lock) applyPinnedPositions()
}

watch(elements, () => {
  if (!cy) return
  cy.json({ elements: elements.value })
  const scanId = graphStore.currentScanId
  const saved = scanId ? graphStore.getSavedViewport(scanId) : null
  if (graphStore.layoutLocked) {
    if (Object.keys(graphStore.pinnedPositions).length > 0) {
      applyPinnedPositions()
    } else {
      releaseAllNodes()
      runLayout(graphStore.layoutMode, { lock: false, fit: !saved })
    }
    if (saved) {
      cy.zoom(saved.zoom)
      cy.pan(saved.pan)
    } else if (scanId) {
      cy.fit(undefined, 40)
    }
    return
  }
  if (graphStore.pinnedPositions && Object.keys(graphStore.pinnedPositions).length > 0) {
    applySavedPositions()
    if (saved) {
      cy.zoom(saved.zoom)
      cy.pan(saved.pan)
    } else if (scanId) {
      cy.fit(undefined, 40)
    }
    return
  }
  releaseAllNodes()
  runLayout(graphStore.layoutMode, { lock: false, fit: !saved })
})

watch(
  () => [graphStore.layoutMode, graphStore.layoutLocked],
  ([mode, locked], [prevMode]) => {
    if (!cy) return
    if (mode !== prevMode) {
      if (locked) {
        applyPinnedPositions()
        return
      }
      releaseAllNodes()
      runLayout(mode as LayoutMode, { lock: false })
    }
  }
)

watch(() => settingsStore.themeId, () => {
  updateCyTheme()
})

let resizeObserver: ResizeObserver | null = null

function onKeydown(e: KeyboardEvent) {
  const target = e.target as HTMLElement | null
  if (target && (target.tagName === 'INPUT' || target.tagName === 'SELECT' || target.tagName === 'TEXTAREA' || target.isContentEditable)) {
    return
  }
  const key = e.key.toLowerCase()
  if (key === 'l') {
    e.preventDefault()
    onToggleLock()
  } else if (key === 'p') {
    e.preventDefault()
    const scanId = scanStore.currentScan?.id
    if (scanId) {
      const scan = scanStore.scans.find(s => s.id === scanId)
      const targetPin = !(scan?.pinned ?? false)
      scanStore.togglePin(scanId, targetPin).catch(() => {})
    }
  } else if (key === 'f') {
    e.preventDefault()
    fitGraph()
  } else if (key === 'r') {
    if (graphStore.layoutLocked) return
    e.preventDefault()
    onRelayout()
  } else if (e.key === '+' || e.key === '=') {
    e.preventDefault()
    zoomGraph(1.2)
  } else if (e.key === '-') {
    e.preventDefault()
    zoomGraph(1 / 1.2)
  } else if (key === '0') {
    e.preventDefault()
    resetZoom()
  }
}

onMounted(() => {
  initCy()
  if (containerRef.value) {
    resizeObserver = new ResizeObserver(() => {
      if (cy) {
        cy.resize()
      }
    })
    resizeObserver.observe(containerRef.value)
  }
  window.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
  window.removeEventListener('keydown', onKeydown)
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
          :disabled="graphStore.layoutLocked"
          @change="(e) => {
            const mode = (e.target as HTMLSelectElement).value as LayoutMode
            graphStore.setLayoutMode(mode)
            const scanId = scanStore.currentScan?.id
            if (scanId) {
              setScanLayout(scanId, mode).catch(() => {})
              graphStore.savePositionsToBackend(scanId).catch(() => {})
            }
          }"
        >
          <option value="tiers">Tiers (by category)</option>
          <option value="tiers-vertical">Tiers (top to bottom)</option>
          <option value="concentric">Concentric (by category)</option>
          <option value="force">Force (cose)</option>
          <option value="none">None (keep positions)</option>
        </select>
        <button
          class="control-btn"
          :class="{ active: graphStore.layoutLocked }"
          :title="graphStore.layoutLocked ? 'Unlock (L)' : 'Lock layout (L)'"
          @click="onToggleLock"
        >
          {{ graphStore.layoutLocked ? '🔒 Locked' : '🔓 Unlock' }}
        </button>
      </div>
      <div class="graph-controls-group">
        <button @click="zoomGraph(1.2)" class="control-btn zoom-btn" title="Zoom in (+)">+</button>
        <button @click="zoomGraph(1 / 1.2)" class="control-btn zoom-btn" title="Zoom out (-)">-</button>
        <button @click="resetZoom" class="control-btn" title="Reset zoom to 100% and center (0)">Reset</button>
        <button @click="fitGraph" class="control-btn" title="Fit to screen (F)">
          Fit
        </button>
        <button
          class="control-btn"
          :disabled="graphStore.layoutLocked"
          :title="graphStore.layoutLocked ? 'Unlock first to rerun layout (L)' : 'Rerun current layout (R)'"
          @click="onRelayout"
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
          🌍 Region ({{ regionFilterLabel }})
        </button>
        <div v-if="regionDropdownOpen" class="filter-dropdown">
          <label v-for="region in graphStore.regions" :key="region" class="filter-option region-filter-option">
            <input
              type="checkbox"
              :checked="graphStore.regionFilter.includes(region)"
              @change="toggleRegion(region)"
            />
            <span class="region-option-text">
              <span>{{ getRegionLabel(region) }}</span>
              <span v-if="getRegionLabel(region) !== region" class="region-code">{{ region }}</span>
            </span>
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
          🏷️ Tags ({{ graphStore.tagFilter.length || 'All' }})
        </button>
        <div v-if="tagDropdownOpen" class="filter-dropdown advanced-dropdown">
          <div class="filter-rule" v-for="tag in graphStore.tagFilter" :key="tag">
            <span class="rule-field">{{ tagKey(tag) }}</span>
            <span class="rule-value">{{ tagValue(tag) }}</span>
            <button class="rule-remove" @click="removeTag(tag)">×</button>
          </div>
          <div class="add-rule" v-if="!showAddTag">
            <button class="add-btn" @click="showAddTag = true">+ Add Filter</button>
          </div>
          <div class="add-rule-form" v-else>
            <select v-model="newTag.key" class="rule-select" @change="newTag.value = ''">
              <option value="" disabled>Select tag key</option>
              <option v-for="key in graphStore.availableTagKeys" :key="key" :value="key">
                {{ key }}
              </option>
            </select>
            <select v-model="newTag.value" class="rule-select" :disabled="!newTag.key">
              <option value="" disabled>Select tag value</option>
              <option v-for="value in graphStore.getTagValues(newTag.key)" :key="value" :value="value">
                {{ value }}
              </option>
            </select>
            <button class="add-btn" @click="addTag" :disabled="!newTag.key || !newTag.value">Add</button>
            <button class="cancel-btn" @click="cancelTag">Cancel</button>
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
                ref="valueDropdownRef"
                class="value-dropdown"
                @mousedown.stop
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
                  @mousedown.stop.prevent="newRule.value = val; valueDropdownOpen = false"
                >
                  {{ val }}
                </div>
              </div>
            </Teleport>
            <button class="add-btn" @click="addRule" :disabled="!newRule.field || !newRule.value">Add</button>
            <button class="cancel-btn" @click="showAddRule = false; valueDropdownOpen = false">Cancel</button>
          </div>
          <button class="clear-btn" @click="graphStore.clearFilterRules()" v-if="graphStore.filterRules.length">
            Clear All
          </button>
        </div>
      </div>
    </div>
    <div v-if="!graphStore.nodes.length" class="empty-state">
      <div class="empty-icon"><Icon name="icon-graph-empty" size="48" /></div>
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
  background: var(--app-graph-bg, #fafafa);
  background: var(--color-bg-subtle);
}

.graph-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: var(--color-white);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.control-btn {
  padding: 6px 12px;
  background: var(--surface-alt, #f5f5f5);
  border: 1px solid var(--border, #ddd);
  background: var(--color-bg-soft);
  border: 1px solid var(--color-border-lighter);
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(12px * var(--app-font-scale));
}

.control-btn:hover {
  background: var(--border, #e0e0e0);
  background: var(--color-border);
}

.control-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: var(--surface-alt, #f5f5f5);
}

.control-btn:disabled:hover {
  background: var(--surface-alt, #f5f5f5);
}

.control-btn.active {
  background: var(--accent, #1976d2);
  color: white;
  border-color: var(--accent, #1976d2);
}

.control-btn.active:hover {
  background: var(--accent, #1565c0);
}

.graph-controls-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.layout-label {
  font-size: calc(11px * var(--app-font-scale));
  color: #666;
  font-weight: 600;
}

.control-select {
  padding: 5px 8px;
  background: white;
  border: 1px solid var(--border, #ddd);
  border-radius: 4px;
  font-size: calc(12px * var(--app-font-scale));
  cursor: pointer;
}

.control-select:hover {
  border-color: var(--accent, #1976d2);
}

.node-count {
  font-size: calc(12px * var(--app-font-scale));
  color: #666;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.topology-view {
  flex: 1;
  min-height: 0;
  width: 100%;
  background: var(--app-graph-bg, #fafafa);
  background: var(--color-bg-subtle);
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
  border: 1px solid var(--border, #ddd);
  border: 1px solid var(--color-border-lighter);
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(12px * var(--app-font-scale));
  box-shadow: 0 2px 4px var(--color-shadow-soft);
}

.overlay-btn:hover {
  background: var(--surface-alt, #f5f5f5);
  background: var(--color-bg-soft);
}

.filter-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  background: white;
  border: 1px solid var(--border, #ddd);
  border: 1px solid var(--color-border-lighter);
  border-radius: 4px;
  box-shadow: 0 2px 8px var(--color-shadow-medium);
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
  font-size: calc(12px * var(--app-font-scale));
}

.filter-option:hover {
  background: var(--surface-alt, #f5f5f5);
  background: var(--color-bg-soft);
}

.region-filter-option {
  min-width: 260px;
}

.region-option-text {
  display: flex;
  flex: 1;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

.region-code {
  flex: 0 0 auto;
  color: var(--color-text-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: calc(11px * var(--app-font-scale));
}

.match-toggle {
  display: flex;
  gap: 4px;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--color-border-lighter);
}

.toggle-btn {
  flex: 1;
  padding: 4px 8px;
  background: var(--surface-alt, #f5f5f5);
  border: 1px solid var(--border, #ddd);
  background: var(--color-bg-soft);
  border: 1px solid var(--color-border-lighter);
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(11px * var(--app-font-scale));
  font-weight: 600;
}

.toggle-btn.active {
  background: var(--accent, #1976d2);
  color: white;
  border-color: var(--accent, #1976d2);
}

.toggle-btn:hover:not(.active) {
  background: var(--border, #e0e0e0);
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

.toggle-btn:hover:not(.active) {
  background: var(--color-border);
}

.advanced-dropdown {
  min-width: 220px;
}

.filter-rule {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  background: var(--surface-alt, #f5f5f5);
  background: var(--color-bg-soft);
  border-radius: 4px;
  margin-bottom: 4px;
  font-size: calc(11px * var(--app-font-scale));
}

.rule-field {
  font-weight: 600;
  color: var(--color-primary);
}

.rule-value {
  flex: 1;
  color: var(--color-text-primary);
}

.rule-remove {
  width: 20px;
  height: 20px;
  background: var(--color-error-light);
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  font-size: calc(14px * var(--app-font-scale));
  line-height: 1;
}

.rule-remove:hover {
  background: var(--color-error-dark-alt);
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

.rule-select,
.rule-input {
  width: 100%;
  box-sizing: border-box;
  padding: 6px 8px;
  border: 1px solid var(--border, #ddd);
  border: 1px solid var(--color-border-lighter);
  border-radius: 4px;
  font-size: calc(11px * var(--app-font-scale));
}

.add-btn {
  padding: 6px 12px;
  background: var(--accent, #1976d2);
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(11px * var(--app-font-scale));
}

.add-btn:hover {
  background: var(--accent, #1565c0);
  background: var(--color-primary-hover);
}

.cancel-btn {
  padding: 6px 12px;
  background: var(--surface-alt, #f5f5f5);
  border: 1px solid var(--border, #ddd);
  background: var(--color-bg-soft);
  border: 1px solid var(--color-border-lighter);
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(11px * var(--app-font-scale));
}

.cancel-btn:hover {
  background: var(--border, #e0e0e0);
}

.rule-input:disabled {
  background: var(--surface-alt, #f5f5f5);
  color: #999;
  background: var(--color-border);
}

.rule-input:disabled {
  background: var(--color-bg-soft);
  color: var(--color-text-muted);
  cursor: not-allowed;
}

.add-btn:disabled {
  background: var(--color-border-input);
  cursor: not-allowed;
}

.value-input-wrapper {
  position: relative;
  width: 100%;
}

.value-dropdown {
  position: fixed;
  background: white;
  border: 1px solid var(--border, #ddd);
  border: 1px solid var(--color-border-lighter);
  border-radius: 4px;
  box-shadow: 0 2px 8px var(--color-shadow-medium);
  max-height: 150px;
  overflow-y: auto;
  z-index: 9999;
}

.value-option {
  padding: 6px 8px;
  cursor: pointer;
  font-size: calc(11px * var(--app-font-scale));
}

.value-option:hover {
  background: var(--color-bg-soft-alt);
}

.tag-section {
  margin-bottom: 8px;
}

.tag-key {
  font-size: calc(10px * var(--app-font-scale));
  font-weight: 600;
  color: var(--color-primary);
  padding: 4px 8px;
  background: var(--surface-alt, #f5f5f5);
  background: var(--color-bg-soft);
  border-radius: 4px;
  margin-bottom: 4px;
}

.filter-help {
  font-size: calc(10px * var(--app-font-scale));
  color: #666;
  padding: 4px 8px;
  background: var(--surface-alt, #f5f5f5);
  font-size: 10px;
  color: var(--color-text-secondary);
  padding: 4px 8px;
  background: var(--color-bg-soft);
  border-radius: 4px;
  margin-top: 4px;
}

.clear-btn {
  width: 100%;
  margin-top: 8px;
  padding: 4px;
  background: var(--surface-alt, #f5f5f5);
  border: 1px solid var(--border, #ddd);
  background: var(--color-bg-soft);
  border: 1px solid var(--color-border-lighter);
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(11px * var(--app-font-scale));
}

.clear-btn:hover {
  background: var(--border, #e0e0e0);
  background: var(--color-border);
}

.empty-state {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  color: var(--color-text-secondary);
  pointer-events: none;
}

.empty-icon {
  font-size: calc(48px * var(--app-font-scale));
  margin-bottom: 16px;
  display: flex;
  justify-content: center;
}

.empty-state h3 {
  margin: 0 0 8px 0;
  font-size: calc(18px * var(--app-font-scale));
  color: #333;
  font-size: 18px;
  color: var(--color-text-primary);
}

.empty-state p {
  margin: 0;
  font-size: calc(14px * var(--app-font-scale));
  max-width: 300px;
}
</style>
