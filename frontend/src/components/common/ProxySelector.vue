<template>
  <div class="components-common-proxy-selector__panel" ref="containerRef">
    <button
      ref="triggerRef"
      type="button"
      @click="toggle"
      :disabled="disabled"
      :class="[
        'select-trigger',
        isOpen && 'select-trigger-open',
        disabled && 'select-trigger-disabled'
      ]"
    >
      <span class="select-value">
        {{ selectedLabel }}
      </span>
      <span class="select-icon">
        <Icon
          name="chevronDown"
          size="md"
          :class="['components-common-proxy-selector__icon-6', isOpen && 'components-common-proxy-selector__icon-7']"
        />
      </span>
    </button>

    <Teleport to="body">
      <Transition name="select-dropdown">
        <div
          v-if="isOpen"
          ref="panelRef"
          class="select-dropdown"
          :style="panelStyle"
          @click.stop
        >
        <!-- Search and Batch Test Header -->
        <div class="select-header">
          <div class="select-search">
            <Icon name="search" size="sm" class="components-common-proxy-selector__icon" />
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="text"
              :placeholder="t('admin.proxies.searchProxies')"
              class="select-search-input"
              @click.stop
            />
          </div>
          <button
            v-if="proxies.length > 0"
            type="button"
            @click.stop="handleBatchTest"
            :disabled="batchTesting"
            class="batch-test-btn"
            :title="t('admin.proxies.batchTest')"
          >
            <svg v-if="batchTesting" class="components-common-proxy-selector__icon-2" fill="none" viewBox="0 0 24 24">
              <circle
                class="components-common-proxy-selector__circle"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="components-common-proxy-selector__path"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            <Icon v-else name="play" size="sm" />
          </button>
        </div>

        <!-- Options list -->
        <div class="select-options">
          <!-- No Proxy option -->
          <div
            @click="selectOption(null)"
            :class="['select-option', modelValue === null && 'select-option-selected']"
          >
            <span class="select-option-label">{{ t('admin.accounts.noProxy') }}</span>
            <Icon v-if="modelValue === null" name="check" size="sm" class="components-common-proxy-selector__icon-3" />
          </div>

          <!-- Proxy options -->
          <div
            v-for="proxy in filteredProxies"
            :key="proxy.id"
            @click="selectOption(proxy.id)"
            :class="['select-option', modelValue === proxy.id && 'select-option-selected']"
          >
            <div class="components-common-proxy-selector__panel-2">
              <div class="components-common-proxy-selector__panel-3">
                <span class="components-common-proxy-selector__text">{{ proxy.name }}</span>
                <!-- Account count badge -->
                <span
                  v-if="proxy.account_count !== undefined"
                  class="components-common-proxy-selector__text-2"
                >
                  {{ proxy.account_count }}
                </span>
                <!-- Test result badges -->
                <template v-if="testResults[proxy.id]">
                  <span
                    v-if="testResults[proxy.id].success"
                    class="components-common-proxy-selector__text-3"
                  >
                    <span v-if="testResults[proxy.id].country">{{
                      testResults[proxy.id].country
                    }}</span>
                    <span v-if="testResults[proxy.id].latency_ms"
                      >{{ testResults[proxy.id].latency_ms }}ms</span
                    >
                  </span>
                  <span
                    v-else
                    class="components-common-proxy-selector__text-4"
                  >
                    {{ t('admin.proxies.testFailed') }}
                  </span>
                </template>
              </div>
              <div class="components-common-proxy-selector__panel-4">
                {{ proxy.protocol }}://{{ proxy.host }}:{{ proxy.port }}
              </div>
            </div>

            <!-- Individual test button -->
            <button
              type="button"
              @click.stop="handleTestProxy(proxy)"
              :disabled="testingProxyIds.has(proxy.id)"
              class="test-btn"
              :title="t('admin.proxies.testConnection')"
            >
              <svg
                v-if="testingProxyIds.has(proxy.id)"
                class="components-common-proxy-selector__icon-4"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="components-common-proxy-selector__circle"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="components-common-proxy-selector__path"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              <Icon v-else name="play" size="xs" />
            </button>

            <Icon
              v-if="modelValue === proxy.id"
              name="check"
              size="sm"
              class="components-common-proxy-selector__icon-5"
            />
          </div>

          <!-- Empty state -->
          <div v-if="filteredProxies.length === 0 && searchQuery" class="select-empty">
            {{ t('common.noOptionsFound') }}
          </div>
        </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import Icon from '@/components/icons/Icon.vue'
import type { Proxy } from '@/types'
import { useFloatingPanel } from '@/composables/useFloatingPanel'

const { t } = useI18n()

interface ProxyTestResult {
  success: boolean
  message: string
  latency_ms?: number
  ip_address?: string
  city?: string
  region?: string
  country?: string
}

interface Props {
  modelValue: string | null
  proxies: Proxy[]
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string | null]
}>()

const isOpen = ref(false)
const searchQuery = ref('')
const containerRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const { panelRef, style: panelStyle } = useFloatingPanel(triggerRef, isOpen, {
  maxWidth: 480,
  align: 'start',
  minComfortableHeight: 220
})
const searchInputRef = ref<HTMLInputElement | null>(null)

// Test state
const testResults = reactive<Record<string, ProxyTestResult>>({})
const testingProxyIds = reactive(new Set<string>())
const batchTesting = ref(false)

