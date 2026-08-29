<template>
  <AppLayout>
    <div class="custom-page-layout">
      <div class="views-user-custom-page-view__panel card">
        <div v-if="loading" class="views-user-custom-page-view__panel-2">
          <div
            class="views-user-custom-page-view__panel-3"
          ></div>
        </div>

        <div
          v-else-if="!menuItem"
          class="views-user-custom-page-view__panel-4"
        >
          <div class="views-user-custom-page-view__panel-5">
            <div
              class="views-user-custom-page-view__panel-6"
            >
              <Icon name="link" size="lg" class="views-user-custom-page-view__icon" />
            </div>
            <h3 class="views-user-custom-page-view__heading">
              {{ t('customPage.notFoundTitle') }}
            </h3>
            <p class="views-user-custom-page-view__description">
              {{ t('customPage.notFoundDesc') }}
            </p>
          </div>
        </div>

        <!-- Markdown mode with TOC -->
        <div v-else-if="isMarkdownMode" class="views-user-custom-page-view__panel-7">
          <!-- TOC Sidebar -->
          <aside
            v-show="tocVisible"
            class="toc-sidebar"
          >
            <div class="toc-header">
              <span class="toc-title">{{ t('customPage.tableOfContents') }}</span>
              <button class="toc-close-btn" @click="tocVisible = false">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 18l-6-6 6-6"/></svg>
              </button>
            </div>
            <nav class="toc-nav">
              <a
                v-for="item in tocItems"
                :key="item.id"
                :href="'#' + item.id"
                class="toc-item"
                :class="[
                  `toc-level-${item.level}`,
                  { 'toc-active': activeHeadingId === item.id }
                ]"
                @click.prevent="scrollToHeading(item.id)"
              >
                {{ item.text }}
              </a>
            </nav>
          </aside>

          <!-- TOC Toggle Button (when collapsed) -->
          <button
            v-show="!tocVisible && tocItems.length > 0"
            class="toc-toggle-btn"
            @click="tocVisible = true"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12h18M3 6h18M3 18h18"/></svg>
            <span class="views-user-custom-page-view__text">{{ t('customPage.tableOfContents') }}</span>
          </button>

          <!-- Content -->
          <div
            ref="markdownContainer"
            class="views-user-custom-page-view__panel-8 markdown-page-content"
            v-html="renderedHtml"
            @scroll="onContentScroll"
          ></div>
        </div>

        <!-- URL not configured -->
        <div v-else-if="!isValidUrl" class="views-user-custom-page-view__panel-4">
          <div class="views-user-custom-page-view__panel-5">
            <div
              class="views-user-custom-page-view__panel-6"
            >
              <Icon name="link" size="lg" class="views-user-custom-page-view__icon" />
            </div>
            <h3 class="views-user-custom-page-view__heading">
              {{ t('customPage.notConfiguredTitle') }}
            </h3>
            <p class="views-user-custom-page-view__description">
              {{ t('customPage.notConfiguredDesc') }}
            </p>
          </div>
        </div>

        <!-- Iframe embed mode -->
        <div v-else class="custom-embed-shell">
          <a
            :href="embeddedUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary btn-sm custom-open-fab"
          >
            <Icon name="externalLink" size="sm" class="views-user-custom-page-view__icon-2" :stroke-width="2" />
            {{ t('customPage.openInNewTab') }}
          </a>
          <iframe
            :src="embeddedUrl"
            class="custom-embed-frame"
            allowfullscreen
          ></iframe>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { buildApiUrl } from '@/api/client'
import { buildEmbeddedUrl, detectTheme } from '@/utils/embedded-url'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

interface TocItem {
  id: string
  text: string
  level: number
}

const { t, locale } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()

const loading = ref(false)
const pageTheme = ref<'light' | 'dark'>('light')
const renderedHtml = ref('')
const markdownContainer = ref<HTMLElement | null>(null)
const tocItems = ref<TocItem[]>([])
const tocVisible = ref(typeof window !== 'undefined' ? window.innerWidth > 768 : true)
const activeHeadingId = ref('')
let themeObserver: MutationObserver | null = null

