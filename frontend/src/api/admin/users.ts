/**
 * Admin Users API endpoints
 * Handles user management for administrators
 */

import { apiClient } from '../client'
import type { AdminUser, UpdateUserRequest, PaginatedResponse, ApiKey } from '@/types'

export interface AdminBindAuthIdentityChannelRequest {
  channel: string
  channel_app_id: string
  channel_subject: string
  metadata?: Record<string, unknown> | null
}

export interface AdminBindAuthIdentityRequest {
  provider_type: string
  provider_key: string
  provider_subject: string
  issuer?: string | null
  metadata?: Record<string, unknown> | null
  channel?: AdminBindAuthIdentityChannelRequest
}

export interface AdminBoundAuthIdentityChannel {
  channel: string
  channel_app_id: string
  channel_subject: string
  metadata: Record<string, unknown> | null
  created_at: string
  updated_at: string
}

export interface AdminBoundAuthIdentity {
  user_id: string
  provider_type: string
  provider_key: string
  provider_subject: string
  verified_at?: string | null
  issuer?: string | null
  metadata: Record<string, unknown> | null
  created_at: string
  updated_at: string
  channel?: AdminBoundAuthIdentityChannel | null
}

export interface BatchUpdateUserLimitsRequest {
  user_ids: string[]
  all?: boolean
  concurrency?: number
  rpm_limit?: number
}

export interface BatchUpdateUserLimitsResponse {
  affected: number
}

export interface InactiveUserFilterRequest {
  max_balance: number
  last_used_before: string
  max_usage_7d: number
}

export interface InactiveUserCandidate {
  id: string
  email: string
  balance: number
  last_used_at: string | null
  usage_7d: number
  created_at: string
}

export interface InactiveUserDeletePreview {
  total: number
  total_balance: number
  total_usage_7d: number
  generated_at: string
  snapshot_token: string
  items: InactiveUserCandidate[]
}

export interface PermanentlyDeleteInactiveUsersRequest extends InactiveUserFilterRequest {
  expected_count: number
  snapshot_token: string
  confirmation: string
}

export interface PermanentlyDeleteInactiveUsersResponse {
  deleted: number
}

export type UserDeletionMode = 'archived' | 'permanently_deleted'

export interface DeleteUserResponse {
  message: string
  mode: UserDeletionMode
}

export interface UserLevelSettings {
	l2_min_spend: number
	l3_min_spend: number
	window_hours: number
}

export interface UserLevelProfile {
	user_id: string
	level: 1 | 2 | 3
	usage_7d: number
	window_from: string
	calculated_at: string
}

export async function getLevelSettings(): Promise<UserLevelSettings> {
	const { data } = await apiClient.get<UserLevelSettings>('/admin/users/level-settings')
	return data
}

export async function updateLevelSettings(input: Pick<UserLevelSettings, 'l2_min_spend' | 'l3_min_spend'>): Promise<UserLevelSettings> {
	const { data } = await apiClient.put<UserLevelSettings>('/admin/users/level-settings', input)
	return data
}

export async function getLevelProfiles(userIds: string[]): Promise<UserLevelProfile[]> {
	if (userIds.length === 0) return []
	const { data } = await apiClient.post<UserLevelProfile[]>('/admin/users/levels/batch', { user_ids: userIds })
	return data
}

/**
 * List all users with pagination
 * @param page - Page number (default: 1)
 * @param pageSize - Items per page (default: 20)
 * @param filters - Optional filters (status, role, search, attributes)
 * @param options - Optional request options (signal)
 * @returns Paginated list of users
 */
export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    status?: 'active' | 'disabled'
    role?: 'admin' | 'user'
    search?: string
    group_name?: string         // fuzzy filter by allowed group name
    api_key_group_id?: string   // filter users by the group their API keys are bound to
    attributes?: Record<string, string>  // attributeId -> value
    include_subscriptions?: boolean
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<PaginatedResponse<AdminUser>> {
  // Build params with attribute filters in attr[id]=value format
  const params: Record<string, any> = {
    page,
    page_size: pageSize,
    status: filters?.status,
    role: filters?.role,
    search: filters?.search,
    group_name: filters?.group_name,
    api_key_group_id: filters?.api_key_group_id,
    include_subscriptions: filters?.include_subscriptions,
    sort_by: filters?.sort_by,
    sort_order: filters?.sort_order
  }

  // Add attribute filters as attr[id]=value
  if (filters?.attributes) {
    for (const [attrId, value] of Object.entries(filters.attributes)) {
      if (value) {
        params[`attr[${attrId}]`] = value
      }
    }
  }
  const { data } = await apiClient.get<PaginatedResponse<AdminUser>>('/admin/users', {
    params,
    signal: options?.signal
  })
  return data
}

export async function listArchived(
  page: number = 1,
  pageSize: number = 20,
  search?: string
): Promise<PaginatedResponse<AdminUser>> {
  const { data } = await apiClient.get<PaginatedResponse<AdminUser>>('/admin/users/archived', {
    params: { page, page_size: pageSize, search: search || undefined }
  })
  return data
}

