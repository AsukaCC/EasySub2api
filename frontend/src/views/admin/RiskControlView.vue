<template>
  <AppLayout>
    <div class="views-admin-risk-control-view__panel">
      <div v-if="loading" class="views-admin-risk-control-view__panel-2">
        <div class="views-admin-risk-control-view__panel-3"></div>
      </div>

      <template v-else>
        <div class="views-admin-risk-control-view__panel-4">
          <div>
            <h1 class="views-admin-risk-control-view__heading">{{ t('admin.riskControl.title') }}</h1>
            <p class="views-admin-risk-control-view__description">{{ t('admin.riskControl.description') }}</p>
          </div>
          <div class="views-admin-risk-control-view__panel-5">
            <button type="button" class="views-admin-risk-control-view__action btn btn-secondary" :disabled="statusLoading" @click="loadStatus(false)">
              <Icon name="refresh" size="sm" :class="statusLoading ? 'views-admin-risk-control-view__icon-7' : ''" />
              {{ t('admin.riskControl.refreshStatus') }}
            </button>
            <button type="button" class="views-admin-risk-control-view__action btn btn-primary" @click="openSettings">
              <Icon name="cog" size="sm" />
              {{ t('admin.riskControl.openSettings') }}
            </button>
          </div>
        </div>

        <RiskControlSystemSettings />

        <div class="views-admin-risk-control-view__panel-6">
          <div
            v-for="item in overviewItems"
            :key="item.key"
            class="views-admin-risk-control-view__panel-7"
          >
            <div class="views-admin-risk-control-view__panel-8">
              <div class="views-admin-risk-control-view__panel-9" :class="item.iconClass">
                <Icon :name="item.icon" size="sm" />
              </div>
              <div class="views-admin-risk-control-view__panel-10">
                <div class="views-admin-risk-control-view__panel-11">
                  <p class="views-admin-risk-control-view__description-2">{{ item.label }}</p>
                  <span
                    v-if="item.badge"
                    class="views-admin-risk-control-view__text"
                    :class="item.badgeClass"
                  >
                    {{ item.badge }}
                  </span>
                </div>
                <div class="views-admin-risk-control-view__panel-12">
                  <p class="views-admin-risk-control-view__description-3">{{ item.value }}</p>
                  <p v-if="item.meta" class="views-admin-risk-control-view__description-4">{{ item.meta }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div
          v-if="showPreBlockRuntimeCard"
          data-test="pre-block-runtime-cards"
          class="views-admin-risk-control-view__panel-13"
        >
          <div data-test="pre-block-sync-card" class="card">
            <div class="views-admin-risk-control-view__panel-14">
              <div>
                <h2 class="views-admin-risk-control-view__heading-2">{{ t('admin.riskControl.preBlockSyncStatus') }}</h2>
                <p class="views-admin-risk-control-view__description">{{ t('admin.riskControl.preBlockSyncHint') }}</p>
              </div>
              <span class="views-admin-risk-control-view__text-2">
                {{ modeLabel(status?.mode ?? configForm.mode) }}
              </span>
            </div>

            <div class="views-admin-risk-control-view__panel-15 card-body">
              <div data-test="pre-block-metric-grid" class="views-admin-risk-control-view__panel-16">
                <div
                  v-for="item in preBlockMetricItems"
                  :key="item.key"
                  class="views-admin-risk-control-view__panel-17"
                  :class="item.class"
                >
                  <p class="views-admin-risk-control-view__description-5">{{ item.label }}</p>
                  <p class="views-admin-risk-control-view__description-6" :class="item.valueClass">{{ item.value }}</p>
                  <p v-if="item.meta" class="views-admin-risk-control-view__description-7">{{ item.meta }}</p>
                </div>
              </div>
            </div>
          </div>

          <div data-test="pre-block-api-key-load-card" class="card">
            <div class="views-admin-risk-control-view__panel-14">
              <div>
                <h2 class="views-admin-risk-control-view__heading-2">{{ t('admin.riskControl.preBlockAPIKeyLoad') }}</h2>
                <p class="views-admin-risk-control-view__description">
                  {{ t('admin.riskControl.preBlockAPIKeyLoadHint') }}
                </p>
              </div>
              <span class="views-admin-risk-control-view__text-2">
                {{ preBlockAPIKeyLoadSummaryText }}
              </span>
            </div>

            <div class="views-admin-risk-control-view__panel-15 card-body">
              <div
                v-if="preBlockAPIKeyLoads.length > 0"
                data-test="pre-block-api-key-load-list"
                class="views-admin-risk-control-view__panel-18"
              >
                <div
                  v-for="item in preBlockAPIKeyLoads"
                  :key="item.key_hash || item.index"
                  class="views-admin-risk-control-view__panel-19"
                >
                  <div class="views-admin-risk-control-view__panel-20">
                    <div class="views-admin-risk-control-view__panel-21">
                      <div class="views-admin-risk-control-view__panel-22">
                        <span class="views-admin-risk-control-view__text-3">#{{ item.index + 1 }}</span>
                        <span class="views-admin-risk-control-view__text-4">{{ item.masked || '-' }}</span>
                        <span class="views-admin-risk-control-view__text-5" :class="apiKeyStatusDotClass(item.status)"></span>
                      </div>
                      <p class="views-admin-risk-control-view__description-8">
                        {{ t('admin.riskControl.preBlockAPIKeyTotals', { total: formatNumber(item.total), success: formatNumber(item.success), errors: formatNumber(item.errors) }) }}
                      </p>
                    </div>
                    <div class="views-admin-risk-control-view__panel-23">
                      <div>
                        <p>{{ t('admin.riskControl.preBlockKeyActiveShort') }}</p>
                        <p class="views-admin-risk-control-view__description-9">{{ formatNumber(item.active) }}</p>
                      </div>
                      <div>
                        <p>{{ t('admin.riskControl.preBlockKeyTotalShort') }}</p>
                        <p class="views-admin-risk-control-view__description-10">{{ formatNumber(item.total) }}</p>
                      </div>
                      <div>
                        <p>{{ t('admin.riskControl.preBlockKeyAvgShort') }}</p>
                        <p class="views-admin-risk-control-view__description-10">{{ formatNumber(item.avg_latency_ms) }} ms</p>
                      </div>
                      <div>
                        <p>{{ t('admin.riskControl.preBlockKeyLastShort') }}</p>
                        <p class="views-admin-risk-control-view__description-10">{{ formatNumber(item.last_latency_ms) }} ms</p>
                      </div>
                    </div>
                  </div>
                  <div class="views-admin-risk-control-view__panel-24">
                    <div class="views-admin-risk-control-view__panel-25" :style="{ width: preBlockAPIKeyLoadWidth(item.total) }"></div>
                  </div>
                </div>
              </div>
              <p v-else class="views-admin-risk-control-view__description-11">
                {{ t('admin.riskControl.preBlockAPIKeyLoadEmpty') }}
              </p>
            </div>
          </div>
        </div>

        <div v-if="showWorkerRuntimeCard" class="card">
          <div class="views-admin-risk-control-view__panel-14">
            <div>
              <h2 class="views-admin-risk-control-view__heading-2">{{ t('admin.riskControl.workerStatus') }}</h2>
              <p class="views-admin-risk-control-view__description">{{ t('admin.riskControl.workerStatusHint') }}</p>
            </div>
            <div class="views-admin-risk-control-view__panel-26">
              <span>{{ t('admin.riskControl.autoRefresh') }}</span>
              <span v-if="status?.last_cleanup_at">
                {{ t('admin.riskControl.lastCleanup', { time: formatDateTime(status.last_cleanup_at) }) }}
              </span>
            </div>
          </div>

          <div class="views-admin-risk-control-view__panel-27 card-body">
            <div class="views-admin-risk-control-view__panel-28">
              <div class="views-admin-risk-control-view__panel-29">
                <div class="views-admin-risk-control-view__panel-30">
                  <div>
                    <p class="views-admin-risk-control-view__description-12">{{ t('admin.riskControl.queueUsage') }}</p>
                    <p class="views-admin-risk-control-view__description-8">
                      {{ formatNumber(status?.queue_length ?? 0) }} / {{ formatNumber(status?.queue_size ?? configForm.queue_size) }}
                    </p>
                  </div>
                  <span class="views-admin-risk-control-view__text-6">{{ queueUsagePercent }}</span>
                </div>
                <div class="views-admin-risk-control-view__panel-31">
                  <div class="views-admin-risk-control-view__panel-32" :style="queueUsageStyle"></div>
                </div>
              </div>

              <div class="views-admin-risk-control-view__panel-33">
                <div class="views-admin-risk-control-view__panel-34">
                  <p class="views-admin-risk-control-view__description-5">{{ t('admin.riskControl.activeWorkers') }}</p>
                  <p class="views-admin-risk-control-view__description-13">{{ status?.active_workers ?? 0 }}</p>
                </div>
                <div class="views-admin-risk-control-view__panel-35">
                  <p class="views-admin-risk-control-view__description-5">{{ t('admin.riskControl.idleWorkers') }}</p>
                  <p class="views-admin-risk-control-view__description-14">{{ status?.idle_workers ?? configForm.worker_count }}</p>
                </div>
                <div class="views-admin-risk-control-view__panel-34">
                  <p class="views-admin-risk-control-view__description-5">{{ t('admin.riskControl.processed') }}</p>
                  <p class="views-admin-risk-control-view__description-13">{{ formatNumber(status?.processed ?? 0) }}</p>
                </div>
                <div class="views-admin-risk-control-view__panel-34">
                  <p class="views-admin-risk-control-view__description-5">{{ t('admin.riskControl.droppedErrors') }}</p>
                  <p class="views-admin-risk-control-view__description-13">{{ formatNumber((status?.dropped ?? 0) + (status?.errors ?? 0)) }}</p>
                </div>
              </div>
            </div>

            <div>
              <div class="views-admin-risk-control-view__panel-36">
                <div>
                  <p class="views-admin-risk-control-view__description-12">{{ t('admin.riskControl.workerPool') }}</p>
                  <p class="views-admin-risk-control-view__description-8">
                    {{ t('admin.riskControl.workerPoolMeta', { active: status?.active_workers ?? 0, idle: status?.idle_workers ?? configForm.worker_count, total: status?.worker_count ?? configForm.worker_count }) }}
                  </p>
                </div>
                <span class="views-admin-risk-control-view__text-7">
                  {{ modeLabel(status?.mode ?? configForm.mode) }}
                </span>
              </div>
              <div class="views-admin-risk-control-view__panel-37">
                <div
                  v-for="worker in workerSlots"
                  :key="worker.id"
                  class="views-admin-risk-control-view__panel-38"
                  :class="workerSlotClass(worker.state)"
                  :title="worker.label"
                >
                  <span class="views-admin-risk-control-view__text-8">#{{ worker.id }}</span>
                  <span class="views-admin-risk-control-view__text-9" :class="workerDotClass(worker.state)"></span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="views-admin-risk-control-view__panel-39">
            <div class="views-admin-risk-control-view__panel-40">
              <div>
                <h2 class="views-admin-risk-control-view__heading-2">{{ t('admin.riskControl.records') }}</h2>
                <p class="views-admin-risk-control-view__description">{{ t('admin.riskControl.recordsHint') }}</p>
              </div>
              <button type="button" class="views-admin-risk-control-view__action btn btn-secondary" :disabled="logsLoading" @click="loadLogs">
                <Icon name="refresh" size="sm" :class="logsLoading ? 'views-admin-risk-control-view__icon-7' : ''" />
                {{ t('admin.riskControl.refresh') }}
              </button>
            </div>

            <div class="views-admin-risk-control-view__panel-41">
              <div class="views-admin-risk-control-view__panel-42">
                <Icon name="filter" size="sm" class="views-admin-risk-control-view__icon" />
                <span class="views-admin-risk-control-view__text-10">{{ t('admin.riskControl.modelFilter') }}</span>
                <span class="views-admin-risk-control-view__text-11">{{ modelFilterSummary }}</span>
              </div>
              <div v-if="modelFilterPreviewModels.length > 0" class="views-admin-risk-control-view__panel-43">
                <span
                  v-for="model in modelFilterPreviewModels"
                  :key="model"
                  class="views-admin-risk-control-view__text-12"
                >
                  {{ model }}
                </span>
                <span v-if="hiddenModelFilterModelCount > 0" class="views-admin-risk-control-view__text-13">
                  +{{ hiddenModelFilterModelCount }}
                </span>
              </div>
            </div>

            <div class="views-admin-risk-control-view__panel-44">
              <Select v-model="filters.result" :options="resultOptions" @change="reloadLogsFromFirstPage" />
              <Select v-model="filters.group_id" :options="groupFilterOptions" @change="reloadLogsFromFirstPage" />
              <Select v-model="filters.endpoint" :options="endpointOptions" @change="reloadLogsFromFirstPage" />
              <input v-model.trim="filters.search" type="search" class="input" :placeholder="t('admin.riskControl.filters.search')" @keyup.enter="reloadLogsFromFirstPage" />
              <input v-model="filters.from" type="datetime-local" class="input" :title="t('admin.riskControl.filters.from')" @change="reloadLogsFromFirstPage" />
              <input v-model="filters.to" type="datetime-local" class="input" :title="t('admin.riskControl.filters.to')" @change="reloadLogsFromFirstPage" />
            </div>
          </div>

          <div class="views-admin-risk-control-view__panel-45">
            <table class="views-admin-risk-control-view__table">
              <thead class="views-admin-risk-control-view__header">
                <tr>
                  <th class="views-admin-risk-control-view__heading-3">{{ t('admin.riskControl.table.time') }}</th>
                  <th class="views-admin-risk-control-view__heading-3">{{ t('admin.riskControl.table.group') }}</th>
                  <th class="views-admin-risk-control-view__heading-3">{{ t('admin.riskControl.table.user') }}</th>
                  <th class="views-admin-risk-control-view__heading-3">{{ t('admin.riskControl.table.apiKey') }}</th>
                  <th class="views-admin-risk-control-view__heading-3">{{ t('admin.riskControl.table.endpoint') }}</th>
                  <th class="views-admin-risk-control-view__heading-3">{{ t('admin.riskControl.table.result') }}</th>
                  <th class="views-admin-risk-control-view__heading-3">{{ t('admin.riskControl.table.highest') }}</th>
                  <th class="views-admin-risk-control-view__heading-3">{{ t('admin.riskControl.table.actionMeta') }}</th>
                  <th class="views-admin-risk-control-view__heading-3">{{ t('admin.riskControl.table.latency') }}</th>
                  <th class="views-admin-risk-control-view__heading-3">{{ t('admin.riskControl.table.input') }}</th>
                </tr>
              </thead>
              <tbody class="views-admin-risk-control-view__body">
                <tr v-if="logsLoading">
                  <td colspan="10" class="views-admin-risk-control-view__cell">{{ t('common.loading') }}</td>
                </tr>
                <tr v-else-if="logs.length === 0">
                  <td colspan="10" class="views-admin-risk-control-view__cell">{{ t('admin.riskControl.emptyLogs') }}</td>
                </tr>
                <template v-else>
                  <tr v-for="row in logs" :key="row.id" class="views-admin-risk-control-view__row">
                    <td class="views-admin-risk-control-view__cell-2">{{ formatDateTime(row.created_at) }}</td>
                    <td class="views-admin-risk-control-view__cell-2">{{ row.group_name || '-' }}</td>
                    <td class="views-admin-risk-control-view__cell-2">
                      <div>{{ row.user_email || '-' }}</div>
                      <div v-if="row.user_id" class="views-admin-risk-control-view__panel-46">UID {{ row.user_id }}</div>
                    </td>
                    <td class="views-admin-risk-control-view__cell-2">{{ row.api_key_name || '-' }}</td>
                    <td class="views-admin-risk-control-view__cell-2">
                      <div>{{ row.endpoint || '-' }}</div>
                      <div class="views-admin-risk-control-view__panel-46">{{ row.provider || '-' }} / {{ row.model || '-' }}</div>
                    </td>
                    <td class="views-admin-risk-control-view__cell-3">
                      <span class="views-admin-risk-control-view__text-14" :class="resultBadgeClass(row)">
                        {{ resultLabel(row) }}
                      </span>
                    </td>
                    <td class="views-admin-risk-control-view__cell-2">
                      <div>{{ row.highest_category || '-' }}</div>
                      <div class="views-admin-risk-control-view__panel-46">{{ percent(row.highest_score) }}</div>
                      <div v-if="row.matched_keyword" class="views-admin-risk-control-view__panel-47" :title="t('admin.riskControl.matchedKeyword') + ': ' + row.matched_keyword">
                        {{ t('admin.riskControl.matchedKeyword') }}: {{ row.matched_keyword }}
                      </div>
                    </td>
                    <td class="views-admin-risk-control-view__cell-2">
                      <div>{{ violationCountText(row) }}</div>
                      <div class="views-admin-risk-control-view__panel-46">
                        {{ row.email_sent ? t('admin.riskControl.emailSent') : t('admin.riskControl.emailNotSent') }}
                        <span v-if="row.auto_banned"> / {{ t('admin.riskControl.autoBanned') }}</span>
                      </div>
                      <button
                        v-if="canUnbanRow(row)"
                        type="button"
                        class="views-admin-risk-control-view__action-2"
                        :disabled="unbanningUserID === row.user_id"
                        @click="unbanUser(row)"
                      >
                        <Icon name="checkCircle" size="xs" :class="unbanningUserID === row.user_id ? 'views-admin-risk-control-view__icon-7' : ''" />
                        {{ unbanningUserID === row.user_id ? t('common.processing') : t('admin.riskControl.unbanUser') }}
                      </button>
                    </td>
                    <td class="views-admin-risk-control-view__cell-2">
                      <div>{{ latencyText(row.upstream_latency_ms) }}</div>
                      <div v-if="row.queue_delay_ms !== null && row.queue_delay_ms !== undefined" class="views-admin-risk-control-view__panel-46">
                        {{ t('admin.riskControl.queueDelay', { ms: row.queue_delay_ms }) }}
                      </div>
                    </td>
                    <td class="views-admin-risk-control-view__cell-4">
                      <button
                        type="button"
                        class="views-admin-risk-control-view__action-3"
                        :title="inputSummaryText(row)"
                        @click="openInputDetail(row)"
                      >
                        <span class="views-admin-risk-control-view__text-15">{{ inputSummaryText(row) }}</span>
                        <Icon name="eye" size="xs" class="views-admin-risk-control-view__icon-2" />
                      </button>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>

          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="onPageChange"
            @update:pageSize="onPageSizeChange"
          />
        </div>
      </template>

      <BaseDialog :show="settingsOpen" :title="t('admin.riskControl.settingsTitle')" width="extra-wide" @close="settingsOpen = false">
        <div class="views-admin-risk-control-view__panel">
          <div class="views-admin-risk-control-view__panel-48">
            <button
              v-for="tab in settingsTabs"
              :key="tab.id"
              type="button"
              class="views-admin-risk-control-view__action-4"
              :class="activeSettingsTab === tab.id ? 'views-admin-risk-control-view__action-15' : 'views-admin-risk-control-view__action-16'"
              @click="activeSettingsTab = tab.id"
            >
              {{ tab.label }}
            </button>
          </div>

          <div v-if="activeSettingsTab === 'basic'" class="views-admin-risk-control-view__panel-49">
            <div class="views-admin-risk-control-view__panel-50">
              <div class="views-admin-risk-control-view__panel-51">
                <div>
                  <p class="views-admin-risk-control-view__description-12">{{ t('admin.riskControl.enabled') }}</p>
                  <p class="views-admin-risk-control-view__description-8">{{ t('admin.riskControl.enabledHint') }}</p>
                </div>
                <Toggle v-model="configForm.enabled" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.riskControl.mode') }}</label>
                <Select v-model="configForm.mode" :options="modeOptions" />
                <p class="views-admin-risk-control-view__description-15">{{ modeDescription(configForm.mode) }}</p>
              </div>
              <div>
                <label class="input-label">{{ t('admin.riskControl.baseUrl') }}</label>
                <input v-model.trim="configForm.base_url" type="url" class="input" placeholder="https://api.openai.com" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.riskControl.model') }}</label>
                <input v-model.trim="configForm.model" type="text" class="input" placeholder="omni-moderation-latest" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.riskControl.timeoutMs') }}</label>
                <input v-model.number="configForm.timeout_ms" type="number" min="500" max="30000" class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.riskControl.retryCount') }}</label>
                <input v-model.number="configForm.retry_count" type="number" min="0" max="5" class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.riskControl.sampleRate') }}</label>
                <div class="views-admin-risk-control-view__panel-52">
                  <input v-model.number="configForm.sample_rate" type="number" min="0" max="100" step="1" class="views-admin-risk-control-view__field input" />
                  <span class="views-admin-risk-control-view__text-16">%</span>
                </div>
              </div>
              <div>
                <label class="input-label">{{ t('admin.riskControl.proxy') }}</label>
                <ProxySelector v-model="configForm.proxy_id" :proxies="proxies" />
                <p class="views-admin-risk-control-view__description-15">{{ t('admin.riskControl.proxyHint') }}</p>
              </div>
            </div>

            <div class="views-admin-risk-control-view__panel-53">
              <div class="views-admin-risk-control-view__panel-54">
                <div class="views-admin-risk-control-view__panel-55">
                  <span class="views-admin-risk-control-view__text-17">
                    <Icon name="key" size="md" />
                  </span>
                  <div>
                    <label class="views-admin-risk-control-view__text-6">{{ t('admin.riskControl.apiKeys') }}</label>
                    <p class="views-admin-risk-control-view__description-16">
                      {{ t('admin.riskControl.apiKeysHint', { count: configForm.api_key_count }) }}
                    </p>
                  </div>
                </div>
                <div class="views-admin-risk-control-view__panel-5">
                  <button
                    type="button"
                    class="views-admin-risk-control-view__action btn btn-secondary"
                    :disabled="apiKeyTesting || inputApiKeyCount === 0 || configForm.clear_api_key"
                    @click="testApiKeys(true)"
                  >
                    <Icon name="beaker" size="sm" :class="apiKeyTesting ? 'views-admin-risk-control-view__icon-8' : ''" />
                    {{ apiKeyTesting ? t('admin.riskControl.testingApiKeys') : t('admin.riskControl.testInputApiKeys') }}
                  </button>
                  <button
                    type="button"
                    class="views-admin-risk-control-view__action btn btn-secondary"
                    :disabled="apiKeyTesting || effectiveStoredApiKeyCount === 0 || pendingDeletedApiKeyCount > 0 || configForm.clear_api_key || configForm.api_keys_mode === 'replace'"
                    @click="testApiKeys(false)"
                  >
                    <Icon name="shield" size="sm" />
                    {{ storedApiKeyTestButtonText }}
                  </button>
                  <button
                    v-if="configForm.api_key_configured"
                    type="button"
                    class="views-admin-risk-control-view__action btn btn-secondary"
                    @click="toggleClearApiKey"
                  >
                    <Icon :name="configForm.clear_api_key ? 'x' : 'trash'" size="sm" />
                    {{ configForm.clear_api_key ? t('admin.riskControl.keepApiKey') : t('admin.riskControl.clearApiKey') }}
                  </button>
                </div>
              </div>

              <div class="views-admin-risk-control-view__panel-56">
                <div class="views-admin-risk-control-view__panel-57">
                  <div class="views-admin-risk-control-view__panel-58">
                    <div class="views-admin-risk-control-view__panel-59">
                      <span class="views-admin-risk-control-view__text-18">{{ t('admin.riskControl.apiKeysWriteMode') }}</span>
                      <span class="views-admin-risk-control-view__text-19">{{ apiKeysModeHint }}</span>
                    </div>
                    <div class="views-admin-risk-control-view__panel-60">
                      <button
                        type="button"
                        class="views-admin-risk-control-view__action-5"
                        :class="configForm.api_keys_mode === 'append' ? 'views-admin-risk-control-view__action-17' : 'views-admin-risk-control-view__action-18'"
                        :disabled="configForm.clear_api_key"
                        @click="setAPIKeysMode('append')"
                      >
                        {{ t('admin.riskControl.apiKeysModeAppend') }}
                      </button>
                      <button
                        type="button"
                        class="views-admin-risk-control-view__action-5"
                        :class="configForm.api_keys_mode === 'replace' ? 'views-admin-risk-control-view__action-19' : 'views-admin-risk-control-view__action-18'"
                        :disabled="configForm.clear_api_key"
                        @click="setAPIKeysMode('replace')"
                      >
                        {{ t('admin.riskControl.apiKeysModeReplace') }}
                      </button>
                    </div>
                  </div>
                  <textarea
                    v-model="configForm.api_keys_text"
                    class="views-admin-risk-control-view__field-2 input"
                    :placeholder="apiKeysPlaceholder"
                    autocomplete="new-password"
                    :disabled="configForm.clear_api_key"
                  ></textarea>
                  <div class="views-admin-risk-control-view__panel-61">
                    <span class="views-admin-risk-control-view__text-20">
                      {{ t('admin.riskControl.inputApiKeyCount', { count: inputApiKeyCount }) }}
                    </span>
                    <span v-if="configForm.api_key_configured" class="views-admin-risk-control-view__text-20">
                      {{ t('admin.riskControl.storedApiKeyCount', { count: configForm.api_key_count }) }}
                    </span>
                    <span v-if="configForm.clear_api_key" class="views-admin-risk-control-view__text-21">
                      {{ t('admin.riskControl.apiKeyWillClear') }}
                    </span>
                    <span v-else-if="pendingDeletedApiKeyCount > 0" class="views-admin-risk-control-view__text-22">
                      {{ t('admin.riskControl.apiKeyPendingDeleteCount', { count: pendingDeletedApiKeyCount }) }}
                    </span>
                    <span v-if="configForm.api_keys_mode === 'replace'" class="views-admin-risk-control-view__text-22">
                      {{ t('admin.riskControl.apiKeysReplaceWarning') }}
                    </span>
                  </div>

                  <div class="views-admin-risk-control-view__panel-62" @paste="handleModerationImagePaste">
                    <div class="views-admin-risk-control-view__panel-36">
                      <div>
                        <p class="views-admin-risk-control-view__text-6">{{ t('admin.riskControl.auditTestInput') }}</p>
                        <p class="views-admin-risk-control-view__description-8">{{ t('admin.riskControl.auditTestInputHint') }}</p>
                      </div>
                      <button
                        v-if="moderationTestPrompt || moderationTestImages.length > 0 || moderationTestResult"
                        type="button"
                        class="views-admin-risk-control-view__action-6"
                        @click="clearModerationTestInput"
                      >
                        <Icon name="x" size="xs" />
                        {{ t('admin.riskControl.clearAuditTest') }}
                      </button>
                    </div>
                    <textarea
                      v-model="moderationTestPrompt"
                      class="views-admin-risk-control-view__field-3 input"
                      :placeholder="t('admin.riskControl.auditTestPromptPlaceholder')"
                    ></textarea>
                    <div
                      class="views-admin-risk-control-view__panel-63"
                      @dragover.prevent
                      @drop.prevent="handleModerationImageDrop"
                    >
                      <div class="views-admin-risk-control-view__panel-64">
                        <div class="views-admin-risk-control-view__panel-65">
                          <Icon name="upload" size="md" class="views-admin-risk-control-view__icon-3" />
                          <div>
                            <p class="views-admin-risk-control-view__description-17">{{ t('admin.riskControl.auditTestImages') }}</p>
                            <p class="views-admin-risk-control-view__description-8">{{ t('admin.riskControl.auditTestImagesHint') }}</p>
                          </div>
                        </div>
                        <label class="views-admin-risk-control-view__label btn btn-secondary">
                          <Icon name="plus" size="sm" />
                          {{ t('admin.riskControl.addAuditTestImage') }}
                          <input type="file" accept="image/*" multiple class="views-admin-risk-control-view__field-4" @change="handleModerationImageUpload" />
                        </label>
                      </div>
                      <div v-if="moderationTestImages.length > 0" class="views-admin-risk-control-view__panel-66">
                        <div
                          v-for="(image, index) in moderationTestImages"
                          :key="image.slice(0, 64) + index"
                          class="views-admin-risk-control-view__panel-67"
                        >
                          <img :src="image" alt="" class="views-admin-risk-control-view__image" />
                          <button
                            type="button"
                            class="views-admin-risk-control-view__action-7"
                            @click="removeModerationTestImage(index)"
                          >
                            <Icon name="x" size="xs" :stroke-width="2" />
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="views-admin-risk-control-view__panel-62">
                  <div class="views-admin-risk-control-view__panel-68">
                    <div class="views-admin-risk-control-view__panel-21">
                      <p class="views-admin-risk-control-view__text-6">{{ t('admin.riskControl.apiKeyHealth') }}</p>
                      <p class="views-admin-risk-control-view__description-8">{{ t('admin.riskControl.apiKeyFreezeRule') }}</p>
                    </div>
                    <span class="views-admin-risk-control-view__text-23">
                      {{ t('admin.riskControl.apiKeyRows', { count: apiKeyRows.length }) }}
                    </span>
                  </div>

                  <div v-if="apiKeyRows.length === 0" class="views-admin-risk-control-view__panel-69">
                    <Icon name="infoCircle" size="lg" class="views-admin-risk-control-view__icon-4" />
                    <p class="views-admin-risk-control-view__description-18">{{ t('admin.riskControl.apiKeyHealthEmpty') }}</p>
                    <p class="views-admin-risk-control-view__description-8">{{ t('admin.riskControl.apiKeyHealthEmptyHint') }}</p>
                  </div>
                  <div v-else class="views-admin-risk-control-view__panel-70">
                    <div class="views-admin-risk-control-view__panel-70" :class="apiKeyRowsExpanded ? 'views-admin-risk-control-view__panel-112' : ''">
                      <div
                        v-for="(row, index) in visibleApiKeyRows"
                        :key="apiKeyRowKey(row, index)"
                        class="views-admin-risk-control-view__panel-71"
                        :class="isStoredApiKeyPendingDelete(row) ? 'views-admin-risk-control-view__panel-113' : 'views-admin-risk-control-view__panel-114'"
                      >
                        <div class="views-admin-risk-control-view__panel-72">
                          <div class="views-admin-risk-control-view__panel-21">
                            <div class="views-admin-risk-control-view__panel-73">
                              <span class="views-admin-risk-control-view__text-24">{{ row.masked || '-' }}</span>
                              <span
                                class="views-admin-risk-control-view__text-25"
                                :class="row.configured ? 'views-admin-risk-control-view__action-15' : 'views-admin-risk-control-view__text-42'"
                              >
                                {{ isStoredApiKeyPendingDelete(row) ? t('admin.riskControl.apiKeyPendingDelete') : row.configured ? t('admin.riskControl.apiKeyConfigured') : t('admin.riskControl.apiKeyTemporary') }}
                              </span>
                            </div>
                            <p class="views-admin-risk-control-view__description-19">{{ apiKeyStatusMeta(row) }}</p>
                          </div>
                          <div class="views-admin-risk-control-view__panel-74">
                            <span class="views-admin-risk-control-view__text-26" :class="apiKeyStatusBadgeClass(row.status)">
                              <span class="views-admin-risk-control-view__text-27" :class="apiKeyStatusDotClass(row.status)"></span>
                              {{ apiKeyStatusLabel(row.status) }}
                            </span>
                            <button
                              v-if="row.configured && !configForm.clear_api_key"
                              type="button"
                              class="views-admin-risk-control-view__action-8"
                              :title="isStoredApiKeyPendingDelete(row) ? t('admin.riskControl.undoDeleteApiKey') : t('admin.riskControl.deleteApiKey')"
                              @click="toggleDeleteStoredApiKey(row)"
                            >
                              <Icon :name="isStoredApiKeyPendingDelete(row) ? 'refresh' : 'trash'" size="xs" />
                            </button>
                          </div>
                        </div>
                        <p v-if="row.last_error" class="views-admin-risk-control-view__description-20">
                          {{ row.last_error }}
                        </p>
                      </div>
                    </div>

                    <div v-if="canToggleApiKeyRows" class="views-admin-risk-control-view__panel-75">
                      <span class="views-admin-risk-control-view__text-28">
                        {{ apiKeyRowsExpanded ? t('admin.riskControl.apiKeyRowsExpanded', { count: apiKeyRows.length }) : t('admin.riskControl.apiKeyRowsCollapsed', { count: hiddenApiKeyRowCount }) }}
                      </span>
                      <button
                        type="button"
                        class="views-admin-risk-control-view__action-9"
                        @click="apiKeyRowsExpanded = !apiKeyRowsExpanded"
                      >
                        <Icon :name="apiKeyRowsExpanded ? 'chevronUp' : 'chevronDown'" size="xs" />
                        {{ apiKeyRowsExpanded ? t('admin.riskControl.collapseApiKeyRows') : t('admin.riskControl.expandApiKeyRows') }}
                      </button>
                    </div>
                  </div>

                  <div v-if="moderationTestResult" class="views-admin-risk-control-view__panel-76">
                    <div class="views-admin-risk-control-view__panel-77">
                      <div>
                        <p class="views-admin-risk-control-view__text-6">{{ t('admin.riskControl.auditTestResult') }}</p>
                        <p class="views-admin-risk-control-view__description-8">
                          {{ t('admin.riskControl.auditTestHighest', { category: moderationTestResult.highest_category || '-', score: percent(moderationTestResult.highest_score) }) }}
                        </p>
                      </div>
                      <span class="views-admin-risk-control-view__text-29" :class="moderationTestResult.flagged ? 'views-admin-risk-control-view__text-43' : 'views-admin-risk-control-view__text-44'">
                        {{ moderationTestResult.flagged ? t('admin.riskControl.auditTestFlagged') : t('admin.riskControl.auditTestPassed') }}
                      </span>
                    </div>
                    <div class="views-admin-risk-control-view__panel-78">
                      <div class="views-admin-risk-control-view__panel-79">
                        <span>{{ t('admin.riskControl.auditTestComposite') }}</span>
                        <span class="views-admin-risk-control-view__text-30">{{ percent(moderationTestResult.composite_score) }}</span>
                      </div>
                      <div class="views-admin-risk-control-view__panel-80">
                        <div class="views-admin-risk-control-view__panel-81" :class="moderationTestResult.flagged ? 'views-admin-risk-control-view__panel-115' : 'views-admin-risk-control-view__panel-116'" :style="{ width: percentWidth(moderationTestResult.composite_score) }"></div>
                      </div>
                    </div>
                    <div class="views-admin-risk-control-view__panel-82">
                      <div v-for="score in moderationScoreRows" :key="score.category">
                        <div class="views-admin-risk-control-view__panel-83">
                          <span class="views-admin-risk-control-view__text-31">{{ score.category }}</span>
                          <span class="views-admin-risk-control-view__text-32">{{ percent(score.score) }} / {{ percent(score.threshold) }}</span>
                        </div>
                        <div class="views-admin-risk-control-view__panel-84">
                          <div class="views-admin-risk-control-view__panel-81" :class="score.hit ? 'views-admin-risk-control-view__panel-115' : 'views-admin-risk-control-view__panel-117'" :style="{ width: percentWidth(score.score) }"></div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-else-if="activeSettingsTab === 'scope'" class="views-admin-risk-control-view__panel-49">
            <div class="views-admin-risk-control-view__panel-4">
              <div>
                <h3 class="views-admin-risk-control-view__heading-4">{{ t('admin.riskControl.groupScope') }}</h3>
                <p class="views-admin-risk-control-view__description">{{ t('admin.riskControl.groupScopeHint') }}</p>
              </div>
              <div class="views-admin-risk-control-view__panel-85">
                <button
                  type="button"
                  class="views-admin-risk-control-view__action-10"
                  :class="configForm.all_groups ? 'views-admin-risk-control-view__action-20' : 'views-admin-risk-control-view__action-21'"
                  @click="configForm.all_groups = true"
                >
                  {{ t('admin.riskControl.allGroups') }}
                </button>
                <button
                  type="button"
                  class="views-admin-risk-control-view__action-10"
                  :class="!configForm.all_groups ? 'views-admin-risk-control-view__action-20' : 'views-admin-risk-control-view__action-21'"
                  @click="configForm.all_groups = false"
                >
                  {{ t('admin.riskControl.selectedGroups') }}
                </button>
              </div>
            </div>

            <div v-if="!configForm.all_groups" class="views-admin-risk-control-view__panel-28">
              <div class="views-admin-risk-control-view__panel-52">
                <Icon name="search" size="sm" class="views-admin-risk-control-view__icon-5" />
                <input v-model.trim="groupSearch" type="search" class="views-admin-risk-control-view__field-5 input" :placeholder="t('admin.riskControl.searchGroups')" />
              </div>
              <div class="views-admin-risk-control-view__panel-86">
                <button
                  v-for="group in filteredGroups"
                  :key="group.id"
                  type="button"
                  class="views-admin-risk-control-view__action-11"
                  :class="isGroupSelected(group.id) ? 'views-admin-risk-control-view__action-22' : 'views-admin-risk-control-view__action-23'"
                  @click="toggleGroup(group.id)"
                >
                  <span class="views-admin-risk-control-view__panel-21">
                    <span class="views-admin-risk-control-view__text-33">{{ group.name }}</span>
                    <span class="views-admin-risk-control-view__text-34">{{ group.platform }}</span>
                  </span>
                  <span
                    class="views-admin-risk-control-view__text-35"
                    :class="isGroupSelected(group.id) ? 'views-admin-risk-control-view__text-45' : 'views-admin-risk-control-view__text-46'"
                  >
                    <Icon name="check" size="xs" :stroke-width="2" />
                  </span>
                </button>
                <p v-if="filteredGroups.length === 0" class="views-admin-risk-control-view__description-21">{{ t('admin.riskControl.noGroups') }}</p>
              </div>
            </div>

            <div class="views-admin-risk-control-view__panel-87">
              <div class="views-admin-risk-control-view__panel-88">
                <div>
                  <h3 class="views-admin-risk-control-view__heading-4">{{ t('admin.riskControl.modelFilter') }}</h3>
                  <p class="views-admin-risk-control-view__description">{{ t('admin.riskControl.modelFilterHint') }}</p>
                </div>
                <span class="views-admin-risk-control-view__text-36">
                  {{ modelFilterSummary }}
                </span>
              </div>

              <div class="views-admin-risk-control-view__panel-89">
                <button
                  v-for="option in modelFilterOptions"
                  :key="option.value"
                  type="button"
                  class="views-admin-risk-control-view__action-12"
                  :class="configForm.model_filter_type === option.value
                    ? 'views-admin-risk-control-view__action-24'
                    : 'views-admin-risk-control-view__action-23'"
                  @click="setModelFilterType(option.value)"
                >
                  <div class="views-admin-risk-control-view__panel-90">
                    <span class="views-admin-risk-control-view__text-8">{{ option.label }}</span>
                    <span
                      class="views-admin-risk-control-view__text-37"
                      :class="configForm.model_filter_type === option.value
                        ? 'views-admin-risk-control-view__text-45'
                        : 'views-admin-risk-control-view__text-46'"
                    >
                      <Icon name="check" size="xs" :stroke-width="2" />
                    </span>
                  </div>
                  <p class="views-admin-risk-control-view__description-19">{{ option.description }}</p>
                </button>
              </div>

              <div v-if="configForm.model_filter_type !== 'all'" class="views-admin-risk-control-view__panel-70">
                <label class="input-label">{{ t('admin.riskControl.modelFilterModels') }}</label>
                <ModelWhitelistSelector v-model="configForm.model_filter_models" />
                <p class="views-admin-risk-control-view__description-5">
                  {{ t('admin.riskControl.modelFilterModelCount', { count: modelFilterModelCount }) }}
                </p>
              </div>
            </div>
          </div>

          <div v-else-if="activeSettingsTab === 'runtime'" class="views-admin-risk-control-view__panel-50">
            <div>
              <label class="input-label">{{ t('admin.riskControl.workerCount') }}</label>
              <input v-model.number="configForm.worker_count" type="number" min="1" max="32" class="input" />
            </div>
            <div>
              <label class="input-label">{{ t('admin.riskControl.queueSize') }}</label>
              <input v-model.number="configForm.queue_size" type="number" min="100" max="100000" class="input" />
            </div>
            <div class="views-admin-risk-control-view__panel-91">
              <div>
                <p class="views-admin-risk-control-view__description-12">{{ t('admin.riskControl.recordNonHits') }}</p>
                <p class="views-admin-risk-control-view__description-8">{{ t('admin.riskControl.recordNonHitsHint') }}</p>
              </div>
              <Toggle v-model="configForm.record_non_hits" />
            </div>
            <div class="views-admin-risk-control-view__panel-92">
              <div class="views-admin-risk-control-view__panel-93">
                <div>
                  <p class="views-admin-risk-control-view__description-12">{{ t('admin.riskControl.preHashCheck') }}</p>
                  <p class="views-admin-risk-control-view__description-8">{{ t('admin.riskControl.preHashCheckHint') }}</p>
                </div>
                <Toggle v-model="configForm.pre_hash_check_enabled" />
              </div>
              <div class="views-admin-risk-control-view__panel-94">
                <div class="views-admin-risk-control-view__panel-95">
                  <div>
                    <p class="views-admin-risk-control-view__description-12">
                      {{ t('admin.riskControl.flaggedHashCount', { count: formatNumber(status?.flagged_hash_count ?? 0) }) }}
                    </p>
                    <p class="views-admin-risk-control-view__description-8">{{ t('admin.riskControl.flaggedHashHint') }}</p>
                  </div>
                  <button
                    type="button"
                    class="views-admin-risk-control-view__action-13 btn btn-secondary"
                    :disabled="hashActionLoading || (status?.flagged_hash_count ?? 0) === 0"
                    @click="clearFlaggedHashes"
                  >
                    <Icon name="trash" size="sm" :class="hashActionLoading ? 'views-admin-risk-control-view__icon-8' : ''" />
                    {{ t('admin.riskControl.clearFlaggedHashes') }}
                  </button>
                </div>
                <div class="views-admin-risk-control-view__panel-96">
                  <input
                    v-model.trim="flaggedHashInput"
                    type="text"
                    class="views-admin-risk-control-view__field-6 input"
                    :placeholder="t('admin.riskControl.flaggedHashPlaceholder')"
                  />
                  <button
                    type="button"
                    class="views-admin-risk-control-view__action-14 btn btn-secondary"
                    :disabled="hashActionLoading || !isFlaggedHashInputValid"
                    @click="deleteFlaggedHash"
                  >
                    <Icon name="trash" size="sm" />
                    {{ t('admin.riskControl.deleteFlaggedHash') }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div v-else-if="activeSettingsTab === 'response'" class="views-admin-risk-control-view__panel-49">
            <div class="views-admin-risk-control-view__panel-50">
              <div>
                <label class="input-label">{{ t('admin.riskControl.blockStatus') }}</label>
                <input v-model.number="configForm.block_status" type="number" min="400" max="599" class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.riskControl.blockMessage') }}</label>
                <input v-model.trim="configForm.block_message" type="text" class="input" />
              </div>
              <div class="views-admin-risk-control-view__panel-51">
                <div>
                  <p class="views-admin-risk-control-view__description-12">{{ t('admin.riskControl.emailOnHit') }}</p>
                  <p class="views-admin-risk-control-view__description-8">{{ t('admin.riskControl.emailOnHitHint') }}</p>
                </div>
                <Toggle v-model="configForm.email_on_hit" />
              </div>
              <div class="views-admin-risk-control-view__panel-51">
                <div>
                  <p class="views-admin-risk-control-view__description-12">{{ t('admin.riskControl.autoBan') }}</p>
                  <p class="views-admin-risk-control-view__description-8">{{ t('admin.riskControl.autoBanHint') }}</p>
                </div>
                <Toggle v-model="configForm.auto_ban_enabled" />
              </div>
              <div class="views-admin-risk-control-view__panel-91">
                <div>
                  <p class="views-admin-risk-control-view__description-12">{{ t('admin.riskControl.cyberPolicyExcludeBan') }}</p>
                  <p class="views-admin-risk-control-view__description-8">{{ t('admin.riskControl.cyberPolicyExcludeBanHint') }}</p>
                </div>
                <Toggle v-model="configForm.cyber_policy_exclude_from_ban_count" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.riskControl.banThreshold') }}</label>
                <input v-model.number="configForm.ban_threshold" type="number" min="1" max="1000" class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.riskControl.violationWindowHours') }}</label>
                <input v-model.number="configForm.violation_window_hours" type="number" min="1" max="8760" class="input" />
              </div>
            </div>
          </div>

          <div v-else-if="activeSettingsTab === 'riskThresholds'" class="views-admin-risk-control-view__panel-49">
            <div class="views-admin-risk-control-view__panel-40">
              <div>
                <h3 class="views-admin-risk-control-view__heading-4">{{ t('admin.riskControl.riskThresholds') }}</h3>
                <p class="views-admin-risk-control-view__description">{{ t('admin.riskControl.riskThresholdsHint') }}</p>
              </div>
              <button
                type="button"
                class="views-admin-risk-control-view__action-14 btn btn-secondary"
                @click="resetRiskThresholds"
              >
                <Icon name="refresh" size="sm" />
                {{ t('admin.riskControl.riskThresholdReset') }}
              </button>
            </div>

            <div class="views-admin-risk-control-view__panel-97">
              <div
                v-for="row in riskThresholdRows"
                :key="row.category"
                class="views-admin-risk-control-view__panel-98"
              >
                <div class="views-admin-risk-control-view__panel-77">
                  <div class="views-admin-risk-control-view__panel-21">
                    <label class="views-admin-risk-control-view__text-33" :for="`risk-threshold-${row.category}`">
                      {{ row.category }}
                    </label>
                    <p class="views-admin-risk-control-view__description-8">
                      {{ t('admin.riskControl.riskThresholdDefault', { value: formatThresholdPercent(row.defaultValue) }) }}
                    </p>
                  </div>
                  <span class="views-admin-risk-control-view__text-38">
                    {{ formatThresholdPercent(row.value) }}
                  </span>
                </div>
                <div class="views-admin-risk-control-view__panel-78">
                  <label class="views-admin-risk-control-view__field-4" :for="`risk-threshold-${row.category}`">
                    {{ t('admin.riskControl.riskThresholdPercent') }}
                  </label>
                  <div class="views-admin-risk-control-view__panel-52">
                    <input
                      :id="`risk-threshold-${row.category}`"
                      v-model.number="configForm.thresholds[row.category]"
                      :data-test="`risk-threshold-${row.category}`"
                      type="number"
                      min="0"
                      max="100"
                      step="0.1"
                      class="views-admin-risk-control-view__field-7 input"
                    />
                    <span class="views-admin-risk-control-view__text-16">%</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-else-if="activeSettingsTab === 'keywords'" class="views-admin-risk-control-view__panel-49">
            <div
              class="views-admin-risk-control-view__panel-99"
              :class="keywordNotice.toneClass"
            >
              <Icon
                :name="keywordNotice.icon"
                size="md"
                :class="keywordNotice.iconClass"
              />
              <div class="views-admin-risk-control-view__panel-100">
                <p class="views-admin-risk-control-view__text-10" :class="keywordNotice.titleClass">{{ keywordNotice.title }}</p>
                <p class="views-admin-risk-control-view__description-8">{{ keywordNotice.description }}</p>
              </div>
            </div>

            <div class="views-admin-risk-control-view__panel-70">
              <label class="input-label">{{ t('admin.riskControl.keywordBlockingMode') }}</label>
              <div class="views-admin-risk-control-view__panel-101">
                <button
                  v-for="option in keywordBlockingModeOptions"
                  :key="option.value"
                  type="button"
                  class="views-admin-risk-control-view__action-12"
                  :class="configForm.keyword_blocking_mode === option.value
                    ? 'views-admin-risk-control-view__action-24'
                    : 'views-admin-risk-control-view__action-23'"
                  @click="configForm.keyword_blocking_mode = option.value"
                >
                  <div class="views-admin-risk-control-view__panel-90">
                    <span class="views-admin-risk-control-view__text-8">{{ option.label }}</span>
                    <span
                      class="views-admin-risk-control-view__text-37"
                      :class="configForm.keyword_blocking_mode === option.value
                        ? 'views-admin-risk-control-view__text-45'
                        : 'views-admin-risk-control-view__text-46'"
                    >
                      <Icon name="check" size="xs" :stroke-width="2" />
                    </span>
                  </div>
                  <p class="views-admin-risk-control-view__description-19">{{ option.description }}</p>
                </button>
              </div>
            </div>

            <div>
              <div class="views-admin-risk-control-view__panel-102">
                <label class="views-admin-risk-control-view__label-2 input-label">{{ t('admin.riskControl.blockedKeywords') }}</label>
                <span class="views-admin-risk-control-view__text-39">
                  {{ t('admin.riskControl.blockedKeywordCount', { count: blockedKeywordCount }) }}
                </span>
              </div>
              <textarea
                v-model="configForm.blocked_keywords_text"
                class="views-admin-risk-control-view__field-8 input"
                :placeholder="t('admin.riskControl.blockedKeywordsPlaceholder')"
                :disabled="configForm.keyword_blocking_mode === 'api_only'"
              ></textarea>
              <p class="views-admin-risk-control-view__description-22">
                {{ t('admin.riskControl.blockedKeywordsLimit', { max: blockedKeywordMax }) }}
              </p>
            </div>
          </div>

          <div v-else class="views-admin-risk-control-view__panel-50">
            <div>
              <label class="input-label">{{ t('admin.riskControl.hitRetentionDays') }}</label>
              <input v-model.number="configForm.hit_retention_days" type="number" min="1" max="3650" class="input" />
            </div>
            <div>
              <label class="input-label">{{ t('admin.riskControl.nonHitRetentionDays') }}</label>
              <input v-model.number="configForm.non_hit_retention_days" type="number" min="1" max="3" class="input" />
            </div>
            <div class="views-admin-risk-control-view__panel-103">
              <div class="views-admin-risk-control-view__panel-104">
                <Icon name="database" size="md" class="views-admin-risk-control-view__icon-6" />
                <span>{{ t('admin.riskControl.cleanupStats', { hit: status?.last_cleanup_deleted_hit ?? 0, nonHit: status?.last_cleanup_deleted_non_hit ?? 0 }) }}</span>
              </div>
            </div>
          </div>
        </div>

        <template #footer>
          <div class="views-admin-risk-control-view__panel-105">
            <button type="button" class="btn btn-secondary" @click="settingsOpen = false">{{ t('common.cancel') }}</button>
            <button type="button" class="views-admin-risk-control-view__action btn btn-primary" :disabled="saving" @click="saveConfig">
              <Icon v-if="saving" name="refresh" size="sm" class="views-admin-risk-control-view__icon-7" />
              <Icon v-else name="check" size="sm" />
              {{ saving ? t('common.saving') : t('admin.riskControl.saveConfig') }}
            </button>
          </div>
        </template>
      </BaseDialog>

      <BaseDialog
        :show="inputDetailRow !== null"
        :title="t('admin.riskControl.inputDetailTitle')"
        width="wide"
        @close="closeInputDetail"
      >
        <div v-if="inputDetailRow" class="views-admin-risk-control-view__panel-49">
          <div class="views-admin-risk-control-view__panel-106">
            <div class="views-admin-risk-control-view__panel-107">
              <p class="views-admin-risk-control-view__description-23">{{ t('admin.riskControl.table.time') }}</p>
              <p class="views-admin-risk-control-view__description-24">{{ formatDateTime(inputDetailRow.created_at) }}</p>
            </div>
            <div class="views-admin-risk-control-view__panel-107">
              <p class="views-admin-risk-control-view__description-23">{{ t('admin.riskControl.table.user') }}</p>
              <p class="views-admin-risk-control-view__description-24">{{ inputDetailRow.user_email || '-' }}</p>
            </div>
            <div class="views-admin-risk-control-view__panel-107">
              <p class="views-admin-risk-control-view__description-23">{{ t('admin.riskControl.table.result') }}</p>
              <span class="views-admin-risk-control-view__text-40" :class="resultBadgeClass(inputDetailRow)">
                {{ resultLabel(inputDetailRow) }}
              </span>
            </div>
            <div class="views-admin-risk-control-view__panel-107">
              <p class="views-admin-risk-control-view__description-23">{{ t('admin.riskControl.table.highest') }}</p>
              <p class="views-admin-risk-control-view__description-24">
                {{ inputDetailRow.highest_category || '-' }} / {{ percent(inputDetailRow.highest_score) }}
              </p>
            </div>
            <div v-if="inputDetailRow.matched_keyword" class="views-admin-risk-control-view__panel-108">
              <p class="views-admin-risk-control-view__description-25">{{ t('admin.riskControl.matchedKeyword') }}</p>
              <p class="views-admin-risk-control-view__description-26" :title="inputDetailRow.matched_keyword">{{ inputDetailRow.matched_keyword }}</p>
            </div>
          </div>

          <div class="views-admin-risk-control-view__panel-109">
            <div class="views-admin-risk-control-view__panel-110">
              <div>
                <p class="views-admin-risk-control-view__text-6">{{ t('admin.riskControl.inputDetailContent') }}</p>
                <p class="views-admin-risk-control-view__description-8">
                  {{ inputDetailRow.endpoint || '-' }} · {{ inputDetailRow.provider || '-' }} / {{ inputDetailRow.model || '-' }}
                </p>
              </div>
              <span v-if="inputDetailRow.group_name" class="views-admin-risk-control-view__text-41">
                {{ inputDetailRow.group_name }}
              </span>
            </div>
            <pre class="views-admin-risk-control-view__pre">{{ inputDetailText }}</pre>
          </div>
        </div>

        <template #footer>
          <div class="views-admin-risk-control-view__panel-111">
            <button type="button" class="btn btn-secondary" @click="closeInputDetail">{{ t('common.close') }}</button>
          </div>
        </template>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import RiskControlSystemSettings from '@/components/admin/RiskControlSystemSettings.vue'
import Pagination from '@/components/common/Pagination.vue'
import ModelWhitelistSelector from '@/components/account/ModelWhitelistSelector.vue'
import ProxySelector from '@/components/common/ProxySelector.vue'
import { adminAPI } from '@/api/admin'
import type {
  ContentModerationAPIKeyLoad,
  ContentModerationAPIKeyStatus,
  ContentModerationConfig,
  ContentModerationLog,
  ContentModerationModelFilter,
  ContentModerationModelFilterType,
  ContentModerationRuntimeStatus,
  ContentModerationTestAuditResult,
  KeywordBlockingMode,
  ModerationMode,
  UpdateContentModerationConfig,
} from '@/api/admin/riskControl'
import type { AdminGroup, Proxy, SelectOption } from '@/types'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime as formatDateTimeValue } from '@/utils/format'

type SettingsTab = 'basic' | 'scope' | 'runtime' | 'response' | 'riskThresholds' | 'retention' | 'keywords'
type WorkerSlotState = 'active' | 'idle' | 'disabled'
type APIKeysWriteMode = 'append' | 'replace'
type OverviewIcon = 'shield' | 'key' | 'users' | 'document'
type OverviewItem = {
  key: string
  label: string
  value: string
  meta: string
  icon: OverviewIcon
  iconClass: string
  badge?: string
  badgeClass?: string
}
type ModerationScoreRow = {
  category: string
  score: number
  threshold: number
  hit: boolean
}
type RiskThresholdRow = {
  category: string
  value: number
  defaultValue: number
}

const maxModerationTestImages = 1
const maxModerationTestImageSize = 8 * 1024 * 1024
const maxVisibleApiKeyRows: number = 3
const blockedKeywordMax = 10000
const riskThresholdDefaults: Record<string, number> = {
  harassment: 98,
  'harassment/threatening': 90,
  hate: 65,
  'hate/threatening': 65,
  illicit: 95,
  'illicit/violent': 95,
  'self-harm': 65,
  'self-harm/intent': 85,
  'self-harm/instructions': 65,
  sexual: 65,
  'sexual/minors': 65,
  violence: 95,
  'violence/graphic': 95,
}
const riskThresholdCategories = Object.keys(riskThresholdDefaults)

const { t } = useI18n()
const appStore = useAppStore()
const defaultBlockMessage = () => t('admin.riskControl.defaultBlockMessage')

const loading = ref(true)
const saving = ref(false)
const logsLoading = ref(false)
const statusLoading = ref(false)
const apiKeyTesting = ref(false)
const hashActionLoading = ref(false)
const unbanningUserID = ref<string | null>(null)
const settingsOpen = ref(false)
const activeSettingsTab = ref<SettingsTab>('basic')
const groupSearch = ref('')
const flaggedHashInput = ref('')
const groups = ref<AdminGroup[]>([])
const proxies = ref<Proxy[]>([])
const logs = ref<ContentModerationLog[]>([])
const status = ref<ContentModerationRuntimeStatus | null>(null)
const testedApiKeyStatuses = ref<ContentModerationAPIKeyStatus[]>([])
const pendingDeleteApiKeyHashes = ref<string[]>([])
const apiKeyRowsExpanded = ref<boolean>(false)
const moderationTestPrompt = ref('')
const moderationTestImages = ref<string[]>([])
const moderationTestResult = ref<ContentModerationTestAuditResult | null>(null)
const inputDetailRow = ref<ContentModerationLog | null>(null)
let statusTimer: number | null = null

const configForm = reactive({
  enabled: false,
  mode: 'pre_block' as ModerationMode,
  base_url: 'https://api.openai.com',
  model: 'omni-moderation-latest',
  proxy_id: null as string | null,
  api_keys_text: '',
  api_key_configured: false,
  api_key_masked: '',
  api_key_count: 0,
  api_key_masks: [] as string[],
  api_key_statuses: [] as ContentModerationAPIKeyStatus[],
  api_keys_mode: 'append' as APIKeysWriteMode,
  clear_api_key: false,
  timeout_ms: 3000,
  retry_count: 2,
  sample_rate: 100,
  all_groups: true,
  group_ids: [] as string[],
  record_non_hits: false,
  worker_count: 4,
  queue_size: 32768,
  block_status: 403,
  block_message: defaultBlockMessage(),
  email_on_hit: true,
  auto_ban_enabled: true,
  cyber_policy_exclude_from_ban_count: false,
  ban_threshold: 10,
  violation_window_hours: 720,
  hit_retention_days: 180,
  non_hit_retention_days: 3,
  pre_hash_check_enabled: false,
  thresholds: { ...riskThresholdDefaults } as Record<string, number>,
  blocked_keywords_text: '',
  keyword_blocking_mode: 'keyword_and_api' as KeywordBlockingMode,
  model_filter_type: 'all' as ContentModerationModelFilterType,
  model_filter_models: [] as string[],
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 1,
})

const filters = reactive({
  result: '',
  group_id: '',
  endpoint: '',
  search: '',
  from: '',
  to: '',
})

const settingsTabs = computed<Array<{ id: SettingsTab; label: string }>>(() => [
  { id: 'basic', label: t('admin.riskControl.tabs.basic') },
  { id: 'scope', label: t('admin.riskControl.tabs.scope') },
  { id: 'runtime', label: t('admin.riskControl.tabs.runtime') },
  { id: 'response', label: t('admin.riskControl.tabs.response') },
  { id: 'riskThresholds', label: t('admin.riskControl.tabs.riskThresholds') },
  { id: 'keywords', label: t('admin.riskControl.tabs.keywords') },
  { id: 'retention', label: t('admin.riskControl.tabs.retention') },
])

const modeOptions = computed<SelectOption[]>(() => [
  { value: 'pre_block', label: t('admin.riskControl.modePreBlock') },
  { value: 'observe', label: t('admin.riskControl.modeObserve') },
  { value: 'off', label: t('admin.riskControl.modeOff') },
])

const keywordBlockingModeOptions = computed<Array<{ value: KeywordBlockingMode; label: string; description: string }>>(() => [
  {
    value: 'keyword_and_api',
    label: t('admin.riskControl.keywordModeKeywordAndApi'),
    description: t('admin.riskControl.keywordModeKeywordAndApiDesc'),
  },
  {
    value: 'keyword_only',
    label: t('admin.riskControl.keywordModeKeywordOnly'),
    description: t('admin.riskControl.keywordModeKeywordOnlyDesc'),
  },
  {
    value: 'api_only',
    label: t('admin.riskControl.keywordModeApiOnly'),
    description: t('admin.riskControl.keywordModeApiOnlyDesc'),
  },
])

const modelFilterOptions = computed<Array<{ value: ContentModerationModelFilterType; label: string; description: string }>>(() => [
  {
    value: 'all',
    label: t('admin.riskControl.modelFilterAll'),
    description: t('admin.riskControl.modelFilterAllDesc'),
  },
  {
    value: 'include',
    label: t('admin.riskControl.modelFilterInclude'),
    description: t('admin.riskControl.modelFilterIncludeDesc'),
  },
  {
    value: 'exclude',
    label: t('admin.riskControl.modelFilterExclude'),
    description: t('admin.riskControl.modelFilterExcludeDesc'),
  },
])

type KeywordNoticeView = {
  title: string
  description: string
  icon: 'infoCircle' | 'exclamationTriangle'
  toneClass: string
  iconClass: string
  titleClass: string
}

const keywordNoticeTones = {
  info: {
    icon: 'infoCircle' as const,
    toneClass: 'views-admin-risk-control-view__state',
    iconClass: 'views-admin-risk-control-view__state-2',
    titleClass: 'views-admin-risk-control-view__state-3',
  },
  warning: {
    icon: 'exclamationTriangle' as const,
    toneClass: 'views-admin-risk-control-view__state-4',
    iconClass: 'views-admin-risk-control-view__state-5',
    titleClass: 'views-admin-risk-control-view__state-6',
  },
}

const keywordNotice = computed<KeywordNoticeView>(() => {
  const strategy = configForm.keyword_blocking_mode
  if (strategy === 'api_only') {
    return {
      ...keywordNoticeTones.info,
      title: t('admin.riskControl.keywordModeApiOnlyNotice'),
      description: t('admin.riskControl.keywordModeApiOnlyDesc'),
    }
  }
  if (configForm.mode !== 'pre_block') {
    return {
      ...keywordNoticeTones.warning,
      title: t('admin.riskControl.blockedKeywordsModeWarning', { mode: modeLabel(configForm.mode) }),
      description: t('admin.riskControl.blockedKeywordsDescription'),
    }
  }
  if (strategy === 'keyword_only') {
    return {
      ...keywordNoticeTones.info,
      title: t('admin.riskControl.keywordModeKeywordOnlyNotice'),
      description: t('admin.riskControl.keywordModeKeywordOnlyDesc'),
    }
  }
  return {
    ...keywordNoticeTones.info,
    title: t('admin.riskControl.blockedKeywordsPreBlockHint'),
    description: t('admin.riskControl.blockedKeywordsDescription'),
  }
})

const resultOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.riskControl.result.all') },
  { value: 'hit', label: t('admin.riskControl.result.hit') },
  { value: 'blocked', label: t('admin.riskControl.result.blocked') },
  { value: 'pass', label: t('admin.riskControl.result.pass') },
  { value: 'error', label: t('admin.riskControl.result.error') },
])

const endpointOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.riskControl.filters.allEndpoints') },
  { value: '/v1/messages', label: '/v1/messages' },
  { value: '/v1/responses', label: '/v1/responses' },
  { value: '/v1/chat/completions', label: '/v1/chat/completions' },
  { value: '/v1beta/models', label: '/v1beta/models' },
  { value: '/v1/images/generations', label: '/v1/images/generations' },
  { value: '/v1/images/edits', label: '/v1/images/edits' },
])

const groupFilterOptions = computed<SelectOption[]>(() => [
  { value: 0, label: t('admin.riskControl.filters.allGroups') },
  ...groups.value.map((group) => ({
    value: group.id,
    label: `${group.name} (${group.platform})`,
  })),
])

const selectedGroupCount = computed(() => String(configForm.group_ids.length))

const modelFilterModelCount = computed(() => configForm.model_filter_models.length)

const modelFilterSummary = computed(() => {
  if (configForm.model_filter_type === 'include') {
    return t('admin.riskControl.modelFilterIncludeSummary', { count: modelFilterModelCount.value })
  }
  if (configForm.model_filter_type === 'exclude') {
    return t('admin.riskControl.modelFilterExcludeSummary', { count: modelFilterModelCount.value })
  }
  return t('admin.riskControl.modelFilterAllSummary')
})

