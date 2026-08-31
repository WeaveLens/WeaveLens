import { describe, it, expect } from 'vitest'
import { CATEGORY_META, getCategoryColor } from '../config/categories'

describe('categories config', () => {
  it('has metadata for all required categories', () => {
    const required = ['compute', 'network', 'database', 'storage', 'security', 'integration', 'other']
    for (const cat of required) {
      expect(CATEGORY_META[cat]).toBeDefined()
      expect(CATEGORY_META[cat].label).toBeTruthy()
      expect(CATEGORY_META[cat].color).toMatch(/^#[0-9a-fA-F]{6}$/)
    }
  })

  it('returns color for known category', () => {
    expect(getCategoryColor('compute')).toBe(CATEGORY_META.compute.color)
  })

  it('returns other color for unknown category', () => {
    expect(getCategoryColor('unknown')).toBe(CATEGORY_META.other.color)
  })
})
