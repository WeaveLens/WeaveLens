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
      return 'Connecting...'
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

const stateIcon = (state: ConnectionState): string => {
  switch (state) {
    case 'connected':
      return '✓'
    case 'connecting':
      return '⟳'
    case 'not_connected':
      return '○'
    case 'authentication_error':
    case 'access_denied':
    case 'configuration_error':
    case 'unknown_error':
      return '✕'
    default:
      return '?'
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
    <div class="panel-header">
      <h3>Cloud Connections</h3>
      <button @click="loadConnectionStatus" class="refresh-btn" :disabled="loading">
        {{ loading ? '⟳' : '↻' }}
      </button>
    </div>

    <div v-if="loading" class="state-container loading">
      <div class="state-icon">⟳</div>
      <span>Checking connection...</span>
    </div>

    <div v-else-if="error" class="state-container error">
      <div class="state-icon">✕</div>
      <div class="state-content">
        <span class="state-label">Error</span>
        <span class="state-message">{{ error }}</span>
      </div>
      <button @click="loadConnectionStatus" class="retry-btn">Retry</button>
    </div>

    <div v-else-if="connection" class="connection-card">
      <div class="provider-header">
        <div class="provider-info">
          <span class="provider-icon">☁</span>
          <span class="provider-name">AWS</span>
        </div>
        <span
          class="connection-status"
          :style="{ color: stateColor(connection.state) }"
        >
          <span class="status-icon">{{ stateIcon(connection.state) }}</span>
          {{ stateLabel(connection.state) }}
        </span>
      </div>

      <div v-if="isConnected" class="connection-details">
        <div class="detail-row">
          <span class="label">Account</span>
          <span class="value">{{ connection.accountId }}</span>
        </div>
        <div class="detail-row">
          <span class="label">Region</span>
          <span class="value">{{ connection.region }}</span>
        </div>
        <div class="detail-row">
          <span class="label">Identity</span>
          <span class="value mono" :title="connection.arn">{{ connection.arn }}</span>
        </div>
        <div v-if="connection.credentialSource" class="detail-row">
          <span class="label">Source</span>
          <span class="value">{{ connection.credentialSource }}</span>
        </div>
      </div>

      <div v-else-if="hasError" class="state-message-full">
        <p>{{ connection.message || 'Unable to connect to AWS.' }}</p>
        <router-link to="/setup/aws" class="setup-link">View Setup Guide →</router-link>
      </div>

      <div v-else class="state-message-full">
        <p>AWS credentials not configured.</p>
        <router-link to="/setup/aws" class="setup-link">View Setup Guide →</router-link>
      </div>
    </div>

    <div v-else class="state-container empty">
      <div class="state-icon">○</div>
      <span>No connection data available</span>
      <button @click="loadConnectionStatus" class="retry-btn">Retry</button>
    </div>
  </div>
</template>

<style scoped>
.cloud-connection {
  padding: 16px;
  border-bottom: 1px solid #e0e0e0;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.panel-header h3 {
  margin: 0;
  font-size: calc(14px * var(--app-font-scale));
  font-weight: 600;
  color: #333;
}

.refresh-btn {
  width: 28px;
  height: 28px;
  border: 1px solid var(--border, #ddd);
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(14px * var(--app-font-scale));
  display: flex;
  align-items: center;
  justify-content: center;
}

.refresh-btn:hover:not(:disabled) {
  background: var(--surface-alt, #f5f5f5);
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.state-container {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--surface-alt, #f5f5f5);
  border-radius: 8px;
  font-size: calc(13px * var(--app-font-scale));
}

.state-container.loading {
  color: #666;
}

.state-container.error {
  background: #ffebee;
  color: #c62828;
  flex-wrap: wrap;
}

.state-container.empty {
  color: #666;
}

.state-icon {
  font-size: calc(20px * var(--app-font-scale));
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border-radius: 50%;
}

.loading .state-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.state-content {
  flex: 1;
  min-width: 0;
}

.state-label {
  font-weight: 600;
  display: block;
}

.state-message {
  font-size: calc(12px * var(--app-font-scale));
  opacity: 0.8;
}

.retry-btn {
  padding: 6px 12px;
  background: #1976d2;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(12px * var(--app-font-scale));
  white-space: nowrap;
}

.retry-btn:hover {
  background: #1565c0;
}

.connection-card {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  overflow: hidden;
}

.provider-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--surface-alt, #f5f5f5);
  border-bottom: 1px solid #e0e0e0;
}

.provider-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.provider-icon {
  font-size: calc(18px * var(--app-font-scale));
}

.provider-name {
  font-weight: 600;
  font-size: calc(14px * var(--app-font-scale));
}

.connection-status {
  font-weight: 500;
  font-size: calc(12px * var(--app-font-scale));
  display: flex;
  align-items: center;
  gap: 4px;
}

.status-icon {
  font-size: calc(14px * var(--app-font-scale));
}

.connection-details {
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  font-size: calc(12px * var(--app-font-scale));
}

.label {
  color: #666;
  font-weight: 500;
}

.value {
  color: #333;
  text-align: right;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mono {
  font-family: monospace;
  font-size: calc(11px * var(--app-font-scale));
}

.state-message-full {
  padding: 16px;
  text-align: center;
}

.state-message-full p {
  margin: 0 0 12px 0;
  color: #666;
  font-size: calc(13px * var(--app-font-scale));
}

.setup-link {
  color: #1976d2;
  text-decoration: none;
  font-size: calc(13px * var(--app-font-scale));
  font-weight: 500;
}

.setup-link:hover {
  text-decoration: underline;
}
</style>