const modelFilterPreviewModels = computed(() => configForm.model_filter_models.slice(0, 6))

const hiddenModelFilterModelCount = computed(() => Math.max(0, configForm.model_filter_models.length - modelFilterPreviewModels.value.length))

const filteredGroups = computed(() => {
  const keyword = groupSearch.value.trim().toLowerCase()
  if (!keyword) return groups.value
  return groups.value.filter((group) => {
    return group.name.toLowerCase().includes(keyword) || String(group.platform).toLowerCase().includes(keyword)
  })
})

const inputApiKeyCount = computed(() => parseApiKeys(configForm.api_keys_text).length)

const blockedKeywordList = computed(() => parseBlockedKeywords(configForm.blocked_keywords_text))

const blockedKeywordCount = computed(() => blockedKeywordList.value.length)

const pendingDeletedApiKeyCount = computed(() => pendingDeleteApiKeyHashes.value.length)

const effectiveStoredApiKeyCount = computed(() => Math.max(0, configForm.api_key_count - pendingDeletedApiKeyCount.value))

const apiKeysPlaceholder = computed(() => (
  configForm.api_keys_mode === 'replace'
    ? t('admin.riskControl.apiKeysPlaceholderReplace')
    : t('admin.riskControl.apiKeysPlaceholder')
))

