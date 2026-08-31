export interface Resource {
  id: string
  name: string
  type: string
  category: string
  arn: string
  region: string
  metadata: Record<string, string>
}

export interface Relationship {
  id: string
  sourceId: string
  targetId: string
  type: string
  metadata: Record<string, string>
}

export interface Scan {
  id: string
  status: string
  region: string
  createdAt: string
  updatedAt: string
}

export interface ScanConfig {
  region: string
}

export interface Graph {
  nodes: Resource[]
  edges: Relationship[]
}

export interface ApiError {
  message: string
}
