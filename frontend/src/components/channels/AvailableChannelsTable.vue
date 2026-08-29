<template>
  <!-- .table-wrapper 是 TablePageLayout 滚动链的挂载点：外层 .table-scroll-container
       负责卡片外观并 overflow-hidden，本层接收 overflow-y-auto 才能在内容超高时滚动。 -->
  <div class="table-wrapper">
    <table
      data-testid="desktop-channels"
      class="components-channels-available-channels-table__table"
    >
      <thead>
        <tr class="components-channels-available-channels-table__row">
          <th class="components-channels-available-channels-table__heading">{{ columns.name }}</th>
          <th class="components-channels-available-channels-table__heading-2">{{ columns.description }}</th>
          <th class="components-channels-available-channels-table__heading-3">{{ columns.platform }}</th>
          <th class="components-channels-available-channels-table__heading-4">{{ columns.groups }}</th>
          <th class="components-channels-available-channels-table__heading-4">{{ columns.supportedModels }}</th>
        </tr>
      </thead>
      <tbody v-if="loading">
        <tr>
          <td colspan="5" class="components-channels-available-channels-table__cell">
            <Icon name="refresh" size="lg" class="components-channels-available-channels-table__icon" />
          </td>
        </tr>
      </tbody>
      <tbody v-else-if="rows.length === 0">
        <tr>
          <td colspan="5" class="components-channels-available-channels-table__cell-2">
            <Icon name="inbox" size="xl" class="components-channels-available-channels-table__icon-2" />
            <p class="components-channels-available-channels-table__description">{{ emptyLabel }}</p>
          </td>
        </tr>
      </tbody>
      <!-- 每个渠道一个 tbody：首行 td rowspan 渠道名，后续行只渲染其余三列。
           tbody 之间强分隔线表达"渠道边界"，tbody 内部用淡分隔线区分平台。 -->
      <tbody
        v-else
        v-for="(channel, chIdx) in rows"
        :key="`${channel.name}-${chIdx}`"
        class="components-channels-available-channels-table__body"
      >
        <tr
          v-for="(section, secIdx) in channel.platforms"
          :key="`${channel.name}-${section.platform}`"
          class="components-channels-available-channels-table__row-2"
          :class="{ 'components-channels-available-channels-table__row-3': secIdx > 0 }"
        >
          <!-- 渠道名：只在第一行渲染并用 rowspan 纵向合并 -->
          <td
            v-if="secIdx === 0"
            :rowspan="channel.platforms.length"
            class="components-channels-available-channels-table__cell-3"
          >
            {{ channel.name }}
          </td>

          <!-- 描述：独立一列，同样用 rowspan 纵向合并 -->
          <td
            v-if="secIdx === 0"
            :rowspan="channel.platforms.length"
            class="components-channels-available-channels-table__cell-4"
          >
            <template v-if="channel.description">{{ channel.description }}</template>
            <span v-else class="components-channels-available-channels-table__text">-</span>
          </td>

          <!-- 平台徽章 -->
          <td class="components-channels-available-channels-table__cell-5">
            <span
              :class="[
                'components-channels-available-channels-table__text-6',
                platformBadgeClass(section.platform),
              ]"
            >
              <PlatformIcon :platform="section.platform as GroupPlatform" size="xs" />
              {{ section.platform }}
            </span>
          </td>

          <!-- 分组：专属分组在前（紫色 shield 行），公开分组在后（灰色 globe 行）。 -->
          <td class="components-channels-available-channels-table__cell-5">
            <div class="components-channels-available-channels-table__panel">
              <div
                v-if="exclusiveGroups(section).length > 0"
                class="components-channels-available-channels-table__panel-2"
              >
                <span
                  class="components-channels-available-channels-table__text-2"
                  :title="t('availableChannels.exclusiveTooltip')"
                >
                  <Icon name="shield" size="xs" class="components-channels-available-channels-table__icon-3" />
                  {{ t('availableChannels.exclusive') }}
                </span>
                <div
                  v-for="g in exclusiveGroups(section)"
                  :key="`ex-${g.id}`"
                  class="components-channels-available-channels-table__panel-3"
                >
                  <GroupBadge
                    :name="g.name"
                    :platform="g.platform as GroupPlatform"
                    :subscription-type="(g.subscription_type || 'standard') as SubscriptionType"
                    :rate-multiplier="g.rate_multiplier"
                    :user-rate-multiplier="userGroupRates[g.id] ?? null"
                    always-show-rate
                  />
                  <span
                    v-if="hasPeakRate(g)"
                    class="components-channels-available-channels-table__text-3"
                    :title="peakRateTitle(g)"
                  >
                    <Icon name="clock" size="xs" class="components-channels-available-channels-table__icon-3" />
                    {{ peakRateLabel(g) }}
                  </span>
                </div>
              </div>
              <div
                v-if="publicGroups(section).length > 0"
                class="components-channels-available-channels-table__panel-2"
              >
                <span
                  class="components-channels-available-channels-table__text-4"
                  :title="t('availableChannels.publicTooltip')"
                >
                  <Icon name="globe" size="xs" class="components-channels-available-channels-table__icon-3" />
                  {{ t('availableChannels.public') }}
                </span>
                <div
                  v-for="g in publicGroups(section)"
                  :key="`pub-${g.id}`"
                  class="components-channels-available-channels-table__panel-3"
                >
                  <GroupBadge
                    :name="g.name"
                    :platform="g.platform as GroupPlatform"
                    :subscription-type="(g.subscription_type || 'standard') as SubscriptionType"
                    :rate-multiplier="g.rate_multiplier"
                    :user-rate-multiplier="userGroupRates[g.id] ?? null"
                    always-show-rate
                  />
                  <span
                    v-if="hasPeakRate(g)"
                    class="components-channels-available-channels-table__text-3"
                    :title="peakRateTitle(g)"
                  >
                    <Icon name="clock" size="xs" class="components-channels-available-channels-table__icon-3" />
                    {{ peakRateLabel(g) }}
                  </span>
                </div>
              </div>
              <span v-if="section.groups.length === 0" class="components-channels-available-channels-table__text-5">-</span>
            </div>
          </td>

          <!-- 支持模型 -->
          <td class="components-channels-available-channels-table__cell-5">
            <div class="components-channels-available-channels-table__panel-4">
              <SupportedModelChip
                v-for="m in section.supported_models"
                :key="`${section.platform}-${m.name}`"
                :model="m"
                :pricing-key-prefix="pricingKeyPrefix"
                :no-pricing-label="noPricingLabel"
                :show-platform="false"
                :platform-hint="section.platform"
              />
              <span v-if="section.supported_models.length === 0" class="components-channels-available-channels-table__text-5">
                {{ noModelsLabel }}
              </span>
            </div>
          </td>
        </tr>
      </tbody>
    </table>

    <div data-testid="mobile-channels" class="components-channels-available-channels-table__panel-5">
      <div v-if="loading" data-testid="mobile-loading" class="components-channels-available-channels-table__cell">
        <Icon name="refresh" size="lg" class="components-channels-available-channels-table__icon" />
      </div>
      <div v-else-if="rows.length === 0" data-testid="mobile-empty" class="components-channels-available-channels-table__cell-2">
        <Icon name="inbox" size="xl" class="components-channels-available-channels-table__icon-2" />
        <p class="components-channels-available-channels-table__description">{{ emptyLabel }}</p>
      </div>
      <section
        v-else
        v-for="(channel, chIdx) in rows"
        :key="`mobile-${channel.name}-${chIdx}`"
        class="components-channels-available-channels-table__section"
      >
        <header class="components-channels-available-channels-table__header">
          <h3 class="components-channels-available-channels-table__heading-5">
            {{ channel.name }}
          </h3>
          <p class="components-channels-available-channels-table__description-2">
            {{ channel.description || '-' }}
          </p>
        </header>

        <div class="components-channels-available-channels-table__panel-6">
          <div
            v-for="section in channel.platforms"
            :key="`mobile-${channel.name}-${section.platform}`"
            class="components-channels-available-channels-table__panel-7"
          >
            <span
              :class="[
                'components-channels-available-channels-table__text-6',
                platformBadgeClass(section.platform),
              ]"
            >
              <PlatformIcon :platform="section.platform as GroupPlatform" size="xs" />
              {{ section.platform }}
            </span>

            <dl class="components-channels-available-channels-table__dl">
              <div class="components-channels-available-channels-table__panel-8">
                <dt class="components-channels-available-channels-table__dt">
                  {{ columns.groups }}
                </dt>
                <dd class="components-channels-available-channels-table__dd">
                  <div
                    v-if="exclusiveGroups(section).length > 0"
                    class="components-channels-available-channels-table__panel-9"
                  >
                    <span
                      class="components-channels-available-channels-table__text-2"
                      :title="t('availableChannels.exclusiveTooltip')"
                    >
                      <Icon name="shield" size="xs" class="components-channels-available-channels-table__icon-3" />
                      {{ t('availableChannels.exclusive') }}
                    </span>
                    <div
                      v-for="g in exclusiveGroups(section)"
                      :key="`mobile-ex-${g.id}`"
                      class="components-channels-available-channels-table__panel-10"
                    >
                      <GroupBadge
                        class="components-channels-available-channels-table__group-badge"
                        :name="g.name"
                        :platform="g.platform as GroupPlatform"
                        :subscription-type="(g.subscription_type || 'standard') as SubscriptionType"
                        :rate-multiplier="g.rate_multiplier"
                        :user-rate-multiplier="userGroupRates[g.id] ?? null"
                        always-show-rate
                      />
                      <span
                        v-if="hasPeakRate(g)"
                        class="components-channels-available-channels-table__text-3"
                        :title="peakRateTitle(g)"
                      >
                        <Icon name="clock" size="xs" class="components-channels-available-channels-table__icon-3" />
                        {{ peakRateLabel(g) }}
                      </span>
                    </div>
                  </div>
                  <div
                    v-if="publicGroups(section).length > 0"
                    class="components-channels-available-channels-table__panel-9"
                  >
                    <span
                      class="components-channels-available-channels-table__text-4"
                      :title="t('availableChannels.publicTooltip')"
                    >
                      <Icon name="globe" size="xs" class="components-channels-available-channels-table__icon-3" />
                      {{ t('availableChannels.public') }}
                    </span>
                    <div
                      v-for="g in publicGroups(section)"
                      :key="`mobile-pub-${g.id}`"
                      class="components-channels-available-channels-table__panel-10"
                    >
                      <GroupBadge
                        class="components-channels-available-channels-table__group-badge"
                        :name="g.name"
                        :platform="g.platform as GroupPlatform"
                        :subscription-type="(g.subscription_type || 'standard') as SubscriptionType"
                        :rate-multiplier="g.rate_multiplier"
                        :user-rate-multiplier="userGroupRates[g.id] ?? null"
                        always-show-rate
                      />
                      <span
                        v-if="hasPeakRate(g)"
                        class="components-channels-available-channels-table__text-3"
                        :title="peakRateTitle(g)"
                      >
                        <Icon name="clock" size="xs" class="components-channels-available-channels-table__icon-3" />
                        {{ peakRateLabel(g) }}
                      </span>
                    </div>
                  </div>
                  <span v-if="section.groups.length === 0" class="components-channels-available-channels-table__text-5">-</span>
                </dd>
              </div>

              <div class="components-channels-available-channels-table__panel-8">
                <dt class="components-channels-available-channels-table__dt">
                  {{ columns.supportedModels }}
                </dt>
                <dd class="components-channels-available-channels-table__dd-2">
                  <SupportedModelChip
                    v-for="m in section.supported_models"
                    :key="`mobile-${section.platform}-${m.name}`"
                    class="components-channels-available-channels-table__supported-model-chip"
                    :model="m"
                    :pricing-key-prefix="pricingKeyPrefix"
                    :no-pricing-label="noPricingLabel"
                    :show-platform="false"
                    :platform-hint="section.platform"
                  />
                  <span v-if="section.supported_models.length === 0" class="components-channels-available-channels-table__text-5">
                    {{ noModelsLabel }}
                  </span>
                </dd>
              </div>
            </dl>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import SupportedModelChip from './SupportedModelChip.vue'
