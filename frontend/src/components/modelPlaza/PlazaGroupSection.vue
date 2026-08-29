<template>
  <section
    class="components-model-plaza-plaza-group-section__section"
    :class="[platformBorderStrongClass(group.platform)]"
  >
    <!-- 分组头部:名称/平台/倍率徽章/专属/订阅徽章 + 描述 -->
    <header class="components-model-plaza-plaza-group-section__header">
      <div class="components-model-plaza-plaza-group-section__panel">
        <GroupBadge
          :name="group.name"
          :platform="group.platform as GroupPlatform"
          :subscription-type="(group.subscription_type || 'standard') as SubscriptionType"
          :rate-multiplier="group.rate_multiplier"
          :user-rate-multiplier="group.user_rate_multiplier ?? null"
          :peak-rate-enabled="group.peak_rate_enabled"
          :peak-start="group.peak_start"
          :peak-end="group.peak_end"
          :peak-rate-multiplier="group.peak_rate_multiplier"
          always-show-rate
        />
        <span
          v-if="group.is_exclusive"
          class="components-model-plaza-plaza-group-section__text"
        >
          <Icon name="shield" size="xs" class="components-model-plaza-plaza-group-section__icon" />
          {{ t('modelPlaza.badges.exclusive') }}
        </span>
        <span
          v-if="group.subscription_type === 'subscription'"
          class="components-model-plaza-plaza-group-section__text-2"
        >
          {{ t('modelPlaza.badges.subscription') }}
        </span>
      </div>
      <p v-if="group.description" class="components-model-plaza-plaza-group-section__description">
        {{ group.description }}
      </p>
      <p
        v-if="peakNote"
        class="components-model-plaza-plaza-group-section__description-2"
      >
        <Icon name="clock" size="xs" class="components-model-plaza-plaza-group-section__icon" />
        {{ peakNote }}
      </p>
    </header>

    <!-- 模型价格表:整行(含 hover 底色/分区底色)顶到卡片边缘,左右留白由表格首列/末列的 padding 提供 -->
    <div>
      <PlazaModelPricingTable
        v-if="group.models.length > 0"
        :models="group.models"
        :platform="group.platform"
        :rate-multiplier="group.rate_multiplier"
        :user-rate-multiplier="group.user_rate_multiplier ?? null"
        :image-rate-independent="group.image_rate_independent"
        :image-rate-multiplier="group.image_rate_multiplier"
      />
      <p v-else class="components-model-plaza-plaza-group-section__description-3">
        {{ t('modelPlaza.detail.noModels') }}
      </p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import PlazaModelPricingTable from './PlazaModelPricingTable.vue'
import type { ModelPlazaGroup } from '@/api/modelPlaza'
import type { GroupPlatform, SubscriptionType } from '@/types'
import { platformBorderStrongClass } from '@/utils/platformColors'
import { hasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  group: ModelPlazaGroup
}>()

const { t } = useI18n()
const appStore = useAppStore()

const peakNote = computed(() => {
  if (!hasPeakRate(props.group)) return ''
  const window = formatPeakRateWindow(
    props.group,
    serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset)
  )
  return t('modelPlaza.detail.peakNote', {
    window,
    multiplier: props.group.peak_rate_multiplier
  })
})
</script>
