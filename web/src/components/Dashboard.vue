<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { useScanStore } from '../stores/scan'
import { useGraphStore } from '../stores/graph'
import { useUiStore } from '../stores/ui'
import TopologyView from './TopologyView.vue'
import ScanPanel from './ScanPanel.vue'
import ResourceDetail from './ResourceDetail.vue'
import Legend from './Legend.vue'
import SearchBar from './SearchBar.vue'
import CloudConnection from './CloudConnection.vue'
import SettingsDropdown from './SettingsDropdown.vue'
import { subscribeScans } from '../api/client'
import type { LayoutMode } from '../stores/graph'

const scanStore = useScanStore()
const graphStore = useGraphStore()
const uiStore = useUiStore()
let unsubscribeScans: (() => void) | null = null

onMounted(async () => {
  unsubscribeScans = subscribeScans((scans) => {
    scanStore.setScans(scans)
    if (scans.length === 0) {
      scanStore.selectScan(null)
      graphStore.clearGraph()
    } else if (scanStore.currentScan && !scans.find(scan => scan.id === scanStore.currentScan?.id)) {
      scanStore.selectScan(null)
      graphStore.clearGraph()
    }
  })

  await scanStore.fetchScans()
})

onUnmounted(() => {
  unsubscribeScans?.()
})

watch(() => scanStore.currentScan?.id, async (scanId) => {
  if (scanId) {
    const prevId = graphStore.currentScanId
    if (prevId && prevId !== scanId) {
      graphStore.saveLayout(prevId)
      graphStore.savePositionsToBackend(prevId).catch(() => {})
    }
    graphStore.clearGraph()
    graphStore.restoreLayout(scanId, {
      mode: scanStore.currentScan?.layout as LayoutMode | undefined,
      locked: scanStore.currentScan?.locked,
    })
    await graphStore.loadPositionsFromBackend(scanId)
    await graphStore.loadGraph(scanId)
    await scanStore.refreshStatus(scanId)
  }
}, { immediate: true })
</script>

<template>
  <div class="app">
    <header class="app-header">
      <div class="header-left">
        <h1>WeaveLens</h1>
        <span class="header-subtitle">Infrastructure Visualization</span>
      </div>
      <nav class="header-actions">
        <router-link to="/setup/aws" class="header-link">Setup Guide</router-link>
        <button @click="uiStore.toggleSidebar" class="header-btn">
          {{ uiStore.sidebarOpen ? 'Hide' : 'Show' }} Panel
        </button>
        <SettingsDropdown />
      </nav>
    </header>

    <div class="app-body">
      <aside v-if="uiStore.sidebarOpen" class="panel left-panel">
        <CloudConnection />
        <ScanPanel />
      </aside>

      <main class="main-content">
        <SearchBar />
        <TopologyView />
        <Legend />
      </main>

      <aside v-if="uiStore.sidebarOpen" class="panel right-panel">
        <ResourceDetail />
      </aside>
    </div>
  </div>
</template>

<style scoped>
.app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100%;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 24px;
  background: var(--app-header-bg);
  color: var(--app-header-text);
  box-shadow: 0 2px 4px var(--color-shadow-soft);
  z-index: 10;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.app-header h1 {
  margin: 0;
  font-size: calc(22px * var(--app-font-scale));
  font-weight: 600;
}

.header-subtitle {
  font-size: calc(13px * var(--app-font-scale));
  opacity: 0.8;
  font-weight: 400;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.header-btn,
.header-link {
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
  color: var(--color-white);
  border-radius: 6px;
  cursor: pointer;
  font-size: calc(13px * var(--app-font-scale));
  transition: background 0.2s;
}

.header-btn:hover {
  background: var(--color-overlay-white-25);
}

.header-link:hover {
  background: var(--color-overlay-white-25);
}

.header-link {
  text-decoration: none;
}

.app-body {
  display: flex;
  flex: 1;
  min-height: 0;
}

.panel {
  width: 300px;
  flex-shrink: 0;
  background: var(--app-panel-bg, #fafafa);
  background: var(--color-bg-subtle);
  overflow-y: auto;
  overflow-x: hidden;
}

.left-panel {
  border-right: 1px solid var(--app-panel-border, #e0e0e0);
}

.right-panel {
  border-left: 1px solid var(--app-panel-border, #e0e0e0);
  border-right: 1px solid var(--color-border);
}

.right-panel {
  border-left: 1px solid var(--color-border);
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

/* Responsive: Smaller screens */
@media (max-width: 1200px) {
  .panel {
    width: 260px;
  }
}

@media (max-width: 900px) {
  .panel {
    width: 240px;
  }
}

/* Very small screens or high zoom */
@media (max-width: 700px) {
  .app-body {
    flex-direction: column;
  }
  .panel {
    width: 100%;
    max-height: 35vh;
    flex-shrink: 0;
  }
  .left-panel {
    border-right: none;
    border-bottom: 1px solid var(--app-panel-border, #e0e0e0);
    border-bottom: 1px solid var(--color-border);
  }
}
</style>