const menuItemId = computed(() => route.params.id as string)

const menuItem = computed(() => {
  const id = menuItemId.value
  const publicItems = appStore.cachedPublicSettings?.custom_menu_items ?? []
  const found = publicItems.find((item) => item.id === id) ?? null
  if (found) return found
  if (authStore.isAdmin) {
    return adminSettingsStore.customMenuItems.find((item) => item.id === id) ?? null
  }
  return null
})

const markdownSlug = computed(() => {
  const item = menuItem.value
  if (!item) return ''
  if (item.page_slug) return item.page_slug
  if (item.url?.startsWith('md:')) return item.url.slice(3)
  return ''
})

const isMarkdownMode = computed(() => !!markdownSlug.value)

const embeddedUrl = computed(() => {
  if (!menuItem.value || isMarkdownMode.value) return ''
  return buildEmbeddedUrl(
    menuItem.value.url,
    authStore.user?.id,
    authStore.token,
    pageTheme.value,
    locale.value,
  )
})

const isValidUrl = computed(() => {
  if (isMarkdownMode.value) return false
  const url = embeddedUrl.value
  return url.startsWith('http://') || url.startsWith('https://')
})

function generateHeadingId(text: string, index: number): string {
  const base = text
    .toLowerCase()
    .replace(/[^\w一-鿿]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return base ? `${base}-${index}` : `heading-${index}`
}

function isRelativeMarkdownAsset(src: string): boolean {
  const trimmed = src.trim()
  if (!trimmed || /^[a-z][a-z0-9+.-]*:/i.test(trimmed) || trimmed.startsWith('//') || trimmed.startsWith('/')) {
    return false
  }
  const [pathPart] = trimmed.split(/([?#].*)/, 2)
  return pathPart
    .split('/')
    .filter((part) => part && part !== '.')
    .every((part) => part !== '..' && !part.includes('\\'))
}

function buildPageImageUrl(slug: string, src: string): string {
  const trimmed = src.trim()
  const [pathPart, suffix = ''] = trimmed.split(/([?#].*)/, 2)
  const encodedPath = pathPart
    .split('/')
    .filter((part) => part && part !== '.')
    .map((part) => encodeURIComponent(part))
    .join('/')
  return buildApiUrl(`/pages/${encodeURIComponent(slug)}/images/${encodedPath}${suffix}`)
}

async function fetchAndRenderMarkdown(slug: string) {
  loading.value = true
  tocItems.value = []
  activeHeadingId.value = ''
  try {
    const resp = await fetch(buildApiUrl(`/pages/${encodeURIComponent(slug)}`), {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {},
    })
    if (!resp.ok) {
      renderedHtml.value = `<p class="views-user-custom-page-view__description-2">${t('common.pageNotFound')}</p>`
      return
    }
    let raw = await resp.text()

    raw = raw.replace(
      /!\[([^\]]*)\]\(([^)]+)\)/g,
      (match, alt, src) => isRelativeMarkdownAsset(src) ? `![${alt}](${buildPageImageUrl(slug, src)})` : match
    )

    const html = marked.parse(raw) as string
    const sanitized = DOMPurify.sanitize(html, {
      ADD_TAGS: ['iframe'],
      ADD_ATTR: ['allowfullscreen', 'frameborder', 'src'],
    })

    // Inject IDs into headings and build TOC
    const toc: TocItem[] = []
    let headingIndex = 0
    const withIds = sanitized.replace(
      /<(h[1-4])[^>]*>(.*?)<\/h[1-4]>/gi,
      (_, tag: string, content: string) => {
        const level = parseInt(tag[1])
        const text = content.replace(/<[^>]+>/g, '').trim()
        const id = generateHeadingId(text, headingIndex++)
        toc.push({ id, text, level })
        return `<${tag} id="${id}">${content}</${tag}>`
      }
    )

    renderedHtml.value = withIds
    tocItems.value = toc
  } catch {
    renderedHtml.value = '<p class="views-user-custom-page-view__description-2">Failed to load page</p>'
  } finally {
    loading.value = false
    await nextTick()
    await nextTick()
    injectCopyButtons()
  }
}

function scrollToHeading(id: string) {
  const container = markdownContainer.value
  if (!container) return
  const el = container.querySelector(`#${CSS.escape(id)}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
    activeHeadingId.value = id
    if (window.innerWidth <= 640) {
      tocVisible.value = false
    }
  }
}

let scrollRafId = 0
function onContentScroll() {
  if (scrollRafId) return
  scrollRafId = requestAnimationFrame(() => {
    scrollRafId = 0
    const container = markdownContainer.value
    if (!container || tocItems.value.length === 0) return

    const containerRect = container.getBoundingClientRect()
    let current = ''

    for (const item of tocItems.value) {
      const el = container.querySelector(`#${CSS.escape(item.id)}`) as HTMLElement | null
      if (el) {
        const elRect = el.getBoundingClientRect()
        if (elRect.top - containerRect.top <= 100) {
          current = item.id
        }
      }
    }
    activeHeadingId.value = current
  })
}

function injectCopyButtons() {
  const container = markdownContainer.value
  if (!container) return

  container.querySelectorAll('pre').forEach((pre) => {
    if (pre.querySelector('.copy-btn')) return
    const btn = document.createElement('button')
    btn.className = 'copy-btn'
    btn.textContent = t('customPage.copyCode')
    btn.addEventListener('click', async () => {
      const code = pre.querySelector('code')?.textContent ?? pre.textContent ?? ''
      try {
        await navigator.clipboard.writeText(code)
        btn.textContent = t('customPage.copiedCode')
        setTimeout(() => { btn.textContent = t('customPage.copyCode') }, 2000)
      } catch {
        btn.textContent = t('customPage.copyCodeFailed')
        setTimeout(() => { btn.textContent = t('customPage.copyCode') }, 2000)
      }
    })
    pre.style.position = 'relative'
    pre.appendChild(btn)
  })
}

watch(markdownSlug, (slug) => {
  if (slug) {
    fetchAndRenderMarkdown(slug)
  } else {
    renderedHtml.value = ''
    tocItems.value = []
  }
}, { immediate: true })

onMounted(async () => {
  pageTheme.value = detectTheme()

  if (typeof document !== 'undefined') {
    themeObserver = new MutationObserver(() => {
      pageTheme.value = detectTheme()
    })
    themeObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class'],
    })
  }

  if (appStore.publicSettingsLoaded) return
  loading.value = true
  try {
    await appStore.fetchPublicSettings()
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (themeObserver) {
    themeObserver.disconnect()
    themeObserver = null
  }
})
</script>

<style scoped>
.custom-page-layout {
  height: calc(100vh - var(--app-shell-height) - 3.25rem);
}

.toc-sidebar {
  width: min(240px, 30%);
  min-width: 160px;
  max-width: 280px;
  overflow: hidden;
}

@media (max-width: 640px) {
  .toc-sidebar {
    position: absolute;
    left: 0;
    top: 0;
    z-index: 20;
    width: 70%;
    max-width: 240px;
    height: 100%;
    box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  }
}

.toc-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--color-border);
}

