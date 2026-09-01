import { apiClient } from './client'

export interface UserLevelDashboard {
  user_id: string
  level: 1 | 2 | 3
  usage_7d: number
  window_hours: number
  window_from: string
  calculated_at: string
  l2_min_spend: number
  l3_min_spend: number
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
