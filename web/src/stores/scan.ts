import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import type { Scan, ScanConfig } from '../types'
import { startScan, getScanStatus, getScans, deleteScan, setScanPinned, clearUnpinnedScans } from '../api/client'

const PIN_STORAGE_KEY = 'weavelens.pinnedScans'

function loadPinnedIds(): Set<string> {
  try {
    const raw = localStorage.getItem(PIN_STORAGE_KEY)
    if (!raw) return new Set()
    const arr = JSON.parse(raw) as unknown
    if (Array.isArray(arr)) {
      return new Set(arr.filter((x): x is string => typeof x === 'string'))
    }
    return new Set()
  } catch {
    return new Set()
  }
}

function savePinnedIds(ids: Set<string>) {
  try {
    localStorage.setItem(PIN_STORAGE_KEY, JSON.stringify([...ids]))
  } catch {
  }
}

function applyPins(scans: Scan[], pinnedIds: Set<string>): Scan[] {
  return scans.map(s => ({ ...s, pinned: pinnedIds.has(s.id) }))
}

export const useScanStore = defineStore('scan', () => {
  const pinnedIds = ref<Set<string>>(loadPinnedIds())
  const scans = ref<Scan[]>([])
  const currentScan = ref<Scan | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const scansRevision = ref(0)

  const maxHistoryCount = 20

  const activeScan = computed(() => scans.value.find(s => s.status === 'RUNNING') ?? null)

  const historyCount = computed(() => scans.value.length)
  const isHistoryFull = computed(() => scans.value.length >= maxHistoryCount)

  watch(pinnedIds, (val) => {
    savePinnedIds(val)
    scans.value = scans.value.map(s => ({ ...s, pinned: val.has(s.id) }))
  }, { deep: true })

  async function fetchScans() {
    loading.value = true
    const requestRevision = scansRevision.value
    try {
      const data = await getScans()
      if (requestRevision !== scansRevision.value) return
      setScans(applyPins(data, pinnedIds.value))
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
      const enriched: Scan = { ...scan, pinned: pinnedIds.value.has(scanId) }
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
    const current = pinned ?? !pinnedIds.value.has(scanId)
    try {
      await setScanPinned(scanId, current)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to set pinned state'
      throw e
    }
    const next = new Set(pinnedIds.value)
    if (current) {
      next.add(scanId)
    } else {
      next.delete(scanId)
    }
    pinnedIds.value = next
    const idx = scans.value.findIndex(s => s.id === scanId)
    if (idx >= 0) {
      scans.value[idx] = { ...scans.value[idx], pinned: current }
    }
  }

  async function clearUnpinned() {
    try {
      const removed = await clearUnpinnedScans()
      const pinnedSet = pinnedIds.value
      scans.value = scans.value.filter(s => pinnedSet.has(s.id))
      if (currentScan.value && !pinnedSet.has(currentScan.value.id)) {
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
    clearUnpinned,
    clearError,
  }
})
