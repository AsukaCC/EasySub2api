<template>
  <AppLayout>
    <TablePageLayout>
      <!-- 筛选与操作合并为一条标准工具条卡:左筛选、右操作 -->
      <template #filters>
        <div class="keys-toolbar card filter-toolbar">
          <div class="keys-toolbar__filters">
            <SearchInput
              v-model="filterSearch"
              :placeholder="t('keys.searchPlaceholder')"
              class="views-user-keys-view__search-input"
              @search="onFilterChange"
            />
            <Select
              :model-value="filterGroupId"
              class="views-user-keys-view__field"
              :options="groupFilterOptions"
              @update:model-value="onGroupFilterChange"
            />
            <Select
              :model-value="filterStatus"
              class="views-user-keys-view__field"
              :options="statusFilterOptions"
              @update:model-value="onStatusFilterChange"
            />
            <EndpointPopover
              v-if="publicSettings?.api_base_url || (publicSettings?.custom_endpoints?.length ?? 0) > 0"
              :api-base-url="publicSettings?.api_base_url || ''"
              :custom-endpoints="publicSettings?.custom_endpoints || []"
            />
          </div>
          <div class="keys-toolbar__actions">
          <button
            @click="loadApiKeys"
            :disabled="loading"
            class="btn btn-secondary"
            :title="t('common.refresh')"
          >
            <Icon name="refresh" size="md" :class="loading ? 'views-user-keys-view__icon-7' : ''" />
          </button>
          <div class="views-user-keys-view__panel-4 filter-toolbar" ref="columnDropdownRef">
            <button
              @click="showColumnDropdown = !showColumnDropdown"
              class="views-user-keys-view__action btn btn-secondary"
              :title="t('keys.columnSettings')"
            >
              <svg class="views-user-keys-view__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 4.5v15m6-15v15m-10.875 0h15.75c.621 0 1.125-.504 1.125-1.125V5.625c0-.621-.504-1.125-1.125-1.125H4.125C3.504 4.5 3 5.004 3 5.625v12.75c0 .621.504 1.125 1.125 1.125z" />
              </svg>
              <span class="views-user-keys-view__text">{{ t('keys.columnSettings') }}</span>
            </button>
            <div
              v-if="showColumnDropdown"
              class="views-user-keys-view__panel-5 dropdown dropdown--menu"
            >
              <button
                v-for="col in toggleableColumns"
                :key="col.key"
                @click="toggleColumn(col.key)"
                class="views-user-keys-view__action-2 dropdown-item"
              >
                <span>{{ col.label }}</span>
                <Icon
                  v-if="isColumnVisible(col.key)"
                  name="check"
                  size="sm"
                  class="views-user-keys-view__icon-2"
                  :stroke-width="2"
                />
              </button>
            </div>
          </div>
          <button @click="showCreateModal = true" class="btn btn-primary" data-tour="keys-create-btn">
            <Icon name="plus" size="md" class="views-user-keys-view__icon-3" />
            {{ t('keys.createKey') }}
          </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="apiKeys"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="created_at"
          default-sort-order="desc"
          @sort="handleSort"
        >
          <template #cell-id="{ value }">
            <span class="views-user-keys-view__text-2">#{{ value }}</span>
          </template>

          <template #cell-key="{ value, row }">
            <div class="views-user-keys-view__panel-6">
              <code class="views-user-keys-view__code code">
                {{ maskApiKey(value) }}
              </code>
              <button
                @click="copyToClipboard(value, row.id)"
                class="views-user-keys-view__action-3"
                :class="
                  copiedKeyId === row.id
                    ? 'views-user-keys-view__action-12'
                    : 'views-user-keys-view__action-13'
                "
                :title="copiedKeyId === row.id ? t('keys.copied') : t('keys.copyToClipboard')"
              >
                <Icon
                  v-if="copiedKeyId === row.id"
                  name="check"
                  size="sm"
                  :stroke-width="2"
                />
                <Icon v-else name="clipboard" size="sm" />
              </button>
            </div>
          </template>

          <template #cell-name="{ value, row }">
            <div class="views-user-keys-view__panel-7">
              <span class="views-user-keys-view__text-3">{{ value }}</span>
              <Icon
                v-if="row.ip_whitelist?.length > 0 || row.ip_blacklist?.length > 0"
                name="shield"
                size="sm"
                class="views-user-keys-view__icon-4"
                :title="t('keys.ipRestrictionEnabled')"
              />
            </div>
          </template>

          <template #cell-group="{ row }">
            <div
              v-if="keyGroupOptions(row).length > 0"
              class="key-groups"
              :title="keyGroupOptions(row).map((group) => group.name).join(', ')"
            >
              <GroupBadge
                v-for="group in keyGroupOptions(row)"
                :key="group.id"
                :name="group.name"
                :platform="group.platform"
                :subscription-type="group.subscription_type"
                :rate-multiplier="group.rate_multiplier"
                :user-rate-multiplier="userGroupRates[group.id] ?? null"
                :effective-rate-multiplier="group.effective_rate_multiplier ?? null"
                :always-show-rate="true"
              />
            </div>
            <span v-else class="views-user-keys-view__text-6">{{ t('keys.noGroup') }}</span>
          </template>

          <template #cell-current_concurrency="{ value }">
            <span
              :class="[
                'views-user-keys-view__text-11',
                (value ?? 0) > 0
                  ? 'views-user-keys-view__text-12'
                  : 'views-user-keys-view__text-13'
              ]"
            >
              {{ value ?? 0 }}
            </span>
          </template>

          <template #cell-usage="{ row }">
            <div class="views-user-keys-view__panel-8">
              <div class="views-user-keys-view__panel-7">
                <span class="views-user-keys-view__text-7">{{ t('keys.today') }}:</span>
                <span class="views-user-keys-view__text-3">
                  {{ formatPoints(usageStats[row.id]?.today_actual_cost ?? 0) }}
                </span>
              </div>
              <div class="views-user-keys-view__panel-9">
                <span class="views-user-keys-view__text-7">{{ t('keys.total') }}:</span>
                <span class="views-user-keys-view__text-3">
                  {{ formatPoints(usageStats[row.id]?.total_actual_cost ?? 0) }}
                </span>
              </div>
              <!-- Quota progress (if quota is set) -->
              <div v-if="row.quota > 0" class="views-user-keys-view__panel-10">
                <div class="views-user-keys-view__panel-7">
                  <span class="views-user-keys-view__text-7">{{ t('keys.quota') }}:</span>
                  <span :class="[
                    'views-user-keys-view__text-14',
                    row.quota_used >= row.quota ? 'views-user-keys-view__text-15' :
                    row.quota_used >= row.quota * 0.8 ? 'views-user-keys-view__text-16' :
                    'views-user-keys-view__text-17'
                  ]">
                    {{ formatPointRange(row.quota_used, row.quota) }}
                  </span>
                </div>
                <div class="views-user-keys-view__panel-11">
                  <div
                    :class="[
                      'views-user-keys-view__panel-31',
                      row.quota_used >= row.quota ? 'views-user-keys-view__panel-32' :
                      row.quota_used >= row.quota * 0.8 ? 'views-user-keys-view__panel-33' :
                      'views-user-keys-view__panel-34'
                    ]"
                    :style="{ width: Math.min((row.quota_used / row.quota) * 100, 100) + '%' }"
                  />
                </div>
              </div>
            </div>
          </template>

          <template #cell-rate_limit="{ row }">
            <div v-if="row.rate_limit_5h > 0 || row.rate_limit_1d > 0 || row.rate_limit_7d > 0" class="views-user-keys-view__panel-12">
              <!-- 5h window -->
              <div v-if="row.rate_limit_5h > 0">
                <div class="views-user-keys-view__panel-13">
                  <span class="views-user-keys-view__text-7">5h</span>
                  <span :class="[
                    'views-user-keys-view__text-18',
                    row.usage_5h >= row.rate_limit_5h ? 'views-user-keys-view__text-15' :
                    row.usage_5h >= row.rate_limit_5h * 0.8 ? 'views-user-keys-view__text-16' :
                    'views-user-keys-view__text-19'
                  ]">
                    {{ formatPointRange(row.usage_5h, row.rate_limit_5h) }}
                  </span>
                </div>
                <div class="views-user-keys-view__panel-14">
                  <div
                    :class="[
                      'views-user-keys-view__panel-31',
                      row.usage_5h >= row.rate_limit_5h ? 'views-user-keys-view__panel-32' :
                      row.usage_5h >= row.rate_limit_5h * 0.8 ? 'views-user-keys-view__panel-33' :
                      'views-user-keys-view__panel-35'
                    ]"
                    :style="{ width: Math.min((row.usage_5h / row.rate_limit_5h) * 100, 100) + '%' }"
                  />
                </div>
                <div v-if="row.reset_5h_at && formatResetTime(row.reset_5h_at)" class="views-user-keys-view__panel-15">
                  ⟳ {{ formatResetTime(row.reset_5h_at) }}
                </div>
              </div>
              <!-- 1d window -->
              <div v-if="row.rate_limit_1d > 0">
                <div class="views-user-keys-view__panel-13">
                  <span class="views-user-keys-view__text-7">1d</span>
                  <span :class="[
                    'views-user-keys-view__text-18',
                    row.usage_1d >= row.rate_limit_1d ? 'views-user-keys-view__text-15' :
                    row.usage_1d >= row.rate_limit_1d * 0.8 ? 'views-user-keys-view__text-16' :
                    'views-user-keys-view__text-19'
                  ]">
                    {{ formatPointRange(row.usage_1d, row.rate_limit_1d) }}
                  </span>
                </div>
                <div class="views-user-keys-view__panel-14">
                  <div
                    :class="[
                      'views-user-keys-view__panel-31',
                      row.usage_1d >= row.rate_limit_1d ? 'views-user-keys-view__panel-32' :
                      row.usage_1d >= row.rate_limit_1d * 0.8 ? 'views-user-keys-view__panel-33' :
                      'views-user-keys-view__panel-35'
                    ]"
                    :style="{ width: Math.min((row.usage_1d / row.rate_limit_1d) * 100, 100) + '%' }"
                  />
                </div>
                <div v-if="row.reset_1d_at && formatResetTime(row.reset_1d_at)" class="views-user-keys-view__panel-15">
                  ⟳ {{ formatResetTime(row.reset_1d_at) }}
                </div>
              </div>
              <!-- 7d window -->
              <div v-if="row.rate_limit_7d > 0">
                <div class="views-user-keys-view__panel-13">
                  <span class="views-user-keys-view__text-7">7d</span>
                  <span :class="[
                    'views-user-keys-view__text-18',
                    row.usage_7d >= row.rate_limit_7d ? 'views-user-keys-view__text-15' :
                    row.usage_7d >= row.rate_limit_7d * 0.8 ? 'views-user-keys-view__text-16' :
                    'views-user-keys-view__text-19'
                  ]">
                    {{ formatPointRange(row.usage_7d, row.rate_limit_7d) }}
                  </span>
                </div>
                <div class="views-user-keys-view__panel-14">
                  <div
                    :class="[
                      'views-user-keys-view__panel-31',
                      row.usage_7d >= row.rate_limit_7d ? 'views-user-keys-view__panel-32' :
                      row.usage_7d >= row.rate_limit_7d * 0.8 ? 'views-user-keys-view__panel-33' :
                      'views-user-keys-view__panel-35'
                    ]"
                    :style="{ width: Math.min((row.usage_7d / row.rate_limit_7d) * 100, 100) + '%' }"
                  />
                </div>
                <div v-if="row.reset_7d_at && formatResetTime(row.reset_7d_at)" class="views-user-keys-view__panel-15">
                  ⟳ {{ formatResetTime(row.reset_7d_at) }}
                </div>
              </div>
              <!-- Reset button -->
              <button
                v-if="row.usage_5h > 0 || row.usage_1d > 0 || row.usage_7d > 0"
                @click.stop="confirmResetRateLimitFromTable(row)"
                class="views-user-keys-view__action-4"
                :title="t('keys.resetRateLimitUsage')"
              >
                <Icon name="refresh" size="xs" />
                {{ t('keys.resetUsage') }}
              </button>
            </div>
            <span v-else class="views-user-keys-view__text-6">-</span>
          </template>

          <template #cell-expires_at="{ value }">
            <span v-if="value" :class="[
              'views-user-keys-view__panel-8',
              new Date(value) < new Date() ? 'views-user-keys-view__text-20' : 'views-user-keys-view__text-21'
            ]">
              {{ formatDateTime(value) }}
            </span>
            <span v-else class="views-user-keys-view__text-6">{{ t('keys.noExpiration') }}</span>
          </template>

          <template #cell-status="{ value }">
            <span :class="[
              'badge',
              value === 'active' ? 'badge-success' :
              value === 'quota_exhausted' ? 'badge-warning' :
              value === 'expired' ? 'badge-danger' :
              'badge-gray'
            ]">
              {{ t('keys.status.' + value) }}
            </span>
          </template>

          <template #cell-last_used_at="{ value }">
            <span v-if="value" class="views-user-keys-view__text-8">
              {{ formatDateTime(value) }}
            </span>
            <span v-else class="views-user-keys-view__text-6">-</span>
          </template>

          <template #cell-last_used_ip="{ value }">
            <span v-if="value" class="views-user-keys-view__text-8">
              {{ value }}
            </span>
            <span v-else class="views-user-keys-view__text-6">-</span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="views-user-keys-view__text-8">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="views-user-keys-view__panel-16">
              <button
                @click="openGroupSelector(row)"
                class="views-user-keys-view__action-5 group/dropdown"
                :title="t('keys.manageGroups')"
              >
                <Icon name="grid" size="sm" />
                <span class="views-user-keys-view__code">{{ t('keys.manageGroups') }}</span>
              </button>
              <!-- Use Key Button -->
              <button
                @click="openUseKeyModal(row)"
                class="views-user-keys-view__action-6"
              >
                <Icon name="terminal" size="sm" />
                <span class="views-user-keys-view__code">{{ t('keys.useKey') }}</span>
              </button>
              <!-- Import to CC Switch Button -->
              <button
                v-if="!publicSettings?.hide_ccs_import_button"
                @click="importToCcswitch(row)"
                class="views-user-keys-view__action-7"
              >
                <Icon name="upload" size="sm" />
                <span class="views-user-keys-view__code">{{ t('keys.importToCcSwitch') }}</span>
              </button>
              <!-- Toggle Status Button -->
              <button
                @click="toggleKeyStatus(row)"
                :class="[
                  'views-user-keys-view__action-14',
                  row.status === 'active'
                    ? 'views-user-keys-view__action-15'
                    : 'views-user-keys-view__action-16'
                ]"
              >
                <Icon v-if="row.status === 'active'" name="ban" size="sm" />
                <Icon v-else name="checkCircle" size="sm" />
                <span class="views-user-keys-view__code">{{ row.status === 'active' ? t('keys.disable') : t('keys.enable') }}</span>
              </button>
              <!-- Edit Button -->
              <button
                @click="editKey(row)"
                class="views-user-keys-view__action-8"
              >
                <Icon name="edit" size="sm" />
                <span class="views-user-keys-view__code">{{ t('common.edit') }}</span>
              </button>
              <!-- Delete Button -->
              <button
                @click="confirmDelete(row)"
                class="views-user-keys-view__action-9"
              >
                <Icon name="trash" size="sm" />
                <span class="views-user-keys-view__code">{{ t('common.delete') }}</span>
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('keys.noKeysYet')"
              :description="t('keys.createFirstKey')"
              :action-text="t('keys.createKey')"
              @action="showCreateModal = true"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Create/Edit Modal -->
    <BaseDialog
      :show="showCreateModal || showEditModal"
      :title="showEditModal ? t('keys.editKey') : t('keys.createKey')"
      width="normal"
      @close="closeModals"
    >
      <form id="key-form" @submit.prevent="handleSubmit" class="views-user-keys-view__form">
        <div>
          <label class="input-label">{{ t('keys.nameLabel') }}</label>
          <input
            v-model="formData.name"
            type="text"
            required
            class="input"
            :placeholder="t('keys.namePlaceholder')"
            data-tour="key-form-name"
          />
        </div>

        <div>
          <label class="input-label">{{ t('keys.groupLabel') }}</label>
          <div class="views-user-keys-view__panel-17" data-tour="key-form-group">
            <div class="views-user-keys-view__panel-18">
              {{ t('keys.selectGroup') }} · {{ formData.group_ids.length }}
            </div>
            <div class="views-user-keys-view__panel-19">
              <label
                v-for="option in groupOptions"
                :key="String(option.value)"
                class="views-user-keys-view__label"
              >
                <input
                  type="checkbox"
                  :checked="formData.group_ids.includes(String(option.value))"
                  class="views-user-keys-view__field-2"
                  @change="formData.group_ids.includes(String(option.value)) ? formData.group_ids.splice(formData.group_ids.indexOf(String(option.value)), 1) : formData.group_ids.push(String(option.value))"
                />
                <GroupOptionItem
                  :name="option.label"
                  :platform="option.platform"
                  :subscription-type="option.subscriptionType"
                  :rate-multiplier="option.rate"
                  :user-rate-multiplier="option.userRate"
                  :effective-rate-multiplier="option.effectiveRate"
                  :peak-rate-enabled="option.peakRateEnabled"
                  :peak-start="option.peakStart"
                  :peak-end="option.peakEnd"
                  :peak-rate-multiplier="option.peakRateMultiplier"
                  :description="option.description"
                  :selected="formData.group_ids.includes(String(option.value))"
                />
              </label>
            </div>
          </div>
        </div>

        <!-- Custom Key Section (only for create) -->
        <div v-if="!showEditModal" class="views-user-keys-view__panel-20">
          <div class="views-user-keys-view__panel-21">
            <label class="views-user-keys-view__label-2 input-label">{{ t('keys.customKeyLabel') }}</label>
            <button
              type="button"
              @click="formData.use_custom_key = !formData.use_custom_key"
              :class="[
                'views-user-keys-view__action-17',
                formData.use_custom_key ? 'views-user-keys-view__action-18' : 'views-user-keys-view__action-19'
              ]"
            >
              <span
                :class="[
                  'views-user-keys-view__text-23',
                  formData.use_custom_key ? 'toggle-thumb--on' : 'views-user-keys-view__text-24'
                ]"
              />
            </button>
          </div>
          <div v-if="formData.use_custom_key">
            <input
              v-model="formData.custom_key"
              type="text"
              class="views-user-keys-view__field-3 input"
              :placeholder="t('keys.customKeyPlaceholder')"
              :class="{ 'views-user-keys-view__field-6': customKeyError }"
            />
            <p v-if="customKeyError" class="views-user-keys-view__description">{{ customKeyError }}</p>
            <p v-else class="input-hint">{{ t('keys.customKeyHint') }}</p>
          </div>
        </div>

        <div v-if="showEditModal">
          <label class="input-label">{{ t('keys.statusLabel') }}</label>
          <Select
            v-model="formData.status"
            :options="statusOptions"
            :placeholder="t('keys.selectStatus')"
          />
        </div>

        <!-- IP Restriction Section -->
        <div class="views-user-keys-view__panel-20">
          <div class="views-user-keys-view__panel-21">
            <label class="views-user-keys-view__label-2 input-label">{{ t('keys.ipRestriction') }}</label>
            <button
              type="button"
              @click="formData.enable_ip_restriction = !formData.enable_ip_restriction"
              :class="[
                'views-user-keys-view__action-17',
                formData.enable_ip_restriction ? 'views-user-keys-view__action-18' : 'views-user-keys-view__action-19'
              ]"
            >
              <span
                :class="[
                  'views-user-keys-view__text-23',
                  formData.enable_ip_restriction ? 'toggle-thumb--on' : 'views-user-keys-view__text-24'
                ]"
              />
            </button>
          </div>

          <div v-if="formData.enable_ip_restriction" class="views-user-keys-view__panel-22">
            <div>
              <label class="input-label">{{ t('keys.ipWhitelist') }}</label>
              <textarea
                v-model="formData.ip_whitelist"
                rows="3"
                class="views-user-keys-view__field-4 input"
                :placeholder="t('keys.ipWhitelistPlaceholder')"
              />
              <p class="input-hint">{{ t('keys.ipWhitelistHint') }}</p>
            </div>

            <div>
              <label class="input-label">{{ t('keys.ipBlacklist') }}</label>
              <textarea
                v-model="formData.ip_blacklist"
                rows="3"
                class="views-user-keys-view__field-4 input"
                :placeholder="t('keys.ipBlacklistPlaceholder')"
              />
              <p class="input-hint">{{ t('keys.ipBlacklistHint') }}</p>
            </div>
          </div>
        </div>

        <!-- Quota Limit Section -->
        <div class="views-user-keys-view__panel-20">
          <label class="input-label">{{ t('keys.quotaLimit') }}</label>
          <!-- Switch commented out - always show input, 0 = unlimited
          <div class="views-user-keys-view__panel-21">
            <label class="views-user-keys-view__label-2 input-label">{{ t('keys.quotaLimit') }}</label>
            <button
              type="button"
              @click="formData.enable_quota = !formData.enable_quota"
              :class="[
                'views-user-keys-view__action-17',
                formData.enable_quota ? 'views-user-keys-view__action-18' : 'views-user-keys-view__action-19'
              ]"
            >
              <span
                :class="[
                  'views-user-keys-view__text-23',
                  formData.enable_quota ? 'toggle-thumb--on' : 'views-user-keys-view__text-24'
                ]"
              />
            </button>
          </div>
          -->

          <div class="views-user-keys-view__panel-23">
            <div>
              <div class="views-user-keys-view__panel-4">
                <Icon name="points" size="sm" class="views-user-keys-view__text-9" />
                <input
                  v-model.number="formData.quota"
                  type="number"
                  step="0.01"
                  min="0"
                  class="views-user-keys-view__field-5 input"
                  :placeholder="t('keys.quotaAmountPlaceholder')"
                />
              </div>
              <p class="input-hint">{{ t('keys.quotaAmountHint') }}</p>
            </div>

            <!-- Quota used display (only in edit mode) -->
            <div v-if="showEditModal && selectedKey && selectedKey.quota > 0">
              <label class="input-label">{{ t('keys.quotaUsed') }}</label>
              <div class="views-user-keys-view__panel-6">
                <div class="views-user-keys-view__panel-24">
                  <span class="views-user-keys-view__text-3">
                    {{ formatPointAmount(selectedKey.quota_used) }}
                  </span>
                  <span class="views-user-keys-view__text-10">/</span>
                  <span class="views-user-keys-view__text-7">
                    {{ formatPoints(selectedKey.quota) }}
                  </span>
                </div>
                <button
                  type="button"
                  @click="confirmResetQuota"
                  class="views-user-keys-view__panel-8 btn btn-secondary"
                  :title="t('keys.resetQuotaUsed')"
                >
                  {{ t('keys.reset') }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Rate Limit Section -->
        <div class="views-user-keys-view__panel-20">
          <div class="views-user-keys-view__panel-21">
            <label class="views-user-keys-view__label-2 input-label">{{ t('keys.rateLimitSection') }}</label>
            <button
              type="button"
              @click="formData.enable_rate_limit = !formData.enable_rate_limit"
              :class="[
                'views-user-keys-view__action-17',
                formData.enable_rate_limit ? 'views-user-keys-view__action-18' : 'views-user-keys-view__action-19'
              ]"
            >
              <span
                :class="[
                  'views-user-keys-view__text-23',
                  formData.enable_rate_limit ? 'toggle-thumb--on' : 'views-user-keys-view__text-24'
                ]"
              />
            </button>
          </div>

          <div v-if="formData.enable_rate_limit" class="views-user-keys-view__panel-22">
            <p class="views-user-keys-view__description-2 input-hint">{{ t('keys.rateLimitHint') }}</p>
            <!-- 5-Hour Limit -->
            <div>
              <label class="input-label">{{ t('keys.rateLimit5h') }}</label>
              <div class="views-user-keys-view__panel-4">
                <Icon name="points" size="sm" class="views-user-keys-view__text-9" />
                <input
                  v-model.number="formData.rate_limit_5h"
                  type="number"
                  step="0.01"
                  min="0"
                  class="views-user-keys-view__field-5 input"
                  :placeholder="'0'"
                />
              </div>
              <!-- Usage info (edit mode only) -->
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_5h > 0" class="views-user-keys-view__panel-25">
                <div class="views-user-keys-view__panel-6">
                  <div class="views-user-keys-view__panel-26">
                    <span :class="[
                      'views-user-keys-view__text-14',
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h ? 'views-user-keys-view__text-15' :
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h * 0.8 ? 'views-user-keys-view__text-16' :
                      'views-user-keys-view__text-17'
                    ]">
                      {{ formatPointAmount(selectedKey.usage_5h) }}
                    </span>
                    <span class="views-user-keys-view__text-10">/</span>
                    <span class="views-user-keys-view__text-7">
                      {{ formatPoints(selectedKey.rate_limit_5h) }}
                    </span>
                  </div>
                </div>
                <div class="views-user-keys-view__panel-11">
                  <div
                    :class="[
                      'views-user-keys-view__panel-31',
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h ? 'views-user-keys-view__panel-32' :
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h * 0.8 ? 'views-user-keys-view__panel-33' :
                      'views-user-keys-view__panel-36'
                    ]"
                    :style="{ width: Math.min((selectedKey.usage_5h / selectedKey.rate_limit_5h) * 100, 100) + '%' }"
                  />
                </div>
              </div>
            </div>

            <!-- Daily Limit -->
            <div>
              <label class="input-label">{{ t('keys.rateLimit1d') }}</label>
              <div class="views-user-keys-view__panel-4">
                <Icon name="points" size="sm" class="views-user-keys-view__text-9" />
                <input
                  v-model.number="formData.rate_limit_1d"
                  type="number"
                  step="0.01"
                  min="0"
                  class="views-user-keys-view__field-5 input"
                  :placeholder="'0'"
                />
              </div>
              <!-- Usage info (edit mode only) -->
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_1d > 0" class="views-user-keys-view__panel-25">
                <div class="views-user-keys-view__panel-6">
                  <div class="views-user-keys-view__panel-26">
                    <span :class="[
                      'views-user-keys-view__text-14',
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d ? 'views-user-keys-view__text-15' :
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d * 0.8 ? 'views-user-keys-view__text-16' :
                      'views-user-keys-view__text-17'
                    ]">
                      {{ formatPointAmount(selectedKey.usage_1d) }}
                    </span>
                    <span class="views-user-keys-view__text-10">/</span>
                    <span class="views-user-keys-view__text-7">
                      {{ formatPoints(selectedKey.rate_limit_1d) }}
                    </span>
                  </div>
                </div>
                <div class="views-user-keys-view__panel-11">
                  <div
                    :class="[
                      'views-user-keys-view__panel-31',
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d ? 'views-user-keys-view__panel-32' :
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d * 0.8 ? 'views-user-keys-view__panel-33' :
                      'views-user-keys-view__panel-36'
                    ]"
                    :style="{ width: Math.min((selectedKey.usage_1d / selectedKey.rate_limit_1d) * 100, 100) + '%' }"
                  />
                </div>
              </div>
            </div>

            <!-- 7-Day Limit -->
            <div>
              <label class="input-label">{{ t('keys.rateLimit7d') }}</label>
              <div class="views-user-keys-view__panel-4">
                <Icon name="points" size="sm" class="views-user-keys-view__text-9" />
                <input
                  v-model.number="formData.rate_limit_7d"
                  type="number"
                  step="0.01"
                  min="0"
                  class="views-user-keys-view__field-5 input"
                  :placeholder="'0'"
                />
              </div>
              <!-- Usage info (edit mode only) -->
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_7d > 0" class="views-user-keys-view__panel-25">
                <div class="views-user-keys-view__panel-6">
                  <div class="views-user-keys-view__panel-26">
                    <span :class="[
                      'views-user-keys-view__text-14',
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d ? 'views-user-keys-view__text-15' :
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d * 0.8 ? 'views-user-keys-view__text-16' :
                      'views-user-keys-view__text-17'
                    ]">
                      {{ formatPointAmount(selectedKey.usage_7d) }}
                    </span>
                    <span class="views-user-keys-view__text-10">/</span>
                    <span class="views-user-keys-view__text-7">
                      {{ formatPoints(selectedKey.rate_limit_7d) }}
                    </span>
                  </div>
                </div>
                <div class="views-user-keys-view__panel-11">
                  <div
                    :class="[
                      'views-user-keys-view__panel-31',
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d ? 'views-user-keys-view__panel-32' :
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d * 0.8 ? 'views-user-keys-view__panel-33' :
                      'views-user-keys-view__panel-36'
                    ]"
                    :style="{ width: Math.min((selectedKey.usage_7d / selectedKey.rate_limit_7d) * 100, 100) + '%' }"
                  />
                </div>
              </div>
            </div>

            <!-- Reset Rate Limit button (edit mode only) -->
            <div v-if="showEditModal && selectedKey && (selectedKey.rate_limit_5h > 0 || selectedKey.rate_limit_1d > 0 || selectedKey.rate_limit_7d > 0)">
              <button
                type="button"
                @click="confirmResetRateLimit"
                class="views-user-keys-view__panel-8 btn btn-secondary"
              >
                {{ t('keys.resetRateLimitUsage') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Expiration Section -->
        <div class="views-user-keys-view__panel-20">
          <div class="views-user-keys-view__panel-21">
            <label class="views-user-keys-view__label-2 input-label">{{ t('keys.expiration') }}</label>
            <button
              type="button"
              @click="formData.enable_expiration = !formData.enable_expiration"
              :class="[
                'views-user-keys-view__action-17',
                formData.enable_expiration ? 'views-user-keys-view__action-18' : 'views-user-keys-view__action-19'
              ]"
            >
              <span
                :class="[
                  'views-user-keys-view__text-23',
                  formData.enable_expiration ? 'toggle-thumb--on' : 'views-user-keys-view__text-24'
                ]"
              />
            </button>
          </div>

          <div v-if="formData.enable_expiration" class="views-user-keys-view__panel-22">
            <!-- Quick select buttons (for both create and edit mode) -->
            <div class="views-user-keys-view__panel-27">
              <button
                v-for="days in ['7', '30', '90']"
                :key="days"
                type="button"
                @click="setExpirationDays(parseInt(days))"
                :class="[
                  'views-user-keys-view__action-20',
                  formData.expiration_preset === days
                    ? 'views-user-keys-view__action-21'
                    : 'views-user-keys-view__action-22'
                ]"
              >
                {{ showEditModal ? t('keys.extendDays', { days }) : t('keys.expiresInDays', { days }) }}
              </button>
              <button
                type="button"
                @click="formData.expiration_preset = 'custom'"
                :class="[
                  'views-user-keys-view__action-20',
                  formData.expiration_preset === 'custom'
                    ? 'views-user-keys-view__action-21'
                    : 'views-user-keys-view__action-22'
                ]"
              >
                {{ t('keys.customDate') }}
              </button>
            </div>

            <!-- Date picker (always show for precise adjustment) -->
            <div>
              <label class="input-label">{{ t('keys.expirationDate') }}</label>
              <input
                v-model="formData.expiration_date"
                type="datetime-local"
                class="input"
              />
              <p class="input-hint">{{ t('keys.expirationDateHint') }}</p>
            </div>

            <!-- Current expiration display (only in edit mode) -->
            <div v-if="showEditModal && selectedKey?.expires_at" class="views-user-keys-view__panel-8">
              <span class="views-user-keys-view__text-7">{{ t('keys.currentExpiration') }}: </span>
              <span class="views-user-keys-view__text-3">
                {{ formatDateTime(selectedKey.expires_at) }}
              </span>
            </div>
          </div>
        </div>
      </form>
      <template #footer>
        <div class="views-user-keys-view__panel-3">
          <button @click="closeModals" type="button" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button
            form="key-form"
            type="submit"
            :disabled="submitting"
            class="btn btn-primary"
            data-tour="key-form-submit"
          >
            <svg
              v-if="submitting"
              class="views-user-keys-view__icon-5"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="views-user-keys-view__circle"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="views-user-keys-view__path"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{
              submitting
                ? t('keys.saving')
                : showEditModal
                  ? t('common.update')
                  : t('common.create')
            }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('keys.deleteKey')"
      :message="t('keys.deleteConfirmMessage', { name: selectedKey?.name })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="handleDelete"
      @cancel="showDeleteDialog = false"
    />

    <!-- Reset Quota Confirmation Dialog -->
    <ConfirmDialog
      :show="showResetQuotaDialog"
      :title="t('keys.resetQuotaTitle')"
      :message="t('keys.resetQuotaConfirmMessage', { name: selectedKey?.name, used: formatPoints(selectedKey?.quota_used) })"
      :confirm-text="t('keys.reset')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="resetQuotaUsed"
      @cancel="showResetQuotaDialog = false"
    />

    <!-- Reset Rate Limit Confirmation Dialog -->
    <ConfirmDialog
      :show="showResetRateLimitDialog"
      :title="t('keys.resetRateLimitTitle')"
      :message="t('keys.resetRateLimitConfirmMessage', { name: selectedKey?.name })"
      :confirm-text="t('keys.reset')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="resetRateLimitUsage"
      @cancel="showResetRateLimitDialog = false"
    />

    <!-- Use Key Modal -->
    <UseKeyModal
      :show="showUseKeyModal"
      :api-key="selectedKey?.key || ''"
      :base-url="publicSettings?.api_base_url || ''"
      :platform="selectedKey?.group?.platform || null"
      :allow-messages-dispatch="selectedKey?.group?.allow_messages_dispatch || false"
      @close="closeUseKeyModal"
    />

    <BaseDialog
      :show="showGroupManager"
      :title="t('keys.manageGroups')"
      width="normal"
      @close="closeGroupSelector"
    >
      <div
        class="views-user-keys-view__panel-20"
      >
        <p class="views-user-keys-view__description-4">{{ selectedKeyForGroup?.name }}</p>
        <GroupTransferPicker
          v-model="pendingGroupIds"
          :groups="groups"
          :available-label="t('keys.availableGroups')"
          :selected-label="t('keys.selectedGroups')"
          :search-placeholder="t('keys.searchGroup')"
          :empty-label="t('keys.noGroupFound')"
        />
        <div class="views-user-keys-view__panel-30">
          <span class="views-user-keys-view__text-5">{{ t('keys.selectedGroupCount', { count: pendingGroupIds.length }) }}</span>
          <button
            type="button"
            class="views-user-keys-view__action-11"
            :disabled="pendingGroupIds.length === 0"
            @click="saveSelectedGroups"
          >
            {{ t('common.save') }}
          </button>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
	import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
	import { useI18n } from 'vue-i18n'
	import { useAppStore } from '@/stores/app'
	import { useOnboardingStore } from '@/stores/onboarding'
	import { useClipboard } from '@/composables/useClipboard'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'

