import type { Scan, ScanConfig, Graph, Resource, Relationship } from '../types'

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
