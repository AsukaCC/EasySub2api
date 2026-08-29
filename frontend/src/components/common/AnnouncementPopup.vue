<template>
  <Teleport to="body">
    <Transition name="popup-fade">
      <div
        v-if="displayedAnnouncement"
        class="announcement-popup modal-overlay"
        role="dialog"
        aria-modal="true"
        :aria-label="displayedAnnouncement.title"
      >
        <div class="announcement-popup__dialog modal-content modal-content--normal" @click.stop>
          <!-- Header -->
          <header class="announcement-popup__header">
            <div class="announcement-popup__meta">
              <span class="announcement-popup__badge">
                <Icon name="bell" size="md" />
              </span>
              <div class="announcement-popup__meta-text">
                <div class="announcement-popup__kicker">
                  <span>{{ t('announcements.title') }}</span>
                  <span v-if="!preview" class="announcement-popup__unread">
                    <span class="announcement-popup__unread-dot" aria-hidden="true"></span>
                    {{ t('announcements.unread') }}
                  </span>
                </div>
                <time class="announcement-popup__time">
                  {{ formatRelativeWithDateTime(displayedAnnouncement.created_at) }}
                </time>
              </div>
            </div>
            <button
              type="button"
              class="announcement-popup__close"
              :aria-label="t('common.close')"
              :title="t('common.close')"
              @click="handleDismiss"
            >
              <Icon name="x" size="md" />
            </button>
          </header>

          <!-- Title + Body -->
          <div class="announcement-popup__body">
            <h2 class="announcement-popup__title">{{ displayedAnnouncement.title }}</h2>
            <div class="announcement-popup__content markdown-body" v-html="renderedContent"></div>
          </div>

          <!-- Footer -->
          <footer class="announcement-popup__footer">
            <button
              data-testid="announcement-popup-dismiss"
              class="btn btn-primary announcement-popup__confirm"
              @click="handleDismiss"
            >
              <Icon :name="preview ? 'x' : 'checkCircle'" size="md" />
              {{ preview ? t('common.close') : t('announcements.markRead') }}
            </button>
          </footer>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeWithDateTime } from '@/utils/format'
import type { Announcement, UserAnnouncement } from '@/types'

type PreviewAnnouncement = Pick<Announcement | UserAnnouncement, 'title' | 'content' | 'created_at'>

const props = withDefaults(defineProps<{
  announcement?: PreviewAnnouncement | null
  preview?: boolean
}>(), {
  announcement: null,
  preview: false,
})

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const announcementStore = useAnnouncementStore()
const displayedAnnouncement = computed(() => (
  props.preview ? props.announcement : announcementStore.currentPopup
))

marked.setOptions({
  breaks: true,
  gfm: true,
})

const renderedContent = computed(() => {
  const content = displayedAnnouncement.value?.content
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

function handleDismiss() {
  if (props.preview) {
    emit('close')
    return
  }
  announcementStore.dismissPopup()
}

// Manage body overflow — only set, never unset (bell component handles restore)
watch(
  displayedAnnouncement,
  (popup) => {
    if (popup) {
      document.body.style.overflow = 'hidden'
    } else if (props.preview) {
      document.body.style.overflow = ''
    }
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  if (props.preview) {
    document.body.style.overflow = ''
  }
})
</script>

<style scoped>
/* 面向用户的公告弹窗:标准玻璃模态(材质由 glass.scss 的 .modal-content 提供)
   + 公告语义的琥珀色头部标识,与右下角悬浮公告同一语言 */
.announcement-popup {
  z-index: 95;
}

.announcement-popup__dialog {
  overflow: hidden;
}

.announcement-popup__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  flex-shrink: 0;
  gap: 0.75rem;
  padding: 1rem 1rem 0.875rem 1.25rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.announcement-popup__meta {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.75rem;
}

.announcement-popup__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: var(--radius-lg);
  color: #d97706;
  background: rgba(245, 158, 11, 0.16);
}

.dark .announcement-popup__badge {
  color: #fbbf24;
  background: rgba(251, 191, 36, 0.14);
}

.announcement-popup__meta-text {
  min-width: 0;
}

.announcement-popup__kicker {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: 650;
}

.announcement-popup__unread {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.125rem 0.5rem;
  border-radius: 999px;
  color: #b45309;
  background: rgba(245, 158, 11, 0.16);
  font-size: var(--font-size-xs);
  font-weight: 600;
}

.dark .announcement-popup__unread {
  color: #fbbf24;
  background: rgba(120, 53, 15, 0.4);
}

.announcement-popup__unread-dot {
  width: 0.375rem;
  height: 0.375rem;
  border-radius: 50%;
  background: currentColor;
}

.announcement-popup__time {
  display: block;
  margin-top: 0.125rem;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.announcement-popup__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 2rem;
  height: 2rem;
  border: 0;
  border-radius: var(--radius-md);
  color: var(--color-text-tertiary);
  background: transparent;
  cursor: pointer;
  transition: color 160ms ease, background-color 160ms ease;
}

.announcement-popup__close:hover {
  color: var(--color-text-primary);
  background: var(--color-surface-hover);
}

.announcement-popup__close:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.32);
}

.announcement-popup__body {
  flex: 1;
  overflow-x: hidden;
  overflow-y: auto;
  padding: 1.25rem 1.5rem;
}

.announcement-popup__title {
  margin: 0 0 0.75rem;
  color: var(--color-text-primary);
  font-size: var(--font-size-xl);
  font-weight: 700;
  line-height: 1.4;
}

.announcement-popup__content {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  line-height: 1.7;
}

.announcement-popup__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  flex-shrink: 0;
  padding: 0.875rem 1.5rem;
  border-top: 1px solid var(--color-border-subtle);
}

.announcement-popup__confirm {
  min-width: 8rem;
}

/* 进出场动画 */
.popup-fade-enter-active {
  transition: opacity 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.popup-fade-leave-active {
  transition: opacity 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.popup-fade-enter-from,
.popup-fade-leave-to {
  opacity: 0;
}

.popup-fade-enter-active .announcement-popup__dialog,
.popup-fade-leave-active .announcement-popup__dialog {
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.25s ease;
}

.popup-fade-enter-from .announcement-popup__dialog {
  transform: scale(0.94) translateY(-12px);
  opacity: 0;
}

.popup-fade-leave-to .announcement-popup__dialog {
  transform: scale(0.96) translateY(-8px);
  opacity: 0;
}

/* 内容滚动条 */
.announcement-popup__body::-webkit-scrollbar {
  width: 8px;
}

.announcement-popup__body::-webkit-scrollbar-track {
  background: transparent;
}

.announcement-popup__body::-webkit-scrollbar-thumb {
  border-radius: 4px;
  background: var(--color-border-strong);
}
</style>
