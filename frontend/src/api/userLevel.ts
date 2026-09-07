import { apiClient } from './client'

export interface UserLevelDashboard {
  user_id: string
  level: number
  configured: boolean
  usage_7d: number
  window_hours: number
  window_from: string
  calculated_at: string
  rules: Array<{
    rule_id: string
    rule_name: string
    window_days: number
    enabled: boolean
    spend: number
    window_from: string
    calculated_at: string
    current_tier_id?: string
    current_tier_name?: string
    current_tier_order: number
    min_spend: number
    default_multiplier?: number | null
    tiers?: Array<{
      id: string
      rule_id: string
      name: string
      sort_order: number
      min_spend: number
      default_multiplier?: number | null
    }>
  }>
  current_tier_ids: string[]
  user_level_multiplier?: number | null
  group_rule_multiplier?: number | null
  effective_source?: string
  /** @deprecated Fixed thresholds are no longer used. */
  l2_min_spend?: number
  /** @deprecated Fixed thresholds are no longer used. */
  l3_min_spend?: number
  level_multiplier?: number | null
  effective_multiplier?: number | null
  multiplier_group?: string
  next_level_multiplier?: number | null
  next_multiplier_group?: string
}

export async function getCurrent(): Promise<UserLevelDashboard> {
  const { data } = await apiClient.get<UserLevelDashboard>('/user/level')
  return data
}

export const userLevelAPI = {
  getCurrent,
}

export default userLevelAPI