import type { UserAvailableChannel, UserAvailableGroup, UserChannelPlatformSection } from '@/api/channels'
import type { GroupPlatform, SubscriptionType } from '@/types'
import { platformBadgeClass } from '@/utils/platformColors'
import { useAppStore } from '@/stores/app'
import { hasPeakRate as groupHasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'

const props = defineProps<{
  columns: {
    name: string
    description: string
    platform: string
    groups: string
    supportedModels: string
  }
  rows: UserAvailableChannel[]
  loading: boolean
  pricingKeyPrefix: string
  noPricingLabel: string
  noModelsLabel: string
  emptyLabel: string
  /** 用户专属倍率（group_id → multiplier）；无专属时由 GroupBadge 仅显示默认倍率。 */
  userGroupRates: Record<string, number>
}>()

// Suppress unused warning — props is accessed via template automatically but
// the explicit reference here keeps the linter from flagging userGroupRates.
void props.userGroupRates

const { t } = useI18n()

function exclusiveGroups(section: UserChannelPlatformSection): UserAvailableGroup[] {
  return section.groups.filter((g) => g.is_exclusive)
}

function publicGroups(section: UserChannelPlatformSection): UserAvailableGroup[] {
  return section.groups.filter((g) => !g.is_exclusive)
}

const appStore = useAppStore()

function hasPeakRate(group: UserAvailableGroup): boolean {
  return groupHasPeakRate(group)
}

function peakRateLabel(group: UserAvailableGroup): string {
  return formatPeakRateWindow(group, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}

function peakRateTitle(group: UserAvailableGroup): string {
  return t('common.peakRateTooltip', { window: peakRateLabel(group) }) + t('common.peakRateImageNote')
}
</script>
