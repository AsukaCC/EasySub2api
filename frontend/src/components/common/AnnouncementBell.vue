<template>
  <div>
    <!-- 铃铛按钮 -->
    <button
      @click="openModal"
      class="components-common-announcement-bell__action"
      :class="{ 'components-common-announcement-bell__action-7': unreadCount > 0 }"
      :aria-label="t('announcements.title')"
    >
      <Icon name="bell" size="md" />
      <!-- 未读红点 -->
      <span
        v-if="unreadCount > 0"
        class="components-common-announcement-bell__text"
      >
        <span class="components-common-announcement-bell__text-2"></span>
        <span class="components-common-announcement-bell__text-3"></span>
      </span>
    </button>

    <!-- 公告列表 Modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div
          v-if="isModalOpen"
          class="components-common-announcement-bell__panel modal-overlay"
          @click="closeModal"
        >
          <div
            class="components-common-announcement-bell__panel-2 modal-content"
            @click.stop
          >
            <!-- Header with Gradient -->
            <div class="components-common-announcement-bell__panel-3">
              <div class="components-common-announcement-bell__panel-4">
                <div>
                  <div class="components-common-announcement-bell__panel-5">
                    <div class="components-common-announcement-bell__panel-6">
                      <Icon name="bell" size="sm" />
                    </div>
                    <h2 class="components-common-announcement-bell__heading">
                      {{ t('announcements.title') }}
                    </h2>
                  </div>
                  <p v-if="unreadCount > 0" class="components-common-announcement-bell__description">
                    <span class="components-common-announcement-bell__text-4">{{ unreadCount }}</span>
                    {{ t('announcements.unread') }}
                  </p>
                </div>
                <div class="components-common-announcement-bell__panel-5">
                  <button
                    v-if="unreadCount > 0"
                    @click="markAllAsRead"
                    :disabled="loading"
                    class="components-common-announcement-bell__action-2"
                  >
                    {{ t('announcements.markAllRead') }}
                  </button>
                  <button
                    @click="closeModal"
                    class="components-common-announcement-bell__action-3"
                    :aria-label="t('common.close')"
                  >
                    <Icon name="x" size="sm" />
                  </button>
                </div>
              </div>
              <!-- Decorative gradient -->
              <div class="components-common-announcement-bell__panel-7"></div>
            </div>

            <!-- Body -->
            <div class="components-common-announcement-bell__panel-8">
              <!-- Loading -->
              <div v-if="loading" class="components-common-announcement-bell__panel-9">
                <div class="components-common-announcement-bell__panel-10">
                  <div class="components-common-announcement-bell__panel-11"></div>
                  <div class="components-common-announcement-bell__panel-12"></div>
                </div>
              </div>

              <!-- Announcements List -->
              <div v-else-if="bellAnnouncements.length > 0">
                <div
                  v-for="item in bellAnnouncements"
                  :key="item.id"
                  class="components-common-announcement-bell__panel-13"
                  :class="{ 'components-common-announcement-bell__panel-46': !item.read_at }"
                  style="min-height: 72px"
                  @click="openDetail(item)"
                >
                  <!-- Status Indicator -->
                  <div class="components-common-announcement-bell__panel-14">
                    <div
                      v-if="!item.read_at"
                      class="components-common-announcement-bell__panel-15"
                    >
                      <!-- Pulse ring -->
                      <span class="components-common-announcement-bell__text-5"></span>
                      <!-- Icon -->
                      <svg class="components-common-announcement-bell__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div
                      v-else
                      class="components-common-announcement-bell__panel-16"
                    >
                      <svg class="components-common-announcement-bell__icon-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                  </div>

                  <!-- Content -->
                  <div class="components-common-announcement-bell__panel-17">
                    <div class="components-common-announcement-bell__panel-18">
                      <h3 class="components-common-announcement-bell__heading-2">
                        {{ item.title }}
                      </h3>
                      <div class="components-common-announcement-bell__panel-19">
                        <time class="components-common-announcement-bell__time">
                          {{ formatRelativeTime(item.created_at) }}
                        </time>
                        <span
                          v-if="!item.read_at"
                          class="components-common-announcement-bell__text-6"
                        >
                          <span class="components-common-announcement-bell__text-7">
                            <span class="components-common-announcement-bell__text-8"></span>
                            <span class="components-common-announcement-bell__text-9"></span>
                          </span>
                          {{ t('announcements.unread') }}
                        </span>
                      </div>
                    </div>

                    <!-- Arrow -->
                    <div class="components-common-announcement-bell__panel-20">
                      <svg
                        class="components-common-announcement-bell__icon-3"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                  </div>

                  <!-- Unread indicator bar -->
                  <div
                    v-if="!item.read_at"
                    class="components-common-announcement-bell__panel-21"
                  ></div>
                </div>
              </div>

              <!-- Empty State -->
              <div v-else class="components-common-announcement-bell__panel-22">
                <div class="components-common-announcement-bell__panel-23">
                  <div class="components-common-announcement-bell__panel-24">
                    <Icon name="inbox" size="xl" class="components-common-announcement-bell__icon-4" />
                  </div>
                  <div class="components-common-announcement-bell__panel-25">
                    <svg class="components-common-announcement-bell__icon-5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                    </svg>
                  </div>
                </div>
                <p class="components-common-announcement-bell__description-2">{{ t('announcements.empty') }}</p>
                <p class="components-common-announcement-bell__description-3">{{ t('announcements.emptyDescription') }}</p>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 公告详情 Modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div
          v-if="detailModalOpen && selectedAnnouncement"
          class="components-common-announcement-bell__panel-26 modal-overlay"
          @click="closeDetail"
        >
          <div
            class="components-common-announcement-bell__panel-27 modal-content"
            @click.stop
          >
            <!-- Header with Decorative Elements -->
            <div class="components-common-announcement-bell__panel-28">
              <!-- Decorative background elements -->
              <div class="components-common-announcement-bell__panel-29"></div>
              <div class="components-common-announcement-bell__panel-30"></div>
              <div class="components-common-announcement-bell__panel-31"></div>

              <div class="components-common-announcement-bell__panel-32">
                <div class="components-common-announcement-bell__panel-33">
                  <!-- Icon and Category -->
                  <div class="components-common-announcement-bell__panel-34">
                    <div class="components-common-announcement-bell__panel-35">
                      <svg class="components-common-announcement-bell__icon-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div class="components-common-announcement-bell__panel-5">
                      <span class="components-common-announcement-bell__text-10">
                        {{ t('announcements.title') }}
                      </span>
                      <span
                        v-if="!selectedAnnouncement.read_at"
                        class="components-common-announcement-bell__text-11"
                      >
                        <span class="components-common-announcement-bell__text-12">
                          <span class="components-common-announcement-bell__text-13"></span>
                          <span class="components-common-announcement-bell__text-14"></span>
                        </span>
                        {{ t('announcements.unread') }}
                      </span>
                    </div>
                  </div>

                  <!-- Title -->
                  <h2 class="components-common-announcement-bell__heading-3">
                    {{ selectedAnnouncement.title }}
                  </h2>

                  <!-- Meta Info -->
                  <div class="components-common-announcement-bell__panel-36">
                    <div class="components-common-announcement-bell__panel-37">
                      <svg class="components-common-announcement-bell__icon-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      <time>{{ formatRelativeWithDateTime(selectedAnnouncement.created_at) }}</time>
                    </div>
                    <div class="components-common-announcement-bell__panel-37">
                      <svg class="components-common-announcement-bell__icon-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                      <span>{{ selectedAnnouncement.read_at ? t('announcements.read') : t('announcements.unread') }}</span>
                    </div>
                  </div>
                </div>

                <!-- Close button -->
                <button
                  @click="closeDetail"
                  class="components-common-announcement-bell__action-4"
                  :aria-label="t('common.close')"
                >
                  <Icon name="x" size="md" />
                </button>
              </div>
            </div>

            <!-- Body with Enhanced Markdown -->
            <div class="components-common-announcement-bell__panel-38">
              <!-- Content with decorative border -->
              <div class="components-common-announcement-bell__panel-10">
                <!-- Decorative left border -->
                <div class="components-common-announcement-bell__panel-39"></div>

                <div class="components-common-announcement-bell__panel-40">
                  <div
                    class="components-common-announcement-bell__panel-41 markdown-body"
                    v-html="renderMarkdown(selectedAnnouncement.content)"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Footer with Actions -->
            <div class="components-common-announcement-bell__panel-42">
              <div class="components-common-announcement-bell__panel-43">
                <div class="components-common-announcement-bell__panel-44">
                  <svg class="components-common-announcement-bell__icon-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span>{{ selectedAnnouncement.read_at ? t('announcements.readStatus') : t('announcements.markReadHint') }}</span>
                </div>
                <div class="components-common-announcement-bell__panel-45">
                  <button
                    @click="closeDetail"
                    class="components-common-announcement-bell__action-5"
                  >
                    {{ t('common.close') }}
                  </button>
                  <button
                    v-if="!selectedAnnouncement.read_at"
                    @click="markAsReadAndClose(selectedAnnouncement.id)"
                    class="components-common-announcement-bell__action-6"
                  >
                    <span class="components-common-announcement-bell__panel-5">
                      <svg class="components-common-announcement-bell__icon-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                      </svg>
                      {{ t('announcements.markRead') }}
                    </span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { storeToRefs } from 'pinia'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAppStore } from '@/stores/app'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeTime, formatRelativeWithDateTime } from '@/utils/format'
