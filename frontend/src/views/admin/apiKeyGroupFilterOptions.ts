import type { AdminGroup } from '@/types'

export interface ApiKeyGroupFilterOption {
  value: string | null
  label: string
  kind?: 'group'
  disabled?: boolean
}

export interface ApiKeyGroupFilterLabels {
  all: string
  exclusive: string
  public: string
  subscription: string
  disabled: string
}

// Sentinel values for section-header rows (prefixed so they never collide with UUID group ids).
// Select.vue generates :key from `${typeof value}:${String(value ?? '')}` — using distinct
// strings avoids the duplicate "object:" keys that null-valued headers would produce.
const HEADER_EXCLUSIVE = '__group_section_exclusive__'
const HEADER_PUBLIC = '__group_section_public__'
const HEADER_SUBSCRIPTION = '__group_section_subscription__'
const HEADER_DISABLED = '__group_section_disabled__'

/**
 * Build options for the "API Key group" filter Select.
 *
 * Active groups are partitioned into exclusive / public / subscription sections,
 * each preceded by a disabled section-header row. Disabled groups are collected
 * into a final "disabled" section so admins can filter users whose keys are still
 * bound to a now-disabled group. Empty sections render no header. The leading
 * "all" option (value null) clears the filter.
 *
 * Section-header rows use prefixed sentinel strings instead of null so
 * that Vue's v-for :key expression produces distinct strings and avoids duplicate-
 * key warnings (fixes F2).
 */
export function buildApiKeyGroupFilterOptions(
  groups: AdminGroup[],
  labels: ApiKeyGroupFilterLabels
): ApiKeyGroupFilterOption[] {
  const exclusive: ApiKeyGroupFilterOption[] = []
  const publicGroups: ApiKeyGroupFilterOption[] = []
  const subscription: ApiKeyGroupFilterOption[] = []
  const disabledGroups: ApiKeyGroupFilterOption[] = []

  for (const grp of groups) {
    const item: ApiKeyGroupFilterOption = { value: grp.id, label: grp.name }
    if (grp.status !== 'active') {
      disabledGroups.push(item)
    } else if (grp.subscription_type === 'subscription') {
      subscription.push(item)
    } else if (grp.is_exclusive) {
      exclusive.push(item)
    } else {
      publicGroups.push(item)
    }
  }

  const options: ApiKeyGroupFilterOption[] = [{ value: null, label: labels.all }]

  const sections: Array<[string, string, ApiKeyGroupFilterOption[]]> = [
    [labels.exclusive,    HEADER_EXCLUSIVE,    exclusive],
    [labels.public,       HEADER_PUBLIC,       publicGroups],
    [labels.subscription, HEADER_SUBSCRIPTION, subscription],
    [labels.disabled,     HEADER_DISABLED,     disabledGroups],
  ]
  for (const [label, headerValue, items] of sections) {
    if (items.length === 0) continue
    options.push({ value: headerValue, label, kind: 'group', disabled: true })
    options.push(...items)
  }
  return options
}