/**
 * Get user by ID
 * @param id - User ID
 * @param includeDeleted - Whether to include soft-deleted users
 * @returns User details
 */
export async function getById(id: string, includeDeleted = false): Promise<AdminUser> {
  const url = includeDeleted ? `/admin/users/${id}?include_deleted=true` : `/admin/users/${id}`
  const { data } = await apiClient.get<AdminUser>(url)
  return data
}

/**
 * Create new user
 * @param userData - User data (email, password, etc.)
 * @returns Created user
 */
export async function create(userData: {
  email: string
  password: string
  username?: string
  notes?: string
  role?: 'admin' | 'user'
  balance?: number
	balance_type?: 'recharge' | 'bonus'
	bonus_validity_days?: number
  concurrency?: number
  rpm_limit?: number
  allowed_groups?: string[] | null
}): Promise<AdminUser> {
  const { data } = await apiClient.post<AdminUser>('/admin/users', userData)
  return data
}

/**
 * Update user
 * @param id - User ID
 * @param updates - Fields to update
 * @returns Updated user
 */
export async function update(id: string, updates: UpdateUserRequest): Promise<AdminUser> {
  const { data } = await apiClient.put<AdminUser>(`/admin/users/${id}`, updates)
  return data
}

/**
 * Delete user
 * @param id - User ID
 * @returns Success confirmation
 */
export async function deleteUser(id: string): Promise<DeleteUserResponse> {
  const { data } = await apiClient.delete<DeleteUserResponse>(`/admin/users/${id}`)
  return data
}

export async function restoreArchivedUser(id: string): Promise<AdminUser> {
  const { data } = await apiClient.post<AdminUser>(`/admin/users/${id}/restore`)
  return data
}

/**
 * Update user balance
 * @param id - User ID
 * @param balance - New balance
 * @param operation - Operation type ('set', 'add', 'subtract')
 * @param notes - Optional notes for the balance adjustment
 * @returns Updated user
 */
export async function updateBalance(
  id: string,
  balance: number,
  operation: 'set' | 'add' | 'subtract' = 'set',
  notes?: string,
  balanceType: 'recharge' | 'bonus' = 'recharge',
  bonusValidityDays = 90,
): Promise<AdminUser> {
  const { data } = await apiClient.post<AdminUser>(`/admin/users/${id}/balance`, {
    balance,
    operation,
	 notes: notes || '',
	 balance_type: balanceType,
	 bonus_validity_days: balanceType === 'bonus' ? bonusValidityDays : undefined,
  })
  return data
}

/**
 * Update user concurrency
 * @param id - User ID
 * @param concurrency - New concurrency limit
 * @returns Updated user
 */
export async function updateConcurrency(id: string, concurrency: number): Promise<AdminUser> {
  return update(id, { concurrency })
}

/** Overwrite concurrency and/or RPM limits for multiple users in one request. */
export async function batchUpdateLimits(
  request: BatchUpdateUserLimitsRequest
): Promise<BatchUpdateUserLimitsResponse> {
  const { data } = await apiClient.post<BatchUpdateUserLimitsResponse>(
    '/admin/users/batch-limits',
    request
  )
  return data
}

export async function previewInactiveUsers(
  request: InactiveUserFilterRequest
): Promise<InactiveUserDeletePreview> {
  const { data } = await apiClient.post<InactiveUserDeletePreview>(
    '/admin/users/inactive/preview',
    request
  )
  return data
}

export async function permanentlyDeleteInactiveUsers(
  request: PermanentlyDeleteInactiveUsersRequest
): Promise<PermanentlyDeleteInactiveUsersResponse> {
  const { data } = await apiClient.post<PermanentlyDeleteInactiveUsersResponse>(
    '/admin/users/inactive/permanent-delete',
    request
  )
  return data
}

/**
 * Toggle user status
 * @param id - User ID
 * @param status - New status
 * @returns Updated user
 */
export async function toggleStatus(id: string, status: 'active' | 'disabled'): Promise<AdminUser> {
  return update(id, { status })
}

/**
 * Get user's API keys
 * @param id - User ID
 * @returns List of user's API keys
 */
export async function getUserApiKeys(id: string): Promise<PaginatedResponse<ApiKey>> {
  const { data } = await apiClient.get<PaginatedResponse<ApiKey>>(`/admin/users/${id}/api-keys`)
  return data
}

/**
 * Get user's usage statistics
 * @param id - User ID
 * @param period - Time period
 * @returns User usage statistics
 */
export async function getUserUsageStats(
  id: string,
  period: string = 'month'
): Promise<{
  total_requests: number
  total_cost: number
  total_tokens: number
}> {
  const { data } = await apiClient.get<{
    total_requests: number
    total_cost: number
    total_tokens: number
  }>(`/admin/users/${id}/usage`, {
    params: { period }
  })
  return data
}

/**
 * Balance history item returned from the API
 */
export interface BalanceHistoryItem {
  id: string
  code: string
  type: string
  value: number
  status: string
  used_by: number | null
  used_at: string | null
  created_at: string
  group_id: string | null
  validity_days: number
  notes: string
  user?: { id: string; email: string } | null
  group?: { id: string; name: string } | null
}

