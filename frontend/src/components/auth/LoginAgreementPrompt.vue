<template>
  <div
    v-if="mode === 'checkbox' && documents.length > 0"
    class="components-auth-login-agreement-prompt__panel"
  >
    <div class="components-auth-login-agreement-prompt__panel-2">
      <input
        id="login-agreement-consent"
        type="checkbox"
        :checked="accepted"
        class="components-auth-login-agreement-prompt__field"
        @change="handleCheckboxChange"
      />
      <div class="components-auth-login-agreement-prompt__panel-3">
        <p class="components-auth-login-agreement-prompt__description">
          <label
            for="login-agreement-consent"
            class="components-auth-login-agreement-prompt__label"
          >
            {{ t('legal.loginAgreementPrompt.checkboxPrefix') }}
          </label>
          <template v-for="(doc, index) in documents" :key="doc.id || doc.title">
            <RouterLink
              :to="documentRoute(doc)"
              target="_blank"
              rel="noopener noreferrer"
              class="components-auth-login-agreement-prompt__router-link"
            >
              {{ doc.title }}
            </RouterLink>
            <span v-if="index < documents.length - 1">{{ t('legal.loginAgreementPrompt.documentSeparator') }}</span>
          </template>
        </p>
      </div>
    </div>
  </div>

  <div
    v-else-if="!accepted && documents.length > 0"
    class="components-auth-login-agreement-prompt__panel-4"
  >
    <div class="components-auth-login-agreement-prompt__panel-5">
      <Icon name="shield" size="sm" class="components-auth-login-agreement-prompt__icon" />
      <div class="components-auth-login-agreement-prompt__panel-3">
        <p class="components-auth-login-agreement-prompt__description-2">{{ t('legal.loginAgreementPrompt.noticeTitle') }}</p>
        <p class="components-auth-login-agreement-prompt__description-3">
          {{ t('legal.loginAgreementPrompt.noticeDescription') }}
        </p>
      </div>
      <button
        type="button"
        class="components-auth-login-agreement-prompt__action"
        @click="emit('open')"
      >
        {{ t('legal.loginAgreementPrompt.viewTerms') }}
      </button>
    </div>
  </div>

  <Teleport to="body">
    <Transition name="agreement-fade">
      <div
        v-if="dialogVisible"
        class="components-auth-login-agreement-prompt__panel-6"
      >
        <div class="components-auth-login-agreement-prompt__panel-7">
          <div class="components-auth-login-agreement-prompt__panel-8">
            <div class="components-auth-login-agreement-prompt__panel-9">
              <span class="components-auth-login-agreement-prompt__text">
                <Icon name="shield" size="md" />
              </span>
              <div class="components-auth-login-agreement-prompt__panel-3">
                <div class="components-auth-login-agreement-prompt__panel-10">
                  <h2 class="components-auth-login-agreement-prompt__heading">
                    {{ t('legal.loginAgreementPrompt.dialogTitle') }}
                  </h2>
                  <span
                    v-if="updatedAt"
                    class="components-auth-login-agreement-prompt__text-2"
                  >
                    {{ updatedAt }}
                  </span>
                </div>
                <p class="components-auth-login-agreement-prompt__description-4">
                  {{
                    t('legal.loginAgreementPrompt.dialogDescription', {
                      date: updatedAt || t('legal.loginAgreementPrompt.recently'),
                    })
                  }}
                </p>
              </div>
            </div>
          </div>

          <div class="components-auth-login-agreement-prompt__panel-11">
            <div class="components-auth-login-agreement-prompt__panel-12">
              <p class="components-auth-login-agreement-prompt__description-5">{{ t('legal.loginAgreementPrompt.relatedDocuments') }}</p>
            </div>
            <div class="components-auth-login-agreement-prompt__panel-13">
              <RouterLink
                v-for="(doc, index) in documents"
                :key="doc.id || doc.title"
                :to="documentRoute(doc)"
                target="_blank"
                rel="noopener noreferrer"
                class="components-auth-login-agreement-prompt__router-link-2"
              >
                <span class="components-auth-login-agreement-prompt__text-3">
                  <Icon :name="documentIcon(index, doc.title)" size="sm" />
                </span>
                <span class="components-auth-login-agreement-prompt__panel-3">
                  <span class="components-auth-login-agreement-prompt__text-4">{{ doc.title }}</span>
                </span>
                <span class="components-auth-login-agreement-prompt__text-5">
                  <Icon name="externalLink" size="sm" />
                </span>
              </RouterLink>
            </div>
          </div>

          <div class="components-auth-login-agreement-prompt__panel-14">
            <div class="components-auth-login-agreement-prompt__panel-15">
              <button
                type="button"
                class="components-auth-login-agreement-prompt__action-2"
                @click="emit('reject')"
              >
                {{ t('legal.loginAgreementPrompt.reject') }}
              </button>
              <button
                type="button"
                class="components-auth-login-agreement-prompt__action-3"
                @click="emit('accept')"
              >
                {{ t('legal.loginAgreementPrompt.accept') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { LoginAgreementDocument } from '@/types'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  accepted: boolean
  documents: LoginAgreementDocument[]
  mode: 'modal' | 'checkbox' | string
  updatedAt?: string
  visible: boolean
}>(), {
  updatedAt: ''
})

const emit = defineEmits<{
  accept: []
  reject: []
  open: []
}>()

const dialogVisible = computed(() => props.visible && documents.value.length > 0)
const documents = computed(() => props.documents.filter((doc) => doc.title.trim()))
const updatedAt = computed(() => props.updatedAt || '')
const accepted = computed(() => props.accepted)
const mode = computed(() => props.mode === 'checkbox' ? 'checkbox' : 'modal')

function documentRoute(doc: LoginAgreementDocument) {
  return {
    name: 'LegalDocument',
    params: {
      documentId: doc.id || doc.title,
    },
  }
}

function handleCheckboxChange(event: Event): void {
  const checked = (event.target as HTMLInputElement).checked
  if (checked) {
    emit('accept')
  } else {
    emit('reject')
  }
}

function documentIcon(index: number, title: string): 'document' | 'shield' | 'globe' | 'cog' {
  const normalizedTitle = title.toLowerCase()
  if (
    normalizedTitle.includes('policy') ||
    normalizedTitle.includes('privacy') ||
    title.includes('政策') ||
    title.includes('隐私')
  ) {
    return 'shield'
  }
  if (
    normalizedTitle.includes('country') ||
    normalizedTitle.includes('region') ||
    title.includes('国家') ||
    title.includes('地区')
  ) {
    return 'globe'
  }
  if (index === 3) {
    return 'cog'
  }
  return 'document'
}
</script>

<style scoped>
.agreement-fade-enter-active,
.agreement-fade-leave-active {
  transition: opacity 0.18s ease;
}

.agreement-fade-enter-from,
.agreement-fade-leave-to {
  opacity: 0;
}

.agreement-fade-enter-active > div,
.agreement-fade-leave-active > div {
  transition: transform 0.18s ease, opacity 0.18s ease;
}

.agreement-fade-enter-from > div,
.agreement-fade-leave-to > div {
  opacity: 0;
  transform: translateY(8px) scale(0.98);
}
</style>
