import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Scan, ScanConfig } from '../types'
import { startScan, getScanStatus, getScans } from '../api/client'

export const useScanStore = defineStore('scan', () => {
  const scans = ref<Scan[]>([])
  const currentScan = ref<Scan | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const activeScan = computed(() => scans.value.find(s => s.status === 'RUNNING') ?? null)

  async function fetchScans() {
    loading.value = true
    try {
      const data = await getScans()
      scans.value = data
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch scans'
    } finally {
      loading.value = false
    }
  }

  async function createScan(config: ScanConfig) {
    loading.value = true
    error.value = null
    try {
      const scan = await startScan(config)
      await fetchScans()
      currentScan.value = scan
      return scan
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to start scan'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function refreshStatus(scanId: string) {
    try {
      const scan = await getScanStatus(scanId)
      const idx = scans.value.findIndex(s => s.id === scanId)
      if (idx >= 0) {
        scans.value[idx] = scan
      } else {
        scans.value.push(scan)
      }
      if (currentScan.value?.id === scanId) {
        currentScan.value = scan
      }
      return scan
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to refresh scan status'
      throw e
    }
  }

  function selectScan(scan: Scan | null) {
    currentScan.value = scan
  }

  function clearError() {
    error.value = null
  }

  return {
    scans,
    currentScan,
    loading,
    error,
    activeScan,
    fetchScans,
    createScan,
    refreshStatus,
    selectScan,
    clearError,
  }
})
