<template>
  <section class="dashboard-announcements card" :aria-label="t('dashboard.announcementCenter.title')">
    <header class="dashboard-announcements__header">
      <div class="dashboard-announcements__title-group">
        <span class="dashboard-announcements__icon"><Icon name="bell" size="md" /></span>
        <div>
          <h2 class="dashboard-announcements__title">{{ t('dashboard.announcementCenter.title') }}</h2>
          <p class="dashboard-announcements__description">{{ t('dashboard.announcementCenter.description') }}</p>
        </div>
      </div>
      <span v-if="unreadCount > 0" class="dashboard-announcements__unread">
        {{ t('dashboard.announcementCenter.unread', { count: unreadCount }) }}
      </span>
    </header>

    <div v-if="loading" class="dashboard-announcements__state">
      <LoadingSpinner size="md" />
    </div>
    <div v-else-if="latestAnnouncements.length === 0" class="dashboard-announcements__state">
      <Icon name="inbox" size="lg" />
      <span>{{ t('announcements.emptyDescription') }}</span>
    </div>
    <div v-else class="dashboard-announcements__list">
      <button
        v-for="announcement in latestAnnouncements"
        :key="announcement.id"
        type="button"
        class="dashboard-announcements__item"
        @click="openAnnouncement(announcement)"
      >
        <span v-if="!announcement.read_at" class="dashboard-announcements__dot" />
        <span class="dashboard-announcements__item-main">
          <span class="dashboard-announcements__item-title">{{ announcement.title }}</span>
          <time class="dashboard-announcements__time">{{ formatRelativeWithDateTime(announcement.created_at) }}</time>
        </span>
        <Icon name="chevronRight" size="sm" class="dashboard-announcements__arrow" />
      </button>
    </div>
  </section>

  <AnnouncementPopup
    v-if="selectedAnnouncement"
    :announcement="selectedAnnouncement"
    preview
    @close="selectedAnnouncement = null"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import AnnouncementPopup from '@/components/common/AnnouncementPopup.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeWithDateTime } from '@/utils/format'
import type { UserAnnouncement } from '@/types'

const { t } = useI18n()
const announcementStore = useAnnouncementStore()
const { dashboardAnnouncements, dashboardUnreadCount, loading } = storeToRefs(announcementStore)
const selectedAnnouncement = ref<UserAnnouncement | null>(null)

const latestAnnouncements = computed(() => dashboardAnnouncements.value.slice(0, 3))
const unreadCount = dashboardUnreadCount

function openAnnouncement(announcement: UserAnnouncement) {
  selectedAnnouncement.value = announcement
  if (!announcement.read_at) {
    void announcementStore.markAsRead(announcement.id)
  }
}

onMounted(() => {
  void announcementStore.fetchAnnouncements()
})
</script>

<style scoped>
.dashboard-announcements {
  overflow: hidden;
}

.dashboard-announcements__header,
.dashboard-announcements__title-group,
.dashboard-announcements__item {
  display: flex;
  align-items: center;
}

.dashboard-announcements__header {
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.dashboard-announcements__title-group {
  min-width: 0;
  gap: 0.75rem;
}

.dashboard-announcements__icon {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: var(--radius-md);
  color: var(--color-text-warning);
  background: var(--color-warning-subtle);
}

.dashboard-announcements__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
  font-weight: 650;
}

.dashboard-announcements__description,
.dashboard-announcements__time {
  margin: 0.125rem 0 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.dashboard-announcements__unread {
  flex: 0 0 auto;
  padding: 0.25rem 0.625rem;
  border-radius: 999px;
  color: var(--color-text-warning);
  background: var(--color-warning-subtle);
  font-size: var(--font-size-xs);
  font-weight: 600;
}

.dashboard-announcements__list {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.dashboard-announcements__item {
  min-width: 0;
  min-height: 4.25rem;
  gap: 0.625rem;
  padding: 0.875rem 1.5rem;
  border: 0;
  border-right: 1px solid var(--color-border-subtle);
  color: inherit;
  text-align: left;
  background: transparent;
  transition: background-color 160ms ease;
}

.dashboard-announcements__item:last-child {
  border-right: 0;
}

.dashboard-announcements__item:hover {
  background: var(--glass-bg-interactive);
  -webkit-backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate-hover));
}

.dashboard-announcements__item:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: -2px;
}

.dashboard-announcements__dot {
  flex: 0 0 auto;
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  background: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-subtle);
}

.dashboard-announcements__item-main {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
}

.dashboard-announcements__item-title {
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dashboard-announcements__arrow {
  flex: 0 0 auto;
  color: var(--color-text-tertiary);
}

.dashboard-announcements__state {
  display: flex;
  min-height: 5rem;
  align-items: center;
  justify-content: center;
  gap: 0.625rem;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
}

@media (max-width: 767px) {
  .dashboard-announcements__header {
    align-items: flex-start;
  }

  .dashboard-announcements__description {
    display: none;
  }

  .dashboard-announcements__list {
    grid-template-columns: 1fr;
  }

  .dashboard-announcements__item {
    border-right: 0;
    border-bottom: 1px solid var(--color-border-subtle);
  }

  .dashboard-announcements__item:last-child {
    border-bottom: 0;
  }
}
</style>
