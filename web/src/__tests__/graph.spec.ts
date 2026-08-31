import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia } from 'pinia'
import { createTestingPinia } from '@pinia/testing'
import { useGraphStore } from '../stores/graph'
import type { Resource, Relationship } from '../types'

describe('graph store', () => {
  beforeEach(() => {
    setActivePinia(createTestingPinia({ stubActions: false }))
  })

  it('initializes with empty state', () => {
    const store = useGraphStore()
    expect(store.nodes).toEqual([])
    expect(store.edges).toEqual([])
    expect(store.selectedResource).toBeNull()
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('computes categories from nodes', () => {
    const store = useGraphStore()
    store.nodes = [
      { id: '1', name: 'a', type: 'EC2', category: 'compute', arn: '', region: '', metadata: {} },
      { id: '2', name: 'b', type: 'VPC', category: 'network', arn: '', region: '', metadata: {} },
      { id: '3', name: 'c', type: 'RDS', category: 'database', arn: '', region: '', metadata: {} },
    ] as Resource[]
    expect(store.categories).toEqual(['compute', 'database', 'network'])
  })

  it('filters nodes by category', () => {
    const store = useGraphStore()
    store.nodes = [
      { id: '1', name: 'a', type: 'EC2', category: 'compute', arn: '', region: '', metadata: {} },
      { id: '2', name: 'b', type: 'VPC', category: 'network', arn: '', region: '', metadata: {} },
    ] as Resource[]
    store.setCategoryFilter('compute')
    expect(store.filteredNodes.map(n => n.id)).toEqual(['1'])
  })

  it('filters nodes by search query', () => {
    const store = useGraphStore()
    store.nodes = [
      { id: '1', name: 'web-server', type: 'EC2', category: 'compute', arn: '', region: '', metadata: {} },
      { id: '2', name: 'db-server', type: 'RDS', category: 'database', arn: '', region: '', metadata: {} },
    ] as Resource[]
    store.setSearch('web')
    expect(store.filteredNodes.map(n => n.id)).toEqual(['1'])
  })

  it('filters edges when nodes are filtered', () => {
    const store = useGraphStore()
    store.nodes = [
      { id: '1', name: 'a', type: 'EC2', category: 'compute', arn: '', region: '', metadata: {} },
      { id: '2', name: 'b', type: 'VPC', category: 'network', arn: '', region: '', metadata: {} },
    ] as Resource[]
    store.edges = [
      { id: 'e1', sourceId: '1', targetId: '2', type: 'in', metadata: {} },
    ] as Relationship[]
    store.setCategoryFilter('compute')
    expect(store.filteredEdges).toEqual([])
  })
})