const apiKeysModeHint = computed(() => (
  configForm.api_keys_mode === 'replace'
    ? t('admin.riskControl.apiKeysModeReplaceHint')
    : t('admin.riskControl.apiKeysModeAppendHint')
))

const hasModerationAuditInput = computed(() => {
  return moderationTestPrompt.value.trim() !== '' || moderationTestImages.value.length > 0
})

const isFlaggedHashInputValid = computed(() => /^[a-fA-F0-9]{64}$/.test(flaggedHashInput.value.trim()))

const storedApiKeyTestButtonText = computed(() => {
  if (apiKeyTesting.value) return t('admin.riskControl.testingApiKeys')
  if (hasModerationAuditInput.value) return t('admin.riskControl.testContentWithStoredApiKey')
  return t('admin.riskControl.testStoredApiKeys')
})

const savedApiKeyRows = computed<ContentModerationAPIKeyStatus[]>(() => {
  const rows = status.value?.api_key_statuses?.length
    ? status.value.api_key_statuses
    : configForm.api_key_statuses
  return Array.isArray(rows) ? rows : []
})

const apiKeyRows = computed<ContentModerationAPIKeyStatus[]>(() => [
  ...savedApiKeyRows.value,
  ...testedApiKeyStatuses.value,
])

const visibleApiKeyRows = computed<ContentModerationAPIKeyStatus[]>(() => {
  if (apiKeyRowsExpanded.value) return apiKeyRows.value
  return apiKeyRows.value.slice(0, maxVisibleApiKeyRows)
})