import type { UserAnnouncement } from '@/types'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()
const announcementStore = useAnnouncementStore()

// Configure marked
marked.setOptions({
  breaks: true,
  gfm: true,
})

// Use store state (storeToRefs for reactivity)
const { bellAnnouncements, loading } = storeToRefs(announcementStore)
const unreadCount = computed(() => announcementStore.unreadCount)

// Local modal state
const isModalOpen = ref(false)
const detailModalOpen = ref(false)
const selectedAnnouncement = ref<UserAnnouncement | null>(null)

// Methods
function renderMarkdown(content: string): string {
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
}

function openModal() {
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
}

function openDetail(announcement: UserAnnouncement) {
  selectedAnnouncement.value = announcement
  detailModalOpen.value = true
  if (!announcement.read_at) {
    markAsRead(announcement.id)
  }
}

function closeDetail() {
  detailModalOpen.value = false
  selectedAnnouncement.value = null
}

async function markAsRead(id: string) {
  try {
    await announcementStore.markAsRead(id)
  } catch (err: any) {
    appStore.showError(err?.message || t('common.unknownError'))
  }
}

async function markAsReadAndClose(id: string) {
  await markAsRead(id)
  appStore.showSuccess(t('announcements.markedAsRead'))
  closeDetail()
}

async function markAllAsRead() {
  try {
    await announcementStore.markAllAsRead()
    appStore.showSuccess(t('announcements.allMarkedAsRead'))
  } catch (err: any) {
    appStore.showError(err?.message || t('common.unknownError'))
  }
}

function handleEscape(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    if (detailModalOpen.value) {
      closeDetail()
    } else if (isModalOpen.value) {
      closeModal()
    }
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleEscape)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleEscape)
  document.body.style.overflow = ''
})

watch(
  [isModalOpen, detailModalOpen, () => announcementStore.currentPopup],
  ([modal, detail, popup]) => {
    document.body.style.overflow = (modal || detail || popup) ? 'hidden' : ''
  }
)
</script>

<style scoped>
/* Modal Animations */
.modal-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-from > div {
  transform: scale(0.94) translateY(-12px);
  opacity: 0;
}

.modal-fade-leave-to > div {
  transform: scale(0.96) translateY(-8px);
  opacity: 0;
}

/* Scrollbar Styling */
.overflow-y-auto::-webkit-scrollbar {
  width: 8px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #cbd5e1, #94a3b8);
  border-radius: 4px;
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #4b5563, #2e2e33);
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(to bottom, #94a3b8, #64748b);
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(to bottom, #6b7280, #4b5563);
}
</style>
