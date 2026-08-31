<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { getConnectionStatus } from '../api/client'
import type { ConnectionStatus, ConnectionState } from '../types'

const connection = ref<ConnectionStatus | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

async function loadConnectionStatus() {
  loading.value = true
  error.value = null
  try {
    connection.value = await getConnectionStatus()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load connection status'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadConnectionStatus()
})

const stateLabel = (state: ConnectionState): string => {
  switch (state) {
    case 'connected':
      return 'Connected'
    case 'connecting':
      return 'Connecting'
    case 'not_connected':
      return 'Not Connected'
    case 'authentication_error':
      return 'Authentication Error'
    case 'access_denied':
      return 'Access Denied'
    case 'configuration_error':
      return 'Configuration Error'
    case 'unknown_error':
      return 'Unknown Error'
    default:
      return 'Unknown'
  }
}

const stateColor = (state: ConnectionState): string => {
  switch (state) {
    case 'connected':
      return '#4CAF50'
    case 'connecting':
      return '#FF9800'
    case 'not_connected':
      return '#9E9E9E'
    case 'authentication_error':
    case 'access_denied':
    case 'configuration_error':
    case 'unknown_error':
      return '#F44336'
    default:
      return '#9E9E9E'
  }
}

const isConnected = computed(() => connection.value?.state === 'connected')
const hasError = computed(() => {
  const state = connection.value?.state
  return state && ['authentication_error', 'access_denied', 'configuration_error', 'unknown_error'].includes(state)
})
</script>

<template>
  <div class="cloud-connection">
    <h3>Cloud Connections</h3>

    <div v-if="loading" class="loading">
      Checking connection...
    </div>

    <div v-else-if="error" class="error">
      {{ error }}
      <button @click="loadConnectionStatus" class="retry-btn">Retry</button>
    </div>

    <div v-else-if="connection" class="connection-card">
      <div class="provider-header">
        <span class="provider-name">AWS</span>
        <span
          class="connection-status"
          :style="{ color: stateColor(connection.state) }"
        >
          {{ stateLabel(connection.state) }}
        </span>
      </div>

      <div v-if="isConnected" class="connection-details">
        <div class="detail-row">
          <span class="label">Account:</span>
          <span class="value">{{ connection.accountId }}</span>
        </div>
        <div class="detail-row">
          <span class="label">Region:</span>
          <span class="value">{{ connection.region }}</span>
        </div>
        <div class="detail-row">
          <span class="label">Identity:</span>
          <span class="value mono">{{ connection.arn }}</span>
        </div>
        <div v-if="connection.credentialSource" class="detail-row">
          <span class="label">Credential Source:</span>
          <span class="value">{{ connection.credentialSource }}</span>
        </div>
      </div>

      <div v-else-if="hasError" class="error-message">
        <p>{{ connection.message }}</p>
        <router-link to="/setup/aws" class="setup-link">View Setup Guide</router-link>
      </div>

      <div v-else class="not-connected">
        <p>AWS credentials not configured.</p>
        <router-link to="/setup/aws" class="setup-link">View Setup Guide</router-link>
      </div>
    </div>

    <div v-else class="no-data">
      No connection data available.
      <button @click="loadConnectionStatus" class="retry-btn">Retry</button>
    </div>
  </div>
</template>

<style scoped>
.cloud-connection {
  padding: 16px;
  border-bottom: 1px solid #e0e0e0;
}

.cloud-connection h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
}

.loading {
  color: #666;
  font-size: 14px;
  padding: 12px 0;
}

.error {
  color: #f44336;
  font-size: 14px;
  padding: 12px 0;
}

.error-message {
  color: #f44336;
  font-size: 13px;
}

.error-message p {
  margin: 0 0 8px 0;
}

.connection-card {
  background: #f5f5f5;
  border-radius: 8px;
  padding: 12px;
}

.provider-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e0e0e0;
}

.provider-name {
  font-weight: 600;
  font-size: 14px;
}

.connection-status {
  font-weight: 500;
  font-size: 13px;
}

.connection-details {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
}

.label {
  font-weight: 500;
  color: #666;
}

.value {
  text-align: right;
  max-width: 200px;
  word-break: break-all;
}

.mono {
  font-family: monospace;
  font-size: 11px;
}

.not-connected {
  text-align: center;
  padding: 16px 0;
}

.not-connected p {
  margin: 0 0 8px 0;
  color: #666;
  font-size: 14px;
}

.setup-link {
  color: #1976d2;
  text-decoration: none;
  font-size: 14px;
  font-weight: 500;
}

.setup-link:hover {
  text-decoration: underline;
}

.retry-btn {
  margin-left: 8px;
  padding: 4px 12px;
  background: #1976d2;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.no-data {
  color: #666;
  font-size: 14px;
  padding: 12px 0;
}
</style>