const hiddenApiKeyRowCount = computed<number>(() => Math.max(0, apiKeyRows.value.length - visibleApiKeyRows.value.length))

const canToggleApiKeyRows = computed<boolean>(() => apiKeyRows.value.length > maxVisibleApiKeyRows)

const activeSavedApiKeyRows = computed<ContentModerationAPIKeyStatus[]>(() => (
  savedApiKeyRows.value.filter((row) => !isStoredApiKeyPendingDelete(row))
))

const apiKeyHealthBadges = computed<Array<{ status: ContentModerationAPIKeyStatus['status']; count: number }>>(() => {
  const counts: Record<ContentModerationAPIKeyStatus['status'], number> = {
    ok: 0,
    error: 0,
    frozen: 0,
    unknown: 0,
  }
  for (const row of activeSavedApiKeyRows.value) {
    counts[row.status] = (counts[row.status] ?? 0) + 1
  }
  if (activeSavedApiKeyRows.value.length === 0 && effectiveStoredApiKeyCount.value > 0) {
    counts.unknown = effectiveStoredApiKeyCount.value
  }
  return (['ok', 'frozen', 'error', 'unknown'] as Array<ContentModerationAPIKeyStatus['status']>)
    .map((item) => ({ status: item, count: counts[item] }))
    .filter((item) => item.count > 0)
})

