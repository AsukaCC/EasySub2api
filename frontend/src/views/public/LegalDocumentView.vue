<template>
  <div class="views-public-legal-document-view__panel">
    <header class="views-public-legal-document-view__header">
      <div class="views-public-legal-document-view__panel-2">
        <RouterLink to="/home" class="views-public-legal-document-view__router-link">
          <template v-if="settings">
            <span class="views-public-legal-document-view__text">
              <img :src="siteLogo || '/logo.svg'" alt="Logo" class="views-public-legal-document-view__image" />
            </span>
            <span class="views-public-legal-document-view__text-2">
              {{ siteName }}
            </span>
          </template>
          <template v-else>
            <span class="views-public-legal-document-view__text-3" aria-hidden="true"></span>
            <span class="views-public-legal-document-view__text-4" aria-hidden="true"></span>
          </template>
        </RouterLink>
        <RouterLink
          to="/login"
          class="views-public-legal-document-view__router-link-2"
        >
          {{ t('home.login') }}
        </RouterLink>
      </div>
    </header>

    <main class="views-public-legal-document-view__main">
      <div v-if="loading" class="views-public-legal-document-view__panel-3">
        <div class="views-public-legal-document-view__panel-4"></div>
      </div>

      <section
        v-else-if="loadError"
        class="views-public-legal-document-view__section card-body"
      >
        <h1 class="views-public-legal-document-view__heading">{{ t('legal.loadFailed') }}</h1>
        <p class="views-public-legal-document-view__description">{{ t('legal.retryLater') }}</p>
      </section>

      <section
        v-else-if="!currentDocument"
        class="views-public-legal-document-view__section-2 card-body"
      >
        <div class="views-public-legal-document-view__panel-5">
          <span class="views-public-legal-document-view__text-5">
            <Icon name="document" size="sm" />
          </span>
          <div>
            <h1 class="views-public-legal-document-view__heading-2">{{ t('legal.notFound') }}</h1>
            <p class="views-public-legal-document-view__description-2">
              {{ t('legal.notFoundDescription') }}
            </p>
          </div>
        </div>
      </section>

      <article v-else>
        <div class="views-public-legal-document-view__panel-6">
          <div class="views-public-legal-document-view__panel-7">
            <span class="views-public-legal-document-view__text-6">
              <Icon :name="documentIcon" size="md" />
            </span>
            <div class="views-public-legal-document-view__panel-8">
              <p class="views-public-legal-document-view__description-3">{{ documentTypeLabel }}</p>
              <h1 class="views-public-legal-document-view__heading-3">
                {{ currentDocument.title }}
              </h1>
              <p v-if="updatedAt" class="views-public-legal-document-view__description-4">
                {{ t('legal.updatedAt', { date: updatedAt }) }}
              </p>
            </div>
          </div>
        </div>

        <div
          v-if="hasContent"
          class="legal-document-content"
          v-html="renderedHtml"
        ></div>
        <div
          v-else
          class="views-public-legal-document-view__panel-9"
        >
          {{ t('legal.empty') }}
        </div>
      </article>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getLocale } from '@/i18n'
import { sanitizeUrl } from '@/utils/url'
import { useAppStore } from '@/stores/app'
import type { LoginAgreementDocument } from '@/types'
import zhAdminCompliance from '../../../../docs/legal/admin-compliance.zh.md?raw'
import enAdminCompliance from '../../../../docs/legal/admin-compliance.en.md?raw'

type LegalDocumentIcon = 'document' | 'shield' | 'globe' | 'cog'

const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const settings = computed(() => appStore.cachedPublicSettings)
const loading = ref(!settings.value)
const loadError = ref(false)

marked.setOptions({
  breaks: true,
  gfm: true,
})

const documentId = computed(() => String(route.params.documentId || ''))
const isAdminComplianceDocument = computed(() => documentId.value === 'admin-compliance')
const documents = computed(() => settings.value?.login_agreement_documents ?? [])
const siteName = computed(() => settings.value?.site_name || 'EasySub2api')
const siteLogo = computed(() => sanitizeUrl(settings.value?.site_logo || '', {
  allowRelative: true,
  allowDataUrl: true,
}))
const updatedAt = computed(() =>
  isAdminComplianceDocument.value ? '' : settings.value?.login_agreement_updated_at || ''
)
const documentTypeLabel = computed(() =>
  isAdminComplianceDocument.value ? t('legal.adminCompliance') : t('legal.loginAgreement')
)

