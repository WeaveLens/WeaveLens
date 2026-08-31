<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useScanStore } from './stores/scan'
import { useGraphStore } from './stores/graph'
import { useUiStore } from './stores/ui'
import TopologyView from './components/TopologyView.vue'
import ScanPanel from './components/ScanPanel.vue'
import ResourceDetail from './components/ResourceDetail.vue'
import Legend from './components/Legend.vue'
import SearchBar from './components/SearchBar.vue'

const scanStore = useScanStore()
const graphStore = useGraphStore()
const uiStore = useUiStore()

onMounted(() => {
  const saved = localStorage.getItem('weavelens_scan_id')
  if (saved) {
    scanStore.selectScan({ id: saved, status: 'UNKNOWN', region: '', updatedAt: '' } as any)
    graphStore.loadGraph(saved)
  }
})

watch(() => scanStore.currentScan?.id, (scanId) => {
  if (scanId) {
    localStorage.setItem('weavelens_scan_id', scanId)
    graphStore.loadGraph(scanId)
  }
})
</script>

<template>
  <div class="app">
    <header class="app-header">
      <h1>WeaveLens</h1>
      <div class="header-actions">
        <button @click="uiStore.toggleSidebar">Toggle Sidebar</button>
        <button @click="uiStore.toggleLegend">Toggle Legend</button>
      </div>
    </header>

    <div class="app-body">
      <aside v-if="uiStore.sidebarOpen" class="left-panel">
        <ScanPanel />
      </aside>

      <main class="main-content">
        <SearchBar />
        <TopologyView />
        <Legend v-if="uiStore.legendOpen" />
      </main>

      <aside v-if="uiStore.sidebarOpen" class="right-panel">
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
}
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #1976d2;
  color: white;
}
.app-header h1 {
  margin: 0;
  font-size: 20px;
}
.header-actions {
  display: flex;
  gap: 8px;
}
.header-actions button {
  padding: 6px 12px;
  background: rgba(255,255,255,0.2);
  border: 1px solid rgba(255,255,255,0.3);
  color: white;
  border-radius: 4px;
  cursor: pointer;
}
.app-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}
.left-panel {
  width: 320px;
  background: white;
  border-right: 1px solid #e0e0e0;
  overflow-y: auto;
}
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.right-panel {
  width: 320px;
  background: white;
  border-left: 1px solid #e0e0e0;
  overflow-y: auto;
}
</style>