const apiKeyHealthSummary = computed(() => {
  if (!configForm.api_key_configured) return ''
  if (apiKeyHealthBadges.value.length === 0) return t('admin.riskControl.apiKeyStatusUnknown')
  return apiKeyHealthBadges.value
    .map((badge) => `${apiKeyStatusLabel(badge.status)} ${badge.count}`)
    .join(' · ')
})

const overviewItems = computed<OverviewItem[]>(() => [
  {
    key: 'status',
    label: t('admin.riskControl.overview.status'),
    value: configForm.enabled ? t('admin.riskControl.overview.enabled') : t('admin.riskControl.overview.disabled'),
    meta: modeLabel(configForm.mode),
    icon: 'shield',
    iconClass: configForm.enabled
      ? 'views-admin-risk-control-view__state-7'
      : 'views-admin-risk-control-view__state-8',
    badge: runtimeBadgeText.value,
    badgeClass: runtimeBadgeClass.value,
  },
  {
    key: 'api-key',
    label: t('admin.riskControl.overview.apiKey'),
    value: configForm.api_key_configured ? t('admin.riskControl.apiKeyCount', { count: configForm.api_key_count }) : t('admin.riskControl.notConfigured'),
    meta: configForm.api_key_configured ? apiKeyHealthSummary.value || configForm.model || '-' : configForm.model || '-',
    icon: 'key',
    iconClass: 'views-admin-risk-control-view__state-9',
  },
  {
    key: 'scope',
    label: t('admin.riskControl.overview.groupScope'),
    value: configForm.all_groups ? t('admin.riskControl.allGroups') : selectedGroupCount.value,
    meta: modelFilterSummary.value,
    icon: 'users',
    iconClass: 'views-admin-risk-control-view__state-10',
  },
  {
    key: 'logs',
    label: t('admin.riskControl.overview.logs'),
    value: formatNumber(pagination.total),
    meta: t('admin.riskControl.overview.currentFilter'),
    icon: 'document',
    iconClass: 'views-admin-risk-control-view__state-11',
  },
])

const moderationScoreRows = computed<ModerationScoreRow[]>(() => {
  const result = moderationTestResult.value
  if (!result) return []
  return Object.entries(result.category_scores || {})
    .map(([category, score]) => {
      const threshold = result.thresholds?.[category] ?? 1
      return {
        category,
        score,
        threshold,
        hit: score >= threshold,
      }
    })
    .sort((a, b) => b.score - a.score)
})

const riskThresholdRows = computed<RiskThresholdRow[]>(() => (
  riskThresholdCategories.map((category) => ({
    category,
    value: configForm.thresholds[category] ?? riskThresholdDefaults[category],
    defaultValue: riskThresholdDefaults[category],
  }))
))

const inputDetailText = computed(() => {
  if (!inputDetailRow.value) return '-'
  return inputDetailRow.value.input_excerpt || inputDetailRow.value.error || '-'
})

const queueUsagePercent = computed(() => `${Math.min(100, Math.max(0, status.value?.queue_usage_percent ?? 0)).toFixed(1)}%`)

const queueUsageStyle = computed(() => ({
  width: queueUsagePercent.value,
}))

const runtimeMode = computed<ModerationMode>(() => status.value?.mode ?? configForm.mode)

const showPreBlockRuntimeCard = computed(() => runtimeMode.value === 'pre_block')

const showWorkerRuntimeCard = computed(() => runtimeMode.value === 'observe')

const preBlockMetricItems = computed(() => [
  {
    key: 'active',
    label: t('admin.riskControl.preBlockActive'),
    value: formatNumber(status.value?.pre_block_active ?? 0),
    meta: t('admin.riskControl.preBlockActiveHint'),
    class: 'views-admin-risk-control-view__render',
    valueClass: 'views-admin-risk-control-view__state-12',
  },
  {
    key: 'checked',
    label: t('admin.riskControl.preBlockChecked'),
    value: formatNumber(status.value?.pre_block_checked ?? 0),
    meta: t('admin.riskControl.preBlockCheckedHint'),
    class: 'views-admin-risk-control-view__render-2',
    valueClass: 'views-admin-risk-control-view__state-13',
  },
  {
    key: 'allowed',
    label: t('admin.riskControl.preBlockAllowed'),
    value: formatNumber(status.value?.pre_block_allowed ?? 0),
    meta: t('admin.riskControl.preBlockAllowedHint'),
    class: 'views-admin-risk-control-view__render-3',
    valueClass: 'views-admin-risk-control-view__state-14',
  },
  {
    key: 'blocked',
    label: t('admin.riskControl.preBlockBlocked'),
    value: formatNumber(status.value?.pre_block_blocked ?? 0),
    meta: t('admin.riskControl.preBlockBlockedHint'),
    class: 'views-admin-risk-control-view__render-4',
    valueClass: 'views-admin-risk-control-view__state-15',
  },
  {
    key: 'errors',
    label: t('admin.riskControl.preBlockErrors'),
    value: formatNumber(status.value?.pre_block_errors ?? 0),
    meta: t('admin.riskControl.preBlockErrorsHint'),
    class: 'views-admin-risk-control-view__render-5',
    valueClass: 'views-admin-risk-control-view__state-16',
  },
  {
    key: 'latency',
    label: t('admin.riskControl.preBlockAvgLatency'),
    value: `${formatNumber(status.value?.pre_block_avg_latency_ms ?? 0)} ms`,
    meta: t('admin.riskControl.preBlockAvgLatencyHint'),
    class: 'views-admin-risk-control-view__render-6',
    valueClass: 'views-admin-risk-control-view__state-17',
  },
])

const preBlockAPIKeyLoads = computed<ContentModerationAPIKeyLoad[]>(() => (
  [...(status.value?.pre_block_api_key_loads ?? [])].sort((a, b) => a.index - b.index)
))

const preBlockAPIKeyMaxTotal = computed(() => Math.max(1, ...preBlockAPIKeyLoads.value.map((item) => item.total || 0)))

const preBlockAPIKeyLoadSummaryText = computed(() => t('admin.riskControl.preBlockAPIKeyLoadSummary', {
  active: formatNumber(status.value?.pre_block_api_key_active ?? 0),
  available: formatNumber(status.value?.pre_block_api_key_available_count ?? 0),
  total: formatNumber(status.value?.pre_block_api_key_total_calls ?? 0),
  workerActive: formatNumber(status.value?.active_workers ?? 0),
  workerTotal: formatNumber(status.value?.worker_count ?? configForm.worker_count),
}))

