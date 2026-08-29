<template>
  <div v-if="activeBanners.length > 0" class="announcement-float" aria-live="polite">
    <TransitionGroup name="announcement-float">
      <section
        v-for="announcement in activeBanners"
        :key="announcement.id"
        class="announcement-float__card"
        :aria-label="announcement.title"
      >
        <header class="announcement-float__header">
          <span class="announcement-float__badge">
            <Icon name="bell" size="sm" />
          </span>
          <h3 class="announcement-float__title">{{ announcement.title }}</h3>
          <button
            type="button"
            class="announcement-float__dismiss"
            :aria-label="t('common.close')"
            :title="t('common.close')"
            @click="dismiss(announcement.id)"
          >
            <Icon name="x" size="sm" />
          </button>
        </header>
        <p v-if="previewFor(announcement)" class="announcement-float__content">
          {{ previewFor(announcement) }}
        </p>
      </section>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useAnnouncementStore } from '@/stores/announcements'
import { useAuthStore } from '@/stores/auth'
import type { UserAnnouncement } from '@/types'

const DISMISSED_STORAGE_KEY = 'announcement_banner_dismissed'

const { t } = useI18n()
const announcementStore = useAnnouncementStore()
const authStore = useAuthStore()
const { bannerAnnouncements, publicBanners } = storeToRefs(announcementStore)

// 未登录访客走公开横幅接口;登录后以用户公告列表为准(含定向公告)
watch(
  () => authStore.isAuthenticated,
  (isAuthenticated) => {
    if (!isAuthenticated) {
      void announcementStore.fetchPublicBanners()
    }
  },
  { immediate: true }
)

function readDismissed(): string[] {
  try {
    const raw = sessionStorage.getItem(DISMISSED_STORAGE_KEY)
    const parsed = raw ? JSON.parse(raw) : []
    return Array.isArray(parsed) ? parsed.filter((v) => typeof v === 'string') : []
  } catch {
    return []
  }
}

const dismissedIds = ref<string[]>(readDismissed())

const activeBanners = computed(() => {
  const source = authStore.isAuthenticated ? bannerAnnouncements.value : publicBanners.value
  return source.filter((a) => !dismissedIds.value.includes(a.id)).slice(0, 3)
})

function dismiss(id: string) {
  dismissedIds.value = [...dismissedIds.value, id]
  try {
    sessionStorage.setItem(DISMISSED_STORAGE_KEY, JSON.stringify(dismissedIds.value))
  } catch {
    // sessionStorage 不可用时仅本次生效
  }
}

function previewFor(announcement: UserAnnouncement): string {
  return announcement.content
    .replace(/[`*_#>[\]~]/g, '')
    .replace(/\s+/g, ' ')
    .trim()
}
</script>

<style scoped>
/* 右下角悬浮公告:玻璃卡片,与顶部导航同一材质语言 */
.announcement-float {
  position: fixed;
  right: 1rem;
  bottom: 1rem;
  left: 1rem;
  z-index: 90; /* 低于 Toast(100),高于页面内容 */
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.75rem;
  pointer-events: none;
}

.announcement-float__card {
  width: min(24rem, 100%);
  overflow: hidden;
  border: 1px solid rgb(245 158 11 / 0.3);
  border-radius: var(--radius-xl);
  background: rgb(255 248 230 / 0.82);
  color: #92400e;
  -webkit-backdrop-filter: blur(20px) saturate(var(--glass-saturate));
  backdrop-filter: blur(20px) saturate(var(--glass-saturate));
  box-shadow:
    0 12px 28px rgb(12 12 14 / 0.14),
    0 1px 0 rgb(255 255 255 / 0.55) inset;
  pointer-events: auto;
}

.announcement-float__header {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.75rem 0.625rem 0.75rem 0.875rem;
}

.announcement-float__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 1.75rem;
  height: 1.75rem;
  border-radius: var(--radius-md);
  background: rgb(245 158 11 / 0.16);
}

.announcement-float__title {
  flex: 1;
  min-width: 0;
  margin: 0;
  overflow: hidden;
  font-size: var(--font-size-sm);
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.announcement-float__dismiss {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 1.75rem;
  height: 1.75rem;
  border: 0;
  border-radius: var(--radius-md);
  background: transparent;
  color: inherit;
  opacity: 0.6;
  cursor: pointer;
  transition: opacity 160ms ease, background-color 160ms ease;
}

.announcement-float__dismiss:hover {
  background: rgb(245 158 11 / 0.16);
  opacity: 1;
}

.announcement-float__dismiss:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px rgb(245 158 11 / 0.35);
  opacity: 1;
}

.announcement-float__content {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  padding: 0 0.875rem 0.875rem;
  font-size: var(--font-size-sm);
  line-height: 1.5;
  opacity: 0.9;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  line-clamp: 3;
}

/* 进出场动画:右下滑入 */
.announcement-float-enter-active,
.announcement-float-leave-active {
  transition: opacity 220ms ease, transform 220ms ease;
}

.announcement-float-enter-from,
.announcement-float-leave-to {
  opacity: 0;
  transform: translateY(0.75rem);
}

.announcement-float-move {
  transition: transform 220ms ease;
}

@media (prefers-reduced-motion: reduce) {
  .announcement-float-enter-active,
  .announcement-float-leave-active,
  .announcement-float-move {
    transition: none;
  }
}

/* 暗色:中性近黑玻璃底 + 琥珀描边与文字 */
.dark .announcement-float__card {
  border-color: rgb(251 191 36 / 0.22);
  background: rgb(23 23 26 / 0.72);
  color: #fcd34d;
  box-shadow:
    0 14px 32px rgb(0 0 0 / 0.32),
    0 1px 0 rgb(255 255 255 / 0.06) inset;
}

.dark .announcement-float__badge,
.dark .announcement-float__dismiss:hover {
  background: rgb(251 191 36 / 0.14);
}
</style>