const { t } = useI18n()
import { keysAPI, authAPI, usageAPI, userGroupsAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
	import DataTable from '@/components/common/DataTable.vue'
	import Pagination from '@/components/common/Pagination.vue'
	import BaseDialog from '@/components/common/BaseDialog.vue'
	import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
	import EmptyState from '@/components/common/EmptyState.vue'
	import Select from '@/components/common/Select.vue'
	import SearchInput from '@/components/common/SearchInput.vue'
	import Icon from '@/components/icons/Icon.vue'
	import UseKeyModal from '@/components/keys/UseKeyModal.vue'
	import EndpointPopover from '@/components/keys/EndpointPopover.vue'
	import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
	import GroupBadge from '@/components/common/GroupBadge.vue'
	import GroupTransferPicker from '@/components/common/GroupTransferPicker.vue'
	import type { ApiKey, Group, PublicSettings, SubscriptionType, GroupPlatform, UpdateApiKeyRequest } from '@/types'
import type { Column } from '@/components/common/types'
import type { BatchApiKeyUsageStats } from '@/api/usage'
import { formatDateTime, formatPointAmount, formatPoints } from '@/utils/format'
import { maskApiKey } from '@/utils/maskApiKey'
import { buildCcSwitchImportDeeplink } from '@/utils/ccswitchImport'

const formatPointRange = (used: number | null | undefined, limit: number | null | undefined): string =>
  `${formatPointAmount(used)} / ${formatPoints(limit)}`

// Helper to format date for datetime-local input
const formatDateTimeLocal = (isoDate: string): string => {
  const date = new Date(isoDate)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

interface GroupOption {
  value: string
  label: string
  description: string | null
  rate: number
  userRate: number | null
  effectiveRate: number | null
  peakRateEnabled: boolean
  peakStart: string
  peakEnd: string
  peakRateMultiplier: number
  subscriptionType: SubscriptionType
  platform: GroupPlatform
}

const appStore = useAppStore()
const onboardingStore = useOnboardingStore()
const { copyToClipboard: clipboardCopy } = useClipboard()

const allColumns = computed<Column[]>(() => [
  { key: 'name', label: t('common.name'), sortable: true },
  { key: 'id', label: t('keys.id'), sortable: true },
  { key: 'key', label: t('keys.apiKey'), sortable: false },
  { key: 'group', label: t('keys.group'), sortable: false },
  { key: 'current_concurrency', label: t('keys.currentConcurrency'), sortable: true },
  { key: 'usage', label: t('keys.usage'), sortable: false },
  { key: 'rate_limit', label: t('keys.rateLimitColumn'), sortable: false },
  { key: 'expires_at', label: t('keys.expiresAt'), sortable: true },
  { key: 'status', label: t('common.status'), sortable: true },
  { key: 'last_used_at', label: t('keys.lastUsedAt'), sortable: true },
  { key: 'last_used_ip', label: t('keys.lastUsedIP'), sortable: false },
  { key: 'created_at', label: t('keys.created'), sortable: true },
  { key: 'actions', label: t('common.actions'), sortable: false }
])

const ALWAYS_VISIBLE_COLUMNS = new Set(['name', 'actions'])
const DEFAULT_HIDDEN_COLUMNS = ['id', 'rate_limit', 'last_used_at', 'last_used_ip']
const HIDDEN_COLUMNS_KEY = 'api-key-hidden-columns'
const COLUMN_SETTINGS_VERSION_KEY = 'api-key-column-settings-version'
const COLUMN_SETTINGS_VERSION = 3
const VERSION_NEW_HIDDEN_COLUMNS: Record<number, string[]> = {
  2: ['last_used_ip'],
  3: ['id']
}

const toggleableColumns = computed(() =>
  allColumns.value.filter((col) => !ALWAYS_VISIBLE_COLUMNS.has(col.key))
)

const hiddenColumns = reactive<Set<string>>(new Set())

const saveColumnsToStorage = () => {
  try {
    localStorage.setItem(HIDDEN_COLUMNS_KEY, JSON.stringify([...hiddenColumns]))
    localStorage.setItem(COLUMN_SETTINGS_VERSION_KEY, String(COLUMN_SETTINGS_VERSION))
  } catch (error) {
    console.error('Failed to save API key table columns:', error)
  }
}

const loadSavedColumns = () => {
  hiddenColumns.clear()
  try {
    const saved = localStorage.getItem(HIDDEN_COLUMNS_KEY)
    if (saved) {
      const parsed = JSON.parse(saved) as string[]
      const validColumnKeys = new Set(allColumns.value.map((col) => col.key))
      parsed
        .filter((key) =>
          typeof key === 'string' &&
          validColumnKeys.has(key) &&
          !ALWAYS_VISIBLE_COLUMNS.has(key)
        )
        .forEach((key) => hiddenColumns.add(key))
      const storedVersion = Number(localStorage.getItem(COLUMN_SETTINGS_VERSION_KEY) ?? '1')
      if (storedVersion < COLUMN_SETTINGS_VERSION) {
        for (let v = storedVersion + 1; v <= COLUMN_SETTINGS_VERSION; v++) {
          for (const key of VERSION_NEW_HIDDEN_COLUMNS[v] ?? []) {
            if (validColumnKeys.has(key) && !ALWAYS_VISIBLE_COLUMNS.has(key)) {
              hiddenColumns.add(key)
            }
          }
        }
        saveColumnsToStorage()
      } else {
        localStorage.setItem(COLUMN_SETTINGS_VERSION_KEY, String(COLUMN_SETTINGS_VERSION))
      }
    } else {
      DEFAULT_HIDDEN_COLUMNS.forEach((key) => hiddenColumns.add(key))
      localStorage.setItem(COLUMN_SETTINGS_VERSION_KEY, String(COLUMN_SETTINGS_VERSION))
    }
  } catch (error) {
    console.error('Failed to load API key table columns:', error)
    DEFAULT_HIDDEN_COLUMNS.forEach((key) => hiddenColumns.add(key))
  }
}

const toggleColumn = (key: string) => {
  if (ALWAYS_VISIBLE_COLUMNS.has(key)) return
  if (hiddenColumns.has(key)) {
    hiddenColumns.delete(key)
  } else {
    hiddenColumns.add(key)
  }
  saveColumnsToStorage()
}

const isColumnVisible = (key: string) => !hiddenColumns.has(key)

const columns = computed<Column[]>(() =>
  allColumns.value.filter((col) => ALWAYS_VISIBLE_COLUMNS.has(col.key) || !hiddenColumns.has(col.key))
)

const apiKeys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
const loading = ref(true)
const submitting = ref(false)
const now = ref(new Date())
let resetTimer: ReturnType<typeof setInterval> | null = null
const usageStats = ref<Record<string, BatchApiKeyUsageStats>>({})
const userGroupRates = ref<Record<string, number>>({})

const pagination = ref({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})
const sortState = ref({
  sort_by: 'created_at',
  sort_order: 'desc' as 'asc' | 'desc'
})

// Filter state
const filterSearch = ref('')
const filterStatus = ref('')
const filterGroupId = ref('')

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteDialog = ref(false)
const showResetQuotaDialog = ref(false)
const showResetRateLimitDialog = ref(false)
const showUseKeyModal = ref(false)
const showColumnDropdown = ref(false)
const selectedKey = ref<ApiKey | null>(null)
const copiedKeyId = ref<string | null>(null)
const groupSelectorKeyId = ref<string | null>(null)
const showGroupManager = ref(false)
const publicSettings = ref<PublicSettings | null>(null)
const columnDropdownRef = ref<HTMLElement | null>(null)
const pendingGroupIds = ref<string[]>([])
let abortController: AbortController | null = null

// Get the currently selected key for group change
const selectedKeyForGroup = computed(() => {
  if (groupSelectorKeyId.value === null) return null
  return apiKeys.value.find((k) => k.id === groupSelectorKeyId.value) || null
})

const keyGroupOptions = (key: ApiKey): Group[] => {
  const ids = key.group_ids?.length ? key.group_ids : (key.group_id ? [key.group_id] : [])
  return ids
    .map((id) => groups.value.find((group) => group.id === id) || (key.group_id === id ? key.group : undefined))
    .filter((group): group is Group => Boolean(group))
}

const formData = ref({
  name: '',
  group_ids: [] as string[],
  status: 'active' as 'active' | 'inactive',
  use_custom_key: false,
  custom_key: '',
  enable_ip_restriction: false,
  ip_whitelist: '',
  ip_blacklist: '',
  // Quota settings (empty = unlimited)
  enable_quota: false,
  quota: null as number | null,
  // Rate limit settings
  enable_rate_limit: false,
  rate_limit_5h: null as number | null,
  rate_limit_1d: null as number | null,
  rate_limit_7d: null as number | null,
  enable_expiration: false,
  expiration_preset: '30' as '7' | '30' | '90' | 'custom',
  expiration_date: ''
})

// 自定义Key验证
const customKeyError = computed(() => {
  if (!formData.value.use_custom_key || !formData.value.custom_key) {
    return ''
  }
  const key = formData.value.custom_key
  if (key.length < 16) {
    return t('keys.customKeyTooShort')
  }
  // 检查字符：只允许字母、数字、下划线、连字符
  if (!/^[a-zA-Z0-9_-]+$/.test(key)) {
    return t('keys.customKeyInvalidChars')
  }
  return ''
})

const statusOptions = computed(() => [
  { value: 'active', label: t('common.active') },
  { value: 'inactive', label: t('common.inactive') }
])

const shouldSubmitEditStatus = (key: ApiKey, status: 'active' | 'inactive') => {
  if (key.status === 'quota_exhausted' || key.status === 'expired') {
    return status === 'active'
  }
  return true
}

// Filter dropdown options
const groupFilterOptions = computed(() => [
  { value: '', label: t('keys.allGroups') },
  { value: 0, label: t('keys.noGroup') },
  ...groups.value.map((g) => ({ value: g.id, label: g.name }))
])

const statusFilterOptions = computed(() => [
  { value: '', label: t('keys.allStatus') },
  { value: 'active', label: t('keys.status.active') },
  { value: 'inactive', label: t('keys.status.inactive') },
  { value: 'quota_exhausted', label: t('keys.status.quota_exhausted') },
  { value: 'expired', label: t('keys.status.expired') }
])

const onFilterChange = () => {
  pagination.value.page = 1
  loadApiKeys()
}

const onGroupFilterChange = (value: string | number | boolean | null) => {
  filterGroupId.value = typeof value === 'string' ? value : ''
  onFilterChange()
}

const onStatusFilterChange = (value: string | number | boolean | null) => {
  filterStatus.value = value as string
  onFilterChange()
}

// Convert groups to Select options format with rate multiplier and subscription type
const groupOptions = computed<GroupOption[]>(() =>
  groups.value.map((group) => ({
    value: group.id,
    label: group.name,
    description: group.description,
    rate: group.rate_multiplier,
    userRate: userGroupRates.value[group.id] ?? null,
    effectiveRate: group.effective_rate_multiplier ?? null,
    peakRateEnabled: group.peak_rate_enabled,
    peakStart: group.peak_start,
    peakEnd: group.peak_end,
    peakRateMultiplier: group.peak_rate_multiplier,
    subscriptionType: group.subscription_type,
    platform: group.platform
  }))
)

const copyToClipboard = async (text: string, keyId: string) => {
  const success = await clipboardCopy(text, t('keys.copied'))
  if (success) {
    copiedKeyId.value = keyId
    setTimeout(() => {
      copiedKeyId.value = null
    }, 800)
  }
}

const isAbortError = (error: unknown) => {
  if (!error || typeof error !== 'object') return false
  const { name, code } = error as { name?: string; code?: string }
  return name === 'AbortError' || code === 'ERR_CANCELED'
}

const loadApiKeys = async () => {
  abortController?.abort()
  const controller = new AbortController()
  abortController = controller
  const { signal } = controller
  loading.value = true
  try {
    // Build filters
    const filters: {
      search?: string
      status?: string
      group_id?: string | string
      sort_by?: string
      sort_order?: 'asc' | 'desc'
    } = {}
    if (filterSearch.value) filters.search = filterSearch.value
    if (filterStatus.value) filters.status = filterStatus.value
    if (filterGroupId.value !== '') filters.group_id = filterGroupId.value
    filters.sort_by = sortState.value.sort_by
    filters.sort_order = sortState.value.sort_order

    const response = await keysAPI.list(pagination.value.page, pagination.value.page_size, filters, {
      signal
    })
    if (signal.aborted) return
    apiKeys.value = response.items.map((key) => ({
      ...key,
      group_ids: key.group_ids?.length ? key.group_ids : (key.group_id ? [key.group_id] : [])
    }))
    pagination.value.total = response.total
    pagination.value.pages = response.pages

    // Load usage stats for all API keys in the list
    if (response.items.length > 0) {
      const keyIds = response.items.map((k) => k.id)
      try {
        const usageResponse = await usageAPI.getDashboardApiKeysUsage(keyIds, { signal })
        if (signal.aborted) return
        usageStats.value = usageResponse.stats
      } catch (e) {
        if (!isAbortError(e)) {
          console.error('Failed to load usage stats:', e)
        }
      }
    }
  } catch (error) {
    if (isAbortError(error)) {
      return
    }
    appStore.showError(t('keys.failedToLoad'))
  } finally {
    if (abortController === controller) {
      loading.value = false
    }
  }
}

const loadGroups = async () => {
  try {
    groups.value = await userGroupsAPI.getAvailable()
  } catch (error) {
    console.error('Failed to load groups:', error)
  }
}

const loadUserGroupRates = async () => {
  try {
    userGroupRates.value = await userGroupsAPI.getUserGroupRates()
  } catch (error) {
    console.error('Failed to load user group rates:', error)
  }
}

const loadPublicSettings = async () => {
  try {
    publicSettings.value = await authAPI.getPublicSettings()
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }
}

const openUseKeyModal = (key: ApiKey) => {
  selectedKey.value = key
  showUseKeyModal.value = true
}

const closeUseKeyModal = () => {
  showUseKeyModal.value = false
  selectedKey.value = null
}

const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadApiKeys()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.page_size = pageSize
  pagination.value.page = 1
  loadApiKeys()
}

const handleSort = (key: string, order: 'asc' | 'desc') => {
  sortState.value.sort_by = key
  sortState.value.sort_order = order
  pagination.value.page = 1
  loadApiKeys()
}

const editKey = (key: ApiKey) => {
  selectedKey.value = key
  const hasIPRestriction = (key.ip_whitelist?.length > 0) || (key.ip_blacklist?.length > 0)
  const hasExpiration = !!key.expires_at
  formData.value = {
    name: key.name,
    group_ids: [...(key.group_ids?.length ? key.group_ids : key.group_id ? [key.group_id] : [])],
    status: key.status === 'quota_exhausted' || key.status === 'expired' ? 'inactive' : key.status,
    use_custom_key: false,
    custom_key: '',
    enable_ip_restriction: hasIPRestriction,
    ip_whitelist: (key.ip_whitelist || []).join('\n'),
    ip_blacklist: (key.ip_blacklist || []).join('\n'),
    enable_quota: key.quota > 0,
    quota: key.quota > 0 ? key.quota : null,
    enable_rate_limit: (key.rate_limit_5h > 0) || (key.rate_limit_1d > 0) || (key.rate_limit_7d > 0),
    rate_limit_5h: key.rate_limit_5h || null,
    rate_limit_1d: key.rate_limit_1d || null,
    rate_limit_7d: key.rate_limit_7d || null,
    enable_expiration: hasExpiration,
    expiration_preset: 'custom',
    expiration_date: key.expires_at ? formatDateTimeLocal(key.expires_at) : ''
  }
  showEditModal.value = true
}

const toggleKeyStatus = async (key: ApiKey) => {
  const newStatus = key.status === 'active' ? 'inactive' : 'active'
  try {
    await keysAPI.toggleStatus(key.id, newStatus)
    appStore.showSuccess(
      newStatus === 'active' ? t('keys.keyEnabledSuccess') : t('keys.keyDisabledSuccess')
    )
    loadApiKeys()
  } catch (error) {
    appStore.showError(t('keys.failedToUpdateStatus'))
  }
}

const openGroupSelector = (key: ApiKey) => {
  if (groupSelectorKeyId.value === key.id) {
    closeGroupSelector()
  } else {
    pendingGroupIds.value = key.group_ids?.length ? [...key.group_ids] : (key.group_id ? [key.group_id] : [])
    groupSelectorKeyId.value = key.id
    showGroupManager.value = true
  }
}

const saveSelectedGroups = async () => {
  const key = selectedKeyForGroup.value
  if (!key || pendingGroupIds.value.length === 0) return

  closeGroupSelector()

  try {
    await keysAPI.update(key.id, {
      group_ids: [...pendingGroupIds.value],
      group_id: pendingGroupIds.value[0]
    })
    appStore.showSuccess(t('keys.groupChangedSuccess'))
    loadApiKeys()
  } catch (error) {
    appStore.showError(t('keys.failedToChangeGroup'))
  }
}

const closeGroupSelector = () => {
  groupSelectorKeyId.value = null
  showGroupManager.value = false
}

const handleDocumentClick = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (columnDropdownRef.value && !columnDropdownRef.value.contains(target)) {
    showColumnDropdown.value = false
  }
}

