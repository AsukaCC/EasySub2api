/**
 * User Announcements API endpoints
 */

import { apiClient } from './client'
import type { UserAnnouncement } from '@/types'

export async function list(unreadOnly: boolean = false): Promise<UserAnnouncement[]> {
  const { data } = await apiClient.get<UserAnnouncement[]>('/announcements', {
    params: unreadOnly ? { unread_only: 1 } : {}
  })
  return data
}

/**
 * List banner announcements visible to anonymous visitors (no auth required).
 */
export async function listPublicBanners(): Promise<UserAnnouncement[]> {
  const { data } = await apiClient.get<UserAnnouncement[]>('/announcements/public')
  return data
}

export async function markRead(id: string): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(`/announcements/${id}/read`)
  return data
}

const announcementsAPI = {
  list,
  listPublicBanners,
  markRead
}

export default announcementsAPI

