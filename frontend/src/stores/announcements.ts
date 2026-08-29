import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { announcementsAPI } from '@/api'
import type { UserAnnouncement } from '@/types'

const THROTTLE_MS = 20 * 60 * 1000 // 20 minutes

export const useAnnouncementStore = defineStore('announcements', () => {
  // State
  const announcements = ref<UserAnnouncement[]>([])
  const loading = ref(false)
  const lastFetchTime = ref(0)
  // 匿名可见的横幅公告(公开接口,不含已读状态);登出时保留,登录后以用户列表为准
  const publicBanners = ref<UserAnnouncement[]>([])
  const lastPublicFetchTime = ref(0)
  const popupQueue = ref<UserAnnouncement[]>([])
  const currentPopup = ref<UserAnnouncement | null>(null)

  // Session-scoped dedup set — not reactive, used as plain lookup only
  let shownPopupIds = new Set<string>()

  // Getters
  const dashboardAnnouncements = computed(() =>
    announcements.value.filter((a) => a.notify_mode === 'silent')
  )
  const bellAnnouncements = computed(() =>
    announcements.value.filter((a) => a.notify_mode !== 'banner')
  )
  const bannerAnnouncements = computed(() =>
    announcements.value.filter((a) => a.notify_mode === 'banner')
  )
  const unreadCount = computed(() =>
    bellAnnouncements.value.filter((a) => !a.read_at).length
  )
  const dashboardUnreadCount = computed(() =>
    dashboardAnnouncements.value.filter((a) => !a.read_at).length
  )

  // Actions
  async function fetchAnnouncements(force = false) {
    const now = Date.now()
    if (!force && lastFetchTime.value > 0 && now - lastFetchTime.value < THROTTLE_MS) {
      return
    }

    // Set immediately to prevent concurrent duplicate requests
    lastFetchTime.value = now

    try {
      loading.value = true
      const all = await announcementsAPI.list(false)
      announcements.value = all.slice(0, 20)
      enqueueNewPopups()
    } catch (err: any) {
      // Revert throttle timestamp on failure so retry is allowed
      lastFetchTime.value = 0
      console.error('Failed to fetch announcements:', err)
    } finally {
      loading.value = false
    }
  }

  async function fetchPublicBanners(force = false) {
    const now = Date.now()
    if (!force && lastPublicFetchTime.value > 0 && now - lastPublicFetchTime.value < THROTTLE_MS) {
      return
    }
    lastPublicFetchTime.value = now

    try {
      publicBanners.value = await announcementsAPI.listPublicBanners()
    } catch (err: any) {
      lastPublicFetchTime.value = 0
      console.error('Failed to fetch public banners:', err)
    }
  }

  function enqueueNewPopups() {
    const newPopups = announcements.value.filter(
      (a) => a.notify_mode === 'popup' && !a.read_at && !shownPopupIds.has(a.id)
    )
    if (newPopups.length === 0) return

    for (const p of newPopups) {
      if (!popupQueue.value.some((q) => q.id === p.id)) {
        popupQueue.value.push(p)
      }
    }

    if (!currentPopup.value) {
      showNextPopup()
    }
  }

  function showNextPopup() {
    if (popupQueue.value.length === 0) {
      currentPopup.value = null
      return
    }
    currentPopup.value = popupQueue.value.shift()!
    shownPopupIds.add(currentPopup.value.id)
  }

  async function dismissPopup() {
    if (!currentPopup.value) return
    const id = currentPopup.value.id
    currentPopup.value = null

    // Mark as read (fire-and-forget, UI already updated)
    markAsRead(id)

    // Show next popup after a short delay
    if (popupQueue.value.length > 0) {
      setTimeout(() => showNextPopup(), 300)
    }
  }

  async function markAsRead(id: string) {
    try {
      await announcementsAPI.markRead(id)
      const ann = announcements.value.find((a) => a.id === id)
      if (ann) {
        ann.read_at = new Date().toISOString()
      }
    } catch (err: any) {
      console.error('Failed to mark announcement as read:', err)
    }
  }

  async function markAllAsRead() {
    const unread = bellAnnouncements.value.filter((a) => !a.read_at)
    if (unread.length === 0) return

    try {
      loading.value = true
      await Promise.all(unread.map((a) => announcementsAPI.markRead(a.id)))
      announcements.value.forEach((a) => {
        if (!a.read_at) {
          a.read_at = new Date().toISOString()
        }
      })
    } catch (err: any) {
      console.error('Failed to mark all as read:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  function reset() {
    announcements.value = []
    lastFetchTime.value = 0
    shownPopupIds = new Set()
    popupQueue.value = []
    currentPopup.value = null
    loading.value = false
  }

  return {
    // State
    announcements,
    publicBanners,
    fetchPublicBanners,
    dashboardAnnouncements,
    bellAnnouncements,
    bannerAnnouncements,
    loading,
    currentPopup,
    // Getters
    unreadCount,
    dashboardUnreadCount,
    // Actions
    fetchAnnouncements,
    dismissPopup,
    markAsRead,
    markAllAsRead,
    reset,
  }
})