function preBlockAPIKeyLoadWidth(total: number): string {
  return `${Math.min(100, Math.max(0, (total / preBlockAPIKeyMaxTotal.value) * 100)).toFixed(1)}%`
}

const workerSlots = computed(() => {
  const total = Math.max(0, status.value?.worker_count ?? configForm.worker_count)
  const active = Math.max(0, status.value?.active_workers ?? 0)
  const enabled = Boolean(status.value?.risk_control_enabled && status.value?.enabled && status.value?.mode !== 'off')
  return Array.from({ length: total }, (_, index) => ({
    id: index + 1,
    state: (!enabled ? 'disabled' : index < active ? 'active' : 'idle') as WorkerSlotState,
    label: !enabled
      ? t('admin.riskControl.workerDisabled')
      : index < active
        ? t('admin.riskControl.workerActive')
        : t('admin.riskControl.workerIdle'),
  }))
})

const runtimeBadgeText = computed(() => {
  if (!status.value?.risk_control_enabled) return t('admin.riskControl.riskSwitchOff')
  if (!configForm.enabled || configForm.mode === 'off') return t('admin.riskControl.overview.disabled')
  return t('admin.riskControl.overview.enabled')
})

const runtimeBadgeClass = computed(() => {
  if (!status.value?.risk_control_enabled || !configForm.enabled || configForm.mode === 'off') {
    return 'views-admin-risk-control-view__state-18'
  }
  return 'views-admin-risk-control-view__text-44'
})

function applyConfig(config: ContentModerationConfig) {
  configForm.enabled = config.enabled
  configForm.mode = config.mode
  configForm.base_url = config.base_url || 'https://api.openai.com'
  configForm.model = config.model || 'omni-moderation-latest'
  configForm.proxy_id = config.proxy_id || null
  configForm.api_keys_text = ''
  configForm.api_key_configured = config.api_key_configured
  configForm.api_key_masked = config.api_key_masked || ''
  configForm.api_key_count = config.api_key_count || 0
  configForm.api_key_masks = Array.isArray(config.api_key_masks) ? [...config.api_key_masks] : []
  configForm.api_key_statuses = Array.isArray(config.api_key_statuses) ? [...config.api_key_statuses] : []
  configForm.api_keys_mode = 'append'
  configForm.clear_api_key = false
  pendingDeleteApiKeyHashes.value = []
  testedApiKeyStatuses.value = []
  apiKeyRowsExpanded.value = false
  configForm.timeout_ms = config.timeout_ms || 3000
  configForm.retry_count = config.retry_count ?? 2
  configForm.sample_rate = config.sample_rate ?? 100
  configForm.all_groups = config.all_groups
  configForm.group_ids = Array.isArray(config.group_ids) ? [...config.group_ids] : []
  configForm.record_non_hits = config.record_non_hits
  configForm.worker_count = config.worker_count || 4
  configForm.queue_size = config.queue_size || 32768
  configForm.block_status = config.block_status || 403
  configForm.block_message = config.block_message || defaultBlockMessage()
  configForm.email_on_hit = config.email_on_hit ?? true
  configForm.auto_ban_enabled = config.auto_ban_enabled ?? true
  configForm.cyber_policy_exclude_from_ban_count = config.cyber_policy_exclude_from_ban_count ?? false
  configForm.ban_threshold = config.ban_threshold || 10
  configForm.violation_window_hours = config.violation_window_hours || 720
  configForm.hit_retention_days = config.hit_retention_days || 180
  configForm.non_hit_retention_days = Math.min(Math.max(config.non_hit_retention_days || 3, 1), 3)
  configForm.pre_hash_check_enabled = config.pre_hash_check_enabled ?? false
  configForm.thresholds = riskThresholdsFromConfig(config.thresholds)
  configForm.blocked_keywords_text = Array.isArray(config.blocked_keywords) ? config.blocked_keywords.join('\n') : ''
  configForm.keyword_blocking_mode = normalizeKeywordBlockingMode(config.keyword_blocking_mode)
  const modelFilter = normalizeModelFilter(config.model_filter)
  configForm.model_filter_type = modelFilter.type
  configForm.model_filter_models = modelFilter.models
}

async function loadAll() {
  loading.value = true
  try {
    const [config, groupItems, runtimeStatus, proxyItems] = await Promise.all([
      adminAPI.riskControl.getConfig(),
      adminAPI.groups.getAll(),
      adminAPI.riskControl.getStatus(),
      // 代理列表加载失败不阻塞风控页面（仅影响下拉可选项）
      adminAPI.proxies.getAll().catch(() => [] as Proxy[]),
    ])
    applyConfig(config)
    groups.value = groupItems
    status.value = runtimeStatus
    proxies.value = proxyItems
    if (Array.isArray(runtimeStatus.api_key_statuses)) {
      configForm.api_key_statuses = [...runtimeStatus.api_key_statuses]
      prunePendingDeleteAPIKeyHashes()
    }
    await loadLogs()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('admin.riskControl.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function loadStatus(silent = true) {
  statusLoading.value = true
  try {
    const runtimeStatus = await adminAPI.riskControl.getStatus()
    status.value = runtimeStatus
    if (Array.isArray(runtimeStatus.api_key_statuses)) {
      configForm.api_key_statuses = [...runtimeStatus.api_key_statuses]
      prunePendingDeleteAPIKeyHashes()
    }
  } catch (err: unknown) {
    if (!silent) {
      appStore.showError(extractApiErrorMessage(err, t('admin.riskControl.statusFailed')))
    }
  } finally {
    statusLoading.value = false
  }
}

async function saveConfig() {
  saving.value = true
  try {
    const modelFilterPayload = buildModelFilterPayload()
    if (modelFilterPayload.type !== 'all' && modelFilterPayload.models.length === 0) {
      appStore.showError(t('admin.riskControl.modelFilterModelsRequired'))
      return
    }
    const payload: UpdateContentModerationConfig = {
      enabled: configForm.enabled,
      mode: configForm.mode,
      base_url: configForm.base_url,
      model: configForm.model,
      proxy_id: configForm.proxy_id,
      timeout_ms: Number(configForm.timeout_ms) || 3000,
      retry_count: Number(configForm.retry_count) || 0,
      sample_rate: Number(configForm.sample_rate) || 0,
      all_groups: configForm.all_groups,
      group_ids: configForm.all_groups ? [] : [...configForm.group_ids],
      record_non_hits: configForm.record_non_hits,
      clear_api_key: configForm.clear_api_key,
      worker_count: Number(configForm.worker_count) || 4,
      queue_size: Number(configForm.queue_size) || 32768,
      block_status: Number(configForm.block_status) || 403,
      block_message: configForm.block_message || defaultBlockMessage(),
      email_on_hit: configForm.email_on_hit,
      auto_ban_enabled: configForm.auto_ban_enabled,
      cyber_policy_exclude_from_ban_count: configForm.cyber_policy_exclude_from_ban_count,
      ban_threshold: Number(configForm.ban_threshold) || 10,
      violation_window_hours: Number(configForm.violation_window_hours) || 720,
      hit_retention_days: Number(configForm.hit_retention_days) || 180,
      non_hit_retention_days: Math.min(Math.max(Number(configForm.non_hit_retention_days) || 3, 1), 3),
      pre_hash_check_enabled: configForm.pre_hash_check_enabled,
      thresholds: buildRiskThresholdPayload(),
      blocked_keywords: blockedKeywordList.value,
      keyword_blocking_mode: configForm.keyword_blocking_mode,
      model_filter: modelFilterPayload,
    }
    const keys = parseApiKeys(configForm.api_keys_text)
    if (!payload.clear_api_key && configForm.api_keys_mode === 'replace' && keys.length === 0) {
      appStore.showError(t('admin.riskControl.apiKeysReplaceNoInput'))
      return
    }
    if (keys.length > 0) {
      payload.api_keys = keys
      payload.api_keys_mode = configForm.api_keys_mode
      payload.clear_api_key = false
    }
    if (!payload.clear_api_key && configForm.api_keys_mode !== 'replace' && pendingDeleteApiKeyHashes.value.length > 0) {
      payload.delete_api_key_hashes = [...pendingDeleteApiKeyHashes.value]
    }

    const updated = await adminAPI.riskControl.updateConfig(payload)
    applyConfig(updated)
    settingsOpen.value = false
    appStore.showSuccess(t('admin.riskControl.saved'))
    await Promise.all([loadStatus(true), loadLogs()])
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('admin.riskControl.saveFailed')))
  } finally {
    saving.value = false
  }
}

async function loadLogs() {
  logsLoading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size,
      result: filters.result || undefined,
      group_id: filters.group_id || undefined,
      endpoint: filters.endpoint || undefined,
      search: filters.search || undefined,
      from: normalizeDateTimeLocal(filters.from),
      to: normalizeDateTimeLocal(filters.to),
    }
    const result = await adminAPI.riskControl.listLogs(params)
    logs.value = result.items
    pagination.total = result.total
    pagination.page = result.page
    pagination.page_size = result.page_size
    pagination.pages = result.pages
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('admin.riskControl.logsFailed')))
  } finally {
    logsLoading.value = false
  }
}

function canUnbanRow(row: ContentModerationLog): boolean {
  return Boolean(row.auto_banned && row.user_id && row.user_status === 'disabled')
}

function inputSummaryText(row: ContentModerationLog): string {
  return row.input_excerpt || row.error || '-'
}

function openInputDetail(row: ContentModerationLog) {
  inputDetailRow.value = row
}

function closeInputDetail() {
  inputDetailRow.value = null
}

async function unbanUser(row: ContentModerationLog) {
  if (!row.user_id || unbanningUserID.value !== null) return
  unbanningUserID.value = row.user_id
  try {
    const result = await adminAPI.riskControl.unbanUser(row.user_id)
    logs.value = logs.value.map((item) => {
      if (item.user_id !== row.user_id) return item
      return { ...item, user_status: result.status }
    })
    appStore.showSuccess(t('admin.riskControl.unbanSuccess'))
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('admin.riskControl.unbanFailed')))
  } finally {
    unbanningUserID.value = null
  }
}

async function deleteFlaggedHash() {
  if (!isFlaggedHashInputValid.value || hashActionLoading.value) return
  hashActionLoading.value = true
  try {
    const result = await adminAPI.riskControl.deleteFlaggedHash(flaggedHashInput.value)
    flaggedHashInput.value = ''
    await loadStatus(true)
    appStore.showSuccess(result.deleted ? t('admin.riskControl.flaggedHashDeleted') : t('admin.riskControl.flaggedHashNotFound'))
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('admin.riskControl.flaggedHashDeleteFailed')))
  } finally {
    hashActionLoading.value = false
  }
}

async function clearFlaggedHashes() {
  if (hashActionLoading.value) return
  const confirmed = window.confirm(t('admin.riskControl.clearFlaggedHashesConfirm'))
  if (!confirmed) return
  hashActionLoading.value = true
  try {
    const result = await adminAPI.riskControl.clearFlaggedHashes()
    await loadStatus(true)
    appStore.showSuccess(t('admin.riskControl.flaggedHashesCleared', { count: result.deleted }))
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('admin.riskControl.flaggedHashesClearFailed')))
  } finally {
    hashActionLoading.value = false
  }
}

function openSettings() {
  activeSettingsTab.value = 'basic'
  settingsOpen.value = true
}

function reloadLogsFromFirstPage() {
  pagination.page = 1
  void loadLogs()
}

function onPageChange(page: number) {
  pagination.page = page
  void loadLogs()
}

function onPageSizeChange(pageSize: number) {
  pagination.page = 1
  pagination.page_size = pageSize
  void loadLogs()
}

function toggleClearApiKey() {
  configForm.clear_api_key = !configForm.clear_api_key
  if (configForm.clear_api_key) {
    configForm.api_keys_text = ''
    configForm.api_keys_mode = 'append'
    testedApiKeyStatuses.value = []
    pendingDeleteApiKeyHashes.value = []
  }
}

function setAPIKeysMode(mode: APIKeysWriteMode) {
  configForm.api_keys_mode = mode
  if (mode === 'replace') {
    pendingDeleteApiKeyHashes.value = []
  }
}

function setModelFilterType(type: ContentModerationModelFilterType) {
  configForm.model_filter_type = type
  if (type === 'all') {
    configForm.model_filter_models = []
  }
}

