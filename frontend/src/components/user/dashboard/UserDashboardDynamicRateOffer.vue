<template>
  <section
    v-if="offers.length > 0"
    class="dashboard-dynamic-rate-offer card"
    :aria-label="t('dashboard.dynamicRateOffer.title')"
  >
    <header class="dashboard-dynamic-rate-offer__header">
      <div class="dashboard-dynamic-rate-offer__title-group">
        <span class="dashboard-dynamic-rate-offer__icon"><Icon name="sparkles" size="md" /></span>
        <div>
          <h2 class="dashboard-dynamic-rate-offer__title">{{ t('dashboard.dynamicRateOffer.title') }}</h2>
          <p class="dashboard-dynamic-rate-offer__description">{{ t('dashboard.dynamicRateOffer.description') }}</p>
        </div>
      </div>
    </header>

    <ul class="dashboard-dynamic-rate-offer__list">
      <li
        v-for="offer in offers"
        :key="`${offer.group_id}:${offer.rule_id}`"
        class="dashboard-dynamic-rate-offer__item"
      >
        <strong class="dashboard-dynamic-rate-offer__group">{{ offer.group_name || offer.rule_name }}</strong>
        <dl class="dashboard-dynamic-rate-offer__times">
          <div>
            <dt>{{ t('dashboard.dynamicRateOffer.startsAt') }}</dt>
            <dd>
              <time :datetime="offer.start_at">{{ formatDateTimeToMinute(offer.start_at) }}</time>
            </dd>
          </div>
          <div>
            <dt>{{ t('dashboard.dynamicRateOffer.endsAt') }}</dt>
            <dd>
              <time :datetime="offer.end_at">{{ formatDateTimeToMinute(offer.end_at) }}</time>
            </dd>
          </div>
        </dl>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getDynamicRateOffers } from '@/api/userLevel'
import Icon from '@/components/icons/Icon.vue'
import type { DynamicRateOffer } from '@/types'
import { formatDateTimeToMinute } from '@/utils/format'

const { t } = useI18n()
const offers = ref<DynamicRateOffer[]>([])

onMounted(async () => {
  try {
    offers.value = await getDynamicRateOffers()
  } catch {
    offers.value = []
  }
})
</script>

<style scoped>
.dashboard-dynamic-rate-offer {
  overflow: hidden;
}

.dashboard-dynamic-rate-offer__header,
.dashboard-dynamic-rate-offer__title-group {
  display: flex;
  align-items: center;
}

.dashboard-dynamic-rate-offer__header {
  gap: 1rem;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.dashboard-dynamic-rate-offer__title-group {
  min-width: 0;
  gap: 0.75rem;
}

.dashboard-dynamic-rate-offer__icon {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: var(--radius-md);
  color: var(--color-text-success);
  background: var(--color-success-subtle);
}

.dashboard-dynamic-rate-offer__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
  font-weight: 650;
}

.dashboard-dynamic-rate-offer__description {
  margin: 0.125rem 0 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.dashboard-dynamic-rate-offer__list {
  display: grid;
  margin: 0;
  padding: 0;
  list-style: none;
}

.dashboard-dynamic-rate-offer__item {
  display: grid;
  gap: 0.5rem;
  padding: 0.875rem 1.5rem 1rem;
}

.dashboard-dynamic-rate-offer__item + .dashboard-dynamic-rate-offer__item {
  border-top: 1px solid var(--color-border-subtle);
}

.dashboard-dynamic-rate-offer__group {
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: 650;
}

.dashboard-dynamic-rate-offer__times {
  display: grid;
  gap: 0.5rem;
  margin: 0;
}

.dashboard-dynamic-rate-offer__times > div {
  display: grid;
  gap: 0.125rem;
}

.dashboard-dynamic-rate-offer__times dt {
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.dashboard-dynamic-rate-offer__times dd {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-weight: 600;
}

@media (max-width: 767px) {
  .dashboard-dynamic-rate-offer__description {
    display: none;
  }
}
</style>
