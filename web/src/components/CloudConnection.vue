<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { getConnectionStatus } from '../api/client'
import type { ConnectionStatus, ConnectionState } from '../types'
import { theme } from '../config/theme'
import Icon from './Icon.vue'

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
      return theme.success
    case 'connecting':
      return theme.warning.DEFAULT
    case 'not_connected':
      return theme.gray.DEFAULT
    case 'authentication_error':
    case 'access_denied':
    case 'configuration_error':
    case 'unknown_error':
      return theme.error.DEFAULT
    default:
      return theme.gray.DEFAULT
  }
}

const stateIcon = (state: ConnectionState): string => {
  switch (state) {
    case 'connected':
      return 'icon-check'
    case 'connecting':
      return 'icon-loading'
    case 'not_connected':
      return 'icon-disconnected'
    case 'authentication_error':
    case 'access_denied':
    case 'configuration_error':
    case 'unknown_error':
      return 'icon-error'
    default:
      return 'icon-error'
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
        <Icon :name="loading ? 'icon-loading' : 'icon-check'" size="18" />
      </button>
    </div>

    <div v-if="loading" class="state-container loading">
      <div class="state-icon">
        <Icon name="icon-loading" size="20" />
      </div>
      <span>Checking connection...</span>
    </div>

    <div v-else-if="error" class="state-container error">
      <div class="state-icon">
        <Icon name="icon-error" size="20" />
      </div>
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
          <Icon :name="stateIcon(connection.state)" size="14" />
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
      <div class="state-icon">
        <Icon name="icon-disconnected" size="20" />
      </div>
      <span>No connection data available</span>
      <button @click="loadConnectionStatus" class="retry-btn">Retry</button>
    </div>
  </div>
</template>

<style scoped>
.cloud-connection {
  padding: 16px;
  border-bottom: 1px solid var(--color-border);
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
  color: var(--color-text-primary);
}

.refresh-btn {
  width: 28px;
  height: 28px;
  border: 1px solid var(--border, #ddd);
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(14px * var(--app-font-scale));
  border: 1px solid var(--color-border-lighter);
  background: var(--color-white);
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.refresh-btn:hover:not(:disabled) {
  background: var(--surface-alt, #f5f5f5);
  background: var(--color-bg-soft);
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
  background: var(--color-bg-soft);
  border-radius: 8px;
  font-size: calc(13px * var(--app-font-scale));
}

.state-container.loading {
  color: var(--color-text-secondary);
}

.state-container.error {
  background: var(--color-error-bg);
  color: var(--color-error-dark);
  flex-wrap: wrap;
}

.state-container.empty {
  color: var(--color-text-secondary);
}

.state-icon {
  font-size: calc(20px * var(--app-font-scale));
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-white);
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
  background: var(--color-primary);
  color: var(--color-white);
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: calc(12px * var(--app-font-scale));
  white-space: nowrap;
}

.retry-btn:hover {
  background: var(--color-primary-hover);
}

.connection-card {
  background: var(--color-white);
  border: 1px solid var(--color-border);
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
  background: var(--color-bg-soft);
  border-bottom: 1px solid var(--color-border);
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
  color: var(--color-text-secondary);
  font-weight: 500;
}

.value {
  color: var(--color-text-primary);
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
  color: var(--color-text-secondary);
  font-size: 13px;
}

.setup-link {
  color: var(--color-primary);
  text-decoration: none;
  font-size: calc(13px * var(--app-font-scale));
  font-weight: 500;
}

.setup-link:hover {
  text-decoration: underline;
}
</style>
