<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useScanStore } from '../stores/scan'
import { useGraphStore } from '../stores/graph'
import { getRegions, type RegionInfo } from '../api/client'

const scanStore = useScanStore()
const graphStore = useGraphStore()
const selectedRegions = ref<string[]>([])
const regions = ref<RegionInfo[]>([])
const regionDropdownOpen = ref(false)
const regionFilterRef = ref<HTMLElement | null>(null)

function handleClickOutside(e: MouseEvent) {
  if (!regionDropdownOpen.value) return
  const target = e.target as Node | null
  if (!target) return
  if (regionFilterRef.value && !regionFilterRef.value.contains(target)) {
    regionDropdownOpen.value = false
  }
}

onMounted(async () => {
  try {
  const fetched = await getRegions()
    regions.value = fetched
  } catch {
    // Use empty list if API fails
  }
  document.addEventListener('mousedown', handleClickOutside, true)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', handleClickOutside, true)
})

function toggleRegion(value: string) {
  const idx = selectedRegions.value.indexOf(value)
  if (idx >= 0) {
    selectedRegions.value.splice(idx, 1)
  } else {
    selectedRegions.value.push(value)
  }
}

function clearRegions() {
  selectedRegions.value = []
}

function getRegionLabel(regionCode?: string): string {
  if (!regionCode) return 'All Regions'
  if (regionCode === 'all') return 'All Regions'
  if (regionCode.includes(',')) {
    const count = regionCode.split(',').length
    return `${count} regions`
  }
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
  const config = selectedRegions.value.length > 0
    ? { regions: [...selectedRegions.value] }
    : { regions: [] as string[] }
  await scanStore.createScan(config)
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
const regionButtonLabel = computed(() => {
  const n = selectedRegions.value.length
  if (n === 0) return 'All Regions'
  if (n === 1) {
    const found = regions.value.find(r => r.value === selectedRegions.value[0])
    return found ? found.label : selectedRegions.value[0]
  }
  return `${n} regions`
})
</script>

<template>
  <div class="scan-panel">
    <div class="panel-header">
      <h3>Scan Configuration</h3>
    </div>

    <form @submit.prevent="handleStartScan" class="scan-form">
      <div class="region-filter-group" ref="regionFilterRef">
        <label class="label-text">Region</label>
        <button
          type="button"
          class="region-btn"
          @click="regionDropdownOpen = !regionDropdownOpen"
        >
          <span>{{ regionButtonLabel }}</span>
          <span class="caret">▾</span>
        </button>
        <div v-if="regionDropdownOpen" class="region-dropdown">
          <div v-if="regions.length === 0" class="region-empty">
            No regions available
          </div>
          <template v-else>
            <label
              v-for="r in regions"
              :key="r.value"
              class="region-option"
            >
              <input
                type="checkbox"
                :checked="selectedRegions.includes(r.value)"
                @change="toggleRegion(r.value)"
              />
              <span>{{ r.label }}</span>
            </label>
            <div v-if="selectedRegions.length > 0" class="region-actions">
              <button type="button" class="clear-btn" @click="clearRegions">
                Clear selection
              </button>
            </div>
          </template>
        </div>
        <div v-if="!regionDropdownOpen" class="region-hint">
          {{ selectedRegions.length === 0
            ? 'No selection = scan all regions'
            : `${selectedRegions.length} selected` }}
        </div>
      </div>
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

.region-filter-group {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.label-text {
  font-size: 12px;
  font-weight: 500;
  color: #666;
}

.region-btn {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 13px;
  background: white;
  cursor: pointer;
  color: #333;
  text-align: left;
}

.region-btn:hover {
  border-color: #bbb;
}

.region-btn:focus {
  outline: none;
  border-color: #1976d2;
}

.caret {
  font-size: 10px;
  color: #999;
}

.region-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 240px;
  overflow-y: auto;
  background: white;
  border: 1px solid #ddd;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 10;
  padding: 4px 0;
}

.region-empty {
  padding: 12px;
  color: #999;
  font-size: 12px;
  text-align: center;
}

.region-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  font-size: 13px;
  cursor: pointer;
  color: #333;
}

.region-option:hover {
  background: #f5f5f5;
}

.region-option input {
  cursor: pointer;
  margin: 0;
}

.region-actions {
  border-top: 1px solid #eee;
  padding: 6px 8px;
  display: flex;
  justify-content: flex-end;
}

.clear-btn {
  background: none;
  border: none;
  color: #1976d2;
  font-size: 12px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
}

.clear-btn:hover {
  background: #e3f2fd;
}

.region-hint {
  font-size: 11px;
  color: #888;
  padding: 0 2px;
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
