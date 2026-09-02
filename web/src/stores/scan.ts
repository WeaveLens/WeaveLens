import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Scan, ScanConfig } from '../types'
import { startScan, getScanStatus, getScans, deleteScan } from '../api/client'

export const useScanStore = defineStore('scan', () => {
  const scans = ref<Scan[]>([])
  const currentScan = ref<Scan | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const scansRevision = ref(0)

  const maxHistoryCount = 20

  const activeScan = computed(() => scans.value.find(s => s.status === 'RUNNING') ?? null)

  const historyCount = computed(() => scans.value.length)
  const isHistoryFull = computed(() => scans.value.length >= maxHistoryCount)

  async function fetchScans() {
    loading.value = true
    const requestRevision = scansRevision.value
    try {
      const data = await getScans()
      if (requestRevision !== scansRevision.value) return
      setScans(data)
      if (currentScan.value && !data.find(s => s.id === currentScan.value!.id)) {
        currentScan.value = null
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch scans'
    } finally {
      loading.value = false
    }
  }

  function setScans(data: Scan[]) {
    scansRevision.value += 1
    scans.value = data
  }

  async function createScan(config: ScanConfig) {
    loading.value = true
    error.value = null
    try {
      const scan = await startScan(config)
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

  async function removeScan(scanId: string) {
    try {
      await deleteScan(scanId)
      scans.value = scans.value.filter(s => s.id !== scanId)
      if (currentScan.value?.id === scanId) {
        currentScan.value = null
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to delete scan'
      throw e
    }
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
    historyCount,
    maxHistoryCount,
    isHistoryFull,
    fetchScans,
    setScans,
    createScan,
    refreshStatus,
    selectScan,
    removeScan,
    clearError,
  }
})
