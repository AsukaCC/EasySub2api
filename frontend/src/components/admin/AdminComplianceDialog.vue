<template>
  <BaseDialog
    :show="visible"
    :title="t('adminCompliance.title')"
    width="wide"
    :close-on-escape="false"
    :close-on-click-outside="false"
    :show-close-button="false"
    :z-index="80"
    @close="noop"
  >
    <div class="components-admin-admin-compliance-dialog__panel">
      <div class="components-admin-admin-compliance-dialog__panel-2">
        <div class="components-admin-admin-compliance-dialog__panel-3">
          <Icon name="exclamationTriangle" size="md" class="components-admin-admin-compliance-dialog__icon" />
          <div class="components-admin-admin-compliance-dialog__panel-4">
            <p class="components-admin-admin-compliance-dialog__description">{{ t('adminCompliance.blockingNotice') }}</p>
            <p class="components-admin-admin-compliance-dialog__description-2">{{ t('adminCompliance.riskNotice') }}</p>
          </div>
        </div>
      </div>

      <div class="components-admin-admin-compliance-dialog__panel-5">
        <section class="components-admin-admin-compliance-dialog__section">
          <div class="legal-document-content" v-html="renderedDocument"></div>
        </section>

        <aside class="components-admin-admin-compliance-dialog__aside">
          <div>
            <p class="components-admin-admin-compliance-dialog__description-3">
              {{ t('adminCompliance.version') }}
            </p>
            <p class="components-admin-admin-compliance-dialog__description-4">
              {{ complianceStore.status?.version || 'v2026.06.10' }}
            </p>
          </div>
          <a
            :href="documentUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="components-admin-admin-compliance-dialog__link"
          >
            <Icon name="externalLink" size="sm" />
            {{ t('adminCompliance.openDocument') }}
          </a>
          <p class="components-admin-admin-compliance-dialog__description-5">
            {{ t('adminCompliance.documentSource') }}
          </p>
        </aside>
      </div>

      <div class="components-admin-admin-compliance-dialog__panel-6">
        <label for="admin-compliance-phrase" class="components-admin-admin-compliance-dialog__label">
          {{ t('adminCompliance.inputLabel') }}
        </label>
        <div class="components-admin-admin-compliance-dialog__panel-7">
          {{ expectedPhrase }}
        </div>
        <Input
          id="admin-compliance-phrase"
          v-model="typedPhrase"
          :placeholder="t('adminCompliance.inputPlaceholder')"
          autocomplete="off"
          :disabled="complianceStore.submitting"
          :error="inputError"
          @enter="submit"
        />
      </div>

      <p class="components-admin-admin-compliance-dialog__description-6">
        {{ t('adminCompliance.legalNote') }}
      </p>
    </div>

    <template #footer>
      <div class="components-admin-admin-compliance-dialog__panel-8">
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="complianceStore.submitting"
          @click="logout"
        >
          {{ t('adminCompliance.logout') }}
        </button>
        <button
          type="button"
          class="btn btn-primary"
          :disabled="!canSubmit || complianceStore.submitting"
          @click="submit"
        >
          <span v-if="complianceStore.submitting">{{ t('common.submitting') }}</span>
          <span v-else>{{ t('adminCompliance.accept') }}</span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Input from '@/components/common/Input.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAdminComplianceStore, useAppStore, useAuthStore } from '@/stores'
import { getLocale } from '@/i18n'
import zhDocument from '../../../../docs/legal/admin-compliance.zh.md?raw'
import enDocument from '../../../../docs/legal/admin-compliance.en.md?raw'

const { t } = useI18n()
const complianceStore = useAdminComplianceStore()
const authStore = useAuthStore()
const appStore = useAppStore()
const typedPhrase = ref('')
const attemptedSubmit = ref(false)

marked.setOptions({
  breaks: true,
  gfm: true,
})

const visible = computed(() => authStore.isAuthenticated && authStore.isAdmin && complianceStore.shouldShow)
const expectedPhrase = computed(() => complianceStore.expectedPhrase)
const canSubmit = computed(() => typedPhrase.value.trim() === expectedPhrase.value)
const currentDocument = computed(() => getLocale() === 'zh' ? zhDocument : enDocument)
const documentUrl = computed(() => {
  if (getLocale() === 'zh') {
    return complianceStore.status?.document_url_zh || 'https://github.com/AsukaCC/EasySub2api/blob/main/docs/legal/admin-compliance.zh.md'
  }
  return complianceStore.status?.document_url_en || 'https://github.com/AsukaCC/EasySub2api/blob/main/docs/legal/admin-compliance.en.md'
})
const inputError = computed(() => {
  if (!attemptedSubmit.value || canSubmit.value) {
    return ''
  }
  return t('adminCompliance.inputMismatch')
})
const renderedDocument = computed(() => {
  const html = marked.parse(currentDocument.value) as string
  return DOMPurify.sanitize(html)
})

watch(expectedPhrase, () => {
  typedPhrase.value = ''
  attemptedSubmit.value = false
})

watch(visible, (isVisible) => {
  if (isVisible) {
    typedPhrase.value = ''
    attemptedSubmit.value = false
  }
})

function noop(): void {
  // 强制确认弹窗不允许通过关闭按钮绕过。
}

async function submit(): Promise<void> {
  attemptedSubmit.value = true
  if (!canSubmit.value) {
    return
  }

  try {
    const status = await complianceStore.accept(typedPhrase.value.trim())
    if (!status.required) {
      appStore.showSuccess(t('adminCompliance.accepted'))
      typedPhrase.value = ''
      attemptedSubmit.value = false
    }
  } catch (error) {
    const message = (error as { message?: string })?.message || t('adminCompliance.acceptFailed')
    appStore.showError(message)
  }
}

async function logout(): Promise<void> {
  await authStore.logout()
  window.location.href = '/login'
}
</script>

<style scoped>
.legal-document-content {
  line-height: 1.75;
  overflow-wrap: anywhere;
  color: inherit;
}

.legal-document-content :deep(h1) {
  margin-bottom: 1rem;
  color: var(--color-text-primary);
  font-size: var(--font-size-2xl);
  font-weight: 700;
}

.legal-document-content :deep(h2) {
  margin: 1.5rem 0 0.75rem;
  color: var(--color-text-primary);
  font-size: var(--font-size-xl);
  font-weight: 600;
}

.legal-document-content :deep(p) {
  margin-bottom: 1rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.legal-document-content :deep(ul),
.legal-document-content :deep(ol) {
  margin-bottom: 1rem;
  padding-left: 1.5rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.legal-document-content :deep(ul) {
  list-style-type: disc;
}

.legal-document-content :deep(ol) {
  list-style-type: decimal;
}

.legal-document-content :deep(li) {
  margin-bottom: 0.25rem;
}

.legal-document-content :deep(strong) {
  color: var(--color-text-primary);
  font-weight: 600;
}
</style>
