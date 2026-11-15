import type { ModelMappingEntry } from '@/composables/useModelWhitelist'

export function isValidModelForwarding(rows: ModelMappingEntry[]): boolean {
  const sources = new Set<string>()
  return rows.every(row => {
    const source = row.from.trim()
    const target = row.to.trim()
    if (!/^[^\s*]+\*?$/.test(source) || !/^[^\s*]+$/.test(target) || sources.has(source)) return false
    sources.add(source)
    return true
  })
}