.toc-title {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-weight: 600;
}

.toc-close-btn {
  padding: 0.25rem;
  border-radius: var(--radius-sm);
  color: var(--color-text-tertiary);
  transition: background-color 150ms ease, color 150ms ease;
}

.toc-close-btn:hover {
  background: rgba(120, 120, 128, 0.14);
  color: var(--color-text-secondary);
}

.toc-nav {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
}

.toc-item {
  display: block;
  overflow: hidden;
  padding: 0.375rem 0.5rem;
  border-radius: var(--radius-sm);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: background-color 150ms ease, color 150ms ease;
}

.toc-item:hover {
  background: rgba(120, 120, 128, 0.14);
  color: var(--color-text-primary);
}

.toc-item.toc-active {
  background: var(--color-primary-subtle);
  color: var(--color-primary);
  font-weight: 500;
}

.toc-level-1 { padding-left: 8px; }
.toc-level-2 { padding-left: 20px; }
.toc-level-3 { padding-left: 32px; }
.toc-level-4 { padding-left: 44px; }

.toc-toggle-btn {
  position: absolute;
  top: 0.5rem;
  left: 0.5rem;
  z-index: 10;
  display: flex;
  align-items: center;
  padding: 0.375rem 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface-elevated);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  box-shadow: var(--shadow-sm);
  cursor: pointer;
  transition: background-color 150ms ease;
}

