import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Scan, ScanConfig } from '../types'
import { startScan, getScanStatus, getScans, deleteScan, setScanPinned, setScanLocked, clearUnpinnedScans } from '../api/client'

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
      currentScan.value = { ...scan, pinned: false }
      return currentScan.value
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
      const existing = scans.value.find(s => s.id === scanId)
      const enriched: Scan = { ...scan, pinned: existing?.pinned ?? scan.pinned ?? false, locked: existing?.locked ?? scan.locked ?? false, layout: existing?.layout ?? scan.layout }
      const idx = scans.value.findIndex(s => s.id === scanId)
      if (idx >= 0) {
        scans.value[idx] = enriched
      } else {
        scans.value.push(enriched)
      }
      if (currentScan.value?.id === scanId) {
        currentScan.value = enriched
      }
      return enriched
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

  async function togglePin(scanId: string, pinned?: boolean) {
    const scan = scans.value.find(s => s.id === scanId)
    const target = pinned ?? !(scan?.pinned ?? false)
    const prev = scan?.pinned ?? false
    if (scan) {
      scan.pinned = target
    }
    try {
      await setScanPinned(scanId, target)
    } catch (e) {
      if (scan) scan.pinned = prev
      error.value = e instanceof Error ? e.message : 'Failed to set pinned state'
      throw e
    }
  }

  async function toggleLocked(scanId: string, locked: boolean) {
    const scan = scans.value.find(s => s.id === scanId)
    const prev = scan?.locked ?? false
    if (scan) {
      scan.locked = locked
    }
    if (currentScan.value?.id === scanId) {
      currentScan.value = { ...currentScan.value, locked }
    }
    try {
      await setScanLocked(scanId, locked)
    } catch (e) {
      if (scan) scan.locked = prev
      if (currentScan.value?.id === scanId) {
        currentScan.value = { ...currentScan.value, locked: prev }
      }
      error.value = e instanceof Error ? e.message : 'Failed to set locked state'
      throw e
    }
  }

  async function clearUnpinned() {
    try {
      const removed = await clearUnpinnedScans()
      const kept = new Set(
        scans.value.filter(s => s.pinned).map(s => s.id)
      )
      scans.value = scans.value.filter(s => s.pinned)
      if (currentScan.value && !kept.has(currentScan.value.id)) {
        currentScan.value = null
      }
      return removed
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to clear unpinned scans'
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
    togglePin,
    toggleLocked,
    clearUnpinned,
    clearError,
  }
})
