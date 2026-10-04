import { apiClient } from '../client'

export interface PointChange {
  id: string
  transaction_id: string
  user_id: string
  email: string
  username: string
  point_type: 'recharge' | 'bonus'
  amount: number
  frozen_amount: number
  balance_before: number | null
  balance_after: number | null
  action: string
  source_type: string
  source_id: string
  notes: string
  created_at: string
}

export interface PointChangeQuery {
  keyword?: string
  user_id?: string
  point_type?: string
  direction?: string
  start_time?: string
  end_time?: string
  page: number
  page_size: number
}

export async function listPointChanges(params: PointChangeQuery) {
  const { data } = await apiClient.get<{ items: PointChange[]; total: number; page: number; page_size: number }>(
    '/admin/users/point-changes', { params }
  )
  return data
}
