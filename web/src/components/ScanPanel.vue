<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useScanStore } from '../stores/scan'
import { useGraphStore } from '../stores/graph'
import { getRegions, type RegionInfo } from '../api/client'

const scanStore = useScanStore()
const graphStore = useGraphStore()
const region = ref('')
const regions = ref<RegionInfo[]>([{ value: '', label: 'All Regions' }])

let pollInterval: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  try {
    const fetched = await getRegions()
    regions.value = [{ value: '', label: 'All Regions' }, ...fetched]
  } catch {
    // Use default regions if API fails
  }

  await scanStore.fetchScans()

  pollInterval = setInterval(() => {
    scanStore.fetchScansSilent()
  }, 1000)
})

onUnmounted(() => {
  if (pollInterval) {
    clearInterval(pollInterval)
  }
})

function getRegionLabel(regionCode?: string): string {
  if (!regionCode) return 'All Regions'
  const found = regions.value.find(r => r.value === regionCode)
  return found ? found.label : regionCode
}

function formatTime(utcString: string): string {
  if (!utcString) return '—'
  const date = new Date(utcString)
  if (Number.isNaN(date.getTime())) return '—'
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'short',
    timeStyle: 'medium',
  }).format(date)
}

function loadHistoricalScan(scanId: string) {
  graphStore.clearGraph()
  graphStore.loadGraph(scanId)
  scanStore.selectScan(scanStore.scans.find(s => s.id === scanId) || null)
}

async function handleStartScan() {
  await scanStore.createScan({ region: region.value.trim() })
}

const statusColor = (status: string) => {
  switch (status) {
    case 'RUNNING':
      return '#4CAF50'
    case 'COMPLETED':
      return '#2196F3'
    case 'FAILED':
      return '#F44336'
    case 'CANCELLED':
      return '#FF9800'
    default:
      return '#9E9E9E'
  }
}

const hasScans = computed(() => scanStore.scans.length > 0)
const isLoading = computed(() => scanStore.loading)
</script>

<template>
  <div class="scan-panel">
    <div class="panel-header">
      <h3>Scan Configuration</h3>
    </div>

    <form @submit.prevent="handleStartScan" class="scan-form">
      <label>
        <span class="label-text">Region</span>
        <select v-model="region" class="region-select">
          <option v-for="r in regions" :key="r.value" :value="r.value">
            {{ r.label }}
          </option>
        </select>
      </label>
      <button type="submit" :disabled="scanStore.loading" class="start-btn">
        <span v-if="scanStore.loading" class="btn-loading">
          <span class="spinner"></span>
          Starting...
        </span>
        <span v-else>Start Scan</span>
      </button>
    </form>

    <div v-if="scanStore.error" class="error-message">
      <span class="error-icon">⚠</span>
      {{ scanStore.error }}
    </div>

    <div v-if="scanStore.currentScan" class="current-scan">
      <h4>Current Scan</h4>
      <div class="scan-card">
        <div class="scan-row">
          <span class="label">ID</span>
          <span class="value mono">{{ scanStore.currentScan.id }}</span>
        </div>
        <div class="scan-row">
          <span class="label">Status</span>
          <span class="value status" :style="{ color: statusColor(scanStore.currentScan.status) }">
            {{ scanStore.currentScan.status }}
          </span>
        </div>
        <div class="scan-row">
          <span class="label">Region</span>
          <span class="value">{{ getRegionLabel(scanStore.currentScan.region) }}</span>
        </div>
        <div class="scan-row">
          <span class="label">Service(s)</span>
          <span class="value">{{ scanStore.currentScan.nodeCount || 0 }}</span>
        </div>
      </div>
    </div>

    <div v-if="isLoading" class="loading-state">
      <span class="spinner"></span>
      Loading history...
    </div>

    <div v-else-if="hasScans" class="scan-history">
      <h4>Scan History</h4>
      <ul>
        <li
          v-for="scan in scanStore.scans"
          :key="scan.id"
          class="history-item"
          :class="{ active: scanStore.currentScan?.id === scan.id }"
          @click="loadHistoricalScan(scan.id)"
        >
          <div class="history-main">
            <span class="history-status" :style="{ color: statusColor(scan.status) }">
              {{ scan.status }}
            </span>
            <span class="history-region">{{ getRegionLabel(scan.region) }}</span>
          </div>
          <div class="history-meta">
            <span class="history-id">{{ scan.id }}</span>
            <span class="history-nodes">{{ scan.nodeCount || 0 }} {{ (scan.nodeCount || 0) <= 1 ? 'service' : 'services' }}</span>
            <span class="history-time" v-show="false">{{ formatTime(scan.createdAt) }}</span>
          </div>
        </li>
      </ul>
    </div>

    <div
      v-else-if="!isLoading && !hasScans"
      class="empty-state"
    >
      <div class="empty-icon">🔍</div>
      <p>No scans yet. Start your first scan above.</p>
    </div>
  </div>
</template>

<style scoped>
.scan-panel {
  padding: 16px;
}

.panel-header {
  margin-bottom: 16px;
}

.panel-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

.scan-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.scan-form label {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.label-text {
  font-size: 12px;
  font-weight: 500;
  color: #666;
}

.region-select {
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 13px;
  background: white;
  cursor: pointer;
}

.region-select:focus {
  outline: none;
  border-color: #1976d2;
}

.start-btn {
  padding: 10px 16px;
  background: #1976d2;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: background 0.2s;
}

.start-btn:hover:not(:disabled) {
  background: #1565c0;
}

.start-btn:disabled {
  background: #90caf9;
  cursor: not-allowed;
}

.btn-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  padding: 10px 12px;
  background: #ffebee;
  border-radius: 6px;
  color: #c62828;
  font-size: 12px;
}

.error-icon {
  font-size: 14px;
}

.current-scan {
  margin-top: 20px;
}

.current-scan h4 {
  margin: 0 0 8px 0;
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.scan-card {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 12px;
}

.scan-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
  font-size: 12px;
}

.scan-row:last-child {
  margin-bottom: 0;
}

.label {
  font-weight: 500;
  color: #666;
}

.value {
  color: #333;
}

.mono {
  font-family: monospace;
  font-size: 11px;
}

.status {
  font-weight: 600;
}

.scan-history {
  margin-top: 20px;
}

.scan-history h4 {
  margin: 0 0 8px 0;
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.scan-history ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.history-item {
  display: flex;
  flex-direction: column;
  padding: 10px 8px;
  border-bottom: 1px solid #eee;
  font-size: 12px;
  cursor: pointer;
  border-radius: 4px;
  transition: background 0.15s;
}

.history-item:hover {
  background: #f5f5f5;
}

.history-item.active {
  background: #e3f2fd;
}

.history-item:last-child {
  border-bottom: none;
}

.history-main {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.history-status {
  font-weight: 500;
}

.history-region {
  color: #666;
}

.history-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  font-size: 10px;
  color: #999;
}

.history-time {
  white-space: nowrap;
  color: #666;
}

.history-id {
  font-family: monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 60%;
}

.history-time {
  white-space: nowrap;
}

.history-nodes {
  font-weight: 500;
}

.empty-state {
  margin-top: 20px;
  text-align: center;
  padding: 20px;
  color: #666;
}

.loading-state {
  margin-top: 20px;
  text-align: center;
  padding: 20px;
  color: #666;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.empty-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.empty-state p {
  margin: 0;
  font-size: 12px;
}
</style>
