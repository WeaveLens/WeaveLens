<script setup lang="ts">
import { ref } from 'vue'
import { useScanStore } from '../stores/scan'

const scanStore = useScanStore()
const region = ref('us-east-1')

async function handleStartScan() {
  if (!region.value.trim()) return
  await scanStore.createScan({ region: region.value.trim() })
  region.value = 'us-east-1'
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
</script>

<template>
  <div class="scan-panel">
    <h3>Scan Configuration</h3>
    <form @submit.prevent="handleStartScan" class="scan-form">
      <label>
        Region
        <input v-model="region" type="text" placeholder="us-east-1" />
      </label>
      <button type="submit" :disabled="scanStore.loading">
        {{ scanStore.loading ? 'Starting...' : 'Start Scan' }}
      </button>
    </form>

    <div v-if="scanStore.error" class="error">
      {{ scanStore.error }}
    </div>

    <div v-if="scanStore.currentScan" class="current-scan">
      <h4>Current Scan</h4>
      <div class="scan-card">
        <div class="scan-row">
          <span class="label">ID:</span>
          <span class="value">{{ scanStore.currentScan.id }}</span>
        </div>
        <div class="scan-row">
          <span class="label">Status:</span>
          <span class="value" :style="{ color: statusColor(scanStore.currentScan.status) }">
            {{ scanStore.currentScan.status }}
          </span>
        </div>
        <div class="scan-row">
          <span class="label">Region:</span>
          <span class="value">{{ scanStore.currentScan.region }}</span>
        </div>
      </div>
    </div>

    <div v-if="scanStore.scans.length" class="scan-history">
      <h4>Scan History</h4>
      <ul>
        <li v-for="scan in scanStore.scans" :key="scan.id">
          <span :style="{ color: statusColor(scan.status) }">{{ scan.status }}</span>
          <span>{{ scan.region }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.scan-panel {
  padding: 16px;
  border-bottom: 1px solid #e0e0e0;
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
  font-size: 14px;
}
.scan-form input {
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
}
.scan-form button {
  padding: 8px 16px;
  background: #1976d2;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
.scan-form button:disabled {
  background: #90caf9;
  cursor: not-allowed;
}
.error {
  color: #f44336;
  margin-top: 8px;
  font-size: 14px;
}
.current-scan {
  margin-top: 16px;
}
.scan-card {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  margin-top: 8px;
}
.scan-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}
.label {
  font-weight: 500;
}
.value {
  font-family: monospace;
}
.scan-history {
  margin-top: 16px;
}
.scan-history ul {
  list-style: none;
  padding: 0;
  margin-top: 8px;
}
.scan-history li {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  border-bottom: 1px solid #eee;
  font-size: 14px;
}
</style>
