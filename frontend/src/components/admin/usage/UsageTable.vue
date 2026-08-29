<template>
  <div :class="flat ? '' : 'card overflow-hidden'">
    <div
      v-if="showIpGeoToolbar"
      class="components-admin-usage-usage-table__panel"
    >
      <span v-if="pendingIpCount > 0" class="components-admin-usage-usage-table__text">
        {{ t('usage.ipGeo.pending', { count: pendingIpCount }) }}
      </span>
      <button
        type="button"
        class="components-admin-usage-usage-table__action"
        :disabled="ipGeoBatchLoading || pendingIpCount === 0"
        @click="handleBatchFetchIpGeo"
      >
        {{ ipGeoBatchLoading ? t('usage.ipGeo.batchFetching') : t('usage.ipGeo.batchFetch') }}
      </button>
    </div>
    <div class="components-admin-usage-usage-table__panel-2">
      <DataTable
        :columns="columns"
        :data="data"
        :loading="loading"
        :server-side-sort="serverSideSort"
        :default-sort-key="defaultSortKey"
        :default-sort-order="defaultSortOrder"
        @sort="(key, order) => $emit('sort', key, order)"
      >
        <template #cell-user="{ row }">
          <div class="components-admin-usage-usage-table__panel-3">
            <button
              v-if="row.user?.email"
              class="components-admin-usage-usage-table__action-2"
              @click="$emit('userClick', row.user_id, row.user?.email)"
              :title="t('admin.usage.clickToViewBalance')"
            >
              {{ row.user.email }}
            </button>
            <span v-else class="components-admin-usage-usage-table__text-2">-</span>
            <span v-if="row.user?.deleted_at" class="components-admin-usage-usage-table__text-3">
              {{ t('admin.usage.userDeletedBadge') }}
            </span>
            <span class="components-admin-usage-usage-table__text-4">#{{ row.user_id }}</span>
          </div>
        </template>

        <template #cell-api_key="{ row }">
          <span class="components-admin-usage-usage-table__text-5">{{ row.api_key?.name || '-' }}</span>
        </template>

        <template #cell-account="{ row }">
          <span class="components-admin-usage-usage-table__text-5">{{ row.account?.name || '-' }}</span>
        </template>

        <template #cell-model="{ row }">
          <div class="components-admin-usage-usage-table__panel-4">
            <div v-if="row.model_mapping_chain && row.model_mapping_chain.includes('→')" class="components-admin-usage-usage-table__panel-5">
              <div v-for="(step, i) in row.model_mapping_chain.split('→')" :key="i"
                   class="components-admin-usage-usage-table__panel-6"
                   :class="i === 0 ? 'components-admin-usage-usage-table__text-2' : 'components-admin-usage-usage-table__panel-31'"
                   :style="i > 0 ? `padding-left: ${i * 0.75}rem` : ''">
                <span v-if="i > 0" class="components-admin-usage-usage-table__text-6">↳</span>{{ step }}
              </div>
            </div>
            <div v-else-if="row.upstream_model && row.upstream_model !== row.model" class="components-admin-usage-usage-table__panel-5">
              <div class="components-admin-usage-usage-table__panel-7">
                {{ row.model }}
              </div>
              <div class="components-admin-usage-usage-table__panel-8">
                <span class="components-admin-usage-usage-table__text-6">↳</span>{{ row.upstream_model }}
              </div>
            </div>
            <span v-else class="components-admin-usage-usage-table__text-2">{{ row.model }}</span>
            <div
              v-if="row.upstream_model_mismatch === true && row.upstream_response_model"
              class="components-admin-usage-usage-table__panel-9"
              :class="isLikelyModelVariant(row) ? 'components-admin-usage-usage-table__panel-32' : 'components-admin-usage-usage-table__panel-33'"
              :title="modelAuditTitle(row)"
            >
              <span class="components-admin-usage-usage-table__text-7">↳ {{ t('usage.upstreamResponseModel') }}:</span>{{ row.upstream_response_model }}
              <span
                class="components-admin-usage-usage-table__text-8"
                :class="isLikelyModelVariant(row)
                  ? 'components-admin-usage-usage-table__text-43'
                  : 'components-admin-usage-usage-table__text-44'"
              >
                {{ isLikelyModelVariant(row) ? t('usage.modelVariant') : t('usage.modelMismatch') }}
              </span>
            </div>
          </div>
        </template>

        <template #cell-reasoning_effort="{ row }">
          <span class="components-admin-usage-usage-table__text-5">
            {{ formatReasoningEffort(row.reasoning_effort) }}
          </span>
        </template>

        <template #cell-endpoint="{ row }">
          <div class="components-admin-usage-usage-table__panel-10">
            <div class="components-admin-usage-usage-table__panel-11">
              <span class="components-admin-usage-usage-table__text-9">{{ t('usage.inbound') }}:</span>
              <span class="components-admin-usage-usage-table__text-10">{{ row.inbound_endpoint?.trim() || '-' }}</span>
            </div>
            <div v-if="showUpstreamEndpoint" class="components-admin-usage-usage-table__panel-11">
              <span class="components-admin-usage-usage-table__text-9">{{ t('usage.upstream') }}:</span>
              <span class="components-admin-usage-usage-table__text-10">{{ row.upstream_endpoint?.trim() || '-' }}</span>
            </div>
          </div>
        </template>

        <template #cell-group="{ row }">
          <span v-if="row.group" class="components-admin-usage-usage-table__text-11">
            {{ row.group.name }}
          </span>
          <span v-else class="components-admin-usage-usage-table__text-12">-</span>
        </template>

        <template #cell-stream="{ row }">
          <span class="components-admin-usage-usage-table__text-13" :class="getRequestTypeBadgeClass(row)">
            {{ getRequestTypeLabel(row) }}
          </span>
        </template>

        <template #cell-billing_mode="{ row }">
          <span class="components-admin-usage-usage-table__text-13" :class="getBillingModeBadgeClass(getDisplayBillingMode(row))">
            {{ getBillingModeLabel(getDisplayBillingMode(row), t) }}
          </span>
        </template>

        <template #cell-tokens="{ row }">
          <!-- 图片生成请求（仅按次计费时显示图片格式） -->
          <div v-if="isImageUsage(row)" class="components-admin-usage-usage-table__panel-12">
            <svg class="components-admin-usage-usage-table__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <span class="components-admin-usage-usage-table__text-2">{{ row.image_count }}{{ t('usage.imageUnit') }}</span>
            <span class="components-admin-usage-usage-table__text-14">({{ formatImageBillingSize(row, t) }})</span>
          </div>
          <!-- Token 请求 -->
          <div v-else class="components-admin-usage-usage-table__panel-12">
            <div class="components-admin-usage-usage-table__panel-13">
              <div class="components-admin-usage-usage-table__panel-14">
                <div class="components-admin-usage-usage-table__panel-15">
                  <Icon name="arrowDown" size="sm" class="components-admin-usage-usage-table__icon-2" />
                  <span class="components-admin-usage-usage-table__text-2">{{ row.input_tokens?.toLocaleString() || 0 }}</span>
                </div>
                <div class="components-admin-usage-usage-table__panel-15">
                  <Icon name="arrowUp" size="sm" class="components-admin-usage-usage-table__icon-3" />
                  <span class="components-admin-usage-usage-table__text-2">{{ row.output_tokens?.toLocaleString() || 0 }}</span>
                </div>
              </div>
              <div v-if="row.cache_read_tokens > 0 || row.cache_creation_tokens > 0" class="components-admin-usage-usage-table__panel-14">
                <div v-if="row.cache_read_tokens > 0" class="components-admin-usage-usage-table__panel-15">
                  <svg class="components-admin-usage-usage-table__icon-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" /></svg>
                  <span class="components-admin-usage-usage-table__text-15">{{ formatCacheTokens(row.cache_read_tokens) }}</span>
                </div>
                <div v-if="row.cache_creation_tokens > 0" class="components-admin-usage-usage-table__panel-15">
                  <svg class="components-admin-usage-usage-table__icon-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                  <span class="components-admin-usage-usage-table__text-16">{{ formatCacheTokens(row.cache_creation_tokens) }}</span>
                  <span v-if="row.cache_creation_1h_tokens > 0" class="components-admin-usage-usage-table__text-17">1h</span>
                  <span v-if="row.cache_ttl_overridden" :title="t('usage.cacheTtlOverriddenHint')" class="components-admin-usage-usage-table__text-18">R</span>
                </div>
              </div>
              <div v-if="hasImageInputTokens(row)" class="components-admin-usage-usage-table__panel-14">
                <div class="components-admin-usage-usage-table__panel-15">
                  <svg class="components-admin-usage-usage-table__icon-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
                  <span class="components-admin-usage-usage-table__text-19">{{ row.image_input_tokens.toLocaleString() }}</span>
                </div>
              </div>
              <div v-if="hasImageOutputTokens(row)" class="components-admin-usage-usage-table__panel-14">
                <div class="components-admin-usage-usage-table__panel-15">
                  <svg class="components-admin-usage-usage-table__icon-7" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
                  <span class="components-admin-usage-usage-table__text-20">{{ row.image_output_tokens.toLocaleString() }}</span>
                </div>
              </div>
            </div>
            <!-- Token Detail Tooltip -->
            <div
              class="components-admin-usage-usage-table__panel-16"
              @mouseenter="showTokenTooltip($event, row)"
              @mouseleave="hideTokenTooltip"
            >
              <div class="components-admin-usage-usage-table__panel-17">
                <Icon name="infoCircle" size="xs" class="components-admin-usage-usage-table__icon-8" />
              </div>
            </div>
          </div>
        </template>

        <template #cell-cost="{ row }">
          <div class="components-admin-usage-usage-table__panel-3">
            <div class="components-admin-usage-usage-table__panel-12">
              <span class="components-admin-usage-usage-table__text-21">{{ formatPoints(row.actual_cost) }}</span>
              <span
                v-if="row.long_context_billing_applied"
                data-testid="long-context-billing-marker"
                class="components-admin-usage-usage-table__text-22"
              >x2</span>
              <!-- Cost Detail Tooltip -->
              <div
                class="components-admin-usage-usage-table__panel-16"
                @mouseenter="showTooltip($event, row)"
                @mouseleave="hideTooltip"
              >
                <div class="components-admin-usage-usage-table__panel-17">
                  <Icon name="infoCircle" size="xs" class="components-admin-usage-usage-table__icon-8" />
                </div>
              </div>
            </div>
            <div v-if="showAccountBilling && row.account_rate_multiplier != null" class="components-admin-usage-usage-table__panel-18">
              A ${{ accountBilled(row).toFixed(6) }}
            </div>
          </div>
        </template>

        <!-- 合并首字/总耗时的健康度列：左侧色条上端随首字档、下端随总耗时档，中段(40%-60%)短渐变过渡，便于纵向扫视整体健康状况 -->
        <template #cell-latency="{ row }">
          <div class="components-admin-usage-usage-table__panel-19">
            <span
              class="components-admin-usage-usage-table__text-23"
              :class="row.first_token_ms != null
                ? ['components-admin-usage-usage-table__text-45', LATENCY_BAR_FROM_CLASSES[firstTokenSeverity(row.first_token_ms)], LATENCY_BAR_TO_CLASSES[durationSeverity(row.duration_ms ?? 0)]]
                : LATENCY_BAR_CLASSES[durationSeverity(row.duration_ms ?? 0)]"
              aria-hidden="true"
            ></span>
            <div class="components-admin-usage-usage-table__panel-20">
              <span class="components-admin-usage-usage-table__text-24">{{ t('usage.latencyFirstToken') }}</span>
              <span v-if="row.first_token_ms != null" class="components-admin-usage-usage-table__text-25" :class="LATENCY_TEXT_CLASSES[firstTokenSeverity(row.first_token_ms)]">{{ formatDuration(row.first_token_ms) }}</span>
              <span v-else class="components-admin-usage-usage-table__text-24">-</span>
              <span class="components-admin-usage-usage-table__text-24">{{ t('usage.latencyDuration') }}</span>
              <span class="components-admin-usage-usage-table__text-25" :class="LATENCY_TEXT_CLASSES[durationSeverity(row.duration_ms ?? 0)]">{{ formatDuration(row.duration_ms) }}</span>
            </div>
          </div>
        </template>

        <template #cell-created_at="{ value }">
          <span class="components-admin-usage-usage-table__text-26">{{ formatDateTime(value) }}</span>
        </template>

        <template #cell-request_id="{ row }">
          <div v-if="row.request_id" class="components-admin-usage-usage-table__panel-21">
            <span class="components-admin-usage-usage-table__text-27" :title="row.request_id">
              {{ row.request_id }}
            </span>
            <button
              type="button"
              class="components-admin-usage-usage-table__action-3"
              :class="copiedRequestId === row.request_id ? 'components-admin-usage-usage-table__action-4' : ''"
              :title="copiedRequestId === row.request_id ? t('keys.copied') : t('keys.copyToClipboard')"
              @click="copyRequestId(row.request_id)"
            >
              <Icon :name="copiedRequestId === row.request_id ? 'check' : 'copy'" size="sm" class="components-admin-usage-usage-table__icon-9" />
            </button>
          </div>
          <span v-else class="components-admin-usage-usage-table__text-12">-</span>
        </template>

        <template #cell-user_agent="{ row }">
          <span v-if="row.user_agent" class="components-admin-usage-usage-table__text-28" :title="row.user_agent">{{ formatUserAgent(row.user_agent) }}</span>
          <span v-else class="components-admin-usage-usage-table__text-12">-</span>
        </template>

        <template #cell-ip_address="{ row }">
          <div v-if="row.ip_address">
            <span class="components-admin-usage-usage-table__text-29">{{ row.ip_address }}</span>
            <IpGeoCell :ip="row.ip_address" />
          </div>
          <span v-else class="components-admin-usage-usage-table__text-12">-</span>
        </template>

        <template #empty><EmptyState :message="t('usage.noRecords')" /></template>
      </DataTable>
    </div>
  </div>

  <!-- Token Tooltip Portal -->
  <Teleport to="body">
    <div
      v-if="tokenTooltipVisible"
      class="components-admin-usage-usage-table__panel-22"
      :style="{
        left: tokenTooltipPosition.x + 'px',
        top: tokenTooltipPosition.y + 'px'
      }"
    >
      <div class="components-admin-usage-usage-table__panel-23">
        <div class="components-admin-usage-usage-table__panel-24">
          <div>
            <div class="components-admin-usage-usage-table__panel-25">{{ t('usage.tokenDetails') }}</div>
            <div v-if="tokenTooltipData && tokenTooltipData.input_tokens > 0 && !hasImageInputTokens(tokenTooltipData)" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('admin.usage.inputTokens') }}</span>
              <span class="components-admin-usage-usage-table__text-30">{{ tokenTooltipData.input_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageInputTokens(tokenTooltipData) && textInputTokens(tokenTooltipData) > 0" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('admin.usage.inputTokens') }}</span>
              <span class="components-admin-usage-usage-table__text-30">{{ textInputTokens(tokenTooltipData).toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageInputTokens(tokenTooltipData)" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageInputTokens') }}</span>
              <span class="components-admin-usage-usage-table__text-31">{{ tokenTooltipData.image_input_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.output_tokens > 0 && !hasImageOutputTokens(tokenTooltipData)" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('admin.usage.outputTokens') }}</span>
              <span class="components-admin-usage-usage-table__text-30">{{ tokenTooltipData.output_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageOutputTokens(tokenTooltipData) && textOutputTokens(tokenTooltipData) > 0" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('admin.usage.outputTokens') }}</span>
              <span class="components-admin-usage-usage-table__text-30">{{ textOutputTokens(tokenTooltipData).toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageOutputTokens(tokenTooltipData)" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageOutputTokens') }}</span>
              <span class="components-admin-usage-usage-table__text-32">{{ tokenTooltipData.image_output_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.cache_creation_tokens > 0">
              <!-- 有 5m/1h 明细时，展开显示 -->
              <template v-if="tokenTooltipData.cache_creation_5m_tokens > 0 || tokenTooltipData.cache_creation_1h_tokens > 0">
                <div v-if="tokenTooltipData.cache_creation_5m_tokens > 0" class="components-admin-usage-usage-table__panel-26">
                  <span class="components-admin-usage-usage-table__text-33">
                    {{ t('admin.usage.cacheCreation5mTokens') }}
                    <span class="components-admin-usage-usage-table__text-34">5m</span>
                  </span>
                  <span class="components-admin-usage-usage-table__text-30">{{ tokenTooltipData.cache_creation_5m_tokens.toLocaleString() }}</span>
                </div>
                <div v-if="tokenTooltipData.cache_creation_1h_tokens > 0" class="components-admin-usage-usage-table__panel-26">
                  <span class="components-admin-usage-usage-table__text-33">
                    {{ t('admin.usage.cacheCreation1hTokens') }}
                    <span class="components-admin-usage-usage-table__text-35">1h</span>
                  </span>
                  <span class="components-admin-usage-usage-table__text-30">{{ tokenTooltipData.cache_creation_1h_tokens.toLocaleString() }}</span>
                </div>
              </template>
              <!-- 无明细时，只显示聚合值 -->
              <div v-else class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('admin.usage.cacheCreationTokens') }}</span>
                <span class="components-admin-usage-usage-table__text-30">{{ tokenTooltipData.cache_creation_tokens.toLocaleString() }}</span>
              </div>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.cache_ttl_overridden" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-33">
                {{ t('usage.cacheTtlOverriddenLabel') }}
                <span class="components-admin-usage-usage-table__text-36">R-{{ tokenTooltipData.cache_creation_1h_tokens > 0 ? '5m' : '1H' }}</span>
              </span>
              <span class="components-admin-usage-usage-table__text-37">{{ tokenTooltipData.cache_creation_1h_tokens > 0 ? t('usage.cacheTtlOverridden1h') : t('usage.cacheTtlOverridden5m') }}</span>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.cache_read_tokens > 0" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('admin.usage.cacheReadTokens') }}</span>
              <span class="components-admin-usage-usage-table__text-30">{{ tokenTooltipData.cache_read_tokens.toLocaleString() }}</span>
            </div>
          </div>
          <div class="components-admin-usage-usage-table__panel-27">
            <span class="components-admin-usage-usage-table__text-14">{{ t('usage.totalTokens') }}</span>
            <span class="components-admin-usage-usage-table__text-38">{{ ((tokenTooltipData?.input_tokens || 0) + (tokenTooltipData?.output_tokens || 0) + (tokenTooltipData?.cache_creation_tokens || 0) + (tokenTooltipData?.cache_read_tokens || 0)).toLocaleString() }}</span>
          </div>
        </div>
        <div class="components-admin-usage-usage-table__panel-28"></div>
      </div>
    </div>
  </Teleport>

  <!-- Cost Tooltip Portal -->
  <Teleport to="body">
    <div
      v-if="tooltipVisible"
      class="components-admin-usage-usage-table__panel-22"
      :style="{
        left: tooltipPosition.x + 'px',
        top: tooltipPosition.y + 'px'
      }"
    >
      <div class="components-admin-usage-usage-table__panel-23">
        <div class="components-admin-usage-usage-table__panel-24">
          <!-- Cost Breakdown -->
          <div class="components-admin-usage-usage-table__panel-29">
            <div class="components-admin-usage-usage-table__panel-25">{{ t('usage.costDetails') }}</div>
            <div v-if="tooltipData && tooltipData.input_cost > 0" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('admin.usage.inputCost') }}</span>
              <span class="components-admin-usage-usage-table__text-30">${{ tooltipData.input_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && hasImageInputCost(tooltipData)" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageInputCost') }}</span>
              <span class="components-admin-usage-usage-table__text-31">${{ tooltipData.image_input_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.output_cost > 0" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('admin.usage.outputCost') }}</span>
              <span class="components-admin-usage-usage-table__text-30">${{ tooltipData.output_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && hasImageOutputCost(tooltipData)" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageOutputCost') }}</span>
              <span class="components-admin-usage-usage-table__text-32">${{ tooltipData.image_output_cost.toFixed(6) }}</span>
            </div>
            <!-- Token billing: show unit prices per 1M tokens -->
            <template v-if="tooltipData && !isImageUsage(tooltipData) && (!tooltipData.billing_mode || tooltipData.billing_mode === BILLING_MODE_TOKEN)">
              <div v-if="tooltipData && textInputTokens(tooltipData) > 0" class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.inputTokenPrice') }}</span>
                <span class="components-admin-usage-usage-table__text-39">{{ formatTokenPricePerMillion(tooltipData.input_cost, textInputTokens(tooltipData)) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && hasImageInputTokens(tooltipData)" class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageInputTokenPrice') }}</span>
                <span class="components-admin-usage-usage-table__text-31">{{ formatTokenPricePerMillion(tooltipData.image_input_cost ?? 0, tooltipData.image_input_tokens) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && tooltipData.output_cost > 0 && textOutputTokens(tooltipData) > 0" class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.outputTokenPrice') }}</span>
                <span class="components-admin-usage-usage-table__text-40">{{ formatTokenPricePerMillion(tooltipData.output_cost, textOutputTokens(tooltipData)) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && hasImageOutputTokens(tooltipData)" class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageOutputTokenPrice') }}</span>
                <span class="components-admin-usage-usage-table__text-32">{{ formatTokenPricePerMillion(tooltipData.image_output_cost ?? 0, tooltipData.image_output_tokens) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
            </template>
            <template v-else-if="tooltipData && isImageUsage(tooltipData)">
              <div class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageCount') }}</span>
                <span class="components-admin-usage-usage-table__text-30">{{ tooltipData.image_count }}{{ t('usage.imageUnit') }}</span>
              </div>
              <div class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageBillingSize') }}</span>
                <span class="components-admin-usage-usage-table__text-30">{{ formatImageBillingSize(tooltipData, t) }}</span>
              </div>
              <div class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageSizeSource') }}</span>
                <span class="components-admin-usage-usage-table__text-30">{{ formatImageSizeSource(tooltipData, t) }}</span>
              </div>
              <div class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageInputSize') }}</span>
                <span class="components-admin-usage-usage-table__text-30">{{ formatImageInputSize(tooltipData, t) }}</span>
              </div>
              <div class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageOutputSize') }}</span>
                <span class="components-admin-usage-usage-table__text-30">{{ formatImageOutputSize(tooltipData, t) }}</span>
              </div>
              <div v-if="formatImageSizeBreakdown(tooltipData)" class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageSizeBreakdown') }}</span>
                <span class="components-admin-usage-usage-table__text-30">{{ formatImageSizeBreakdown(tooltipData) }}</span>
              </div>
              <div class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageUnitPrice') }}</span>
                <span class="components-admin-usage-usage-table__text-39">${{ imageUnitPrice(tooltipData).toFixed(6) }}</span>
              </div>
              <div class="components-admin-usage-usage-table__panel-26">
                <span class="components-admin-usage-usage-table__text-14">{{ t('usage.imageTotalPrice') }}</span>
                <span class="components-admin-usage-usage-table__text-30">${{ tooltipData.total_cost?.toFixed(6) || '0.000000' }}</span>
              </div>
            </template>
            <div v-else class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('usage.unitPrice') }}</span>
              <span class="components-admin-usage-usage-table__text-39">${{ tooltipData?.total_cost?.toFixed(6) || '0.000000' }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.cache_creation_cost > 0" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('admin.usage.cacheCreationCost') }}</span>
              <span class="components-admin-usage-usage-table__text-30">${{ tooltipData.cache_creation_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.cache_read_cost > 0" class="components-admin-usage-usage-table__panel-26">
              <span class="components-admin-usage-usage-table__text-14">{{ t('admin.usage.cacheReadCost') }}</span>
              <span class="components-admin-usage-usage-table__text-30">${{ tooltipData.cache_read_cost.toFixed(6) }}</span>
            </div>
          </div>
          <!-- Rate and Summary -->
          <div class="components-admin-usage-usage-table__panel-30">
            <span class="components-admin-usage-usage-table__text-14">{{ t('usage.serviceTier') }}</span>
            <span class="components-admin-usage-usage-table__text-41">{{ getUsageServiceTierLabel(tooltipData?.service_tier, t) }}</span>
          </div>
          <div class="components-admin-usage-usage-table__panel-30">
            <span class="components-admin-usage-usage-table__text-14">{{ t('usage.rate') }}</span>
            <span class="components-admin-usage-usage-table__text-38">{{ formatMultiplier(tooltipData?.rate_multiplier || 1) }}x</span>
          </div>
          <div class="components-admin-usage-usage-table__panel-30">
            <span class="components-admin-usage-usage-table__text-14">{{ t('usage.original') }}</span>
            <span class="components-admin-usage-usage-table__text-30">${{ tooltipData?.total_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <div class="components-admin-usage-usage-table__panel-30">
            <span class="components-admin-usage-usage-table__text-14">{{ t('usage.userBilled') }}</span>
            <span class="components-admin-usage-usage-table__text-42">{{ formatPoints(tooltipData?.actual_cost) }}</span>
          </div>
          <!-- Account billing (separated from user billing) -->
          <template v-if="showAccountBilling">
            <div class="components-admin-usage-usage-table__panel-27">
              <span class="components-admin-usage-usage-table__text-14">{{ t('usage.accountMultiplier') }}</span>
              <span class="components-admin-usage-usage-table__text-38">{{ formatMultiplier(tooltipData?.account_rate_multiplier ?? 1) }}x</span>
            </div>
            <div class="components-admin-usage-usage-table__panel-30">
              <span class="components-admin-usage-usage-table__text-14">{{ t('usage.accountBilled') }}</span>
              <span class="components-admin-usage-usage-table__text-42">
                ${{ accountBilled({
                  total_cost: tooltipData?.total_cost,
                  account_stats_cost: tooltipData?.account_stats_cost,
                  account_rate_multiplier: tooltipData?.account_rate_multiplier,
                }).toFixed(6) }}
              </span>
            </div>
          </template>
        </div>
        <div class="components-admin-usage-usage-table__panel-28"></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { formatDateTime, formatPoints, formatReasoningEffort } from '@/utils/format'
import { formatCacheTokens, formatMultiplier } from '@/utils/formatters'
import { formatTokenPricePerMillion } from '@/utils/usagePricing'
import { getUsageServiceTierLabel } from '@/utils/usageServiceTier'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import {
  LATENCY_BAR_CLASSES,
  LATENCY_BAR_FROM_CLASSES,
  LATENCY_BAR_TO_CLASSES,
  LATENCY_TEXT_CLASSES,
  durationSeverity,
  firstTokenSeverity,
} from '@/utils/latencyHealth'
import {
  BILLING_MODE_TOKEN,
  getBillingModeLabel,
  getBillingModeBadgeClass,
  isImageUsage,
  getDisplayBillingMode,
  imageUnitPrice,
} from '@/utils/billingMode'
import {
  formatImageBillingSize,
  formatImageInputSize,
  formatImageOutputSize,
  formatImageSizeBreakdown,
  formatImageSizeSource,
  hasImageOutputTokens,
  textOutputTokens,
  hasImageOutputCost,
  hasImageInputTokens,
  textInputTokens,
  hasImageInputCost,
} from '@/utils/imageUsage'

/** Compute the account-billed cost for display: (account_stats_cost ?? total_cost) * rate_multiplier */
function accountBilled(row: { total_cost?: number | null; account_stats_cost?: number | null; account_rate_multiplier?: number | null }): number {
  const base = row.account_stats_cost != null ? row.account_stats_cost : (row.total_cost ?? 0)
  const result = base * (row.account_rate_multiplier ?? 1)
  return Number.isNaN(result) ? 0 : result
}


import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import IpGeoCell from '@/components/common/IpGeoCell.vue'
import Icon from '@/components/icons/Icon.vue'
import { fetchBatch, getEntry } from '@/utils/ipGeoLookup'
import type { AdminUsageLog } from '@/types'
import type { Column } from '@/components/common/types'

interface Props {
  data: AdminUsageLog[]
  loading?: boolean
  columns: Column[]
  serverSideSort?: boolean
  defaultSortKey?: string
  defaultSortOrder?: 'asc' | 'desc'
  showAccountBilling?: boolean
  showUpstreamEndpoint?: boolean
  /** 嵌入统一卡片内使用：去掉自身卡片外观 */
  flat?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  serverSideSort: false,
  defaultSortKey: '',
  defaultSortOrder: 'asc',
  showAccountBilling: true,
  showUpstreamEndpoint: true,
  flat: false
})
const emit = defineEmits<{
  userClick: [userID: string, email?: string]
  sort: [key: string, order: 'asc' | 'desc']
  ipGeoBatchFailed: []
}>()
const { t } = useI18n()
const appStore = useAppStore()
const copiedRequestId = ref<string | null>(null)
const showAccountBilling = props.showAccountBilling
const showUpstreamEndpoint = props.showUpstreamEndpoint
const ipGeoBatchLoading = ref(false)

const showIpGeoToolbar = computed(() => props.columns.some((col) => col.key === 'ip_address'))

const sentUpstreamModel = (row: AdminUsageLog): string => row.upstream_model?.trim() || row.model?.trim() || ''

const normalizeModelVariant = (model: string): string => model
  .trim()
  .toLowerCase()
  .replace(/-latest$/, '')
  .replace(/-\d{4}-\d{2}-\d{2}$/, '')
  .replace(/-\d{8}$/, '')

const isLikelyModelVariant = (row: AdminUsageLog): boolean => {
  const sent = sentUpstreamModel(row)
  const response = row.upstream_response_model?.trim() || ''
  return sent !== '' && response !== '' && normalizeModelVariant(sent) === normalizeModelVariant(response)
}

const modelAuditTitle = (row: AdminUsageLog): string => [
  `${t('usage.requestedModel')}: ${row.model || '-'}`,
  `${t('usage.sentUpstreamModel')}: ${sentUpstreamModel(row) || '-'}`,
  `${t('usage.upstreamResponseModel')}: ${row.upstream_response_model || '-'}`,
].join('\n')

const currentPageIps = computed(() =>
  Array.from(new Set(props.data.map((row) => row.ip_address).filter((ip): ip is string => Boolean(ip))))
)

const pendingIpCount = computed(() => {
  if (!showIpGeoToolbar.value) return 0
  return currentPageIps.value.filter((ip) => {
    const status = getEntry(ip).status
    return status === 'idle' || status === 'error'
  }).length
})

const handleBatchFetchIpGeo = async () => {
  ipGeoBatchLoading.value = true
  try {
    const ok = await fetchBatch(currentPageIps.value)
    if (!ok) emit('ipGeoBatchFailed')
  } finally {
    ipGeoBatchLoading.value = false
  }
}

const copyRequestId = async (requestId: string) => {
  try {
    await navigator.clipboard.writeText(requestId)
    copiedRequestId.value = requestId
    appStore.showSuccess(t('admin.usage.requestIdCopied'))
    window.setTimeout(() => {
      if (copiedRequestId.value === requestId) copiedRequestId.value = null
    }, 2000)
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

// Tooltip state - cost
const tooltipVisible = ref(false)
const tooltipPosition = ref({ x: 0, y: 0 })
const tooltipData = ref<AdminUsageLog | null>(null)

// Tooltip state - token
const tokenTooltipVisible = ref(false)
const tokenTooltipPosition = ref({ x: 0, y: 0 })
const tokenTooltipData = ref<AdminUsageLog | null>(null)

const getRequestTypeLabel = (row: AdminUsageLog): string => {
  const requestType = resolveUsageRequestType(row)
  if (requestType === 'cyber') return t('usage.cyber')
  if (requestType === 'live') return t('usage.live')
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
}

const getRequestTypeBadgeClass = (row: AdminUsageLog): string => {
  const requestType = resolveUsageRequestType(row)
  if (requestType === 'cyber') return 'components-admin-usage-usage-table__state'
  if (requestType === 'live') return 'components-admin-usage-usage-table__state-2'
  if (requestType === 'ws_v2') return 'components-admin-usage-usage-table__state-3'
  if (requestType === 'stream') return 'components-admin-usage-usage-table__state-4'
  if (requestType === 'sync') return 'components-admin-usage-usage-table__state-5'
  return 'components-admin-usage-usage-table__state-6'
}



const formatUserAgent = (ua: string): string => {
  return ua
}

// 超过 1 分钟简化为 "Xm Ys"，免去人工换算（超过 1 小时再进位为 "Xh Ym"）
const formatDuration = (ms: number | null | undefined): string => {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(2)}s`
  const totalSec = Math.round(ms / 1000)
  if (totalSec < 3600) return `${Math.floor(totalSec / 60)}m ${totalSec % 60}s`
  return `${Math.floor(totalSec / 3600)}h ${Math.floor((totalSec % 3600) / 60)}m`
}

// Cost tooltip functions
const showTooltip = (event: MouseEvent, row: AdminUsageLog) => {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  tooltipData.value = row
  tooltipPosition.value.x = rect.right + 8
  tooltipPosition.value.y = rect.top + rect.height / 2
  tooltipVisible.value = true
}

const hideTooltip = () => {
  tooltipVisible.value = false
  tooltipData.value = null
}

// Token tooltip functions
const showTokenTooltip = (event: MouseEvent, row: AdminUsageLog) => {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  tokenTooltipData.value = row
  tokenTooltipPosition.value.x = rect.right + 8
  tokenTooltipPosition.value.y = rect.top + rect.height / 2
  tokenTooltipVisible.value = true
}

const hideTokenTooltip = () => {
  tokenTooltipVisible.value = false
  tokenTooltipData.value = null
}
</script>