const confirmDelete = (key: ApiKey) => {
  selectedKey.value = key
  showDeleteDialog.value = true
}

const handleSubmit = async () => {
  // Validate at least one group is selected
  if (formData.value.group_ids.length === 0) {
    appStore.showError(t('keys.groupRequired'))
    return
  }

  // Validate custom key if enabled
  if (!showEditModal.value && formData.value.use_custom_key) {
    if (!formData.value.custom_key) {
      appStore.showError(t('keys.customKeyRequired'))
      return
    }
    if (customKeyError.value) {
      appStore.showError(customKeyError.value)
      return
    }
  }

  // Parse IP lists only if IP restriction is enabled
  const parseIPList = (text: string): string[] =>
    text.split('\n').map(ip => ip.trim()).filter(ip => ip.length > 0)
  const ipWhitelist = formData.value.enable_ip_restriction ? parseIPList(formData.value.ip_whitelist) : []
  const ipBlacklist = formData.value.enable_ip_restriction ? parseIPList(formData.value.ip_blacklist) : []

  // Calculate quota value (null/empty/0 = unlimited, stored as 0)
  const quota = formData.value.quota && formData.value.quota > 0 ? formData.value.quota : 0

  // Calculate expiration
  let expiresInDays: number | undefined
  let expiresAt: string | null | undefined
  if (formData.value.enable_expiration && formData.value.expiration_date) {
    if (!showEditModal.value) {
      // Create mode: calculate days from date
      const expDate = new Date(formData.value.expiration_date)
      const now = new Date()
      const diffDays = Math.ceil((expDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
      expiresInDays = diffDays > 0 ? diffDays : 1
    } else {
      // Edit mode: use custom date directly
      expiresAt = new Date(formData.value.expiration_date).toISOString()
    }
  } else if (showEditModal.value) {
    // Edit mode: if expiration disabled or date cleared, send empty string to clear
    expiresAt = ''
  }

  // Calculate rate limit values (send 0 when toggle is off)
  const rateLimitData = formData.value.enable_rate_limit ? {
    rate_limit_5h: formData.value.rate_limit_5h && formData.value.rate_limit_5h > 0 ? formData.value.rate_limit_5h : 0,
    rate_limit_1d: formData.value.rate_limit_1d && formData.value.rate_limit_1d > 0 ? formData.value.rate_limit_1d : 0,
    rate_limit_7d: formData.value.rate_limit_7d && formData.value.rate_limit_7d > 0 ? formData.value.rate_limit_7d : 0,
  } : { rate_limit_5h: 0, rate_limit_1d: 0, rate_limit_7d: 0 }

  submitting.value = true
  try {
    if (showEditModal.value && selectedKey.value) {
      const updates: UpdateApiKeyRequest = {
        name: formData.value.name,
        group_ids: [...formData.value.group_ids],
        group_id: formData.value.group_ids[0],
        ip_whitelist: ipWhitelist,
        ip_blacklist: ipBlacklist,
        quota: quota,
        expires_at: expiresAt,
        rate_limit_5h: rateLimitData.rate_limit_5h,
        rate_limit_1d: rateLimitData.rate_limit_1d,
        rate_limit_7d: rateLimitData.rate_limit_7d,
      }
      if (shouldSubmitEditStatus(selectedKey.value, formData.value.status)) {
        updates.status = formData.value.status
      }
      await keysAPI.update(selectedKey.value.id, updates)
      appStore.showSuccess(t('keys.keyUpdatedSuccess'))
    } else {
      const customKey = formData.value.use_custom_key ? formData.value.custom_key : undefined
      await keysAPI.create(
        formData.value.name,
        formData.value.group_ids,
        customKey,
        ipWhitelist,
        ipBlacklist,
        quota,
        expiresInDays,
        rateLimitData
      )
      appStore.showSuccess(t('keys.keyCreatedSuccess'))
      // Only advance tour if active, on submit step, and creation succeeded
      if (onboardingStore.isCurrentStep('[data-tour="key-form-submit"]')) {
        onboardingStore.nextStep(500)
      }
    }
    closeModals()
    loadApiKeys()
  } catch (error: any) {
    const errorMsg = error.response?.data?.detail || t('keys.failedToSave')
    appStore.showError(errorMsg)
    // Don't advance tour on error
  } finally {
    submitting.value = false
  }
}

/**
 * 处理删除 API Key 的操作
 * 优化：错误处理改进，优先显示后端返回的具体错误消息（如权限不足等），
 * 若后端未返回消息则显示默认的国际化文本
 */
const handleDelete = async () => {
  if (!selectedKey.value) return

  try {
    await keysAPI.delete(selectedKey.value.id)
    appStore.showSuccess(t('keys.keyDeletedSuccess'))
    showDeleteDialog.value = false
    loadApiKeys()
  } catch (error: any) {
    // 优先使用后端返回的错误消息，提供更具体的错误信息给用户
    const errorMsg = error?.message || t('keys.failedToDelete')
    appStore.showError(errorMsg)
  }
}

const closeModals = () => {
  showCreateModal.value = false
  showEditModal.value = false
  selectedKey.value = null
  formData.value = {
    name: '',
    group_ids: [],
    status: 'active',
    use_custom_key: false,
    custom_key: '',
    enable_ip_restriction: false,
    ip_whitelist: '',
    ip_blacklist: '',
    enable_quota: false,
    quota: null,
    enable_rate_limit: false,
    rate_limit_5h: null,
    rate_limit_1d: null,
    rate_limit_7d: null,
    enable_expiration: false,
    expiration_preset: '30',
    expiration_date: ''
  }
}

// Show reset quota confirmation dialog
const confirmResetQuota = () => {
  showResetQuotaDialog.value = true
}

// Set expiration date based on quick select days
const setExpirationDays = (days: number) => {
  formData.value.expiration_preset = days.toString() as '7' | '30' | '90'
  const expDate = new Date()
  expDate.setDate(expDate.getDate() + days)
  formData.value.expiration_date = formatDateTimeLocal(expDate.toISOString())
}

// Reset quota used for an API key
const resetQuotaUsed = async () => {
  if (!selectedKey.value) return
  showResetQuotaDialog.value = false
  try {
    await keysAPI.update(selectedKey.value.id, { reset_quota: true })
    appStore.showSuccess(t('keys.quotaResetSuccess'))
    // Update local state
    if (selectedKey.value) {
      selectedKey.value.quota_used = 0
    }
  } catch (error: any) {
    const errorMsg = error.response?.data?.detail || t('keys.failedToResetQuota')
    appStore.showError(errorMsg)
  }
}

// Show reset rate limit confirmation dialog (from edit modal)
const confirmResetRateLimit = () => {
  showResetRateLimitDialog.value = true
}

// Show reset rate limit confirmation dialog (from table row)
const confirmResetRateLimitFromTable = (row: ApiKey) => {
  selectedKey.value = row
  showResetRateLimitDialog.value = true
}

// Reset rate limit usage for an API key
const resetRateLimitUsage = async () => {
  if (!selectedKey.value) return
  showResetRateLimitDialog.value = false
  try {
    await keysAPI.update(selectedKey.value.id, { reset_rate_limit_usage: true })
    appStore.showSuccess(t('keys.rateLimitResetSuccess'))
    // Refresh key data
    await loadApiKeys()
    // Update the editing key with fresh data
    const refreshedKey = apiKeys.value.find(k => k.id === selectedKey.value!.id)
    if (refreshedKey) {
      selectedKey.value = refreshedKey
    }
  } catch (error: any) {
    const errorMsg = error.response?.data?.detail || t('keys.failedToResetRateLimit')
    appStore.showError(errorMsg)
  }
}

const importToCcswitch = (row: ApiKey) => {
  executeCcsImport(row)
}

const executeCcsImport = (row: ApiKey) => {
  const baseUrl = publicSettings.value?.api_base_url || window.location.origin
  const platform = row.group?.platform || 'anthropic'

  const usageScript = `({
    request: {
      url: "{{baseUrl}}/v1/usage",
      method: "GET",
      headers: { "Authorization": "Bearer {{apiKey}}" }
    },
    extractor: function(response) {
      const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance;
      const unit = response?.unit ?? response?.quota?.unit ?? "points";
      return {
        isValid: response?.is_active ?? response?.isValid ?? true,
        remaining,
        unit
      };
    }
  })`
  const providerName = (publicSettings.value?.site_name || 'EasySub2api').trim() || 'EasySub2api'
  const deeplink = buildCcSwitchImportDeeplink({
    baseUrl,
    platform,
    clientType: 'claude',
    providerName,
    apiKey: row.key,
    usageScript,
    codexWebsocketEnabled: row.group?.ccs_codex_ws_enabled === true
  })

  try {
    window.open(deeplink, '_self')

    // Check if the protocol handler worked by detecting if we're still focused
    setTimeout(() => {
      if (document.hasFocus()) {
        // Still focused means the protocol handler likely failed
        appStore.showError(t('keys.ccSwitchNotInstalled'))
      }
    }, 100)
  } catch (error) {
    appStore.showError(t('keys.ccSwitchNotInstalled'))
  }
}

function formatResetTime(resetAt: string | null): string {
  if (!resetAt) return ''
  const diff = new Date(resetAt).getTime() - now.value.getTime()
  if (diff <= 0) return t('keys.resetNow')
  const days = Math.floor(diff / 86400000)
  const hours = Math.floor((diff % 86400000) / 3600000)
  const mins = Math.floor((diff % 3600000) / 60000)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

onMounted(() => {
  loadSavedColumns()
  loadApiKeys()
  loadGroups()
  loadUserGroupRates()
  loadPublicSettings()
  document.addEventListener('click', handleDocumentClick)
  resetTimer = setInterval(() => { now.value = new Date() }, 60000)
})

onUnmounted(() => {
  document.removeEventListener('click', handleDocumentClick)
  if (resetTimer) clearInterval(resetTimer)
})
</script>

<style scoped>
/* 工具条:筛选与操作合并为一行(与 /orders 筛选卡同标准) */
.keys-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.875rem 1.25rem;
}

.keys-toolbar__filters {
  display: flex;
  flex: 1 1 auto;
  flex-wrap: wrap;
  align-items: center;
  min-width: 0;
  gap: 0.75rem;
}

.keys-toolbar__actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.75rem;
  margin-left: auto;
}

.key-groups {
  display: flex;
  max-width: 32rem;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.375rem;
}
</style>