async function testApiKeys(useInputKeys: boolean) {
  const keys = useInputKeys ? parseApiKeys(configForm.api_keys_text) : []
  if (useInputKeys && keys.length === 0) {
    appStore.showError(t('admin.riskControl.apiKeyTestNoInput'))
    return
  }
  apiKeyTesting.value = true
  try {
    const result = await adminAPI.riskControl.testAPIKeys({
      api_keys: keys,
      base_url: configForm.base_url,
      model: configForm.model,
      timeout_ms: Number(configForm.timeout_ms) || 3000,
      proxy_id: configForm.proxy_id,
      prompt: moderationTestPrompt.value,
      images: moderationTestImages.value,
    })
    moderationTestResult.value = result.audit_result ?? null
    if (useInputKeys) {
      testedApiKeyStatuses.value = result.items.map((item) => ({ ...item, configured: false }))
    } else {
      mergeConfiguredAPIKeyStatuses(result.items)
      testedApiKeyStatuses.value = []
      await loadStatus(true)
    }
    appStore.showSuccess(t('admin.riskControl.apiKeyTestDone', { count: result.items.length }))
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('admin.riskControl.apiKeyTestFailed')))
  } finally {
    apiKeyTesting.value = false
  }
}

function mergeConfiguredAPIKeyStatuses(items: ContentModerationAPIKeyStatus[]) {
  if (!hasModerationAuditInput.value || configForm.api_key_statuses.length === 0) {
    configForm.api_key_statuses = items
    return
  }
  const updates = new Map(items.map((item) => [item.key_hash, item]))
  configForm.api_key_statuses = configForm.api_key_statuses.map((item) => updates.get(item.key_hash) ?? item)
}

function toggleDeleteStoredApiKey(row: ContentModerationAPIKeyStatus) {
  if (!row.configured || !row.key_hash) return
  const index = pendingDeleteApiKeyHashes.value.indexOf(row.key_hash)
  if (index >= 0) {
    pendingDeleteApiKeyHashes.value.splice(index, 1)
    return
  }
  pendingDeleteApiKeyHashes.value.push(row.key_hash)
}

function isStoredApiKeyPendingDelete(row: ContentModerationAPIKeyStatus): boolean {
  return row.configured && row.key_hash !== '' && pendingDeleteApiKeyHashes.value.includes(row.key_hash)
}

function prunePendingDeleteAPIKeyHashes() {
  const currentHashes = new Set(savedApiKeyRows.value.map((row) => row.key_hash).filter(Boolean))
  pendingDeleteApiKeyHashes.value = pendingDeleteApiKeyHashes.value.filter((hash) => currentHashes.has(hash))
}

function clearModerationTestInput() {
  moderationTestPrompt.value = ''
  moderationTestImages.value = []
  moderationTestResult.value = null
}

function removeModerationTestImage(index: number) {
  moderationTestImages.value.splice(index, 1)
}

async function handleModerationImageUpload(event: Event) {
  const input = event.target as HTMLInputElement
  await addModerationTestFiles(input.files)
  input.value = ''
}

async function handleModerationImageDrop(event: DragEvent) {
  await addModerationTestFiles(event.dataTransfer?.files ?? null)
}

async function handleModerationImagePaste(event: ClipboardEvent) {
  const files = Array.from(event.clipboardData?.files ?? []).filter((file) => file.type.startsWith('image/'))
  if (files.length === 0) return
  event.preventDefault()
  await addModerationTestFiles(files)
}

async function addModerationTestFiles(files: FileList | File[] | null) {
  if (!files) return
  const items = Array.from(files).filter((file) => file.type.startsWith('image/'))
  for (const file of items) {
    if (moderationTestImages.value.length >= maxModerationTestImages) {
      appStore.showError(t('admin.riskControl.auditTestImageLimit', { count: maxModerationTestImages }))
      return
    }
    if (file.size > maxModerationTestImageSize) {
      appStore.showError(t('admin.riskControl.auditTestImageTooLarge'))
      continue
    }
    try {
      moderationTestImages.value.push(await fileToDataURL(file))
    } catch {
      appStore.showError(t('admin.riskControl.auditTestImageReadFailed'))
    }
  }
}

function fileToDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(file)
  })
}

function toggleGroup(groupID: string) {
  const index = configForm.group_ids.indexOf(groupID)
  if (index >= 0) {
    configForm.group_ids.splice(index, 1)
  } else {
    configForm.group_ids.push(groupID)
  }
}

function isGroupSelected(groupID: string): boolean {
  return configForm.group_ids.includes(groupID)
}

function modeLabel(mode: ModerationMode): string {
  const found = modeOptions.value.find((option) => option.value === mode)
  return found?.label ?? mode
}

function modeDescription(mode: ModerationMode): string {
  const descriptions: Record<ModerationMode, string> = {
    pre_block: t('admin.riskControl.modePreBlockDesc'),
    observe: t('admin.riskControl.modeObserveDesc'),
    off: t('admin.riskControl.modeOffDesc'),
  }
  return descriptions[mode] ?? ''
}

function resultLabel(row: ContentModerationLog): string {
  if (row.action === 'cyber_policy') return t('admin.riskControl.action.cyberPolicy')
  if (row.action === 'keyword_block') return t('admin.riskControl.action.keywordBlock')
  if (row.action === 'block') return t('admin.riskControl.action.block')
  if (row.action === 'error' || row.error) return t('admin.riskControl.action.error')
  if (row.flagged) return t('admin.riskControl.result.hit')
  return t('admin.riskControl.result.pass')
}

function resultBadgeClass(row: ContentModerationLog): string {
  if (row.action === 'block' || row.action === 'keyword_block' || row.action === 'cyber_policy') return 'views-admin-risk-control-view__state-19'
  if (row.action === 'error' || row.error) return 'views-admin-risk-control-view__state-20'
  if (row.flagged) return 'views-admin-risk-control-view__state-21'
  return 'views-admin-risk-control-view__state-22'
}

function workerSlotClass(state: WorkerSlotState): string {
  if (state === 'active') {
    return 'views-admin-risk-control-view__state-23'
  }
  if (state === 'idle') {
    return 'views-admin-risk-control-view__state-24'
  }
  return 'views-admin-risk-control-view__state-25'
}

function workerDotClass(state: WorkerSlotState): string {
  if (state === 'active') return 'status-fill--info'
  if (state === 'idle') return 'status-fill--success'
  return 'views-admin-risk-control-view__state-26'
}

function percent(value: number): string {
  if (!Number.isFinite(value)) return '-'
  return `${(value * 100).toFixed(1)}%`
}

function percentWidth(value: number): string {
  if (!Number.isFinite(value)) return '0%'
  return `${Math.min(100, Math.max(0, value * 100)).toFixed(1)}%`
}

function latencyText(value: number | null): string {
  if (value === null || value === undefined) return '-'
  return `${value} ms`
}

function apiKeyRowKey(row: ContentModerationAPIKeyStatus, index: number): string {
  return `${row.configured ? 'saved' : 'test'}-${row.key_hash || index}`
}

function apiKeyStatusLabel(statusValue: ContentModerationAPIKeyStatus['status']): string {
  const labels: Record<ContentModerationAPIKeyStatus['status'], string> = {
    ok: t('admin.riskControl.apiKeyStatusOk'),
    error: t('admin.riskControl.apiKeyStatusError'),
    frozen: t('admin.riskControl.apiKeyStatusFrozen'),
    unknown: t('admin.riskControl.apiKeyStatusUnknown'),
  }
  return labels[statusValue] ?? labels.unknown
}

function apiKeyStatusBadgeClass(statusValue: ContentModerationAPIKeyStatus['status']): string {
  const classes: Record<ContentModerationAPIKeyStatus['status'], string> = {
    ok: 'views-admin-risk-control-view__text-44',
    error: 'views-admin-risk-control-view__state-27',
    frozen: 'views-admin-risk-control-view__text-43',
    unknown: 'views-admin-risk-control-view__state-18',
  }
  return classes[statusValue] ?? classes.unknown
}

function apiKeyStatusDotClass(statusValue: ContentModerationAPIKeyStatus['status']): string {
  const classes: Record<ContentModerationAPIKeyStatus['status'], string> = {
    ok: 'status-fill--success',
    error: 'status-fill--warning',
    frozen: 'status-fill--danger',
    unknown: 'status-fill--neutral',
  }
  return classes[statusValue] ?? classes.unknown
}

function apiKeyStatusMeta(row: ContentModerationAPIKeyStatus): string {
  const parts: string[] = []
  parts.push(t('admin.riskControl.apiKeyFailureCount', { count: row.failure_count || 0 }))
  if (row.last_latency_ms > 0) {
    parts.push(t('admin.riskControl.apiKeyLatency', { ms: row.last_latency_ms }))
  }
  if (row.last_http_status > 0) {
    parts.push(t('admin.riskControl.apiKeyHTTPStatus', { status: row.last_http_status }))
  }
  if (row.frozen_until) {
    parts.push(t('admin.riskControl.apiKeyFrozenUntil', { time: formatDateTime(row.frozen_until) }))
  } else if (row.last_checked_at) {
    parts.push(t('admin.riskControl.apiKeyLastChecked', { time: formatDateTime(row.last_checked_at) }))
  } else {
    parts.push(t('admin.riskControl.apiKeyNotTested'))
  }
  return parts.join(' / ')
}

function parseApiKeys(value: string): string[] {
  return value
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter((item, index, arr) => item && arr.indexOf(item) === index)
}

function normalizeKeywordBlockingMode(value: unknown): KeywordBlockingMode {
  if (value === 'keyword_only' || value === 'api_only' || value === 'keyword_and_api') {
    return value
  }
  return 'keyword_and_api'
}

function normalizeModelFilter(value: unknown): ContentModerationModelFilter {
  if (!value || typeof value !== 'object') {
    return { type: 'all', models: [] }
  }
  const raw = value as Partial<ContentModerationModelFilter>
  const type = normalizeModelFilterType(raw.type)
  const models = type === 'all' ? [] : normalizeModelNames(raw.models)
  return { type, models }
}

function normalizeModelFilterType(value: unknown): ContentModerationModelFilterType {
  if (value === 'include' || value === 'exclude' || value === 'all') {
    return value
  }
  return 'all'
}

function normalizeModelNames(models: unknown): string[] {
  if (!Array.isArray(models)) return []
  const seen = new Set<string>()
  const out: string[] = []
  for (const item of models) {
    const model = String(item ?? '').trim()
    if (!model) continue
    const key = model.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(model)
  }
  return out
}

function buildModelFilterPayload(): ContentModerationModelFilter {
  const type = normalizeModelFilterType(configForm.model_filter_type)
  if (type === 'all') {
    return { type: 'all', models: [] }
  }
  return {
    type,
    models: normalizeModelNames(configForm.model_filter_models),
  }
}

function riskThresholdsFromConfig(thresholds: Record<string, number> | null | undefined): Record<string, number> {
  const out: Record<string, number> = { ...riskThresholdDefaults }
  for (const category of riskThresholdCategories) {
    const value = thresholds?.[category]
    if (Number.isFinite(value)) {
      out[category] = clampPercent(Number(value) * 100)
    }
  }
  return out
}

function buildRiskThresholdPayload(): Record<string, number> {
  const payload: Record<string, number> = {}
  for (const category of riskThresholdCategories) {
    payload[category] = Number((clampPercent(configForm.thresholds[category]) / 100).toFixed(4))
  }
  return payload
}

function resetRiskThresholds() {
  configForm.thresholds = { ...riskThresholdDefaults }
}

function clampPercent(value: unknown): number {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) {
    return 0
  }
  return Math.min(100, Math.max(0, numeric))
}

function formatThresholdPercent(value: number): string {
  return `${clampPercent(value).toFixed(1)}%`
}

function parseBlockedKeywords(value: string): string[] {
  const seen = new Set<string>()
  const out: string[] = []
  for (const line of value.split(/\r?\n/)) {
    const kw = line.trim()
    if (!kw) continue
    const key = kw.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(kw)
  }
  return out
}

function violationCountText(row: ContentModerationLog): string {
  if (!row.flagged) return '-'
  if (row.violation_count === 0) return t('admin.riskControl.violationNotCounted')
  return t('admin.riskControl.violationCount', { count: row.violation_count || 1 })
}

function normalizeDateTimeLocal(value: string): string | undefined {
  if (!value) return undefined
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return undefined
  return date.toISOString()
}

function formatDateTime(value: string): string {
  return formatDateTimeValue(value) || '-'
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat().format(value)
}

onMounted(() => {
  void loadAll()
  statusTimer = window.setInterval(() => {
    void loadStatus(true)
  }, 15000)
})

onUnmounted(() => {
  if (statusTimer !== null) {
    window.clearInterval(statusTimer)
    statusTimer = null
  }
})
</script>
