import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Resource, Relationship } from '../types'
import { getGraph, getResource, getRelationships } from '../api/client'
import { useQuery } from '@tanstack/vue-query'

export const useGraphStore = defineStore('graph', () => {
  const nodes = ref<Resource[]>([])
  const edges = ref<Relationship[]>([])
  const selectedResource = ref<Resource | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const searchQuery = ref('')
  const categoryFilter = ref<string | null>(null)

  const categories = computed(() => {
    const cats = new Set(nodes.value.map(n => n.category))
    return Array.from(cats).sort()
  })

  const filteredNodes = computed(() => {
    let result = nodes.value
    if (categoryFilter.value) {
      result = result.filter(n => n.category === categoryFilter.value)
    }
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      result = result.filter(n =>
        n.name.toLowerCase().includes(q) ||
        n.type.toLowerCase().includes(q) ||
        n.id.toLowerCase().includes(q)
      )
    }
    return result
  })

  const filteredEdges = computed(() => {
    const nodeIds = new Set(filteredNodes.value.map(n => n.id))
    return edges.value.filter(e => nodeIds.has(e.sourceId) && nodeIds.has(e.targetId))
  })

  const selectedRelationships = computed(() => {
    if (!selectedResource.value) return []
    return edges.value.filter(
      e => e.sourceId === selectedResource.value!.id || e.targetId === selectedResource.value!.id
    )
  })

  async function loadGraph(scanId: string) {
    loading.value = true
    error.value = null
    try {
      const data = await getGraph(scanId)
      nodes.value = data.nodes
      edges.value = data.edges
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load graph'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function loadResourceDetail(resourceId: string) {
    try {
      const resource = await getResource(resourceId)
      selectedResource.value = resource
      return resource
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load resource'
      throw e
    }
  }

  async function loadRelationships(resourceId: string) {
    try {
      return await getRelationships(resourceId)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load relationships'
      throw e
    }
  }

  function selectResource(resource: Resource | null) {
    selectedResource.value = resource
  }

  function setSearch(query: string) {
    searchQuery.value = query
  }

  function setCategoryFilter(category: string | null) {
    categoryFilter.value = category
  }

  function clearSelection() {
    selectedResource.value = null
  }

  return {
    nodes,
    edges,
    selectedResource,
    loading,
    error,
    searchQuery,
    categoryFilter,
    categories,
    filteredNodes,
    filteredEdges,
    selectedRelationships,
    loadGraph,
    loadResourceDetail,
    loadRelationships,
    selectResource,
    setSearch,
    setCategoryFilter,
    clearSelection,
  }
})

export function useGraphQuery(scanId: string) {
  return useQuery({
    queryKey: ['graph', scanId],
    queryFn: () => getGraph(scanId),
    enabled: !!scanId,
  })
}
