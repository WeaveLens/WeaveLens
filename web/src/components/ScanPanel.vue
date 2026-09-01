<script setup lang="ts">
import { ref, computed } from 'vue'
import { useScanStore } from '../stores/scan'

const scanStore = useScanStore()
const region = ref('')

const regions = [
  { value: '', label: 'All Regions' },
  // US Regions
  { value: 'us-east-1', label: 'US East (N. Virginia)' },
  { value: 'us-east-2', label: 'US East (Ohio)' },
  { value: 'us-west-1', label: 'US West (N. California)' },
  { value: 'us-west-2', label: 'US West (Oregon)' },
  // GovCloud
  { value: 'us-gov-east-1', label: 'AWS GovCloud (US-East)' },
  { value: 'us-gov-west-1', label: 'AWS GovCloud (US-West)' },
  // Canada
  { value: 'ca-central-1', label: 'Canada (Central)' },
  { value: 'ca-west-1', label: 'Canada West (Calgary)' },
  // South America
  { value: 'sa-east-1', label: 'South America (São Paulo)' },
  // Europe
  { value: 'eu-west-1', label: 'Europe (Ireland)' },
  { value: 'eu-west-2', label: 'Europe (London)' },
  { value: 'eu-west-3', label: 'Europe (Paris)' },
  { value: 'eu-central-1', label: 'Europe (Frankfurt)' },
  { value: 'eu-central-2', label: 'Europe (Zurich)' },
  { value: 'eu-south-1', label: 'Europe (Milan)' },
  { value: 'eu-south-2', label: 'Europe (Spain)' },
  { value: 'eu-north-1', label: 'Europe (Stockholm)' },
  // Middle East
  { value: 'me-south-1', label: 'Middle East (Bahrain)' },
  { value: 'me-central-1', label: 'Middle East (UAE)' },
  { value: 'il-central-1', label: 'Israel (Tel Aviv)' },
  // Africa
  { value: 'af-south-1', label: 'Africa (Cape Town)' },
  // Asia Pacific
  { value: 'ap-southeast-1', label: 'Asia Pacific (Singapore)' },
  { value: 'ap-southeast-2', label: 'Asia Pacific (Sydney)' },
  { value: 'ap-southeast-3', label: 'Asia Pacific (Jakarta)' },
  { value: 'ap-southeast-4', label: 'Asia Pacific (Melbourne)' },
  { value: 'ap-southeast-5', label: 'Asia Pacific (Malaysia)' },
  { value: 'ap-south-1', label: 'Asia Pacific (Mumbai)' },
  { value: 'ap-south-2', label: 'Asia Pacific (Hyderabad)' },
  { value: 'ap-northeast-1', label: 'Asia Pacific (Tokyo)' },
  { value: 'ap-northeast-2', label: 'Asia Pacific (Seoul)' },
  { value: 'ap-northeast-3', label: 'Asia Pacific (Osaka)' },
  { value: 'ap-east-1', label: 'Asia Pacific (Hong Kong)' },
  { value: 'cn-north-1', label: 'China (Beijing)' },
  { value: 'cn-northwest-1', label: 'China (Ningxia)' },
]

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
          <span class="value">{{ scanStore.currentScan.region }}</span>
        </div>
      </div>
    </div>

    <div v-if="hasScans" class="scan-history">
      <h4>Scan History</h4>
      <ul>
        <li v-for="scan in scanStore.scans" :key="scan.id" class="history-item">
          <span class="history-status" :style="{ color: statusColor(scan.status) }">
            {{ scan.status }}
          </span>
          <span class="history-region">{{ scan.region }}</span>
        </li>
      </ul>
    </div>

    <div v-else-if="!scanStore.currentScan" class="empty-state">
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
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
  font-size: 12px;
}

.history-item:last-child {
  border-bottom: none;
}

.history-status {
  font-weight: 500;
}

.history-region {
  color: #666;
}

.empty-state {
  margin-top: 20px;
  text-align: center;
  padding: 20px;
  color: #666;
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
