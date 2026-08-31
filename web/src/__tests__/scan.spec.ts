import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia } from 'pinia'
import { createTestingPinia } from '@pinia/testing'
import { useScanStore } from '../stores/scan'

describe('scan store', () => {
  beforeEach(() => {
    setActivePinia(createTestingPinia({ stubActions: false }))
  })

  it('initializes with empty state', () => {
    const store = useScanStore()
    expect(store.scans).toEqual([])
    expect(store.currentScan).toBeNull()
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('clears error', () => {
    const store = useScanStore()
    store.error = 'some error'
    store.clearError()
    expect(store.error).toBeNull()
  })

  it('selects scan', () => {
    const store = useScanStore()
    const scan = { id: 'scan-1', status: 'RUNNING', region: 'us-east-1' }
    store.selectScan(scan as any)
    expect(store.currentScan).toEqual(scan)
  })
})
