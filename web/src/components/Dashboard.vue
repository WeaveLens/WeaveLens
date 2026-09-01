<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useScanStore } from '../stores/scan'
import { useGraphStore } from '../stores/graph'
import { useUiStore } from '../stores/ui'
import TopologyView from './TopologyView.vue'
import ScanPanel from './ScanPanel.vue'
import ResourceDetail from './ResourceDetail.vue'
import Legend from './Legend.vue'
import SearchBar from './SearchBar.vue'
import CloudConnection from './CloudConnection.vue'

const scanStore = useScanStore()
const graphStore = useGraphStore()
const uiStore = useUiStore()

onMounted(async () => {
  await scanStore.fetchScans()
  const saved = localStorage.getItem('weavelens_scan_id')
  if (!saved) return

  const savedScan = scanStore.scans.find(scan => scan.id === saved)
  if (!savedScan) {
    localStorage.removeItem('weavelens_scan_id')
    scanStore.selectScan(null)
    return
  }

  scanStore.selectScan(savedScan)
  await graphStore.loadGraph(saved)
  await scanStore.refreshStatus(saved)
})

watch(() => scanStore.currentScan?.id, async (scanId) => {
  if (scanId) {
    localStorage.setItem('weavelens_scan_id', scanId)
    graphStore.clearGraph()
    await graphStore.loadGraph(scanId)
    await scanStore.refreshStatus(scanId)
  }
})
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
        <button @click="uiStore.toggleLegend" class="header-btn">
          {{ uiStore.legendOpen ? 'Hide' : 'Show' }} Legend
        </button>
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
        <Legend v-if="uiStore.legendOpen" />
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
  background: linear-gradient(135deg, #1976d2 0%, #1565c0 100%);
  color: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
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
  font-size: 22px;
  font-weight: 600;
}

.header-subtitle {
  font-size: 13px;
  opacity: 0.8;
  font-weight: 400;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.header-btn {
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  transition: background 0.2s;
}

.header-btn:hover {
  background: rgba(255, 255, 255, 0.25);
}

.header-link {
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: white;
  border-radius: 6px;
  cursor: pointer;
  text-decoration: none;
  font-size: 13px;
  transition: background 0.2s;
}

.header-link:hover {
  background: rgba(255, 255, 255, 0.25);
}

.app-body {
  display: flex;
  flex: 1;
  min-height: 0;
}

.panel {
  width: 300px;
  flex-shrink: 0;
  background: #fafafa;
  overflow-y: auto;
  overflow-x: hidden;
}

.left-panel {
  border-right: 1px solid #e0e0e0;
}

.right-panel {
  border-left: 1px solid #e0e0e0;
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
    border-bottom: 1px solid #e0e0e0;
  }
}
</style>