const selectedProxy = computed(() => {
  if (props.modelValue === null) return null
  return props.proxies.find((p) => p.id === props.modelValue) || null
})

const selectedLabel = computed(() => {
  if (!selectedProxy.value) {
    return t('admin.accounts.noProxy')
  }
  const proxy = selectedProxy.value
  return `${proxy.name} (${proxy.protocol}://${proxy.host}:${proxy.port})`
})

const filteredProxies = computed(() => {
  if (!searchQuery.value) {
    return props.proxies
  }
  const query = searchQuery.value.toLowerCase()
  return props.proxies.filter((proxy) => {
    const name = proxy.name.toLowerCase()
    const host = proxy.host.toLowerCase()
    return name.includes(query) || host.includes(query)
  })
})

const toggle = () => {
  if (props.disabled) return
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    nextTick(() => {
      searchInputRef.value?.focus()
    })
  }
}

const selectOption = (value: string | null) => {
  emit('update:modelValue', value)
  isOpen.value = false
  searchQuery.value = ''
}

const handleTestProxy = async (proxy: Proxy) => {
  if (testingProxyIds.has(proxy.id)) return

  testingProxyIds.add(proxy.id)
  try {
    const result = await adminAPI.proxies.testProxy(proxy.id)
    testResults[proxy.id] = result
  } catch (error: any) {
    testResults[proxy.id] = {
      success: false,
      message: error.response?.data?.detail || 'Test failed'
    }
  } finally {
    testingProxyIds.delete(proxy.id)
  }
}

const handleBatchTest = async () => {
  if (batchTesting.value || props.proxies.length === 0) return

  batchTesting.value = true

  // Test all proxies in parallel
  const testPromises = props.proxies.map(async (proxy) => {
    testingProxyIds.add(proxy.id)
    try {
      const result = await adminAPI.proxies.testProxy(proxy.id)
      testResults[proxy.id] = result
    } catch (error: any) {
      testResults[proxy.id] = {
        success: false,
        message: error.response?.data?.detail || 'Test failed'
      }
    } finally {
      testingProxyIds.delete(proxy.id)
    }
  })

  await Promise.all(testPromises)
  batchTesting.value = false
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as Node
  if (containerRef.value?.contains(target) || panelRef.value?.contains(target)) return
  isOpen.value = false
  searchQuery.value = ''
}

const handleEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && isOpen.value) {
    isOpen.value = false
    searchQuery.value = ''
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleEscape)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEscape)
})
</script>

<style scoped>
.select-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  width: 100%;
  padding: 0.625rem 1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--glass-field-bg);
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: border-color 200ms ease, box-shadow 200ms ease;
}

.select-trigger:hover {
  border-color: var(--color-border-strong);
}

.select-trigger:focus {
  border-color: var(--color-primary);
  outline: none;
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.2);
}

.select-trigger-open {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.2);
}

.select-trigger-disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.select-value {
  flex: 1;
  overflow: hidden;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.select-icon {
  flex-shrink: 0;
  color: var(--color-text-tertiary);
}

.select-dropdown {
  position: absolute;
  z-index: 100;
  width: 100%;
  margin-top: 0.5rem;
  overflow: hidden;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-layer-floating-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  box-shadow: var(--glass-shadow-hover), 0 1px 0 var(--glass-highlight) inset;
}

.select-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.select-search {
  display: flex;
  flex: 1;
  align-items: center;
  gap: 0.5rem;
}

.select-search-input {
  flex: 1;
  background: transparent;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
}

.select-search-input:focus {
  outline: none;
}

.select-search-input::placeholder {
  color: var(--color-text-tertiary);
}

.batch-test-btn {
  flex-shrink: 0;
  padding: 0.375rem;
  border-radius: var(--radius-md);
  color: var(--color-text-tertiary);
  transition: background-color 150ms ease, color 150ms ease;
}

.batch-test-btn:hover:not(:disabled) {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}

.batch-test-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.select-options {
  max-height: 15rem;
  overflow-y: auto;
  padding: 0.25rem 0;
}

.select-option {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.5rem 0.875rem;
  margin: 0 0.25rem 0.125rem;
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  transition: color 150ms ease, background-color 150ms ease, border-color 150ms ease, box-shadow 150ms ease;
}

.select-option:hover {
  color: var(--color-text-primary);
  border-color: var(--glass-border);
  background-color: var(--glass-bg-interactive-hover);
  -webkit-backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}

.select-option-selected {
  background-color: var(--color-primary-subtle);
  border-color: var(--color-primary-border);
  color: var(--color-text-brand);
  font-weight: var(--font-weight-semibold);
}

.dark .select-option-selected {
  color: var(--color-text-brand);
  background-color: var(--color-primary-subtle);
}

.select-option-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.select-empty {
  padding: 2rem 1rem;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
  text-align: center;
}

.test-btn {
  flex-shrink: 0;
  padding: 0.25rem;
  border-radius: var(--radius-sm);
  color: var(--color-text-tertiary);
  transition: background-color 150ms ease, color 150ms ease;
}

.test-btn:hover:not(:disabled) {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}

.test-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

/* Dropdown animation */
.select-dropdown-enter-active,
.select-dropdown-leave-active {
  transition: all 0.2s ease;
}

.select-dropdown-enter-from,
.select-dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
