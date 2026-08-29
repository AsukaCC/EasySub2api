<template>
  <div class="card quick-actions">
    <div class="quick-actions__header">
      <h2 class="quick-actions__title">{{ t('dashboard.quickActions') }}</h2>
    </div>
    <div class="quick-actions__list">
      <button
        v-for="action in actions"
        :key="action.route"
        type="button"
        class="quick-actions__item"
        :class="`quick-actions__item--${action.tone}`"
        @click="router.push(action.route)"
      >
        <span class="quick-actions__icon-box">
          <Icon :name="action.icon" size="lg" />
        </span>
        <span class="quick-actions__text">
          <span class="quick-actions__label">{{ t(action.labelKey) }}</span>
          <span class="quick-actions__hint">{{ t(action.hintKey) }}</span>
        </span>
        <Icon name="chevronRight" size="md" class="quick-actions__chevron" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const router = useRouter()
const { t } = useI18n()

const actions = [
  {
    route: '/keys',
    icon: 'key',
    tone: 'blue',
    labelKey: 'dashboard.createApiKey',
    hintKey: 'dashboard.generateNewKey',
  },
  {
    route: '/usage',
    icon: 'chart',
    tone: 'green',
    labelKey: 'dashboard.viewUsage',
    hintKey: 'dashboard.checkDetailedLogs',
  },
  {
    route: '/redeem',
    icon: 'gift',
    tone: 'amber',
    labelKey: 'dashboard.redeemCode',
    hintKey: 'dashboard.addBalanceWithCode',
  },
] as const
</script>

<style scoped>
.quick-actions__header {
  padding: 0.75rem 1.25rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.quick-actions__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
  font-weight: 650;
}

.quick-actions__list {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  grid-auto-rows: 1fr;
  gap: 0.75rem;
  padding: 0.75rem;
}

@media (min-width: 768px) {
  .quick-actions__list {
    grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
  }
}

.quick-actions__item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-height: 3.5rem;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-bg-thin);
  color: var(--color-text-primary);
  text-align: left;
  cursor: pointer;
  transition:
    border-color 160ms ease,
    background-color 160ms ease,
    box-shadow 160ms ease,
    transform 160ms ease;
}

.quick-actions__item:hover {
  border-color: rgb(10 132 255 / 0.22);
  background: var(--glass-bg-thick);
  box-shadow:
    0 8px 18px rgb(12 12 14 / 0.08),
    0 1px 0 var(--glass-highlight) inset;
  transform: translateY(-1px);
}

.quick-actions__item:active {
  transform: translateY(0);
  box-shadow: none;
  transition-duration: 60ms;
}

.quick-actions__item:focus-visible {
  outline: 2px solid rgb(10 132 255 / 0.5);
  outline-offset: 2px;
}

.quick-actions__icon-box {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  transition: transform 160ms ease;
}

.quick-actions__item:hover .quick-actions__icon-box {
  transform: scale(1.06);
}

.quick-actions__item--blue .quick-actions__icon-box {
  border-color: rgb(10 132 255 / 0.18);
  background: rgb(10 132 255 / 0.1);
  color: rgb(10 132 255);
}

.quick-actions__item--green .quick-actions__icon-box {
  border-color: rgb(5 150 105 / 0.18);
  background: rgb(5 150 105 / 0.1);
  color: rgb(5 150 105);
}

.quick-actions__item--amber .quick-actions__icon-box {
  border-color: rgb(217 119 6 / 0.18);
  background: rgb(217 119 6 / 0.1);
  color: rgb(217 119 6);
}

.quick-actions__text {
  display: grid;
  flex: 1;
  min-width: 0;
  gap: 0.125rem;
}

.quick-actions__label {
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: 600;
  line-height: 1.25rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quick-actions__hint {
  overflow: hidden;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
  line-height: 1rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quick-actions__chevron {
  flex: 0 0 auto;
  color: var(--color-text-tertiary);
  transition:
    color 160ms ease,
    transform 160ms ease;
}

.quick-actions__item:hover .quick-actions__chevron {
  transform: translateX(2px);
}

.quick-actions__item--blue:hover .quick-actions__chevron {
  color: rgb(10 132 255);
}

.quick-actions__item--green:hover .quick-actions__chevron {
  color: rgb(16 185 129);
}

.quick-actions__item--amber:hover .quick-actions__chevron {
  color: rgb(245 158 11);
}
</style>

<style>
/* 暗色覆盖放在非 scoped 块:Vue scoped 编译器在生产构建中会丢弃
   `:global(.dark) ...` 规则(与 SettingsView 中的处理一致)。 */
.dark .quick-actions__item:hover {
  border-color: rgb(10 132 255 / 0.35);
  box-shadow:
    0 12px 26px rgb(0 0 0 / 0.22),
    0 1px 0 var(--glass-highlight) inset;
}

.dark .quick-actions__item--blue .quick-actions__icon-box {
  border-color: rgb(58 162 255 / 0.22);
  background: rgb(58 162 255 / 0.12);
  color: rgb(58 162 255);
}

.dark .quick-actions__item--green .quick-actions__icon-box {
  border-color: rgb(52 211 153 / 0.22);
  background: rgb(52 211 153 / 0.12);
  color: rgb(52 211 153);
}

.dark .quick-actions__item--amber .quick-actions__icon-box {
  border-color: rgb(251 191 36 / 0.22);
  background: rgb(251 191 36 / 0.12);
  color: rgb(251 191 36);
}

.dark .quick-actions__item--blue:hover .quick-actions__chevron {
  color: rgb(58 162 255);
}

.dark .quick-actions__item--green:hover .quick-actions__chevron {
  color: rgb(52 211 153);
}

.dark .quick-actions__item--amber:hover .quick-actions__chevron {
  color: rgb(251 191 36);
}
</style>
