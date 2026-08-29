/**
 * Admin API Keys API endpoints
 * Handles API key management for administrators
 */

import { apiClient } from '../client'
import type { ApiKey } from '@/types'

export interface UpdateApiKeyGroupResult {
  api_key: ApiKey
  auto_granted_group_access: boolean
  granted_group_id?: string
  granted_group_name?: string
}

export interface UpdateApiKeyGroupsResult extends UpdateApiKeyGroupResult {}

/**
 * Update an API key's group binding
 * @param id - API Key ID
 * @param groupId - Group UUID, or null to unbind
 * @returns Updated API key with auto-grant info
 */
export async function updateApiKeyGroup(id: string, groupId: string | null): Promise<UpdateApiKeyGroupResult> {
  const { data } = await apiClient.put<UpdateApiKeyGroupResult>(`/admin/api-keys/${id}`, {
    group_id: groupId
  })
  return data
}

export async function updateApiKeyGroups(id: string, groupIds: string[]): Promise<UpdateApiKeyGroupsResult> {
  const { data } = await apiClient.put<UpdateApiKeyGroupsResult>(`/admin/api-keys/${id}`, {
    group_ids: groupIds
  })
  return data
}

export const apiKeysAPI = {
  updateApiKeyGroup,
  updateApiKeyGroups
}

export default apiKeysAPI
