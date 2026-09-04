import type { Scan, ScanConfig, Graph, Resource, Relationship, ConnectionStatus } from '../types'

const API_BASE = '/api'

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Unknown error' }))
    throw new Error((error as { message: string }).message)
  }
  return response.json()
}

export async function startScan(config: ScanConfig): Promise<Scan> {
  const response = await fetch(`${API_BASE}/scans`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  })
  return handleResponse<Scan>(response)
}

export async function getScanStatus(scanId: string): Promise<Scan> {
  const response = await fetch(`${API_BASE}/scans/${encodeURIComponent(scanId)}/status`)
  return handleResponse<Scan>(response)
}

export async function deleteScan(scanId: string): Promise<void> {
  const response = await fetch(`${API_BASE}/scans/${encodeURIComponent(scanId)}`, {
    method: 'DELETE',
  })
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Unknown error' }))
    throw new Error((error as { message: string }).message)
  }
}

export async function setScanPinned(scanId: string, pinned: boolean): Promise<void> {
  const response = await fetch(`${API_BASE}/scans/${encodeURIComponent(scanId)}/pin`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pinned }),
  })
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Unknown error' }))
    throw new Error((error as { message: string }).message)
  }
}

export async function setScanLocked(scanId: string, locked: boolean): Promise<void> {
  const response = await fetch(`${API_BASE}/scans/${encodeURIComponent(scanId)}/lock`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ locked }),
  })
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Unknown error' }))
    throw new Error((error as { message: string }).message)
  }
}

export async function clearUnpinnedScans(): Promise<number> {
  const response = await fetch(`${API_BASE}/scans/clear-unpinned`, {
    method: 'POST',
  })
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Unknown error' }))
    throw new Error((error as { message: string }).message)
  }
  const data = await response.json().catch(() => ({ removed: 0 }))
  return (data as { removed: number }).removed
}

export async function getGraph(scanId: string): Promise<Graph> {
  const response = await fetch(`${API_BASE}/scans/${encodeURIComponent(scanId)}/graph`)
  return handleResponse<Graph>(response)
}

export async function getResource(resourceId: string): Promise<Resource> {
  const response = await fetch(`${API_BASE}/resources/${encodeURIComponent(resourceId)}`)
  return handleResponse<Resource>(response)
}

export async function getRelationships(resourceId: string): Promise<Relationship[]> {
  const response = await fetch(`${API_BASE}/resources/${encodeURIComponent(resourceId)}/relationships`)
  return handleResponse<Relationship[]>(response)
}

export async function getScans(): Promise<Scan[]> {
  const response = await fetch(`${API_BASE}/scans`)
  return handleResponse<Scan[]>(response)
}

export interface RegionInfo {
  value: string
  label: string
}

export async function getRegions(): Promise<RegionInfo[]> {
  const response = await fetch(`${API_BASE}/regions`)
  return handleResponse<RegionInfo[]>(response)
}

export async function getConnectionStatus(): Promise<ConnectionStatus> {
  const response = await fetch(`${API_BASE}/connection`)
  return handleResponse<ConnectionStatus>(response)
}

type ScanListener = (scans: Scan[]) => void

class ScanStreamManager {
  private eventSource: EventSource | null = null
  private listeners: Set<ScanListener> = new Set()
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null

  subscribe(onUpdate: ScanListener): () => void {
    this.listeners.add(onUpdate)

    if (this.listeners.size === 1) {
      this.connect()
    } else {
      getScans().then((scans) => this.notify(scans)).catch(() => {})
    }

    return () => {
      this.listeners.delete(onUpdate)
      if (this.listeners.size === 0) {
        this.disconnect()
      }
    }
  }

  private connect() {
    if (this.eventSource) return

    this.eventSource = new EventSource(`${API_BASE}/scans/stream`)

    this.eventSource.onmessage = (event) => {
      try {
        const scans = JSON.parse(event.data) as Scan[]
        this.notify(scans)
      } catch {
        // Ignore parse errors
      }
    }

    this.eventSource.onerror = () => {
      this.eventSource?.close()
      this.eventSource = null
      this.scheduleReconnect()
    }
  }

  private disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.eventSource?.close()
    this.eventSource = null
  }

  private scheduleReconnect() {
    if (this.reconnectTimer) return
    if (this.listeners.size === 0) return
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, 2000)
  }

  private notify(scans: Scan[]) {
    this.listeners.forEach((fn) => fn(scans))
  }
}

const scanStreamManager = new ScanStreamManager()

export function subscribeScans(onUpdate: (scans: Scan[]) => void): () => void {
  return scanStreamManager.subscribe(onUpdate)
}
