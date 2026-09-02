/**
 * Admin Affiliate API endpoints
 * Manage per-user affiliate (邀请返利) configurations:
 * exclusive invite codes (overrides aff_code) and exclusive rebate rates.
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export interface AffiliateAdminEntry {
  user_id: string
  email: string
  username: string
  aff_code: string
  aff_code_custom: boolean
  aff_rebate_rate_percent?: number | null
  aff_count: number
}

export interface ListAffiliateUsersParams {
  page?: number
  page_size?: number
  search?: string
}

export interface ListAffiliateRecordsParams {
  page?: number
  page_size?: number
  search?: string
  start_at?: string
  end_at?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
  timezone?: string
}

export interface AffiliateInviteRecord {
  inviter_id: string
  inviter_email: string
  inviter_username: string
  invitee_id: string
  invitee_email: string
  invitee_username: string
  aff_code: string
  total_rebate: number
  created_at: string
}

export interface AffiliateRebateRecord {
  order_id: string
  out_trade_no: string
  inviter_id: string
  inviter_email: string
  inviter_username: string
  invitee_id: string
  invitee_email: string
  invitee_username: string
  recipient_id: string
  recipient_email: string
  recipient_username: string
  rebate_recipient: 'inviter' | 'invitee'
  order_amount: number
  pay_amount: number
  rebate_amount: number
  reserved_reversal_points: number
  reversed_points: number
  net_rebate_points: number
  payment_type: string
  order_status: string
  created_at: string
}

export interface AffiliateTransferRecord {
  ledger_id: string
  user_id: string
  user_email: string
  username: string
  amount: number
  balance_after?: number | null
  available_quota_after?: number | null
  frozen_quota_after?: number | null
  history_quota_after?: number | null
  snapshot_available: boolean
  created_at: string
}

export interface AffiliateUserOverview {
  user_id: string
  email: string
  username: string
  aff_code: string
  rebate_rate_percent: number
  invited_count: number
  rebated_invitee_count: number
  available_quota: number
  history_quota: number
}

export interface UpdateAffiliateUserRequest {
  aff_code?: string
  aff_rebate_rate_percent?: number | null
  /** Set true to explicitly clear the per-user rate (sets it to NULL). */
  clear_rebate_rate?: boolean
}

export interface BatchSetRateRequest {
  user_ids: string[]
  aff_rebate_rate_percent?: number | null
  /** Set true to clear rates instead of setting. */
  clear?: boolean
}

export interface SimpleUser {
  id: string
  email: string
  username: string
}

export interface AffiliateBindingRewardConfig {
  inviter_points: number
  inviter_validity_days: number
  invitee_points: number
  invitee_validity_days: number
}

export interface AffiliateRewardBackfillPreview {
  config: AffiliateBindingRewardConfig
  eligible_relations: number
  estimated_inviter_grants: number
  estimated_invitee_grants: number
  estimated_inviter_points: number
  estimated_invitee_points: number
  preview_token: string
}

export interface AffiliateRewardBackfillRun {
  id: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  config: AffiliateBindingRewardConfig
  eligible_relations: number
  processed_relations: number
  inviter_grants: number
  invitee_grants: number
  inviter_points_granted: number
  invitee_points_granted: number
  error_message?: string
  created_at: string
  started_at?: string
  completed_at?: string
  updated_at: string
}

export async function listUsers(
  params: ListAffiliateUsersParams = {},
): Promise<PaginatedResponse<AffiliateAdminEntry>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateAdminEntry>>(
    '/admin/affiliates/users',
    {
      params: {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        search: params.search ?? '',
      },
    },
  )
  return data
}

export async function lookupUsers(q: string): Promise<SimpleUser[]> {
  const { data } = await apiClient.get<SimpleUser[]>(
    '/admin/affiliates/users/lookup',
    { params: { q } },
  )
  return data
}

export async function updateUserSettings(
  userId: string,
  payload: UpdateAffiliateUserRequest,
): Promise<{ user_id: string }> {
  const { data } = await apiClient.put<{ user_id: string }>(
    `/admin/affiliates/users/${userId}`,
    payload,
  )
  return data
}

export async function clearUserSettings(
  userId: string,
): Promise<{ user_id: string }> {
  const { data } = await apiClient.delete<{ user_id: string }>(
    `/admin/affiliates/users/${userId}`,
  )
  return data
}

export async function batchSetRate(
  payload: BatchSetRateRequest,
): Promise<{ affected: number }> {
  const { data } = await apiClient.post<{ affected: number }>(
    '/admin/affiliates/users/batch-rate',
    payload,
  )
  return data
}

function recordParams(params: ListAffiliateRecordsParams = {}) {
  return {
    page: params.page ?? 1,
    page_size: params.page_size ?? 20,
    search: params.search ?? '',
    start_at: params.start_at || undefined,
    end_at: params.end_at || undefined,
    sort_by: params.sort_by || undefined,
    sort_order: params.sort_order || undefined,
    timezone: params.timezone || undefined,
  }
}

export async function listInviteRecords(
  params: ListAffiliateRecordsParams = {},
): Promise<PaginatedResponse<AffiliateInviteRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateInviteRecord>>(
    '/admin/affiliates/invites',
    { params: recordParams(params) },
  )
  return data
}

export async function listRebateRecords(
  params: ListAffiliateRecordsParams = {},
): Promise<PaginatedResponse<AffiliateRebateRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateRebateRecord>>(
    '/admin/affiliates/rebates',
    { params: recordParams(params) },
  )
  return data
}

export async function listTransferRecords(
  params: ListAffiliateRecordsParams = {},
): Promise<PaginatedResponse<AffiliateTransferRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateTransferRecord>>(
    '/admin/affiliates/transfers',
    { params: recordParams(params) },
  )
  return data
}

export async function getUserOverview(
  userId: string,
): Promise<AffiliateUserOverview> {
  const { data } = await apiClient.get<AffiliateUserOverview>(
    `/admin/affiliates/users/${userId}/overview`,
  )
  return data
}

export async function previewRewardBackfill(): Promise<AffiliateRewardBackfillPreview> {
  const { data } = await apiClient.get<AffiliateRewardBackfillPreview>(
    '/admin/affiliates/reward-backfill/preview',
  )
  return data
}

export async function startRewardBackfill(
  previewToken: string,
): Promise<AffiliateRewardBackfillRun> {
  const { data } = await apiClient.post<AffiliateRewardBackfillRun>(
    '/admin/affiliates/reward-backfill',
    { preview_token: previewToken, confirm: true },
  )
  return data
}

export async function getRewardBackfill(id: string): Promise<AffiliateRewardBackfillRun> {
  const { data } = await apiClient.get<AffiliateRewardBackfillRun>(
    `/admin/affiliates/reward-backfill/${id}`,
  )
  return data
}

export const affiliatesAPI = {
  listUsers,
  lookupUsers,
  updateUserSettings,
  clearUserSettings,
  batchSetRate,
  listInviteRecords,
  listRebateRecords,
  listTransferRecords,
  getUserOverview,
  previewRewardBackfill,
  startRewardBackfill,
  getRewardBackfill,
}

export default affiliatesAPI
