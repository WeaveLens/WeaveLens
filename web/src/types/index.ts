export interface Resource {
  id: string
  name: string
  type: string
  category: string
  arn: string
  region: string
  metadata: Record<string, string>
  tags: Record<string, string>
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
  nodeCount: number
  edgeCount: number
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

export type ConnectionState =
  | 'connected'
  | 'connecting'
  | 'not_connected'
  | 'authentication_error'
  | 'access_denied'
  | 'configuration_error'
  | 'unknown_error'

export interface ConnectionStatus {
  state: ConnectionState
  accountId: string
  arn: string
  region: string
  credentialSource: string
  message: string
}
