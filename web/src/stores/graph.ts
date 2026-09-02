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
  const regionFilter = ref<string[]>([])
  const typeFilter = ref<string[]>([])
  const tagFilter = ref<string[]>([])
  const filterRules = ref<FilterRule[]>([])

  interface FilterRule {
    id: string
    field: string
    value: string
  }

  const categories = computed(() => {
    const cats = new Set(nodes.value.map(n => n.category))
    return Array.from(cats).sort()
  })

  const regions = computed(() => {
    const regs = new Set(nodes.value.map(n => n.region).filter(Boolean))
    return Array.from(regs).sort() as string[]
  })

  const types = computed(() => {
    const t = new Set(nodes.value.map(n => n.type).filter(Boolean))
    return Array.from(t).sort() as string[]
  })

  const availableTags = computed(() => {
    const tags = new Set<string>()
    nodes.value.forEach(n => {
      Object.entries(n.tags || {}).forEach(([k, v]) => tags.add(`${k}=${v}`))
    })
    return Array.from(tags).sort()
  })

  const availableTagKeys = computed(() => {
    const keys = new Set<string>()
    nodes.value.forEach(n => {
      Object.keys(n.tags || {}).forEach(k => keys.add(k))
    })
    return Array.from(keys).sort()
  })

  function getTagValues(tagKey: string): string[] {
    const values = new Set<string>()
    nodes.value.forEach(n => {
      const val = n.tags?.[tagKey]
      if (val) values.add(val)
    })
    return Array.from(values).sort()
  }

  const availableFields = computed(() => {
    const fields = new Set<string>()
    fields.add('category')
    fields.add('name')
    nodes.value.forEach(n => {
      Object.keys(n.metadata || {}).forEach(k => fields.add(`meta:${k}`))
    })
    return Array.from(fields).sort()
  })

  function getAvailableValues(field: string): string[] {
    const values = new Set<string>()
    nodes.value.forEach(n => {
      if (field === 'category') {
        values.add(n.category)
      } else if (field === 'name') {
        values.add(n.name)
      } else if (field.startsWith('meta:')) {
        const metaKey = field.slice(5)
        const metaValue = n.metadata?.[metaKey]
        if (metaValue) values.add(metaValue)
      }
    })
    return Array.from(values).sort()
  }

  const filteredNodes = computed(() => {
    let result = nodes.value
    if (categoryFilter.value) {
      result = result.filter(n => n.category === categoryFilter.value)
    }
    if (regionFilter.value.length > 0) {
      result = result.filter(n => regionFilter.value.includes(n.region))
    }
    if (typeFilter.value.length > 0) {
      result = result.filter(n => typeFilter.value.includes(n.type))
    }
    if (tagFilter.value.length > 0) {
      const grouped = new Map<string, string[]>()
      tagFilter.value.forEach(t => {
        const [k, ...rest] = t.split('=')
        const v = rest.join('=')
        const arr = grouped.get(k) || []
        arr.push(v)
        grouped.set(k, arr)
      })
      result = result.filter(n => {
        for (const [key, values] of grouped) {
          const matches = values.some(v => n.tags?.[key] === v)
          if (!matches) return false
        }
        return true
      })
    }
    if (filterRules.value.length > 0) {
      const grouped = new Map<string, FilterRule[]>()
      filterRules.value.forEach(rule => {
        const arr = grouped.get(rule.field) || []
        arr.push(rule)
        grouped.set(rule.field, arr)
      })
      result = result.filter(n => {
        for (const [field, rules] of grouped) {
          const matches = rules.some(rule => {
            const value = rule.value.toLowerCase()
            if (field === 'category') return n.category?.toLowerCase() === value
            if (field === 'name') return n.name?.toLowerCase().includes(value)
            if (field.startsWith('meta:')) {
              const metaKey = field.slice(5)
              return n.metadata?.[metaKey]?.toLowerCase() === value
            }
            return true
          })
          if (!matches) return false
        }
        return true
      })
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

  function setRegionFilter(regions: string[]) {
    regionFilter.value = regions
  }

  function setTypeFilter(types: string[]) {
    typeFilter.value = types
  }

  function setTagFilter(tags: string[]) {
    tagFilter.value = tags
  }

  function addFilterRule(field: string, value: string) {
    filterRules.value.push({ id: `${Date.now()}`, field, value })
  }

  function removeFilterRule(id: string) {
    filterRules.value = filterRules.value.filter(r => r.id !== id)
  }

  function clearFilterRules() {
    filterRules.value = []
  }

  function clearSelection() {
    selectedResource.value = null
  }

  function clearGraph() {
    nodes.value = []
    edges.value = []
    selectedResource.value = null
    error.value = null
  }

  return {
    nodes,
    edges,
    selectedResource,
    loading,
    error,
    searchQuery,
    categoryFilter,
    regionFilter,
    typeFilter,
    tagFilter,
    filterRules,
    categories,
    regions,
    types,
    availableTags,
    availableTagKeys,
    getTagValues,
    availableFields,
    getAvailableValues,
    filteredNodes,
    filteredEdges,
    selectedRelationships,
    loadGraph,
    loadResourceDetail,
    loadRelationships,
    selectResource,
    setSearch,
    setCategoryFilter,
    setRegionFilter,
    setTypeFilter,
    setTagFilter,
    addFilterRule,
    removeFilterRule,
    clearFilterRules,
    clearSelection,
    clearGraph,
  }
})

export function useGraphQuery(scanId: string) {
  return useQuery({
    queryKey: ['graph', scanId],
    queryFn: () => getGraph(scanId),
    enabled: !!scanId,
  })
}