// Balance history response extends pagination with total_recharged summary
export interface BalanceHistoryResponse extends PaginatedResponse<BalanceHistoryItem> {
  total_recharged: number
}

/**
 * Get user's balance/concurrency change history
 * @param id - User ID
 * @param page - Page number
 * @param pageSize - Items per page
 * @param type - Optional type filter (balance, affiliate_balance, admin_balance, concurrency, admin_concurrency, subscription)
 * @returns Paginated balance history with total_recharged
 */
export async function getUserBalanceHistory(
  id: string,
  page: number = 1,
  pageSize: number = 20,
  type?: string
): Promise<BalanceHistoryResponse> {
  const params: Record<string, any> = { page, page_size: pageSize }
  if (type) params.type = type
  const { data } = await apiClient.get<BalanceHistoryResponse>(
    `/admin/users/${id}/balance-history`,
    { params }
  )
  return data
}

/**
 * Replace user's exclusive group
 * @param userId - User ID
 * @param oldGroupId - Current group ID to replace
 * @param newGroupId - New group ID to replace with
 * @returns Number of migrated keys
 */
export async function replaceGroup(
  userId: string,
  oldGroupId: string,
  newGroupId: string
): Promise<{ migrated_keys: number }> {
  const { data } = await apiClient.post<{ migrated_keys: number }>(
    `/admin/users/${userId}/replace-group`,
    { old_group_id: oldGroupId, new_group_id: newGroupId }
  )
  return data
}

export async function bindUserAuthIdentity(
  userId: string,
  input: AdminBindAuthIdentityRequest
): Promise<AdminBoundAuthIdentity> {
  const { data } = await apiClient.post<AdminBoundAuthIdentity>(
    `/admin/users/${userId}/auth-identities`,
    input
  )
  return data
}

/**
 * Platform quota types
 */
export type PlatformQuotaPlatform = 'anthropic' | 'openai' | 'gemini' | 'antigravity' | 'grok'
export type PlatformQuotaWindow = 'daily' | 'weekly' | 'monthly'

export interface PlatformQuotaItem {
  platform: PlatformQuotaPlatform
  daily_limit_usd: number | null
  daily_limit_points?: number | null
  weekly_limit_usd: number | null
  weekly_limit_points?: number | null
  monthly_limit_usd: number | null
  monthly_limit_points?: number | null
  daily_usage_usd: number
  daily_usage_points?: number
  weekly_usage_usd: number
  weekly_usage_points?: number
  monthly_usage_usd: number
  monthly_usage_points?: number
  daily_window_start?: string | null
  weekly_window_start?: string | null
  monthly_window_start?: string | null
  daily_window_resets_at?: string | null
  weekly_window_resets_at?: string | null
  monthly_window_resets_at?: string | null
}

export interface PlatformQuotaUpdateItem {
  platform: PlatformQuotaPlatform
  daily_limit_points: number | null
  weekly_limit_points: number | null
  monthly_limit_points: number | null
  /** @deprecated Compatibility aliases for older clients. */
  daily_limit_usd?: number | null
  /** @deprecated Compatibility aliases for older clients. */
  weekly_limit_usd?: number | null
  /** @deprecated Compatibility aliases for older clients. */
  monthly_limit_usd?: number | null
}

export interface PlatformQuotasResponse {
  platform_quotas: PlatformQuotaItem[]
}

/**
 * Get user's platform quotas
 */
export async function getPlatformQuotas(id: string): Promise<PlatformQuotasResponse> {
  const { data } = await apiClient.get<PlatformQuotasResponse>(
    `/admin/users/${id}/platform-quotas`
  )
  return data
}

/**
 * Replace user's platform quotas (全量替换)
 */
export async function updatePlatformQuotas(
  id: string,
  quotas: PlatformQuotaUpdateItem[]
): Promise<PlatformQuotasResponse> {
  const { data } = await apiClient.put<PlatformQuotasResponse>(
    `/admin/users/${id}/platform-quotas`,
    { quotas }
  )
  return data
}

/**
 * Reset a single (platform, window) usage immediately
 */
export async function resetPlatformQuotaWindow(
  id: string,
  platform: PlatformQuotaPlatform,
  window: PlatformQuotaWindow
): Promise<PlatformQuotasResponse> {
  const { data } = await apiClient.post<PlatformQuotasResponse>(
    `/admin/users/${id}/platform-quotas/reset`,
    { platform, window }
  )
  return data
}

export const usersAPI = {
  list,
  listArchived,
  getLevelSettings,
  updateLevelSettings,
  getLevelProfiles,
  getById,
  create,
  update,
  delete: deleteUser,
  restoreArchivedUser,
  updateBalance,
  updateConcurrency,
  batchUpdateLimits,
  previewInactiveUsers,
  permanentlyDeleteInactiveUsers,
  toggleStatus,
  getUserApiKeys,
  getUserUsageStats,
  getUserBalanceHistory,
  replaceGroup,
  bindUserAuthIdentity,
  getPlatformQuotas,
  updatePlatformQuotas,
  resetPlatformQuotaWindow,
}

export default usersAPI