.toc-toggle-btn:hover {
  background: var(--color-surface-hover);
}

.custom-embed-shell {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  border-radius: var(--radius-xl);
}

.custom-open-fab {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  z-index: 10;
  box-shadow: var(--shadow-sm);
}

.custom-embed-frame {
  display: block;
  margin: 0;
  width: 100%;
  height: 100%;
  border: 0;
  border-radius: 0;
  box-shadow: none;
  background: transparent;
}
</style>

<style>
.markdown-page-content {
  line-height: 1.7;
  color: inherit;
}
.markdown-page-content h1 {
  margin: 2rem 0 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--color-border);
  font-size: var(--font-size-3xl);
  font-weight: 700;
}
.markdown-page-content h2 {
  margin: 1.5rem 0 0.75rem;
  font-size: var(--font-size-2xl);
  font-weight: 700;
}
.markdown-page-content h3 {
  margin: 1.25rem 0 0.5rem;
  font-size: var(--font-size-xl);
  font-weight: 600;
}
.markdown-page-content h4 {
  margin: 1rem 0 0.5rem;
  font-size: var(--font-size-lg);
  font-weight: 600;
}
.markdown-page-content p {
  margin-bottom: 1rem;
}
.markdown-page-content ul {
  margin-bottom: 1rem;
  padding-left: 1.5rem;
  list-style-type: disc;
}
.markdown-page-content ol {
  margin-bottom: 1rem;
  padding-left: 1.5rem;
  list-style-type: decimal;
}
.markdown-page-content li {
  margin-bottom: 0.25rem;
}
.markdown-page-content a {
  color: var(--color-primary);
  text-decoration: underline;
}
.markdown-page-content a:hover {
  color: var(--color-primary-hover);
}
.markdown-page-content blockquote {
  margin: 1rem 0;
  padding-left: 1rem;
  border-left: 4px solid var(--color-border-strong);
  color: var(--color-text-tertiary);
  font-style: italic;
}
.markdown-page-content img {
  max-width: 100%;
  height: auto;
  margin: 1rem 0;
  border-radius: var(--radius-md);
}
.markdown-page-content table {
  width: 100%;
  margin: 1rem 0;
  border-collapse: collapse;
}
.markdown-page-content th {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface-muted);
  font-weight: 600;
  text-align: left;
}
.markdown-page-content td {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border);
}
.markdown-page-content code {
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  background: rgba(120, 120, 128, 0.14);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: var(--font-size-sm);
}
.markdown-page-content pre {
  position: relative;
  margin: 1rem 0;
  padding: 1rem;
  overflow-x: auto;
  border-radius: var(--radius-md);
  background: #0c0c0e;
  color: #f1f5f9;
}
.markdown-page-content pre code {
  padding: 0;
  background: transparent;
  color: inherit;
}
.markdown-page-content hr {
  margin: 1.5rem 0;
  border-color: var(--color-border);
}

.copy-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 4px 10px;
  font-size: var(--font-size-xs);
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.15);
  color: #e2e8f0;
  border: 1px solid rgba(255, 255, 255, 0.2);
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s, background 0.2s;
  font-family: inherit;
}
.copy-btn:hover { background: rgba(255, 255, 255, 0.25); }
pre:hover .copy-btn { opacity: 1; }
</style>