const currentDocument = computed<LoginAgreementDocument | null>(() => {
  if (isAdminComplianceDocument.value) {
    return {
      id: 'admin-compliance',
      title: t('adminCompliance.title'),
      content_md: getLocale() === 'zh' ? zhAdminCompliance : enAdminCompliance
    }
  }
  const id = documentId.value
  if (!id) {
    return null
  }
  return documents.value.find((doc) => doc.id === id) ?? null
})

const hasContent = computed(() => Boolean(currentDocument.value?.content_md?.trim()))

const renderedHtml = computed(() => {
  const content = currentDocument.value?.content_md?.trim() || ''
  if (!content) {
    return ''
  }
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

const documentIcon = computed<LegalDocumentIcon>(() => {
  const title = currentDocument.value?.title || ''
  if (title.includes('政策') || title.includes('隐私')) {
    return 'shield'
  }
  if (title.includes('国家') || title.includes('地区')) {
    return 'globe'
  }
  if (title.includes('特定')) {
    return 'cog'
  }
  return 'document'
})

onMounted(async () => {
  loadError.value = false
  const loadedSettings = await appStore.fetchPublicSettings()
  if (!loadedSettings) {
    loadError.value = true
  }
  loading.value = false
})
</script>

<style scoped>
.legal-document-content {
  line-height: 1.75;
  overflow-wrap: anywhere;
  color: inherit;
}

.legal-document-content :deep(h1) {
  margin: 2rem 0 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--color-border);
  font-size: var(--font-size-3xl);
  font-weight: 700;
}

.legal-document-content :deep(h2) {
  margin: 1.75rem 0 0.75rem;
  font-size: var(--font-size-2xl);
  font-weight: 700;
}

.legal-document-content :deep(h3) {
  margin: 1.5rem 0 0.5rem;
  font-size: var(--font-size-xl);
  font-weight: 600;
}

.legal-document-content :deep(h4) {
  margin: 1.25rem 0 0.5rem;
  font-size: var(--font-size-lg);
  font-weight: 600;
}

.legal-document-content :deep(p) {
  margin-bottom: 1rem;
  color: var(--color-text-secondary);
}

.legal-document-content :deep(a) {
  color: var(--color-primary);
  text-decoration: underline;
  text-underline-offset: 4px;
}

.legal-document-content :deep(a:hover) {
  color: var(--color-primary-hover);
}

.legal-document-content :deep(ul) {
  margin-bottom: 1rem;
  padding-left: 1.5rem;
  list-style-type: disc;
}

.legal-document-content :deep(ol) {
  margin-bottom: 1rem;
  padding-left: 1.5rem;
  list-style-type: decimal;
}

.legal-document-content :deep(li) {
  margin-bottom: 0.25rem;
  color: var(--color-text-secondary);
}

.legal-document-content :deep(blockquote) {
  margin: 1.25rem 0;
  padding-left: 1rem;
  border-left: 4px solid var(--color-border-strong);
  color: var(--color-text-tertiary);
}

.legal-document-content :deep(code) {
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  background: rgba(120, 120, 128, 0.14);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: var(--font-size-sm);
}

.legal-document-content :deep(pre) {
  margin: 1.25rem 0;
  padding: 1rem;
  overflow-x: auto;
  border-radius: var(--radius-md);
  background: #0c0c0e;
  color: #f1f5f9;
}

.legal-document-content :deep(pre code) {
  padding: 0;
  background: transparent;
  color: inherit;
}

.legal-document-content :deep(table) {
  display: block;
  width: 100%;
  margin: 1.25rem 0;
  overflow-x: auto;
  border-collapse: collapse;
}

.legal-document-content :deep(th) {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface-muted);
  font-weight: 600;
  text-align: left;
}

.legal-document-content :deep(td) {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border);
}

.legal-document-content :deep(img) {
  max-width: 100%;
  height: auto;
  margin: 1.25rem 0;
  border-radius: var(--radius-md);
}

.legal-document-content :deep(hr) {
  margin: 1.75rem 0;
  border-color: var(--color-border);
}
</style>
