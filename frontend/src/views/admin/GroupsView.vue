<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div
          class="groups-toolbar views-admin-groups-view__panel"
        >
          <!-- Left: fuzzy search + filters (can wrap to multiple lines) -->
          <div class="groups-toolbar__filters views-admin-groups-view__panel-2">
            <div class="views-admin-groups-view__panel-3">
              <Icon
                name="search"
                size="md"
                class="views-admin-groups-view__icon"
              />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('admin.groups.searchGroups')"
                class="views-admin-groups-view__field input"
                @input="handleSearch"
              />
            </div>
            <div class="groups-toolbar__platforms views-admin-groups-view__panel-4" role="tablist" :aria-label="t('admin.groups.allPlatforms')">
              <div class="views-admin-groups-view__panel-5">
                <button
                  v-for="option in platformFilterOptions"
                  :key="option.value || 'all'"
                  type="button"
                  role="tab"
                  :aria-selected="filters.platform === option.value"
                  :class="[
                    'views-admin-groups-view__action-22',
                    filters.platform === option.value
                      ? 'views-admin-groups-view__action-23'
                      : 'views-admin-groups-view__action-24'
                  ]"
                  @click="selectPlatformFilter(option.value)"
                >
                  <PlatformIcon v-if="option.value" :platform="option.value as GroupPlatform" size="xs" />
                  {{ option.label }}
                </button>
              </div>
            </div>
            <Select
              v-model="filters.status"
              :options="statusOptions"
              :placeholder="t('admin.groups.allStatus')"
              class="views-admin-groups-view__field-2"
              @change="handleFilterChange"
            />
            <Select
              v-model="filters.is_exclusive"
              :options="exclusiveOptions"
              :placeholder="t('admin.groups.allGroups')"
              class="views-admin-groups-view__field-3"
              @change="handleFilterChange"
            />
          </div>

          <!-- Right: actions -->
          <div
            class="groups-toolbar__actions views-admin-groups-view__panel-6"
          >
            <button
              @click="loadGroups"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh')"
            >
              <Icon
                name="refresh"
                size="md"
                :class="loading ? 'views-admin-groups-view__icon-8' : ''"
              />
            </button>
            <div class="views-admin-groups-view__panel-7" ref="columnDropdownRef">
              <button
                @click="showColumnDropdown = !showColumnDropdown"
                class="btn btn-secondary"
                :title="t('admin.groups.columnSettings')"
              >
                <Icon name="grid" size="md" class="views-admin-groups-view__icon-2" />
                <span class="views-admin-groups-view__text">{{
                  t("admin.groups.columnSettings")
                }}</span>
              </button>
              <div
                v-if="showColumnDropdown"
                class="views-admin-groups-view__panel-8"
              >
                <button
                  v-for="col in toggleableColumns"
                  :key="col.key"
                  @click="toggleColumn(col.key)"
                  class="views-admin-groups-view__action"
                >
                  <span>{{ col.label }}</span>
                  <Icon
                    v-if="isColumnVisible(col.key)"
                    name="check"
                    size="sm"
                    class="views-admin-groups-view__icon-3"
                    :stroke-width="2"
                  />
                </button>
              </div>
            </div>
            <button
              @click="openSortModal"
              class="btn btn-secondary"
              :title="t('admin.groups.sortOrder')"
            >
              <Icon name="arrowsUpDown" size="md" class="views-admin-groups-view__icon-2" />
              {{ t("admin.groups.sortOrder") }}
            </button>
            <button
              @click="openCreateModal"
              class="btn btn-primary"
              data-tour="groups-create-btn"
            >
              <Icon name="plus" size="md" class="views-admin-groups-view__icon-2" />
              {{ t("admin.groups.createGroup") }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="groups"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="sort_order"
          default-sort-order="asc"
          @sort="handleSort"
        >
          <template #cell-name="{ value }">
            <span class="views-admin-groups-view__text-2">{{
              value
            }}</span>
          </template>

          <template #cell-id="{ value }">
            <span class="views-admin-groups-view__text-3"
              >#{{ value }}</span
            >
          </template>

          <template #cell-platform="{ value }">
            <span
              :class="[
                'views-admin-groups-view__text-26',
                value === 'anthropic'
                  ? 'views-admin-groups-view__text-27'
                  : value === 'openai'
                    ? 'views-admin-groups-view__text-28'
                    : value === 'grok'
                        ? 'views-admin-groups-view__text-30'
                        : value === 'kimi'
                          ? 'views-admin-groups-view__text-31'
                          : value === 'zhipu'
                            ? 'views-admin-groups-view__text-32'
                            : value === 'deepseek'
                              ? 'views-admin-groups-view__text-33'
                              : 'views-admin-groups-view__text-34',
              ]"
            >
              <PlatformIcon :platform="value" size="xs" />
              {{ t("admin.groups.platforms." + value) }}
            </span>
          </template>

          <template #cell-billing_type="{ row }">
            <div class="views-admin-groups-view__panel-9">
              <!-- Type Badge -->
              <span
                :class="[
                  'views-admin-groups-view__text-35',
                  row.subscription_type === 'subscription'
                    ? 'views-admin-groups-view__text-36'
                    : 'views-admin-groups-view__text-37',
                ]"
              >
                {{
                  row.subscription_type === "subscription"
                    ? t("admin.groups.subscription.subscription")
                    : t("admin.groups.subscription.standard")
                }}
              </span>
              <!-- Subscription Limits - compact single line -->
              <div
                v-if="row.subscription_type === 'subscription'"
                class="views-admin-groups-view__panel-10"
              >
                <div
                  v-if="
                    groupLimitPoints(row, 'daily') ||
                    groupLimitPoints(row, 'weekly') ||
                    groupLimitPoints(row, 'monthly')
                  "
                  class="views-admin-groups-view__panel-11"
                >
                  <span v-if="groupLimitPoints(row, 'daily')" class="views-admin-groups-view__text-4">
                    <span
                      v-if="isUsageSummaryPending(row.id)"
                      class="views-admin-groups-view__text-5"
                      >—</span
                    >
                    <span
                      v-else
                      :class="
                        getQuotaUsageClass(
                          usageMap.get(row.id)?.today_cost ?? 0,
                          groupLimitPoints(row, 'daily')
                        )
                      "
                      >{{
                        formatPoints(usageMap.get(row.id)?.today_cost ?? 0)
                      }}</span
                    >
                    <span class="views-admin-groups-view__text-6">
                      / {{ formatPoints(groupLimitPoints(row, 'daily')) }}/{{
                        t("admin.groups.limitDay")
                      }}</span
                    >
                  </span>
                  <span
                    v-if="
                      groupLimitPoints(row, 'daily') &&
                      (groupLimitPoints(row, 'weekly') || groupLimitPoints(row, 'monthly'))
                    "
                    class="views-admin-groups-view__text-7"
                    >·</span
                  >
                  <span v-if="groupLimitPoints(row, 'weekly')" class="views-admin-groups-view__text-4"
                    >{{ formatPoints(groupLimitPoints(row, 'weekly')) }}/{{
                      t("admin.groups.limitWeek")
                    }}</span
                  >
                  <span
                    v-if="groupLimitPoints(row, 'weekly') && groupLimitPoints(row, 'monthly')"
                    class="views-admin-groups-view__text-7"
                    >·</span
                  >
                  <span v-if="groupLimitPoints(row, 'monthly')" class="views-admin-groups-view__text-4"
                    >{{ formatPoints(groupLimitPoints(row, 'monthly')) }}/{{
                      t("admin.groups.limitMonth")
                    }}</span
                  >
                </div>
                <span v-else class="views-admin-groups-view__text-6">{{
                  t("admin.groups.subscription.noLimit")
                }}</span>
                <div class="views-admin-groups-view__text-6">
                  {{ t("admin.groups.usageTotal") }}
                  <span class="views-admin-groups-view__text-8"
                    >{{
                      isUsageSummaryPending(row.id)
                        ? "—"
                        : formatPoints(usageMap.get(row.id)?.total_cost ?? 0)
                    }}</span
                  >
                </div>
              </div>
            </div>
          </template>

          <template #cell-rate_multiplier="{ value }">
            <span class="views-admin-groups-view__text-9"
              >{{ value }}x</span
            >
          </template>

          <template #cell-is_exclusive="{ value }">
            <span :class="['badge', value ? 'badge-primary' : 'badge-gray']">
              {{
                value ? t("admin.groups.exclusive") : t("admin.groups.public")
              }}
            </span>
          </template>

          <template #cell-account_count="{ row }">
            <div class="views-admin-groups-view__panel-12">
              <div>
                <span class="views-admin-groups-view__text-10">{{
                  t("admin.groups.accountsAvailable")
                }}</span>
                <span
                  class="views-admin-groups-view__text-11"
                  >{{ row.active_account_count || 0 }}</span
                >
                <span
                  class="views-admin-groups-view__text-12"
                  >{{ t("admin.groups.accountsUnit") }}</span
                >
              </div>
              <div v-if="row.rate_limited_account_count">
                <span class="views-admin-groups-view__text-10">{{
                  t("admin.groups.accountsRateLimited")
                }}</span>
                <span
                  class="views-admin-groups-view__text-13"
                  >{{ row.rate_limited_account_count }}</span
                >
                <span
                  class="views-admin-groups-view__text-12"
                  >{{ t("admin.groups.accountsUnit") }}</span
                >
              </div>
              <div>
                <span class="views-admin-groups-view__text-10">{{
                  t("admin.groups.accountsTotal")
                }}</span>
                <span
                  class="views-admin-groups-view__text-14"
                  >{{ row.account_count || 0 }}</span
                >
                <span
                  class="views-admin-groups-view__text-12"
                  >{{ t("admin.groups.accountsUnit") }}</span
                >
              </div>
            </div>
          </template>

          <template #cell-capacity="{ row }">
            <GroupCapacityBadge
              v-if="capacityMap.get(row.id)"
              :concurrency-used="capacityMap.get(row.id)!.concurrencyUsed"
              :concurrency-max="capacityMap.get(row.id)!.concurrencyMax"
              :sessions-used="capacityMap.get(row.id)!.sessionsUsed"
              :sessions-max="capacityMap.get(row.id)!.sessionsMax"
              :rpm-used="capacityMap.get(row.id)!.rpmUsed"
              :rpm-max="capacityMap.get(row.id)!.rpmMax"
            />
            <span v-else class="views-admin-groups-view__text-15">—</span>
          </template>

          <template #cell-usage="{ row }">
            <div v-if="isUsageSummaryPending(row.id)" class="views-admin-groups-view__text-15">—</div>
            <div v-else class="views-admin-groups-view__panel-12">
              <div class="views-admin-groups-view__text-10">
                <span class="views-admin-groups-view__text-6">{{
                  t("admin.groups.usageToday")
                }}</span>
                <span class="views-admin-groups-view__text-14">
                  {{ formatPoints(usageMap.get(row.id)?.today_cost ?? 0) }}
                </span>
              </div>
              <div class="views-admin-groups-view__text-10">
                <span class="views-admin-groups-view__text-6">{{
                  t("admin.groups.usageYesterday")
                }}</span>
                <span class="views-admin-groups-view__text-14">
                  {{ formatPoints(usageMap.get(row.id)?.yesterday_cost ?? 0) }}
                </span>
              </div>
              <div class="views-admin-groups-view__text-10">
                <span class="views-admin-groups-view__text-6">{{
                  t("admin.groups.usageTotal")
                }}</span>
                <span class="views-admin-groups-view__text-14">
                  {{ formatPoints(usageMap.get(row.id)?.total_cost ?? 0) }}
                </span>
              </div>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span
              :class="[
                'badge',
                value === 'active' ? 'badge-success' : 'badge-danger',
              ]"
            >
              {{ t("admin.accounts.status." + value) }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="views-admin-groups-view__panel-13">
              <button
                @click="handleEdit(row)"
                class="views-admin-groups-view__action-2"
              >
                <Icon name="edit" size="sm" />
                <span class="views-admin-groups-view__text-16">{{ t("common.edit") }}</span>
              </button>
              <button
                data-testid="group-duplicate"
                :title="
                  duplicatingGroupIds.has(row.id)
                    ? t('admin.groups.duplicating')
                    : t('admin.groups.duplicate')
                "
                :disabled="duplicatingGroupIds.has(row.id)"
                @click="handleDuplicate(row)"
                class="views-admin-groups-view__action-3"
              >
                <Icon name="copy" size="sm" />
                <span class="views-admin-groups-view__text-16">
                  {{
                    duplicatingGroupIds.has(row.id)
                      ? t("admin.groups.duplicating")
                      : t("admin.groups.duplicate")
                  }}
                </span>
              </button>
              <button
                v-if="row.platform === 'composite'"
                @click="handleCompositeRoutes(row)"
                class="views-admin-groups-view__action-4"
              >
                <Icon name="swap" size="sm" />
                <span class="views-admin-groups-view__text-16">{{
                  t("admin.groups.compositeRoutes.action")
                }}</span>
              </button>
              <button
                @click="handleRateMultipliers(row)"
                class="views-admin-groups-view__action-5"
              >
                <Icon name="dollar" size="sm" />
                <span class="views-admin-groups-view__text-16">{{
                  t("admin.groups.rateMultipliers")
                }}</span>
              </button>
              <button
                @click="handleLevelRates(row)"
                class="views-admin-groups-view__action-5"
              >
                <Icon name="chart" size="sm" />
                <span class="views-admin-groups-view__text-16">{{ t("admin.groups.levelRates.action") }}</span>
              </button>
              <button
                @click="handleRPMOverrides(row)"
                class="views-admin-groups-view__action-6"
              >
                <Icon name="bolt" size="sm" />
                <span class="views-admin-groups-view__text-16">{{
                  t("admin.groups.rpmOverrides")
                }}</span>
              </button>
              <button
                @click="handleDelete(row)"
                class="views-admin-groups-view__action-7"
              >
                <Icon name="trash" size="sm" />
                <span class="views-admin-groups-view__text-16">{{ t("common.delete") }}</span>
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.groups.noGroupsYet')"
              :description="t('admin.groups.createFirstGroup')"
              :action-text="t('admin.groups.createGroup')"
              @action="openCreateModal"
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

    <!-- Create Group Modal -->
    <BaseDialog
      :show="showCreateModal"
      :title="t('admin.groups.createGroup')"
      width="normal"
      @close="closeCreateModal"
    >
      <form
        id="create-group-form"
        @submit.prevent="handleCreateGroup"
        class="views-admin-groups-view__form"
      >
        <div>
          <label class="input-label">{{ t("admin.groups.form.name") }}</label>
          <input
            v-model="createForm.name"
            type="text"
            required
            class="input"
            :placeholder="t('admin.groups.enterGroupName')"
            data-tour="group-form-name"
          />
        </div>
        <div>
          <label class="input-label">{{
            t("admin.groups.form.description")
          }}</label>
          <textarea
            v-model="createForm.description"
            rows="3"
            class="input"
            :placeholder="t('admin.groups.optionalDescription')"
          ></textarea>
        </div>
        <div>
          <label class="input-label">{{
            t("admin.groups.form.platform")
          }}</label>
          <div class="views-admin-groups-view__panel-14" role="tablist" :aria-label="t('admin.groups.form.platform')" data-tour="group-form-platform">
            <div class="views-admin-groups-view__panel-5">
              <button
                v-for="option in platformOptions"
                :key="option.value"
                type="button"
                role="tab"
                :aria-selected="createForm.platform === option.value"
                :class="[
                  'views-admin-groups-view__action-25',
                  createForm.platform === option.value
                    ? 'views-admin-groups-view__action-23'
                    : 'views-admin-groups-view__action-24'
                ]"
                @click="selectCreatePlatform(option.value as GroupPlatform)"
              >
                <PlatformIcon :platform="option.value as GroupPlatform" size="xs" />
                {{ option.label }}
              </button>
            </div>
          </div>
          <p class="input-hint">{{ t("admin.groups.platformHint") }}</p>
        </div>
        <!-- 从分组复制账号 -->
        <div v-if="copyAccountsGroupOptions.length > 0">
          <div class="views-admin-groups-view__panel-15">
            <label class="views-admin-groups-view__label">
              {{ t("admin.groups.copyAccounts.title") }}
            </label>
            <div class="views-admin-groups-view__panel-16">
              <Icon
                name="questionCircle"
                size="sm"
                :stroke-width="2"
                class="views-admin-groups-view__icon-4"
              />
              <div
                class="views-admin-groups-view__panel-17"
              >
                <div
                  class="views-admin-groups-view__panel-18"
                >
                  <p class="views-admin-groups-view__description">
                    {{ t("admin.groups.copyAccounts.tooltip") }}
                  </p>
                  <div
                    class="views-admin-groups-view__panel-19"
                  ></div>
                </div>
              </div>
            </div>
          </div>
          <!-- 已选分组标签 -->
          <div
            v-if="createForm.copy_accounts_from_group_ids.length > 0"
            class="views-admin-groups-view__panel-20"
          >
            <span
              v-for="groupId in createForm.copy_accounts_from_group_ids"
              :key="groupId"
              class="views-admin-groups-view__text-17"
            >
              {{
                copyAccountsGroupOptions.find((o) => o.value === groupId)
                  ?.label || `#${groupId}`
              }}
              <button
                type="button"
                @click="
                  createForm.copy_accounts_from_group_ids =
                    createForm.copy_accounts_from_group_ids.filter(
                      (id) => id !== groupId,
                    )
                "
                class="views-admin-groups-view__action-8"
              >
                <Icon name="x" size="xs" />
              </button>
            </span>
          </div>
          <!-- 分组选择下拉 -->
          <Select
            :model-value="null"
            :options="[
              { value: null, label: t('admin.groups.copyAccounts.selectPlaceholder') },
              ...copyAccountsGroupOptions.map(opt => ({ ...opt, disabled: createForm.copy_accounts_from_group_ids.includes(opt.value) }))
            ]"
            @change="
              (value) => {
                const val = String(value ?? '');
                if (
                  val &&
                  !createForm.copy_accounts_from_group_ids.includes(val)
                ) {
                  createForm.copy_accounts_from_group_ids.push(val);
                }
              }
            "
          />
          <p class="input-hint">{{ t("admin.groups.copyAccounts.hint") }}</p>
        </div>
        <div>
          <label class="input-label">{{
            t("admin.groups.form.rateMultiplier")
          }}</label>
          <input
            v-model.number="createForm.rate_multiplier"
            type="number"
            step="0.001"
            min="0.001"
            required
            class="input"
            data-tour="group-form-multiplier"
          />
          <p class="input-hint">{{ t("admin.groups.rateMultiplierHint") }}</p>
        </div>
        <div>
          <label class="input-label">{{ t("admin.groups.form.rpmLimit") }}</label>
          <input
            v-model.number="createForm.rpm_limit"
            type="number"
            min="0"
            step="1"
            class="input"
            :placeholder="t('admin.groups.form.rpmLimitPlaceholder')"
          />
          <p class="input-hint">{{ t("admin.groups.form.rpmLimitHint") }}</p>
        </div>
        <ReasoningEffortPolicyFields
          v-if="supportsReasoningEffortPolicyPlatform(createForm.platform)"
          ref="createReasoningEffortPolicyRef"
          id-prefix="create-group-reasoning"
          :platform="createForm.platform"
          v-model:max-effort="createForm.max_reasoning_effort"
          v-model:mappings="createForm.reasoning_effort_mappings"
        />
        <div
          v-if="createForm.subscription_type !== 'subscription'"
          data-tour="group-form-exclusive"
        >
          <div class="views-admin-groups-view__panel-15">
            <label class="views-admin-groups-view__label">
              {{ t("admin.groups.form.exclusive") }}
            </label>
            <!-- Help Tooltip -->
            <div class="views-admin-groups-view__panel-16">
              <Icon
                name="questionCircle"
                size="sm"
                :stroke-width="2"
                class="views-admin-groups-view__icon-4"
              />
              <!-- Tooltip Popover -->
              <div
                class="views-admin-groups-view__panel-17"
              >
                <div
                  class="views-admin-groups-view__panel-18"
                >
                  <p class="views-admin-groups-view__description-2">
                    {{ t("admin.groups.exclusiveTooltip.title") }}
                  </p>
                  <p class="views-admin-groups-view__description-3">
                    {{ t("admin.groups.exclusiveTooltip.description") }}
                  </p>
                  <div class="views-admin-groups-view__panel-21">
                    <p class="views-admin-groups-view__description">
                      <span
                        class="views-admin-groups-view__text-18"
                        ><Icon name="lightbulb" size="xs" />
                        {{ t("admin.groups.exclusiveTooltip.example") }}</span
                      >
                      {{ t("admin.groups.exclusiveTooltip.exampleContent") }}
                    </p>
                  </div>
                  <!-- Arrow -->
                  <div
                    class="views-admin-groups-view__panel-19"
                  ></div>
                </div>
              </div>
            </div>
          </div>
          <div class="views-admin-groups-view__panel-22">
            <button
              type="button"
              @click="createForm.is_exclusive = !createForm.is_exclusive"
              :class="[
                'views-admin-groups-view__action-26',
                createForm.is_exclusive
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28',
              ]"
            >
              <span
                :class="[
                  'views-admin-groups-view__text-39',
                  createForm.is_exclusive ? 'toggle-thumb--on' : 'toggle-thumb--off',
                ]"
              />
            </button>
            <span class="views-admin-groups-view__text-19">
              {{
                createForm.is_exclusive
                  ? t("admin.groups.exclusive")
                  : t("admin.groups.public")
              }}
            </span>
          </div>
        </div>

        <!-- Subscription Configuration -->
        <div class="views-admin-groups-view__panel-23">
          <div>
            <label class="input-label">{{
              t("admin.groups.subscription.type")
            }}</label>
            <Select
              v-model="createForm.subscription_type"
              :options="subscriptionTypeOptions"
            />
            <p class="input-hint">
              {{ t("admin.groups.subscription.typeHint") }}
            </p>
          </div>

          <!-- Subscription limits (only show when subscription type is selected) -->
          <div
            v-if="createForm.subscription_type === 'subscription'"
            class="views-admin-groups-view__panel-24"
          >
            <div>
              <label class="input-label">{{
                t("admin.groups.subscription.dailyLimit")
              }}</label>
              <input
                v-model.number="createForm.daily_limit_points"
                type="number"
                step="0.01"
                min="0"
                class="input"
                :placeholder="t('admin.groups.subscription.noLimit')"
              />
            </div>
            <div>
              <label class="input-label">{{
                t("admin.groups.subscription.weeklyLimit")
              }}</label>
              <input
                v-model.number="createForm.weekly_limit_points"
                type="number"
                step="0.01"
                min="0"
                class="input"
                :placeholder="t('admin.groups.subscription.noLimit')"
              />
            </div>
            <div>
              <label class="input-label">{{
                t("admin.groups.subscription.monthlyLimit")
              }}</label>
              <input
                v-model.number="createForm.monthly_limit_points"
                type="number"
                step="0.01"
                min="0"
                class="input"
                :placeholder="t('admin.groups.subscription.noLimit')"
              />
            </div>
          </div>
        </div>

        <div class="views-admin-groups-view__panel-25">
          <div class="views-admin-groups-view__panel-26">
            <div>
              <label class="views-admin-groups-view__label">
                {{ t("admin.groups.modelsList.title") }}
              </label>
              <p class="views-admin-groups-view__description-4">
                {{ t("admin.groups.modelsList.hint") }}
              </p>
            </div>
            <button
              type="button"
              @click="createModelsListState.enabled = !createModelsListState.enabled"
              :class="[
                'views-admin-groups-view__action-29',
                createModelsListState.enabled
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28',
              ]"
            >
              <span
                :class="[
                  'views-admin-groups-view__text-39',
                  createModelsListState.enabled ? 'toggle-thumb--on' : 'toggle-thumb--off',
                ]"
              />
            </button>
          </div>
          <div
            v-if="createModelsListState.enabled"
            class="views-admin-groups-view__panel-27"
          >
            <div
              v-if="!createModelsListLoading && createModelsListState.items.length > 0"
              class="views-admin-groups-view__panel-28"
            >
              <span class="views-admin-groups-view__text-10">
                {{
                  t("admin.groups.modelsList.selectedSummary", {
                    selected: createModelsListSelectedCount,
                    total: createModelsListState.items.length,
                  })
                }}
              </span>
              <div class="views-admin-groups-view__panel-29">
                <button
                  type="button"
                  class="views-admin-groups-view__action-9"
                  @click="selectAllModelsListItems(createModelsListState)"
                >
                  {{ t("admin.groups.modelsList.selectAll") }}
                </button>
                <button
                  type="button"
                  class="views-admin-groups-view__action-10"
                  @click="invertModelsListSelection(createModelsListState)"
                >
                  {{ t("admin.groups.modelsList.invertSelection") }}
                </button>
              </div>
            </div>
            <div
              class="views-admin-groups-view__panel-30"
            >
              <p v-if="createModelsListLoading" class="views-admin-groups-view__description-5">
                {{ t("admin.groups.modelsList.loading") }}
              </p>
              <p
                v-else-if="createModelsListState.items.length === 0"
                class="views-admin-groups-view__description-5"
              >
                {{ t("admin.groups.modelsList.empty") }}
              </p>
              <div
                v-for="(item, index) in createModelsListState.items"
                :key="item.id"
                class="views-admin-groups-view__panel-31"
              >
                <input
                  v-model="item.selected"
                  type="checkbox"
                  class="views-admin-groups-view__field-4"
                />
                <span class="views-admin-groups-view__text-20">
                  {{ item.id }}
                </span>
                <button
                  type="button"
                  :disabled="index === 0"
                  class="views-admin-groups-view__action-11"
                  @click="moveCreateModelsListItem(index, index - 1)"
                >
                  <Icon name="arrowUp" size="sm" />
                </button>
                <button
                  type="button"
                  :disabled="index === createModelsListState.items.length - 1"
                  class="views-admin-groups-view__action-11"
                  @click="moveCreateModelsListItem(index, index + 1)"
                >
                  <Icon name="arrowDown" size="sm" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 图片生成计费配置 -->
        <div
          v-if="supportsImagePricingPlatform(createForm.platform)"
          class="views-admin-groups-view__panel-25"
        >
          <label
            class="views-admin-groups-view__label-2"
          >
            {{ t(imagePricingI18nKey(createForm.platform, "title")) }}
          </label>
          <p class="views-admin-groups-view__description-6">
            {{ t(imagePricingI18nKey(createForm.platform, "description")) }}
          </p>
          <div class="views-admin-groups-view__panel-32">
            <label class="views-admin-groups-view__label-3">
              <input
                v-model="createForm.allow_image_generation"
                type="checkbox"
                class="views-admin-groups-view__field-5"
              />
              {{ t(imagePricingI18nKey(createForm.platform, "allowImageGeneration")) }}
            </label>
            <label class="views-admin-groups-view__label-3">
              <input
                v-model="createForm.image_rate_independent"
                type="checkbox"
                class="views-admin-groups-view__field-5"
              />
              {{ t(imagePricingI18nKey(createForm.platform, "independentMultiplier")) }}
            </label>
          </div>
          <div
            v-if="createForm.image_rate_independent"
            class="views-admin-groups-view__panel-33"
          >
            <label class="input-label">{{
              t(imagePricingI18nKey(createForm.platform, "imageMultiplier"))
            }}</label>
            <input
              v-model.number="createForm.image_rate_multiplier"
              type="number"
              step="0.0001"
              min="0"
              class="input"
              placeholder="1"
            />
          </div>
          <div class="views-admin-groups-view__panel-34">
            <div>
              <label class="input-label">1K ($)</label>
              <input
                v-model.number="createForm.image_price_1k"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getImagePricePlaceholder(createForm.platform, 'image_price_1k')"
              />
            </div>
            <div>
              <label class="input-label">2K ($)</label>
              <input
                v-model.number="createForm.image_price_2k"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getImagePricePlaceholder(createForm.platform, 'image_price_2k')"
              />
            </div>
            <div>
              <label class="input-label">4K ($)</label>
              <input
                v-model.number="createForm.image_price_4k"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getImagePricePlaceholder(createForm.platform, 'image_price_4k')"
              />
            </div>
          </div>
          <p class="views-admin-groups-view__description-7">
            {{ t(imagePricingI18nKey(createForm.platform, "modeHint")) }}
          </p>
          <div class="views-admin-groups-view__panel-35">
            <div class="views-admin-groups-view__panel-36">
              {{ t(imagePricingI18nKey(createForm.platform, "finalPricePreview")) }}
            </div>
            <div class="views-admin-groups-view__panel-37">
              <div
                v-for="item in createImageFinalPricePreview"
                :key="item.label"
              >
                {{ item.label }}: {{ item.value }}
              </div>
            </div>
          </div>
        </div>

        <!-- 视频生成计费配置（仅 Grok 平台） -->
        <div
          v-if="supportsVideoPricingPlatform(createForm.platform)"
          class="views-admin-groups-view__panel-25"
        >
          <label
            class="views-admin-groups-view__label-2"
          >
            {{ t(videoPricingI18nKey("title")) }}
          </label>
          <p class="views-admin-groups-view__description-6">
            {{ t(videoPricingI18nKey("description")) }}
          </p>
          <div class="views-admin-groups-view__panel-33">
            <label class="views-admin-groups-view__label-3">
              <input
                v-model="createForm.video_rate_independent"
                type="checkbox"
                class="views-admin-groups-view__field-5"
              />
              {{ t(videoPricingI18nKey("independentMultiplier")) }}
            </label>
          </div>
          <div
            v-if="createForm.video_rate_independent"
            class="views-admin-groups-view__panel-33"
          >
            <label class="input-label">{{
              t(videoPricingI18nKey("videoMultiplier"))
            }}</label>
            <input
              v-model.number="createForm.video_rate_multiplier"
              type="number"
              step="0.0001"
              min="0"
              class="input"
              placeholder="1"
            />
          </div>
          <div class="views-admin-groups-view__panel-34">
            <div>
              <label class="input-label">480p ($/s)</label>
              <input
                v-model.number="createForm.video_price_480p"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getVideoPricePlaceholder(createForm.platform, 'video_price_480p')"
              />
            </div>
            <div>
              <label class="input-label">720p ($/s)</label>
              <input
                v-model.number="createForm.video_price_720p"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getVideoPricePlaceholder(createForm.platform, 'video_price_720p')"
              />
            </div>
            <div>
              <label class="input-label">1080p ($/s)</label>
              <input
                v-model.number="createForm.video_price_1080p"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getVideoPricePlaceholder(createForm.platform, 'video_price_1080p')"
              />
            </div>
          </div>
          <div
            class="views-admin-groups-view__panel-38"
            data-testid="create-grok-video-model-prices"
          >
            <p class="views-admin-groups-view__label">
              {{ t("admin.groups.videoPricing.modelOverridesTitle") }}
            </p>
            <p class="views-admin-groups-view__description-4">
              {{ t("admin.groups.videoPricing.modelOverridesDescription") }}
            </p>
            <div class="views-admin-groups-view__panel-40">
              <div
                v-for="family in videoModelPriceFamilyRows(createForm.video_model_prices)"
                :key="family.key"
                class="views-admin-groups-view__panel-41"
              >
                <div class="views-admin-groups-view__panel-42">
                  {{ family.label }}
                </div>
                <label
                  v-for="resolution in grokVideoPriceResolutions"
                  :key="resolution.key"
                  class="views-admin-groups-view__label-5"
                >
                  <span class="views-admin-groups-view__text-21">
                    {{ resolution.label }} ($/s)
                  </span>
                  <input
                    v-model.number="createForm.video_model_prices[family.key][resolution.key]"
                    type="number"
                    step="0.001"
                    min="0"
                    class="input"
                    :data-testid="`create-grok-video-price-${family.key}-${resolution.key}`"
                  />
                </label>
              </div>
            </div>
          </div>
          <p class="views-admin-groups-view__description-7">
            {{ t(videoPricingI18nKey("modeHint")) }}
          </p>
          <div class="views-admin-groups-view__panel-35">
            <div class="views-admin-groups-view__panel-36">
              {{ t(videoPricingI18nKey("finalPricePreview")) }}
            </div>
            <div class="views-admin-groups-view__panel-37">
              <div
                v-for="item in createVideoFinalPricePreview"
                :key="item.label"
              >
                {{ item.label }}: {{ item.value }}
              </div>
            </div>
          </div>
        </div>

        <!-- 高峰时段倍率配置（仅订阅类型分组） -->
        <div v-if="createForm.subscription_type === 'subscription'" class="views-admin-groups-view__panel-25">
          <div class="views-admin-groups-view__panel-32">
            <label class="views-admin-groups-view__label-3">
              <input
                v-model="createForm.peak_rate_enabled"
                type="checkbox"
                class="views-admin-groups-view__field-5"
              />
              <span>{{ t("admin.groups.peakRate.enable") }}</span>
            </label>
          </div>
          <div
            v-if="createForm.peak_rate_enabled"
            class="views-admin-groups-view__panel-43"
          >
            <div>
              <label class="input-label">{{ t("admin.groups.peakRate.peakStart") }}</label>
              <input
                v-model="createForm.peak_start"
                type="time"
                class="input"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.peakRate.peakEnd") }}</label>
              <input
                v-model="createForm.peak_end"
                type="time"
                class="input"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.peakRate.peakMultiplier") }}</label>
              <input
                v-model.number="createForm.peak_rate_multiplier"
                type="number"
                step="0.001"
                min="0"
                class="input"
                placeholder="1"
                :title="t('admin.groups.peakRate.multiplierHint')"
              />
            </div>
          </div>
        </div>

        <!-- 分组利润控制（五个平台 token 请求） -->
        <div v-if="isProfitControlPlatform(createForm.platform)" class="views-admin-groups-view__panel-25">
          <label class="views-admin-groups-view__label-3">
            <input
              v-model="createForm.profit_control_enabled"
              type="checkbox"
              class="views-admin-groups-view__field-5"
            />
            <span>{{ t("admin.groups.profitControl.enable") }}</span>
          </label>
          <p class="views-admin-groups-view__description-10">
            {{
              createForm.profit_control_enabled
                ? t("admin.groups.profitControl.enabledHint")
                : t("admin.groups.profitControl.disabledHint")
            }}
          </p>
          <div
            v-if="createForm.profit_control_enabled"
            class="views-admin-groups-view__panel-44"
          >
            <div>
              <label class="input-label">{{ t("admin.groups.profitControl.minMargin") }}</label>
              <input
                v-model.number="createForm.profit_min_margin_percent"
                type="number"
                step="0.1"
                min="0"
                max="99.99"
                class="input"
                placeholder="0"
                :title="t('admin.groups.profitControl.minMarginHint')"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.profitControl.safetyBuffer") }}</label>
              <input
                v-model.number="createForm.profit_safety_buffer_percent"
                type="number"
                step="0.1"
                min="0"
                max="99.99"
                class="input"
                placeholder="0"
                :title="t('admin.groups.profitControl.safetyBufferHint')"
              />
            </div>
          </div>
        </div>

        <!-- Claude Code 客户端限制（仅 anthropic 平台） -->
        <div v-if="createForm.platform === 'anthropic'" class="views-admin-groups-view__panel-25">
          <div class="views-admin-groups-view__panel-15">
            <label class="views-admin-groups-view__label">
              {{ t("admin.groups.claudeCode.title") }}
            </label>
            <!-- Help Tooltip -->
            <div class="views-admin-groups-view__panel-16">
              <Icon
                name="questionCircle"
                size="sm"
                :stroke-width="2"
                class="views-admin-groups-view__icon-4"
              />
              <div
                class="views-admin-groups-view__panel-17"
              >
                <div
                  class="views-admin-groups-view__panel-18"
                >
                  <p class="views-admin-groups-view__description">
                    {{ t("admin.groups.claudeCode.tooltip") }}
                  </p>
                  <div
                    class="views-admin-groups-view__panel-19"
                  ></div>
                </div>
              </div>
            </div>
          </div>
          <div class="views-admin-groups-view__panel-22">
            <button
              type="button"
              @click="
                createForm.claude_code_only = !createForm.claude_code_only
              "
              :class="[
                'views-admin-groups-view__action-26',
                createForm.claude_code_only
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28',
              ]"
            >
              <span
                :class="[
                  'views-admin-groups-view__text-39',
                  createForm.claude_code_only
                    ? 'toggle-thumb--on'
                    : 'toggle-thumb--off',
                ]"
              />
            </button>
            <span class="views-admin-groups-view__text-19">
              {{
                createForm.claude_code_only
                  ? t("admin.groups.claudeCode.enabled")
                  : t("admin.groups.claudeCode.disabled")
              }}
            </span>
          </div>
          <!-- 降级分组选择（仅当启用 claude_code_only 时显示） -->
          <div v-if="createForm.claude_code_only" class="views-admin-groups-view__panel-46">
            <label class="input-label">{{
              t("admin.groups.claudeCode.fallbackGroup")
            }}</label>
            <Select
              v-model="createForm.fallback_group_id"
              :options="fallbackGroupOptions"
              :placeholder="t('admin.groups.claudeCode.noFallback')"
            />
            <p class="input-hint">
              {{ t("admin.groups.claudeCode.fallbackHint") }}
            </p>
          </div>
        </div>

        <!-- Codex 网页搜索按次计费（仅 openai 平台） -->
        <div
          v-if="createForm.platform === 'openai'"
          class="views-admin-groups-view__panel-47"
        >
          <h4 class="views-admin-groups-view__heading">
            {{ t("admin.groups.webSearchPricing.title") }}
          </h4>
          <div>
            <label class="input-label">{{
              t("admin.groups.webSearchPricing.pricePerCall")
            }}</label>
            <input
              v-model.number="createForm.web_search_price_per_call"
              type="number"
              step="0.001"
              min="0"
              placeholder="0.01"
              class="input"
            />
            <p class="input-hint">
              {{ t("admin.groups.webSearchPricing.pricePerCallHint") }}
            </p>
            <div
              class="views-admin-groups-view__panel-48"
            >
              {{
                t("admin.groups.webSearchPricing.finalPricePreview", {
                  price: createWebSearchFinalPricePreview,
                })
              }}
            </div>
          </div>
        </div>


        <div class="views-admin-groups-view__panel-49">
          <div class="views-admin-groups-view__panel-50">
            <div>
              <h4 class="views-admin-groups-view__label">{{ t("admin.groups.modelPricing.title") }}</h4>
              <p class="views-admin-groups-view__description-4">{{ t("admin.groups.modelPricing.description") }}</p>
            </div>
            <button type="button" class="btn btn-secondary" @click="addGroupPricing(createForm.model_pricing)">
              <Icon name="plus" size="sm" class="views-admin-groups-view__icon-5" />{{ t("admin.groups.modelPricing.add") }}
            </button>
          </div>
          <label class="views-admin-groups-view__label-7">
            <input v-model="createForm.long_context_pricing_enabled" type="checkbox" class="views-admin-groups-view__field-7" />
            <span><span class="views-admin-groups-view__text-22">{{ t("admin.groups.modelPricing.longContext") }}</span><span class="views-admin-groups-view__text-23">{{ t("admin.groups.modelPricing.longContextHint") }}</span></span>
          </label>
          <div class="views-admin-groups-view__panel-51">
            <PricingEntryCard v-for="(entry, index) in createForm.model_pricing" :key="index" :entry="entry" :platform="createForm.platform" hide-token-intervals @update="createForm.model_pricing[index] = $event" @remove="createForm.model_pricing.splice(index, 1)" />
          </div>
        </div>

        <!-- Grok Voice 显式定价（仅 grok 平台） -->
        <div
          v-if="createForm.platform === 'grok'"
          class="views-admin-groups-view__panel-47"
        >
          <h4 class="views-admin-groups-view__heading-2">
            {{ t("admin.groups.explicitPricing.title") }}
          </h4>
          <p class="views-admin-groups-view__description-6">
            {{ t("admin.groups.explicitPricing.description") }}
          </p>
          <div class="views-admin-groups-view__panel-52">
            <div>
              <label class="input-label">{{ t("admin.groups.explicitPricing.searchPricePer1k") }}</label>
              <input
                v-model.number="createForm.search_price_per_1k"
                type="number"
                step="0.000001"
                min="0"
                class="input"
                :placeholder="t('admin.groups.explicitPricing.pricePlaceholder')"
                data-testid="create-search-price"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.voicePricing.audioRealtimePerMin") }}</label>
              <input
                v-model.number="createForm.audio_realtime_price_per_min"
                type="number"
                step="0.000001"
                min="0"
                class="input"
                :placeholder="t('admin.groups.voicePricing.pricePlaceholder')"
                data-testid="create-audio-realtime-price"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.voicePricing.audioTtsPerMillionChars") }}</label>
              <input
                v-model.number="createForm.audio_tts_price_per_million_chars"
                type="number"
                step="0.000001"
                min="0"
                class="input"
                :placeholder="t('admin.groups.voicePricing.pricePlaceholder')"
                data-testid="create-audio-tts-price"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.voicePricing.audioSttPerHour") }}</label>
              <input
                v-model.number="createForm.audio_stt_price_per_hour"
                type="number"
                step="0.000001"
                min="0"
                class="input"
                :placeholder="t('admin.groups.voicePricing.pricePlaceholder')"
                data-testid="create-audio-stt-price"
              />
            </div>
          </div>
        </div>
        <!-- OpenAI Live 开关（仅 openai 平台） -->
        <div
          v-if="createForm.platform === 'openai'"
          class="views-admin-groups-view__panel-47"
        >
          <h4 class="views-admin-groups-view__heading">
            {{ t("admin.groups.openaiLive.title") }}
          </h4>
          <div class="views-admin-groups-view__panel-53">
            <label class="views-admin-groups-view__label-8">{{
              t("admin.groups.openaiLive.allow")
            }}</label>
            <button
              type="button"
              @click="toggleLive('create')"
              class="views-admin-groups-view__action-12"
              :class="
                createForm.allow_live
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28'
              "
            >
              <span
                class="views-admin-groups-view__text-24"
                :class="createForm.allow_live ? 'toggle-thumb--on' : 'toggle-thumb--off'"
              />
            </button>
          </div>
          <p class="views-admin-groups-view__description-11">
            {{ t("admin.groups.openaiLive.hint") }}
          </p>
        </div>

        <!-- OpenAI Messages 调度配置（仅 openai 平台） -->
        <div
          v-if="createForm.platform === 'openai'"
          class="views-admin-groups-view__panel-47"
        >
          <h4 class="views-admin-groups-view__heading">
            {{ t("admin.groups.openaiMessages.title") }}
          </h4>

          <!-- 允许 Messages 调度开关 -->
          <div class="views-admin-groups-view__panel-53">
            <label class="views-admin-groups-view__label-8">{{
              t("admin.groups.openaiMessages.allowDispatch")
            }}</label>
            <button
              type="button"
              @click="
                createForm.allow_messages_dispatch =
                  !createForm.allow_messages_dispatch
              "
              class="views-admin-groups-view__action-12"
              :class="
                createForm.allow_messages_dispatch
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28'
              "
            >
              <span
                class="views-admin-groups-view__text-24"
                :class="
                  createForm.allow_messages_dispatch
                    ? 'toggle-thumb--on'
                    : 'toggle-thumb--off'
                "
              />
            </button>
          </div>
          <p class="views-admin-groups-view__description-11">
            {{ t("admin.groups.openaiMessages.allowDispatchHint") }}
          </p>

          <div v-if="createForm.allow_messages_dispatch" class="views-admin-groups-view__panel-46">
            <div
              class="views-admin-groups-view__panel-54"
            >
              <div
                class="views-admin-groups-view__panel-55"
              >
                <div class="views-admin-groups-view__panel-56">
                  <div class="views-admin-groups-view__panel-57"></div>
                  <label
                    class="views-admin-groups-view__label-9"
                    >{{
                      t("admin.groups.openaiMessages.familyMappingTitle")
                    }}</label
                  >
                </div>
                <p class="views-admin-groups-view__description-4">
                  {{ t("admin.groups.openaiMessages.familyMappingHint") }}
                </p>
              </div>
              <div class="views-admin-groups-view__panel-58">
                <div class="views-admin-groups-view__panel-59">
                  <div>
                    <label class="input-label">{{
                      t("admin.groups.openaiMessages.opusModel")
                    }}</label>
                    <input
                      v-model="createForm.opus_mapped_model"
                      type="text"
                      :placeholder="
                        t('admin.groups.openaiMessages.opusModelPlaceholder')
                      "
                      class="input"
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.groups.openaiMessages.sonnetModel")
                    }}</label>
                    <input
                      v-model="createForm.sonnet_mapped_model"
                      type="text"
                      :placeholder="
                        t('admin.groups.openaiMessages.sonnetModelPlaceholder')
                      "
                      class="input"
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.groups.openaiMessages.haikuModel")
                    }}</label>
                    <input
                      v-model="createForm.haiku_mapped_model"
                      type="text"
                      :placeholder="
                        t('admin.groups.openaiMessages.haikuModelPlaceholder')
                      "
                      class="input"
                    />
                  </div>
                </div>
              </div>
            </div>

            <div
              class="views-admin-groups-view__panel-60"
            >
              <div
                class="views-admin-groups-view__panel-61"
              >
                <div class="views-admin-groups-view__panel-62">
                  <div>
                    <div class="views-admin-groups-view__panel-56">
                      <div class="views-admin-groups-view__panel-63"></div>
                      <label
                        class="views-admin-groups-view__label-10"
                        >{{
                          t("admin.groups.openaiMessages.exactMappingTitle")
                        }}</label
                      >
                    </div>
                    <p
                      class="views-admin-groups-view__description-12"
                    >
                      {{ t("admin.groups.openaiMessages.exactMappingHint") }}
                    </p>
                  </div>
                </div>
              </div>

              <div class="views-admin-groups-view__panel-64">
                <div
                  v-if="createForm.exact_model_mappings.length === 0"
                  class="views-admin-groups-view__panel-65"
                >
                  <span>{{
                    t("admin.groups.openaiMessages.noExactMappings")
                  }}</span>
                  <button
                    type="button"
                    @click="addCreateMessagesDispatchMapping"
                    class="views-admin-groups-view__action-13"
                  >
                    <Icon name="plus" size="sm" />
                    {{ t("admin.groups.openaiMessages.addExactMapping") }}
                  </button>
                </div>

                <div v-else class="views-admin-groups-view__panel-66">
                  <div
                    v-for="row in createForm.exact_model_mappings"
                    :key="getCreateMessagesDispatchRowKey(row)"
                    class="views-admin-groups-view__panel-67"
                  >
                    <div class="views-admin-groups-view__panel-68">
                      <div
                        class="views-admin-groups-view__panel-69"
                      >
                        <div>
                          <label class="input-label">{{
                            t("admin.groups.openaiMessages.claudeModel")
                          }}</label>
                          <input
                            v-model="row.claude_model"
                            type="text"
                            :placeholder="
                              t(
                                'admin.groups.openaiMessages.claudeModelPlaceholder',
                              )
                            "
                            class="views-admin-groups-view__field-8 input"
                          />
                        </div>
                        <div
                          class="views-admin-groups-view__panel-70"
                        >
                          <Icon
                            name="arrowRight"
                            size="sm"
                            class="views-admin-groups-view__icon-6"
                          />
                        </div>
                        <div>
                          <label class="input-label">{{
                            t("admin.groups.openaiMessages.targetModel")
                          }}</label>
                          <input
                            v-model="row.target_model"
                            type="text"
                            :placeholder="
                              t(
                                'admin.groups.openaiMessages.targetModelPlaceholder',
                              )
                            "
                            class="views-admin-groups-view__field-8 input"
                          />
                        </div>
                      </div>
                      <button
                        type="button"
                        @click="removeCreateMessagesDispatchMapping(row)"
                        class="views-admin-groups-view__action-14"
                        :title="
                          t('admin.groups.openaiMessages.removeExactMapping')
                        "
                      >
                        <Icon name="trash" size="sm" />
                      </button>
                    </div>
                  </div>

                  <button
                    type="button"
                    @click="addCreateMessagesDispatchMapping"
                    class="views-admin-groups-view__action-15"
                  >
                    <Icon name="plus" size="sm" />
                    {{ t("admin.groups.openaiMessages.addExactMapping") }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 账号过滤控制 -->
        <div
          v-if="
            ['openai', 'anthropic'].includes(
              createForm.platform,
            )
          "
          class="views-admin-groups-view__panel-71"
        >
          <h4 class="views-admin-groups-view__heading">
            {{ t("admin.groups.accountFilters.title") }}
          </h4>

          <!-- require_oauth_only toggle -->
          <div class="views-admin-groups-view__panel-53">
            <div>
              <label class="views-admin-groups-view__label-8"
                >{{ t("admin.groups.accountFilters.oauthOnly") }}</label
              >
              <p class="views-admin-groups-view__description-13">
                {{
                  createForm.require_oauth_only
                    ? t("admin.groups.accountFilters.oauthOnlyEnabled")
                    : t("admin.groups.accountFilters.disabled")
                }}
              </p>
            </div>
            <button
              type="button"
              @click="
                createForm.require_oauth_only = !createForm.require_oauth_only
              "
              class="views-admin-groups-view__action-12"
              :class="
                createForm.require_oauth_only
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28'
              "
            >
              <span
                class="views-admin-groups-view__text-24"
                :class="
                  createForm.require_oauth_only
                    ? 'toggle-thumb--on'
                    : 'toggle-thumb--off'
                "
              />
            </button>
          </div>

          <!-- require_privacy_set toggle -->
          <div class="views-admin-groups-view__panel-53">
            <div>
              <label class="views-admin-groups-view__label-8"
                >{{ t("admin.groups.accountFilters.privacySetOnly") }}</label
              >
              <p class="views-admin-groups-view__description-13">
                {{
                  createForm.require_privacy_set
                    ? t("admin.groups.accountFilters.privacySetOnlyEnabled")
                    : t("admin.groups.accountFilters.disabled")
                }}
              </p>
            </div>
            <button
              type="button"
              @click="
                createForm.require_privacy_set = !createForm.require_privacy_set
              "
              class="views-admin-groups-view__action-12"
              :class="
                createForm.require_privacy_set
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28'
              "
            >
              <span
                class="views-admin-groups-view__text-24"
                :class="
                  createForm.require_privacy_set
                    ? 'toggle-thumb--on'
                    : 'toggle-thumb--off'
                "
              />
            </button>
          </div>
        </div>

        <!-- 无效请求兜底（仅 anthropic 平台，且非订阅分组） -->
        <div
          v-if="
            createForm.platform === 'anthropic' &&
            createForm.subscription_type !== 'subscription'
          "
          class="views-admin-groups-view__panel-25"
        >
          <label class="input-label">{{
            t("admin.groups.invalidRequestFallback.title")
          }}</label>
          <Select
            v-model="createForm.fallback_group_id_on_invalid_request"
            :options="invalidRequestFallbackOptions"
            :placeholder="t('admin.groups.invalidRequestFallback.noFallback')"
          />
          <p class="input-hint">
            {{ t("admin.groups.invalidRequestFallback.hint") }}
          </p>
        </div>

        <!-- 模型路由配置（仅 anthropic 平台） -->
        <div v-if="createForm.platform === 'anthropic'" class="views-admin-groups-view__panel-25">
          <div class="views-admin-groups-view__panel-15">
            <label class="views-admin-groups-view__label">
              {{ t("admin.groups.modelRouting.title") }}
            </label>
            <!-- Help Tooltip -->
            <div class="views-admin-groups-view__panel-16">
              <Icon
                name="questionCircle"
                size="sm"
                :stroke-width="2"
                class="views-admin-groups-view__icon-4"
              />
              <div
                class="views-admin-groups-view__panel-72"
              >
                <div
                  class="views-admin-groups-view__panel-18"
                >
                  <p class="views-admin-groups-view__description">
                    {{ t("admin.groups.modelRouting.tooltip") }}
                  </p>
                  <div
                    class="views-admin-groups-view__panel-19"
                  ></div>
                </div>
              </div>
            </div>
          </div>
          <!-- 启用开关 -->
          <div class="views-admin-groups-view__panel-73">
            <button
              type="button"
              @click="
                createForm.model_routing_enabled =
                  !createForm.model_routing_enabled
              "
              :class="[
                'views-admin-groups-view__action-26',
                createForm.model_routing_enabled
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28',
              ]"
            >
              <span
                :class="[
                  'views-admin-groups-view__text-39',
                  createForm.model_routing_enabled
                    ? 'toggle-thumb--on'
                    : 'toggle-thumb--off',
                ]"
              />
            </button>
            <span class="views-admin-groups-view__text-19">
              {{
                createForm.model_routing_enabled
                  ? t("admin.groups.modelRouting.enabled")
                  : t("admin.groups.modelRouting.disabled")
              }}
            </span>
          </div>
          <p
            v-if="!createForm.model_routing_enabled"
            class="views-admin-groups-view__description-6"
          >
            {{ t("admin.groups.modelRouting.disabledHint") }}
          </p>
          <p v-else class="views-admin-groups-view__description-6">
            {{ t("admin.groups.modelRouting.noRulesHint") }}
          </p>
          <!-- 路由规则列表（仅在启用时显示） -->
          <div v-if="createForm.model_routing_enabled" class="views-admin-groups-view__panel-66">
            <div
              v-for="rule in createModelRoutingRules"
              :key="getCreateRuleRenderKey(rule)"
              class="views-admin-groups-view__panel-74"
            >
              <div class="views-admin-groups-view__panel-75">
                <div class="views-admin-groups-view__panel-76">
                  <div>
                    <label class="views-admin-groups-view__text-16 input-label">{{
                      t("admin.groups.modelRouting.modelPattern")
                    }}</label>
                    <input
                      v-model="rule.pattern"
                      type="text"
                      class="views-admin-groups-view__field-9 input"
                      :placeholder="
                        t('admin.groups.modelRouting.modelPatternPlaceholder')
                      "
                    />
                  </div>
                  <div>
                    <label class="views-admin-groups-view__text-16 input-label">{{
                      t("admin.groups.modelRouting.accounts")
                    }}</label>
                    <!-- 已选账号标签 -->
                    <div
                      v-if="rule.accounts.length > 0"
                      class="views-admin-groups-view__panel-20"
                    >
                      <span
                        v-for="account in rule.accounts"
                        :key="account.id"
                        class="views-admin-groups-view__text-17"
                      >
                        {{ account.name }}
                        <button
                          type="button"
                          @click="removeSelectedAccount(rule, account.id)"
                          class="views-admin-groups-view__action-8"
                        >
                          <Icon name="x" size="xs" />
                        </button>
                      </span>
                    </div>
                    <!-- 账号搜索输入框 -->
                    <div class="views-admin-groups-view__panel-7 account-search-container">
                      <input
                        v-model="
                          accountSearchKeyword[getCreateRuleSearchKey(rule)]
                        "
                        type="text"
                        class="views-admin-groups-view__field-9 input"
                        :placeholder="
                          t(
                            'admin.groups.modelRouting.searchAccountPlaceholder',
                          )
                        "
                        @input="searchAccountsByRule(rule)"
                        @focus="onAccountSearchFocus(rule)"
                      />
                      <!-- 搜索结果下拉框 -->
                      <div
                        v-if="
                          showAccountDropdown[getCreateRuleSearchKey(rule)] &&
                          accountSearchResults[getCreateRuleSearchKey(rule)]
                            ?.length > 0
                        "
                        class="views-admin-groups-view__panel-77"
                      >
                        <button
                          v-for="account in accountSearchResults[
                            getCreateRuleSearchKey(rule)
                          ]"
                          :key="account.id"
                          type="button"
                          @click="selectAccount(rule, account)"
                          class="views-admin-groups-view__action-16"
                          :class="{
                            'views-admin-groups-view__action-30': rule.accounts.some(
                              (a) => a.id === account.id,
                            ),
                          }"
                          :disabled="
                            rule.accounts.some((a) => a.id === account.id)
                          "
                        >
                          <span>{{ account.name }}</span>
                          <span class="views-admin-groups-view__text-25"
                            >#{{ account.id }}</span
                          >
                        </button>
                      </div>
                    </div>
                    <p class="views-admin-groups-view__description-14">
                      {{ t("admin.groups.modelRouting.accountsHint") }}
                    </p>
                  </div>
                </div>
                <button
                  type="button"
                  @click="removeCreateRoutingRule(rule)"
                  class="views-admin-groups-view__action-17"
                  :title="t('admin.groups.modelRouting.removeRule')"
                >
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </div>
          </div>
          <!-- 添加规则按钮（仅在启用时显示） -->
          <button
            v-if="createForm.model_routing_enabled"
            type="button"
            @click="addCreateRoutingRule"
            class="views-admin-groups-view__action-18"
          >
            <Icon name="plus" size="sm" />
            {{ t("admin.groups.modelRouting.addRule") }}
          </button>
        </div>
      </form>

      <template #footer>
        <div class="views-admin-groups-view__panel-78">
          <button
            @click="closeCreateModal"
            type="button"
            class="btn btn-secondary"
          >
            {{ t("common.cancel") }}
          </button>
          <button
            type="submit"
            form="create-group-form"
            :disabled="submitting"
            class="btn btn-primary"
            data-tour="group-form-submit"
          >
            <svg
              v-if="submitting"
              class="views-admin-groups-view__icon-7"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="views-admin-groups-view__circle"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="views-admin-groups-view__path"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{ submitting ? t("admin.groups.creating") : t("common.create") }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Edit Group Modal -->
    <BaseDialog
      :show="showEditModal"
      :title="t('admin.groups.editGroup')"
      width="normal"
      @close="closeEditModal"
    >
      <form
        v-if="editingGroup"
        id="edit-group-form"
        @submit.prevent="handleUpdateGroup"
        class="views-admin-groups-view__form"
      >
        <div>
          <label class="input-label">{{ t("admin.groups.form.name") }}</label>
          <input
            v-model="editForm.name"
            type="text"
            required
            class="input"
            data-tour="edit-group-form-name"
          />
        </div>
        <div>
          <label class="input-label">{{
            t("admin.groups.form.description")
          }}</label>
          <textarea
            v-model="editForm.description"
            rows="3"
            class="input"
          ></textarea>
        </div>
        <div>
          <label class="input-label">{{
            t("admin.groups.form.platform")
          }}</label>
          <div class="views-admin-groups-view__panel-14" role="tablist" :aria-label="t('admin.groups.form.platform')" data-tour="group-form-platform">
            <div class="views-admin-groups-view__panel-79">
              <button
                v-for="option in platformOptions"
                :key="option.value"
                type="button"
                role="tab"
                disabled
                :aria-selected="editForm.platform === option.value"
                :class="[
                  'views-admin-groups-view__action-31',
                  editForm.platform === option.value
                    ? 'views-admin-groups-view__action-23'
                    : 'views-admin-groups-view__text-10'
                ]"
              >
                <PlatformIcon :platform="option.value as GroupPlatform" size="xs" />
                {{ option.label }}
              </button>
            </div>
          </div>
          <p class="input-hint">{{ t("admin.groups.platformNotEditable") }}</p>
        </div>
        <!-- 从分组复制账号（编辑时） -->
        <div v-if="copyAccountsGroupOptionsForEdit.length > 0">
          <div class="views-admin-groups-view__panel-15">
            <label class="views-admin-groups-view__label">
              {{ t("admin.groups.copyAccounts.title") }}
            </label>
            <div class="views-admin-groups-view__panel-16">
              <Icon
                name="questionCircle"
                size="sm"
                :stroke-width="2"
                class="views-admin-groups-view__icon-4"
              />
              <div
                class="views-admin-groups-view__panel-17"
              >
                <div
                  class="views-admin-groups-view__panel-18"
                >
                  <p class="views-admin-groups-view__description">
                    {{ t("admin.groups.copyAccounts.tooltipEdit") }}
                  </p>
                  <div
                    class="views-admin-groups-view__panel-19"
                  ></div>
                </div>
              </div>
            </div>
          </div>
          <!-- 已选分组标签 -->
          <div
            v-if="editForm.copy_accounts_from_group_ids.length > 0"
            class="views-admin-groups-view__panel-20"
          >
            <span
              v-for="groupId in editForm.copy_accounts_from_group_ids"
              :key="groupId"
              class="views-admin-groups-view__text-17"
            >
              {{
                copyAccountsGroupOptionsForEdit.find((o) => o.value === groupId)
                  ?.label || `#${groupId}`
              }}
              <button
                type="button"
                @click="
                  editForm.copy_accounts_from_group_ids =
                    editForm.copy_accounts_from_group_ids.filter(
                      (id) => id !== groupId,
                    )
                "
                class="views-admin-groups-view__action-8"
              >
                <Icon name="x" size="xs" />
              </button>
            </span>
          </div>
          <!-- 分组选择下拉 -->
          <Select
            :model-value="null"
            :options="[
              { value: null, label: t('admin.groups.copyAccounts.selectPlaceholder') },
              ...copyAccountsGroupOptionsForEdit.map(opt => ({ ...opt, disabled: editForm.copy_accounts_from_group_ids.includes(opt.value) }))
            ]"
            @change="
              (value) => {
                const val = String(value ?? '');
                if (
                  val &&
                  !editForm.copy_accounts_from_group_ids.includes(val)
                ) {
                  editForm.copy_accounts_from_group_ids.push(val);
                }
              }
            "
          />
          <p class="input-hint">
            {{ t("admin.groups.copyAccounts.hintEdit") }}
          </p>
        </div>
        <div>
          <label class="input-label">{{
            t("admin.groups.form.rateMultiplier")
          }}</label>
          <input
            v-model.number="editForm.rate_multiplier"
            type="number"
            step="0.001"
            min="0.001"
            required
            class="input"
            data-tour="group-form-multiplier"
          />
        </div>
        <div>
          <label class="input-label">{{ t("admin.groups.form.rpmLimit") }}</label>
          <input
            v-model.number="editForm.rpm_limit"
            type="number"
            min="0"
            step="1"
            class="input"
            :placeholder="t('admin.groups.form.rpmLimitPlaceholder')"
          />
          <p class="input-hint">{{ t("admin.groups.form.rpmLimitHint") }}</p>
        </div>
        <ReasoningEffortPolicyFields
          v-if="supportsReasoningEffortPolicyPlatform(editForm.platform)"
          ref="editReasoningEffortPolicyRef"
          id-prefix="edit-group-reasoning"
          :platform="editForm.platform"
          v-model:max-effort="editForm.max_reasoning_effort"
          v-model:mappings="editForm.reasoning_effort_mappings"
        />
        <div v-if="editForm.subscription_type !== 'subscription'">
          <div class="views-admin-groups-view__panel-15">
            <label class="views-admin-groups-view__label">
              {{ t("admin.groups.form.exclusive") }}
            </label>
            <!-- Help Tooltip -->
            <div class="views-admin-groups-view__panel-16">
              <Icon
                name="questionCircle"
                size="sm"
                :stroke-width="2"
                class="views-admin-groups-view__icon-4"
              />
              <!-- Tooltip Popover -->
              <div
                class="views-admin-groups-view__panel-17"
              >
                <div
                  class="views-admin-groups-view__panel-18"
                >
                  <p class="views-admin-groups-view__description-2">
                    {{ t("admin.groups.exclusiveTooltip.title") }}
                  </p>
                  <p class="views-admin-groups-view__description-3">
                    {{ t("admin.groups.exclusiveTooltip.description") }}
                  </p>
                  <div class="views-admin-groups-view__panel-21">
                    <p class="views-admin-groups-view__description">
                      <span
                        class="views-admin-groups-view__text-18"
                        ><Icon name="lightbulb" size="xs" />
                        {{ t("admin.groups.exclusiveTooltip.example") }}</span
                      >
                      {{ t("admin.groups.exclusiveTooltip.exampleContent") }}
                    </p>
                  </div>
                  <!-- Arrow -->
                  <div
                    class="views-admin-groups-view__panel-19"
                  ></div>
                </div>
              </div>
            </div>
          </div>
          <div class="views-admin-groups-view__panel-22">
            <button
              type="button"
              @click="editForm.is_exclusive = !editForm.is_exclusive"
              :class="[
                'views-admin-groups-view__action-26',
                editForm.is_exclusive
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28',
              ]"
            >
              <span
                :class="[
                  'views-admin-groups-view__text-39',
                  editForm.is_exclusive ? 'toggle-thumb--on' : 'toggle-thumb--off',
                ]"
              />
            </button>
            <span class="views-admin-groups-view__text-19">
              {{
                editForm.is_exclusive
                  ? t("admin.groups.exclusive")
                  : t("admin.groups.public")
              }}
            </span>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t("admin.groups.form.status") }}</label>
          <Select v-model="editForm.status" :options="editStatusOptions" />
        </div>

        <!-- Subscription Configuration -->
        <div class="views-admin-groups-view__panel-23">
          <div>
            <label class="input-label">{{
              t("admin.groups.subscription.type")
            }}</label>
            <Select
              v-model="editForm.subscription_type"
              :options="subscriptionTypeOptions"
              :disabled="true"
            />
            <p class="input-hint">
              {{ t("admin.groups.subscription.typeNotEditable") }}
            </p>
          </div>

          <!-- Subscription limits (only show when subscription type is selected) -->
          <div
            v-if="editForm.subscription_type === 'subscription'"
            class="views-admin-groups-view__panel-24"
          >
            <div>
              <label class="input-label">{{
                t("admin.groups.subscription.dailyLimit")
              }}</label>
              <input
                v-model.number="editForm.daily_limit_points"
                type="number"
                step="0.01"
                min="0"
                class="input"
                :placeholder="t('admin.groups.subscription.noLimit')"
              />
            </div>
            <div>
              <label class="input-label">{{
                t("admin.groups.subscription.weeklyLimit")
              }}</label>
              <input
                v-model.number="editForm.weekly_limit_points"
                type="number"
                step="0.01"
                min="0"
                class="input"
                :placeholder="t('admin.groups.subscription.noLimit')"
              />
            </div>
            <div>
              <label class="input-label">{{
                t("admin.groups.subscription.monthlyLimit")
              }}</label>
              <input
                v-model.number="editForm.monthly_limit_points"
                type="number"
                step="0.01"
                min="0"
                class="input"
                :placeholder="t('admin.groups.subscription.noLimit')"
              />
            </div>
          </div>
        </div>

        <div class="views-admin-groups-view__panel-25">
          <div class="views-admin-groups-view__panel-26">
            <div>
              <label class="views-admin-groups-view__label">
                {{ t("admin.groups.modelsList.title") }}
              </label>
              <p class="views-admin-groups-view__description-4">
                {{ t("admin.groups.modelsList.hint") }}
              </p>
            </div>
            <button
              type="button"
              @click="editModelsListState.enabled = !editModelsListState.enabled"
              :class="[
                'views-admin-groups-view__action-29',
                editModelsListState.enabled
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28',
              ]"
            >
              <span
                :class="[
                  'views-admin-groups-view__text-39',
                  editModelsListState.enabled ? 'toggle-thumb--on' : 'toggle-thumb--off',
                ]"
              />
            </button>
          </div>
          <div
            v-if="editModelsListState.enabled"
            class="views-admin-groups-view__panel-27"
          >
            <div
              v-if="!editModelsListLoading && editModelsListState.items.length > 0"
              class="views-admin-groups-view__panel-28"
            >
              <span class="views-admin-groups-view__text-10">
                {{
                  t("admin.groups.modelsList.selectedSummary", {
                    selected: editModelsListSelectedCount,
                    total: editModelsListState.items.length,
                  })
                }}
              </span>
              <div class="views-admin-groups-view__panel-29">
                <button
                  type="button"
                  class="views-admin-groups-view__action-9"
                  @click="selectAllModelsListItems(editModelsListState)"
                >
                  {{ t("admin.groups.modelsList.selectAll") }}
                </button>
                <button
                  type="button"
                  class="views-admin-groups-view__action-10"
                  @click="invertModelsListSelection(editModelsListState)"
                >
                  {{ t("admin.groups.modelsList.invertSelection") }}
                </button>
              </div>
            </div>
            <div
              class="views-admin-groups-view__panel-30"
            >
              <p v-if="editModelsListLoading" class="views-admin-groups-view__description-5">
                {{ t("admin.groups.modelsList.loading") }}
              </p>
              <p
                v-else-if="editModelsListState.items.length === 0"
                class="views-admin-groups-view__description-5"
              >
                {{ t("admin.groups.modelsList.empty") }}
              </p>
              <div
                v-for="(item, index) in editModelsListState.items"
                :key="item.id"
                class="views-admin-groups-view__panel-31"
              >
                <input
                  v-model="item.selected"
                  type="checkbox"
                  class="views-admin-groups-view__field-4"
                />
                <span class="views-admin-groups-view__text-20">
                  {{ item.id }}
                </span>
                <button
                  type="button"
                  :disabled="index === 0"
                  class="views-admin-groups-view__action-11"
                  @click="moveEditModelsListItem(index, index - 1)"
                >
                  <Icon name="arrowUp" size="sm" />
                </button>
                <button
                  type="button"
                  :disabled="index === editModelsListState.items.length - 1"
                  class="views-admin-groups-view__action-11"
                  @click="moveEditModelsListItem(index, index + 1)"
                >
                  <Icon name="arrowDown" size="sm" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 图片生成计费配置 -->
        <div
          v-if="supportsImagePricingPlatform(editForm.platform)"
          class="views-admin-groups-view__panel-25"
        >
          <label
            class="views-admin-groups-view__label-2"
          >
            {{ t(imagePricingI18nKey(editForm.platform, "title")) }}
          </label>
          <p class="views-admin-groups-view__description-6">
            {{ t(imagePricingI18nKey(editForm.platform, "description")) }}
          </p>
          <div class="views-admin-groups-view__panel-32">
            <label class="views-admin-groups-view__label-3">
              <input
                v-model="editForm.allow_image_generation"
                type="checkbox"
                class="views-admin-groups-view__field-5"
              />
              {{ t(imagePricingI18nKey(editForm.platform, "allowImageGeneration")) }}
            </label>
            <label class="views-admin-groups-view__label-3">
              <input
                v-model="editForm.image_rate_independent"
                type="checkbox"
                class="views-admin-groups-view__field-5"
              />
              {{ t(imagePricingI18nKey(editForm.platform, "independentMultiplier")) }}
            </label>
          </div>
          <div
            v-if="editForm.image_rate_independent"
            class="views-admin-groups-view__panel-33"
          >
            <label class="input-label">{{
              t(imagePricingI18nKey(editForm.platform, "imageMultiplier"))
            }}</label>
            <input
              v-model.number="editForm.image_rate_multiplier"
              type="number"
              step="0.0001"
              min="0"
              class="input"
              placeholder="1"
            />
          </div>
          <div class="views-admin-groups-view__panel-34">
            <div>
              <label class="input-label">1K ($)</label>
              <input
                v-model.number="editForm.image_price_1k"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getImagePricePlaceholder(editForm.platform, 'image_price_1k')"
              />
            </div>
            <div>
              <label class="input-label">2K ($)</label>
              <input
                v-model.number="editForm.image_price_2k"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getImagePricePlaceholder(editForm.platform, 'image_price_2k')"
              />
            </div>
            <div>
              <label class="input-label">4K ($)</label>
              <input
                v-model.number="editForm.image_price_4k"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getImagePricePlaceholder(editForm.platform, 'image_price_4k')"
              />
            </div>
          </div>
          <p class="views-admin-groups-view__description-7">
            {{ t(imagePricingI18nKey(editForm.platform, "modeHint")) }}
          </p>
          <div class="views-admin-groups-view__panel-35">
            <div class="views-admin-groups-view__panel-36">
              {{ t(imagePricingI18nKey(editForm.platform, "finalPricePreview")) }}
            </div>
            <div class="views-admin-groups-view__panel-37">
              <div
                v-for="item in editImageFinalPricePreview"
                :key="item.label"
              >
                {{ item.label }}: {{ item.value }}
              </div>
            </div>
          </div>
        </div>

        <!-- 视频生成计费配置（仅 Grok 平台） -->
        <div
          v-if="supportsVideoPricingPlatform(editForm.platform)"
          class="views-admin-groups-view__panel-25"
        >
          <label
            class="views-admin-groups-view__label-2"
          >
            {{ t(videoPricingI18nKey("title")) }}
          </label>
          <p class="views-admin-groups-view__description-6">
            {{ t(videoPricingI18nKey("description")) }}
          </p>
          <div class="views-admin-groups-view__panel-33">
            <label class="views-admin-groups-view__label-3">
              <input
                v-model="editForm.video_rate_independent"
                type="checkbox"
                class="views-admin-groups-view__field-5"
              />
              {{ t(videoPricingI18nKey("independentMultiplier")) }}
            </label>
          </div>
          <div
            v-if="editForm.video_rate_independent"
            class="views-admin-groups-view__panel-33"
          >
            <label class="input-label">{{
              t(videoPricingI18nKey("videoMultiplier"))
            }}</label>
            <input
              v-model.number="editForm.video_rate_multiplier"
              type="number"
              step="0.0001"
              min="0"
              class="input"
              placeholder="1"
            />
          </div>
          <div class="views-admin-groups-view__panel-34">
            <div>
              <label class="input-label">480p ($/s)</label>
              <input
                v-model.number="editForm.video_price_480p"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getVideoPricePlaceholder(editForm.platform, 'video_price_480p')"
              />
            </div>
            <div>
              <label class="input-label">720p ($/s)</label>
              <input
                v-model.number="editForm.video_price_720p"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getVideoPricePlaceholder(editForm.platform, 'video_price_720p')"
              />
            </div>
            <div>
              <label class="input-label">1080p ($/s)</label>
              <input
                v-model.number="editForm.video_price_1080p"
                type="number"
                step="0.001"
                min="0"
                class="input"
                :placeholder="getVideoPricePlaceholder(editForm.platform, 'video_price_1080p')"
              />
            </div>
          </div>
          <div
            class="views-admin-groups-view__panel-38"
            data-testid="edit-grok-video-model-prices"
          >
            <p class="views-admin-groups-view__label">
              {{ t("admin.groups.videoPricing.modelOverridesTitle") }}
            </p>
            <p class="views-admin-groups-view__description-4">
              {{ t("admin.groups.videoPricing.modelOverridesDescription") }}
            </p>
            <div class="views-admin-groups-view__panel-40">
              <div
                v-for="family in videoModelPriceFamilyRows(editForm.video_model_prices)"
                :key="family.key"
                class="views-admin-groups-view__panel-41"
              >
                <div class="views-admin-groups-view__panel-42">
                  {{ family.label }}
                </div>
                <label
                  v-for="resolution in grokVideoPriceResolutions"
                  :key="resolution.key"
                  class="views-admin-groups-view__label-5"
                >
                  <span class="views-admin-groups-view__text-21">
                    {{ resolution.label }} ($/s)
                  </span>
                  <input
                    v-model.number="editForm.video_model_prices[family.key][resolution.key]"
                    type="number"
                    step="0.001"
                    min="0"
                    class="input"
                    :data-testid="`edit-grok-video-price-${family.key}-${resolution.key}`"
                  />
                </label>
              </div>
            </div>
          </div>
          <p class="views-admin-groups-view__description-7">
            {{ t(videoPricingI18nKey("modeHint")) }}
          </p>
          <div class="views-admin-groups-view__panel-35">
            <div class="views-admin-groups-view__panel-36">
              {{ t(videoPricingI18nKey("finalPricePreview")) }}
            </div>
            <div class="views-admin-groups-view__panel-37">
              <div
                v-for="item in editVideoFinalPricePreview"
                :key="item.label"
              >
                {{ item.label }}: {{ item.value }}
              </div>
            </div>
          </div>
        </div>

        <!-- 高峰时段倍率配置（仅订阅类型分组） -->
        <div v-if="editForm.subscription_type === 'subscription'" class="views-admin-groups-view__panel-25">
          <div class="views-admin-groups-view__panel-32">
            <label class="views-admin-groups-view__label-3">
              <input
                v-model="editForm.peak_rate_enabled"
                type="checkbox"
                class="views-admin-groups-view__field-5"
              />
              <span>{{ t("admin.groups.peakRate.enable") }}</span>
            </label>
          </div>
          <div
            v-if="editForm.peak_rate_enabled"
            class="views-admin-groups-view__panel-43"
          >
            <div>
              <label class="input-label">{{ t("admin.groups.peakRate.peakStart") }}</label>
              <input
                v-model="editForm.peak_start"
                type="time"
                class="input"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.peakRate.peakEnd") }}</label>
              <input
                v-model="editForm.peak_end"
                type="time"
                class="input"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.peakRate.peakMultiplier") }}</label>
              <input
                v-model.number="editForm.peak_rate_multiplier"
                type="number"
                step="0.001"
                min="0"
                class="input"
                placeholder="1"
                :title="t('admin.groups.peakRate.multiplierHint')"
              />
            </div>
          </div>
        </div>

        <!-- 分组利润控制（五个平台 token 请求） -->
        <div v-if="isProfitControlPlatform(editForm.platform)" class="views-admin-groups-view__panel-25">
          <label class="views-admin-groups-view__label-3">
            <input
              v-model="editForm.profit_control_enabled"
              type="checkbox"
              class="views-admin-groups-view__field-5"
            />
            <span>{{ t("admin.groups.profitControl.enable") }}</span>
          </label>
          <p class="views-admin-groups-view__description-10">
            {{
              editForm.profit_control_enabled
                ? t("admin.groups.profitControl.enabledHint")
                : t("admin.groups.profitControl.disabledHint")
            }}
          </p>
          <div
            v-if="editForm.profit_control_enabled"
            class="views-admin-groups-view__panel-44"
          >
            <div>
              <label class="input-label">{{ t("admin.groups.profitControl.minMargin") }}</label>
              <input
                v-model.number="editForm.profit_min_margin_percent"
                type="number"
                step="0.1"
                min="0"
                max="99.99"
                class="input"
                placeholder="0"
                :title="t('admin.groups.profitControl.minMarginHint')"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.profitControl.safetyBuffer") }}</label>
              <input
                v-model.number="editForm.profit_safety_buffer_percent"
                type="number"
                step="0.1"
                min="0"
                max="99.99"
                class="input"
                placeholder="0"
                :title="t('admin.groups.profitControl.safetyBufferHint')"
              />
            </div>
          </div>
        </div>

        <!-- Claude Code 客户端限制（仅 anthropic 平台） -->
        <div v-if="editForm.platform === 'anthropic'" class="views-admin-groups-view__panel-25">
          <div class="views-admin-groups-view__panel-15">
            <label class="views-admin-groups-view__label">
              {{ t("admin.groups.claudeCode.title") }}
            </label>
            <!-- Help Tooltip -->
            <div class="views-admin-groups-view__panel-16">
              <Icon
                name="questionCircle"
                size="sm"
                :stroke-width="2"
                class="views-admin-groups-view__icon-4"
              />
              <div
                class="views-admin-groups-view__panel-17"
              >
                <div
                  class="views-admin-groups-view__panel-18"
                >
                  <p class="views-admin-groups-view__description">
                    {{ t("admin.groups.claudeCode.tooltip") }}
                  </p>
                  <div
                    class="views-admin-groups-view__panel-19"
                  ></div>
                </div>
              </div>
            </div>
          </div>
          <div class="views-admin-groups-view__panel-22">
            <button
              type="button"
              @click="editForm.claude_code_only = !editForm.claude_code_only"
              :class="[
                'views-admin-groups-view__action-26',
                editForm.claude_code_only
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28',
              ]"
            >
              <span
                :class="[
                  'views-admin-groups-view__text-39',
                  editForm.claude_code_only ? 'toggle-thumb--on' : 'toggle-thumb--off',
                ]"
              />
            </button>
            <span class="views-admin-groups-view__text-19">
              {{
                editForm.claude_code_only
                  ? t("admin.groups.claudeCode.enabled")
                  : t("admin.groups.claudeCode.disabled")
              }}
            </span>
          </div>
          <!-- 降级分组选择（仅当启用 claude_code_only 时显示） -->
          <div v-if="editForm.claude_code_only" class="views-admin-groups-view__panel-46">
            <label class="input-label">{{
              t("admin.groups.claudeCode.fallbackGroup")
            }}</label>
            <Select
              v-model="editForm.fallback_group_id"
              :options="fallbackGroupOptionsForEdit"
              :placeholder="t('admin.groups.claudeCode.noFallback')"
            />
            <p class="input-hint">
              {{ t("admin.groups.claudeCode.fallbackHint") }}
            </p>
          </div>
        </div>

        <!-- Codex 网页搜索按次计费（仅 openai 平台） -->
        <div
          v-if="editForm.platform === 'openai'"
          class="views-admin-groups-view__panel-47"
        >
          <h4 class="views-admin-groups-view__heading">
            {{ t("admin.groups.webSearchPricing.title") }}
          </h4>
          <div>
            <label class="input-label">{{
              t("admin.groups.webSearchPricing.pricePerCall")
            }}</label>
            <input
              v-model.number="editForm.web_search_price_per_call"
              type="number"
              step="0.001"
              min="0"
              placeholder="0.01"
              class="input"
            />
            <p class="input-hint">
              {{ t("admin.groups.webSearchPricing.pricePerCallHint") }}
            </p>
            <div
              class="views-admin-groups-view__panel-48"
            >
              {{
                t("admin.groups.webSearchPricing.finalPricePreview", {
                  price: editWebSearchFinalPricePreview,
                })
              }}
            </div>
          </div>
        </div>


        <div class="views-admin-groups-view__panel-49">
          <div class="views-admin-groups-view__panel-50">
            <div>
              <h4 class="views-admin-groups-view__label">{{ t("admin.groups.modelPricing.title") }}</h4>
              <p class="views-admin-groups-view__description-4">{{ t("admin.groups.modelPricing.description") }}</p>
            </div>
            <button type="button" class="btn btn-secondary" @click="addGroupPricing(editForm.model_pricing)">
              <Icon name="plus" size="sm" class="views-admin-groups-view__icon-5" />{{ t("admin.groups.modelPricing.add") }}
            </button>
          </div>
          <label class="views-admin-groups-view__label-7">
            <input v-model="editForm.long_context_pricing_enabled" type="checkbox" class="views-admin-groups-view__field-7" />
            <span><span class="views-admin-groups-view__text-22">{{ t("admin.groups.modelPricing.longContext") }}</span><span class="views-admin-groups-view__text-23">{{ t("admin.groups.modelPricing.longContextHint") }}</span></span>
          </label>
          <div class="views-admin-groups-view__panel-51">
            <PricingEntryCard v-for="(entry, index) in editForm.model_pricing" :key="index" :entry="entry" :platform="editForm.platform" hide-token-intervals @update="editForm.model_pricing[index] = $event" @remove="editForm.model_pricing.splice(index, 1)" />
          </div>
        </div>

        <!-- Grok Voice 显式定价（仅 grok 平台） -->
        <div
          v-if="editForm.platform === 'grok'"
          class="views-admin-groups-view__panel-47"
        >
          <h4 class="views-admin-groups-view__heading-2">
            {{ t("admin.groups.explicitPricing.title") }}
          </h4>
          <p class="views-admin-groups-view__description-6">
            {{ t("admin.groups.explicitPricing.description") }}
          </p>
          <div class="views-admin-groups-view__panel-52">
            <div>
              <label class="input-label">{{ t("admin.groups.explicitPricing.searchPricePer1k") }}</label>
              <input
                v-model.number="editForm.search_price_per_1k"
                type="number"
                step="0.000001"
                min="0"
                class="input"
                :placeholder="t('admin.groups.explicitPricing.pricePlaceholder')"
                data-testid="edit-search-price"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.voicePricing.audioRealtimePerMin") }}</label>
              <input
                v-model.number="editForm.audio_realtime_price_per_min"
                type="number"
                step="0.000001"
                min="0"
                class="input"
                :placeholder="t('admin.groups.voicePricing.pricePlaceholder')"
                data-testid="edit-audio-realtime-price"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.voicePricing.audioTtsPerMillionChars") }}</label>
              <input
                v-model.number="editForm.audio_tts_price_per_million_chars"
                type="number"
                step="0.000001"
                min="0"
                class="input"
                :placeholder="t('admin.groups.voicePricing.pricePlaceholder')"
                data-testid="edit-audio-tts-price"
              />
            </div>
            <div>
              <label class="input-label">{{ t("admin.groups.voicePricing.audioSttPerHour") }}</label>
              <input
                v-model.number="editForm.audio_stt_price_per_hour"
                type="number"
                step="0.000001"
                min="0"
                class="input"
                :placeholder="t('admin.groups.voicePricing.pricePlaceholder')"
                data-testid="edit-audio-stt-price"
              />
            </div>
          </div>
        </div>
        <!-- OpenAI Live 开关（仅 openai 平台） -->
        <div
          v-if="editForm.platform === 'openai'"
          class="views-admin-groups-view__panel-47"
        >
          <h4 class="views-admin-groups-view__heading">
            {{ t("admin.groups.openaiLive.title") }}
          </h4>
          <div class="views-admin-groups-view__panel-53">
            <label class="views-admin-groups-view__label-8">{{
              t("admin.groups.openaiLive.allow")
            }}</label>
            <button
              type="button"
              @click="toggleLive('edit')"
              class="views-admin-groups-view__action-12"
              :class="
                editForm.allow_live
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28'
              "
            >
              <span
                class="views-admin-groups-view__text-24"
                :class="editForm.allow_live ? 'toggle-thumb--on' : 'toggle-thumb--off'"
              />
            </button>
          </div>
          <p class="views-admin-groups-view__description-11">
            {{ t("admin.groups.openaiLive.hint") }}
          </p>
        </div>

        <!-- OpenAI Messages 调度配置（仅 openai 平台） -->
        <div
          v-if="editForm.platform === 'openai'"
          class="views-admin-groups-view__panel-47"
        >
          <h4 class="views-admin-groups-view__heading">
            {{ t("admin.groups.openaiMessages.title") }}
          </h4>

          <!-- 允许 Messages 调度开关 -->
          <div class="views-admin-groups-view__panel-53">
            <label class="views-admin-groups-view__label-8">{{
              t("admin.groups.openaiMessages.allowDispatch")
            }}</label>
            <button
              type="button"
              @click="
                editForm.allow_messages_dispatch =
                  !editForm.allow_messages_dispatch
              "
              class="views-admin-groups-view__action-12"
              :class="
                editForm.allow_messages_dispatch
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28'
              "
            >
              <span
                class="views-admin-groups-view__text-24"
                :class="
                  editForm.allow_messages_dispatch
                    ? 'toggle-thumb--on'
                    : 'toggle-thumb--off'
                "
              />
            </button>
          </div>
          <p class="views-admin-groups-view__description-11">
            {{ t("admin.groups.openaiMessages.allowDispatchHint") }}
          </p>

          <div v-if="editForm.allow_messages_dispatch" class="views-admin-groups-view__panel-46">
            <div
              class="views-admin-groups-view__panel-54"
            >
              <div
                class="views-admin-groups-view__panel-55"
              >
                <div class="views-admin-groups-view__panel-56">
                  <div class="views-admin-groups-view__panel-57"></div>
                  <label
                    class="views-admin-groups-view__label-9"
                    >{{
                      t("admin.groups.openaiMessages.familyMappingTitle")
                    }}</label
                  >
                </div>
                <p class="views-admin-groups-view__description-4">
                  {{ t("admin.groups.openaiMessages.familyMappingHint") }}
                </p>
              </div>
              <div class="views-admin-groups-view__panel-58">
                <div class="views-admin-groups-view__panel-59">
                  <div>
                    <label class="input-label">{{
                      t("admin.groups.openaiMessages.opusModel")
                    }}</label>
                    <input
                      v-model="editForm.opus_mapped_model"
                      type="text"
                      :placeholder="
                        t('admin.groups.openaiMessages.opusModelPlaceholder')
                      "
                      class="input"
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.groups.openaiMessages.sonnetModel")
                    }}</label>
                    <input
                      v-model="editForm.sonnet_mapped_model"
                      type="text"
                      :placeholder="
                        t('admin.groups.openaiMessages.sonnetModelPlaceholder')
                      "
                      class="input"
                    />
                  </div>
                  <div>
                    <label class="input-label">{{
                      t("admin.groups.openaiMessages.haikuModel")
                    }}</label>
                    <input
                      v-model="editForm.haiku_mapped_model"
                      type="text"
                      :placeholder="
                        t('admin.groups.openaiMessages.haikuModelPlaceholder')
                      "
                      class="input"
                    />
                  </div>
                </div>
              </div>
            </div>

            <div
              class="views-admin-groups-view__panel-60"
            >
              <div
                class="views-admin-groups-view__panel-61"
              >
                <div class="views-admin-groups-view__panel-62">
                  <div>
                    <div class="views-admin-groups-view__panel-56">
                      <div class="views-admin-groups-view__panel-63"></div>
                      <label
                        class="views-admin-groups-view__label-10"
                        >{{
                          t("admin.groups.openaiMessages.exactMappingTitle")
                        }}</label
                      >
                    </div>
                    <p
                      class="views-admin-groups-view__description-12"
                    >
                      {{ t("admin.groups.openaiMessages.exactMappingHint") }}
                    </p>
                  </div>
                </div>
              </div>

              <div class="views-admin-groups-view__panel-64">
                <div
                  v-if="editForm.exact_model_mappings.length === 0"
                  class="views-admin-groups-view__panel-65"
                >
                  <span>{{
                    t("admin.groups.openaiMessages.noExactMappings")
                  }}</span>
                  <button
                    type="button"
                    @click="addEditMessagesDispatchMapping"
                    class="views-admin-groups-view__action-13"
                  >
                    <Icon name="plus" size="sm" />
                    {{ t("admin.groups.openaiMessages.addExactMapping") }}
                  </button>
                </div>

                <div v-else class="views-admin-groups-view__panel-66">
                  <div
                    v-for="row in editForm.exact_model_mappings"
                    :key="getEditMessagesDispatchRowKey(row)"
                    class="views-admin-groups-view__panel-67"
                  >
                    <div class="views-admin-groups-view__panel-68">
                      <div
                        class="views-admin-groups-view__panel-69"
                      >
                        <div>
                          <label class="input-label">{{
                            t("admin.groups.openaiMessages.claudeModel")
                          }}</label>
                          <input
                            v-model="row.claude_model"
                            type="text"
                            :placeholder="
                              t(
                                'admin.groups.openaiMessages.claudeModelPlaceholder',
                              )
                            "
                            class="views-admin-groups-view__field-8 input"
                          />
                        </div>
                        <div
                          class="views-admin-groups-view__panel-70"
                        >
                          <Icon
                            name="arrowRight"
                            size="sm"
                            class="views-admin-groups-view__icon-6"
                          />
                        </div>
                        <div>
                          <label class="input-label">{{
                            t("admin.groups.openaiMessages.targetModel")
                          }}</label>
                          <input
                            v-model="row.target_model"
                            type="text"
                            :placeholder="
                              t(
                                'admin.groups.openaiMessages.targetModelPlaceholder',
                              )
                            "
                            class="views-admin-groups-view__field-8 input"
                          />
                        </div>
                      </div>
                      <button
                        type="button"
                        @click="removeEditMessagesDispatchMapping(row)"
                        class="views-admin-groups-view__action-14"
                        :title="
                          t('admin.groups.openaiMessages.removeExactMapping')
                        "
                      >
                        <Icon name="trash" size="sm" />
                      </button>
                    </div>
                  </div>

                  <button
                    type="button"
                    @click="addEditMessagesDispatchMapping"
                    class="views-admin-groups-view__action-15"
                  >
                    <Icon name="plus" size="sm" />
                    {{ t("admin.groups.openaiMessages.addExactMapping") }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 账号过滤控制 -->
        <div
          v-if="
            ['openai', 'anthropic'].includes(
              editForm.platform,
            )
          "
          class="views-admin-groups-view__panel-71"
        >
          <h4 class="views-admin-groups-view__heading">
            {{ t("admin.groups.accountFilters.title") }}
          </h4>

          <!-- require_oauth_only toggle -->
          <div class="views-admin-groups-view__panel-53">
            <div>
              <label class="views-admin-groups-view__label-8"
                >{{ t("admin.groups.accountFilters.oauthOnly") }}</label
              >
              <p class="views-admin-groups-view__description-13">
                {{
                  editForm.require_oauth_only
                    ? t("admin.groups.accountFilters.oauthOnlyEnabled")
                    : t("admin.groups.accountFilters.disabled")
                }}
              </p>
            </div>
            <button
              type="button"
              @click="
                editForm.require_oauth_only = !editForm.require_oauth_only
              "
              class="views-admin-groups-view__action-12"
              :class="
                editForm.require_oauth_only
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28'
              "
            >
              <span
                class="views-admin-groups-view__text-24"
                :class="
                  editForm.require_oauth_only
                    ? 'toggle-thumb--on'
                    : 'toggle-thumb--off'
                "
              />
            </button>
          </div>

          <!-- require_privacy_set toggle -->
          <div class="views-admin-groups-view__panel-53">
            <div>
              <label class="views-admin-groups-view__label-8"
                >{{ t("admin.groups.accountFilters.privacySetOnly") }}</label
              >
              <p class="views-admin-groups-view__description-13">
                {{
                  editForm.require_privacy_set
                    ? t("admin.groups.accountFilters.privacySetOnlyEnabled")
                    : t("admin.groups.accountFilters.disabled")
                }}
              </p>
            </div>
            <button
              type="button"
              @click="
                editForm.require_privacy_set = !editForm.require_privacy_set
              "
              class="views-admin-groups-view__action-12"
              :class="
                editForm.require_privacy_set
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28'
              "
            >
              <span
                class="views-admin-groups-view__text-24"
                :class="
                  editForm.require_privacy_set
                    ? 'toggle-thumb--on'
                    : 'toggle-thumb--off'
                "
              />
            </button>
          </div>
        </div>

        <!-- 无效请求兜底（仅 anthropic 平台，且非订阅分组） -->
        <div
          v-if="
            editForm.platform === 'anthropic' &&
            editForm.subscription_type !== 'subscription'
          "
          class="views-admin-groups-view__panel-25"
        >
          <label class="input-label">{{
            t("admin.groups.invalidRequestFallback.title")
          }}</label>
          <Select
            v-model="editForm.fallback_group_id_on_invalid_request"
            :options="invalidRequestFallbackOptionsForEdit"
            :placeholder="t('admin.groups.invalidRequestFallback.noFallback')"
          />
          <p class="input-hint">
            {{ t("admin.groups.invalidRequestFallback.hint") }}
          </p>
        </div>

        <!-- 模型路由配置（仅 anthropic 平台） -->
        <div v-if="editForm.platform === 'anthropic'" class="views-admin-groups-view__panel-25">
          <div class="views-admin-groups-view__panel-15">
            <label class="views-admin-groups-view__label">
              {{ t("admin.groups.modelRouting.title") }}
            </label>
            <!-- Help Tooltip -->
            <div class="views-admin-groups-view__panel-16">
              <Icon
                name="questionCircle"
                size="sm"
                :stroke-width="2"
                class="views-admin-groups-view__icon-4"
              />
              <div
                class="views-admin-groups-view__panel-72"
              >
                <div
                  class="views-admin-groups-view__panel-18"
                >
                  <p class="views-admin-groups-view__description">
                    {{ t("admin.groups.modelRouting.tooltip") }}
                  </p>
                  <div
                    class="views-admin-groups-view__panel-19"
                  ></div>
                </div>
              </div>
            </div>
          </div>
          <!-- 启用开关 -->
          <div class="views-admin-groups-view__panel-73">
            <button
              type="button"
              @click="
                editForm.model_routing_enabled = !editForm.model_routing_enabled
              "
              :class="[
                'views-admin-groups-view__action-26',
                editForm.model_routing_enabled
                  ? 'views-admin-groups-view__action-27'
                  : 'views-admin-groups-view__action-28',
              ]"
            >
              <span
                :class="[
                  'views-admin-groups-view__text-39',
                  editForm.model_routing_enabled
                    ? 'toggle-thumb--on'
                    : 'toggle-thumb--off',
                ]"
              />
            </button>
            <span class="views-admin-groups-view__text-19">
              {{
                editForm.model_routing_enabled
                  ? t("admin.groups.modelRouting.enabled")
                  : t("admin.groups.modelRouting.disabled")
              }}
            </span>
          </div>
          <p
            v-if="!editForm.model_routing_enabled"
            class="views-admin-groups-view__description-6"
          >
            {{ t("admin.groups.modelRouting.disabledHint") }}
          </p>
          <p v-else class="views-admin-groups-view__description-6">
            {{ t("admin.groups.modelRouting.noRulesHint") }}
          </p>
          <!-- 路由规则列表（仅在启用时显示） -->
          <div v-if="editForm.model_routing_enabled" class="views-admin-groups-view__panel-66">
            <div
              v-for="rule in editModelRoutingRules"
              :key="getEditRuleRenderKey(rule)"
              class="views-admin-groups-view__panel-74"
            >
              <div class="views-admin-groups-view__panel-75">
                <div class="views-admin-groups-view__panel-76">
                  <div>
                    <label class="views-admin-groups-view__text-16 input-label">{{
                      t("admin.groups.modelRouting.modelPattern")
                    }}</label>
                    <input
                      v-model="rule.pattern"
                      type="text"
                      class="views-admin-groups-view__field-9 input"
                      :placeholder="
                        t('admin.groups.modelRouting.modelPatternPlaceholder')
                      "
                    />
                  </div>
                  <div>
                    <label class="views-admin-groups-view__text-16 input-label">{{
                      t("admin.groups.modelRouting.accounts")
                    }}</label>
                    <!-- 已选账号标签 -->
                    <div
                      v-if="rule.accounts.length > 0"
                      class="views-admin-groups-view__panel-20"
                    >
                      <span
                        v-for="account in rule.accounts"
                        :key="account.id"
                        class="views-admin-groups-view__text-17"
                      >
                        {{ account.name }}
                        <button
                          type="button"
                          @click="removeSelectedAccount(rule, account.id, true)"
                          class="views-admin-groups-view__action-8"
                        >
                          <Icon name="x" size="xs" />
                        </button>
                      </span>
                    </div>
                    <!-- 账号搜索输入框 -->
                    <div class="views-admin-groups-view__panel-7 account-search-container">
                      <input
                        v-model="
                          accountSearchKeyword[getEditRuleSearchKey(rule)]
                        "
                        type="text"
                        class="views-admin-groups-view__field-9 input"
                        :placeholder="
                          t(
                            'admin.groups.modelRouting.searchAccountPlaceholder',
                          )
                        "
                        @input="searchAccountsByRule(rule, true)"
                        @focus="onAccountSearchFocus(rule, true)"
                      />
                      <!-- 搜索结果下拉框 -->
                      <div
                        v-if="
                          showAccountDropdown[getEditRuleSearchKey(rule)] &&
                          accountSearchResults[getEditRuleSearchKey(rule)]
                            ?.length > 0
                        "
                        class="views-admin-groups-view__panel-77"
                      >
                        <button
                          v-for="account in accountSearchResults[
                            getEditRuleSearchKey(rule)
                          ]"
                          :key="account.id"
                          type="button"
                          @click="selectAccount(rule, account, true)"
                          class="views-admin-groups-view__action-16"
                          :class="{
                            'views-admin-groups-view__action-30': rule.accounts.some(
                              (a) => a.id === account.id,
                            ),
                          }"
                          :disabled="
                            rule.accounts.some((a) => a.id === account.id)
                          "
                        >
                          <span>{{ account.name }}</span>
                          <span class="views-admin-groups-view__text-25"
                            >#{{ account.id }}</span
                          >
                        </button>
                      </div>
                    </div>
                    <p class="views-admin-groups-view__description-14">
                      {{ t("admin.groups.modelRouting.accountsHint") }}
                    </p>
                  </div>
                </div>
                <button
                  type="button"
                  @click="removeEditRoutingRule(rule)"
                  class="views-admin-groups-view__action-17"
                  :title="t('admin.groups.modelRouting.removeRule')"
                >
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </div>
          </div>
          <!-- 添加规则按钮（仅在启用时显示） -->
          <button
            v-if="editForm.model_routing_enabled"
            type="button"
            @click="addEditRoutingRule"
            class="views-admin-groups-view__action-18"
          >
            <Icon name="plus" size="sm" />
            {{ t("admin.groups.modelRouting.addRule") }}
          </button>
        </div>
      </form>

      <template #footer>
        <div class="views-admin-groups-view__panel-78">
          <button
            @click="closeEditModal"
            type="button"
            class="btn btn-secondary"
          >
            {{ t("common.cancel") }}
          </button>
          <button
            type="submit"
            form="edit-group-form"
            :disabled="submitting"
            class="btn btn-primary"
            data-tour="group-form-submit"
          >
            <svg
              v-if="submitting"
              class="views-admin-groups-view__icon-7"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="views-admin-groups-view__circle"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="views-admin-groups-view__path"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{ submitting ? t("admin.groups.updating") : t("common.update") }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.groups.deleteGroup')"
      :message="deleteConfirmMessage"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />

    <ConfirmDialog
      :show="showUnsupportedLiveConfirm"
      :title="t('admin.groups.openaiLive.unsupportedTitle')"
      :message="t('admin.groups.openaiLive.unsupportedMessage')"
      :confirm-text="t('admin.groups.openaiLive.enableAnyway')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmUnsupportedLive"
      @cancel="cancelUnsupportedLive"
    />

    <!-- Sort Order Modal -->
    <BaseDialog
      :show="showSortModal"
      :title="t('admin.groups.sortOrder')"
      width="normal"
      @close="closeSortModal"
    >
      <div class="views-admin-groups-view__panel-80">
        <p class="views-admin-groups-view__text-19">
          {{ t("admin.groups.sortOrderHint") }}
        </p>
        <VueDraggable
          v-model="sortableGroups"
          :animation="200"
          class="views-admin-groups-view__panel-45"
        >
          <div
            v-for="group in sortableGroups"
            :key="group.id"
            class="views-admin-groups-view__panel-81"
          >
            <div class="views-admin-groups-view__panel-82">
              <Icon name="menu" size="md" />
            </div>
            <div class="views-admin-groups-view__panel-83">
              <div class="views-admin-groups-view__text-2">
                {{ group.name }}
              </div>
              <div class="views-admin-groups-view__description-5">
                <span
                  :class="[
                    'views-admin-groups-view__text-40',
                    group.platform === 'anthropic'
                      ? 'views-admin-groups-view__text-27'
                      : group.platform === 'openai'
                        ? 'views-admin-groups-view__text-28'
                        : group.platform === 'grok'
                            ? 'views-admin-groups-view__text-30'
                            : group.platform === 'kimi'
                              ? 'views-admin-groups-view__text-31'
                              : group.platform === 'zhipu'
                                ? 'views-admin-groups-view__text-32'
                                : group.platform === 'deepseek'
                                  ? 'views-admin-groups-view__text-33'
                                  : 'views-admin-groups-view__text-34',
                  ]"
                >
                  {{ t("admin.groups.platforms." + group.platform) }}
                </span>
              </div>
            </div>
            <div class="views-admin-groups-view__panel-84">#{{ group.id }}</div>
          </div>
        </VueDraggable>
      </div>

      <template #footer>
        <div class="views-admin-groups-view__panel-78">
          <button
            @click="closeSortModal"
            type="button"
            class="btn btn-secondary"
          >
            {{ t("common.cancel") }}
          </button>
          <button
            @click="saveSortOrder"
            :disabled="sortSubmitting"
            class="btn btn-primary"
          >
            <svg
              v-if="sortSubmitting"
              class="views-admin-groups-view__icon-7"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="views-admin-groups-view__circle"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="views-admin-groups-view__path"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{ sortSubmitting ? t("common.saving") : t("common.save") }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Composite Routes Modal -->
    <BaseDialog
      :show="showCompositeRoutesModal"
      :title="
        compositeRoutesGroup
          ? t('admin.groups.compositeRoutes.titleWithGroup', {
              name: compositeRoutesGroup.name,
            })
          : t('admin.groups.compositeRoutes.title')
      "
      width="wide"
      @close="closeCompositeRoutesModal"
    >
      <div class="views-admin-groups-view__panel-85">
        <section class="views-admin-groups-view__section">
          <div class="views-admin-groups-view__panel-26">
            <h3 class="views-admin-groups-view__heading-3">
              {{ t("admin.groups.compositeRoutes.routes") }}
            </h3>
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="compositeRoutesLoading"
              @click="loadCompositeRoutes"
            >
              <Icon
                name="refresh"
                size="sm"
                :class="compositeRoutesLoading ? 'views-admin-groups-view__icon-8' : ''"
              />
            </button>
          </div>

          <div
            class="views-admin-groups-view__panel-86"
          >
            <div
              v-if="compositeRoutesLoading"
              class="views-admin-groups-view__panel-87"
            >
              {{ t("common.loading") }}
            </div>
            <div
              v-else-if="compositeRoutes.length === 0"
              class="views-admin-groups-view__panel-87"
            >
              {{ t("admin.groups.compositeRoutes.empty") }}
            </div>
            <div v-else class="views-admin-groups-view__panel-14">
              <table class="views-admin-groups-view__table">
                <thead class="views-admin-groups-view__header">
                  <tr>
                    <th class="views-admin-groups-view__heading-4">
                      {{ t("admin.groups.compositeRoutes.publicModel") }}
                    </th>
                    <th class="views-admin-groups-view__heading-4">
                      {{ t("admin.groups.compositeRoutes.target") }}
                    </th>
                    <th class="views-admin-groups-view__heading-4">
                      {{ t("admin.groups.compositeRoutes.scope") }}
                    </th>
                    <th class="views-admin-groups-view__heading-5">
                      {{ t("admin.groups.columns.actions") }}
                    </th>
                  </tr>
                </thead>
                <tbody class="views-admin-groups-view__body">
                  <tr
                    v-for="route in compositeRoutes"
                    :key="route.id"
                    :class="!route.enabled && 'views-admin-groups-view__row'"
                  >
                    <td class="views-admin-groups-view__cell">
                      <div class="views-admin-groups-view__panel-88">
                        {{ route.public_model }}
                      </div>
                      <div class="views-admin-groups-view__panel-89">
                        <span class="badge badge-gray">{{
                          compositeRouteMatchLabel(route.match_type)
                        }}</span>
                        <span
                          v-if="!route.enabled"
                          class="badge badge-danger"
                        >
                          {{ t("admin.accounts.status.inactive") }}
                        </span>
                      </div>
                    </td>
                    <td class="views-admin-groups-view__heading-4">
                      <div class="views-admin-groups-view__panel-90">
                        <PlatformIcon :platform="route.target_platform" size="xs" />
                        <span>{{ formatCompositePlatform(route.target_platform) }}</span>
                      </div>
                      <div class="views-admin-groups-view__panel-91">
                        {{ route.upstream_model || route.public_model }}
                      </div>
                    </td>
                    <td class="views-admin-groups-view__heading-4">
                      <div class="views-admin-groups-view__panel-92">
                        {{ formatCompositeEndpoint(route.endpoint) }}
                      </div>
                      <div class="views-admin-groups-view__description-5">
                        {{ t("admin.groups.compositeRoutes.priority") }}:
                        {{ route.priority }}
                      </div>
                    </td>
                    <td class="views-admin-groups-view__heading-4">
                      <div class="views-admin-groups-view__panel-93">
                        <button
                          type="button"
                          class="views-admin-groups-view__action-19"
                          :title="t('common.edit')"
                          @click="editCompositeRoute(route)"
                        >
                          <Icon name="edit" size="sm" />
                        </button>
                        <button
                          type="button"
                          class="views-admin-groups-view__action-20"
                          :title="t('common.delete')"
                          @click="deleteCompositeRoute(route)"
                        >
                          <Icon name="trash" size="sm" />
                        </button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <section class="views-admin-groups-view__form">
          <form class="views-admin-groups-view__panel-66" @submit.prevent="saveCompositeRoute">
            <div class="views-admin-groups-view__panel-94">
              <h3 class="views-admin-groups-view__heading-3">
                {{
                  compositeRouteEditingId
                    ? t("admin.groups.compositeRoutes.editRoute")
                    : t("admin.groups.compositeRoutes.addRoute")
                }}
              </h3>
              <button
                v-if="compositeRouteEditingId"
                type="button"
                class="views-admin-groups-view__action-21"
                @click="resetCompositeRouteForm"
              >
                {{ t("common.cancel") }}
              </button>
            </div>

            <div>
              <label class="input-label">{{
                t("admin.groups.compositeRoutes.publicModel")
              }}</label>
              <input
                v-model.trim="compositeRouteForm.public_model"
                type="text"
                class="input"
                required
                placeholder="openrouter/gpt-5"
              />
            </div>

            <div class="views-admin-groups-view__panel-95">
              <div>
                <label class="input-label">{{
                  t("admin.groups.compositeRoutes.matchType")
                }}</label>
                <Select
                  v-model="compositeRouteForm.match_type"
                  :options="compositeRouteMatchOptions"
                />
              </div>
              <div>
                <label class="input-label">{{
                  t("admin.groups.compositeRoutes.endpoint")
                }}</label>
                <Select
                  v-model="compositeRouteForm.endpoint"
                  :options="compositeRouteEndpointOptions"
                />
              </div>
            </div>

            <div class="views-admin-groups-view__panel-95">
              <div>
                <label class="input-label">{{
                  t("admin.groups.compositeRoutes.targetPlatform")
                }}</label>
                <div class="views-admin-groups-view__panel-14" role="tablist" :aria-label="t('admin.groups.compositeRoutes.targetPlatform')">
                  <div class="views-admin-groups-view__panel-5">
                    <button
                      v-for="option in compositeRoutePlatformOptions"
                      :key="option.value"
                      type="button"
                      role="tab"
                      :aria-selected="compositeRouteForm.target_platform === option.value"
                      :class="[
                        'views-admin-groups-view__action-32',
                        compositeRouteForm.target_platform === option.value
                          ? 'views-admin-groups-view__action-23'
                          : 'views-admin-groups-view__action-24'
                      ]"
                      @click="compositeRouteForm.target_platform = option.value as ConcreteGroupPlatform"
                    >
                      <PlatformIcon :platform="option.value as GroupPlatform" size="xs" />
                      {{ option.label }}
                    </button>
                  </div>
                </div>
              </div>
              <div>
                <label class="input-label">{{
                  t("admin.groups.compositeRoutes.priority")
                }}</label>
                <input
                  v-model.number="compositeRouteForm.priority"
                  type="number"
                  min="1"
                  step="1"
                  class="input"
                />
              </div>
            </div>

            <div>
              <label class="input-label">{{
                t("admin.groups.compositeRoutes.upstreamModel")
              }}</label>
              <input
                v-model.trim="compositeRouteForm.upstream_model"
                type="text"
                class="input"
                placeholder="gpt-5"
              />
              <p class="views-admin-groups-view__description-4">
                {{ t("admin.groups.compositeRoutes.upstreamModelHint") }}
              </p>
            </div>

            <div>
              <label class="input-label">{{
                t("admin.groups.compositeRoutes.notes")
              }}</label>
              <textarea
                v-model.trim="compositeRouteForm.notes"
                rows="2"
                class="input"
              ></textarea>
            </div>

            <div class="views-admin-groups-view__panel-94">
              <label class="views-admin-groups-view__label-3">
                <input
                  v-model="compositeRouteForm.enabled"
                  type="checkbox"
                  class="views-admin-groups-view__field-6"
                />
                {{ t("admin.groups.compositeRoutes.enabled") }}
              </label>
              <button
                type="submit"
                class="btn btn-primary"
                :disabled="compositeRouteSaving"
              >
                <Icon
                  v-if="!compositeRouteSaving"
                  name="check"
                  size="sm"
                  class="views-admin-groups-view__icon-2"
                />
                {{ compositeRouteEditingId ? t("common.update") : t("common.create") }}
              </button>
            </div>
          </form>

          <div class="views-admin-groups-view__panel-96">
            <h3 class="views-admin-groups-view__heading-6">
              {{ t("admin.groups.compositeRoutes.preview") }}
            </h3>
            <div class="views-admin-groups-view__panel-66">
              <input
                v-model.trim="compositePreviewModel"
                type="text"
                class="input"
                placeholder="openrouter/gpt-5"
                @keyup.enter="previewCompositeRoute"
              />
              <div class="views-admin-groups-view__panel-97">
                <Select
                  v-model="compositePreviewEndpoint"
                  :options="compositeRouteEndpointOptions"
                  class="views-admin-groups-view__field-10"
                />
                <button
                  type="button"
                  class="btn btn-secondary"
                  :disabled="compositePreviewLoading || !compositePreviewModel"
                  @click="previewCompositeRoute"
                >
                  <Icon name="play" size="sm" />
                </button>
              </div>

              <div
                v-if="compositePreviewDecision"
                class="views-admin-groups-view__panel-98"
              >
                <div class="views-admin-groups-view__panel-99">
                  <span
                    :class="[
                      'badge',
                      compositePreviewDecision.matched
                        ? 'badge-success'
                        : 'badge-danger',
                    ]"
                  >
                    {{
                      compositePreviewDecision.matched
                        ? t("admin.groups.compositeRoutes.matched")
                        : t("admin.groups.compositeRoutes.notMatched")
                    }}
                  </span>
                  <span class="badge badge-gray">
                    {{
                      compositeRouteSourceLabel(
                        compositePreviewDecision.source,
                      )
                    }}
                  </span>
                </div>
                <div
                  v-if="compositePreviewDecision.matched"
                  class="views-admin-groups-view__panel-100"
                >
                  <div>
                    {{ t("admin.groups.compositeRoutes.targetPlatform") }}:
                    {{
                      formatCompositePlatform(
                        compositePreviewDecision.target_platform,
                      )
                    }}
                  </div>
                  <div class="views-admin-groups-view__panel-101">
                    {{ t("admin.groups.compositeRoutes.upstreamModel") }}:
                    {{ compositePreviewDecision.upstream_model }}
                  </div>
                </div>
                <div
                  v-else
                  class="views-admin-groups-view__text-10"
                >
                  {{ compositePreviewDecision.reason }}
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="views-admin-groups-view__panel-102">
          <button
            type="button"
            class="btn btn-secondary"
            @click="closeCompositeRoutesModal"
          >
            {{ t("common.close") }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Group Rate Multipliers Modal -->
    <GroupRateMultipliersModal
      :show="showRateMultipliersModal"
      :group="rateMultipliersGroup"
      @close="showRateMultipliersModal = false"
      @success="loadGroups"
    />

    <GroupLevelRatesModal
      :show="showLevelRatesModal"
      :group="levelRatesGroup"
      @close="showLevelRatesModal = false"
      @success="loadGroups"
    />

    <!-- Group RPM Overrides Modal -->
    <GroupRPMOverridesModal
      :show="showRPMOverridesModal"
      :group="rpmOverridesGroup"
      @close="showRPMOverridesModal = false"
      @success="loadGroups"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useAppStore } from "@/stores/app";
import { useOnboardingStore } from "@/stores/onboarding";
import { adminAPI } from "@/api/admin";
import type {
  AdminGroup,
  CompositeModelRoute,
  CompositeModelRouteInput,
  CompositeRouteDecision,
  CompositeRouteEndpoint,
  CompositeRouteMatchType,
  GroupPlatform,
  SubscriptionType,
} from "@/types";
import type { Column } from "@/components/common/types";
import AppLayout from "@/components/layout/AppLayout.vue";
import TablePageLayout from "@/components/layout/TablePageLayout.vue";
import DataTable from "@/components/common/DataTable.vue";
import Pagination from "@/components/common/Pagination.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import ConfirmDialog from "@/components/common/ConfirmDialog.vue";
import EmptyState from "@/components/common/EmptyState.vue";
import Select from "@/components/common/Select.vue";
import PlatformIcon from "@/components/common/PlatformIcon.vue";
import Icon from "@/components/icons/Icon.vue";
import GroupRateMultipliersModal from "@/components/admin/group/GroupRateMultipliersModal.vue";
import GroupLevelRatesModal from "@/components/admin/group/GroupLevelRatesModal.vue";
import GroupRPMOverridesModal from "@/components/admin/group/GroupRPMOverridesModal.vue";
import GroupCapacityBadge from "@/components/common/GroupCapacityBadge.vue";
import ReasoningEffortPolicyFields from "@/components/admin/group/ReasoningEffortPolicyFields.vue";
import PricingEntryCard from "@/components/admin/channel/PricingEntryCard.vue";
import type { PricingFormEntry } from "@/components/admin/channel/types";
import {
  apiIntervalsToForm,
  createDefaultTimePricingForm,
  formIntervalsToAPI,
  mTokToPerToken,
  perTokenToMTok,
  toNullableNumber,
} from "@/components/admin/channel/types";
import type { ChannelModelPricing } from "@/api/admin/channels";
import { VueDraggable } from "vue-draggable-plus";
import { createStableObjectKeyResolver } from "@/utils/stableObjectKey";
import { extractApiErrorMessage } from "@/utils/apiError";
import { useKeyedDebouncedSearch } from "@/composables/useKeyedDebouncedSearch";
import { getPersistedPageSize } from "@/composables/usePersistedPageSize";
import { formatPoints } from "@/utils/format";
import {
  createDefaultMessagesDispatchFormState,
  messagesDispatchConfigToFormState,
  messagesDispatchFormStateToConfig,
  resetMessagesDispatchFormState,
  type MessagesDispatchMappingRow,
} from "./groupsMessagesDispatch";
import {
  buildModelsListConfig,
  createModelsListState as createInitialModelsListState,
  invertModelsListSelection,
  moveModelsListItem,
  selectAllModelsListItems,
  setModelsListCandidates,
} from "./groupsModelsList";
import { createModelsListCandidatesTracker } from "./groupsModelsListCandidates";
import {
  isProfitControlPlatform,
  profitPercentToDecimal,
  profitDecimalToPercent,
  validateProfitControlFormState,
  type ProfitControlFormState,
} from "./groupsProfitControl";
import {
  normalizeReasoningEffortForPlatform,
  reasoningEffortMappingsToAPI,
  reasoningEffortMappingsToRows,
  supportsReasoningEffortPolicyPlatform,
  type ReasoningEffortMappingRow,
} from "./groupsReasoningEffort";
import {
  getDefaultImagePreviewPrice,
  getDefaultVideoPreviewPrice,
  getImagePricePlaceholder,
  getVideoPricePlaceholder,
  imagePricingI18nKey,
  supportsImagePricingPlatform,
  supportsVideoPricingPlatform,
  videoPricingI18nKey,
} from "./groupsImagePricing";
import {
  createVideoModelPricesForm,
  grokVideoPriceResolutions,
  serializeVideoModelPrices,
  videoModelPriceFamilyRows,
} from "./groupsVideoModelPricing";

const emptyGroupPricing = (): PricingFormEntry => ({
  models: [],
  billing_mode: "token",
  input_price: null,
  output_price: null,
  cache_write_price: null,
  cache_read_price: null,
  image_input_price: null,
  image_output_price: null,
  per_request_price: null,
  intervals: [],
  time_pricing: createDefaultTimePricingForm(),
});

const addGroupPricing = (entries: PricingFormEntry[]) =>
  entries.push(emptyGroupPricing());

const groupPricingFromAPI = (
  pricing: ChannelModelPricing[] | undefined,
): PricingFormEntry[] =>
  (pricing || []).map((entry) => ({
    models: entry.models || [],
    billing_mode: entry.billing_mode || "token",
    input_price: perTokenToMTok(entry.input_price),
    output_price: perTokenToMTok(entry.output_price),
    cache_write_price: perTokenToMTok(entry.cache_write_price),
    cache_read_price: perTokenToMTok(entry.cache_read_price),
    image_input_price: perTokenToMTok(entry.image_input_price),
    image_output_price: perTokenToMTok(entry.image_output_price),
    per_request_price: entry.per_request_price,
    intervals: apiIntervalsToForm(entry.intervals || []),
    time_pricing: createDefaultTimePricingForm(),
  }));

const groupPricingToAPI = (
  pricing: PricingFormEntry[],
  platform: string,
): ChannelModelPricing[] =>
  pricing
    .filter((entry) => entry.models.length > 0)
    .map((entry) => ({
      platform,
      models: entry.models,
      billing_mode: entry.billing_mode,
      input_price: mTokToPerToken(entry.input_price),
      output_price: mTokToPerToken(entry.output_price),
      cache_write_price: mTokToPerToken(entry.cache_write_price),
      cache_read_price: mTokToPerToken(entry.cache_read_price),
      image_input_price: mTokToPerToken(entry.image_input_price),
      image_output_price: mTokToPerToken(entry.image_output_price),
      per_request_price: toNullableNumber(entry.per_request_price),
      intervals:
        entry.billing_mode === "token"
          ? []
          : formIntervalsToAPI(entry.intervals || []),
      time_pricing: null,
    }));

const { t } = useI18n();
const appStore = useAppStore();
const onboardingStore = useOnboardingStore();

const ALWAYS_VISIBLE_COLUMNS = new Set(["name", "actions"]);
// Default hidden columns (hidden on first load / after schema bumps).
const DEFAULT_HIDDEN_COLUMNS = ["id"];
const HIDDEN_COLUMNS_KEY = "group-hidden-columns";
// Bump when adding new default-hidden columns so existing admins pick them up once.
const COLUMN_SETTINGS_VERSION_KEY = "group-column-settings-version";
const COLUMN_SETTINGS_VERSION = 2;
const VERSION_NEW_HIDDEN_COLUMNS: Record<number, string[]> = {
  2: ["id"],
};

const allColumns = computed<Column[]>(() => [
  { key: "name", label: t("admin.groups.columns.name"), sortable: true },
  { key: "id", label: t("admin.groups.columns.id"), sortable: true },
  {
    key: "platform",
    label: t("admin.groups.columns.platform"),
    sortable: true,
  },
  {
    key: "billing_type",
    label: t("admin.groups.columns.billingType"),
    sortable: true,
  },
  {
    key: "rate_multiplier",
    label: t("admin.groups.columns.rateMultiplier"),
    sortable: true,
  },
  {
    key: "is_exclusive",
    label: t("admin.groups.columns.type"),
    sortable: true,
  },
  {
    key: "account_count",
    label: t("admin.groups.columns.accounts"),
    sortable: true,
  },
  {
    key: "capacity",
    label: t("admin.groups.columns.capacity"),
    sortable: false,
  },
  { key: "usage", label: t("admin.groups.columns.usage"), sortable: false },
  { key: "status", label: t("admin.groups.columns.status"), sortable: true },
  { key: "actions", label: t("admin.groups.columns.actions"), sortable: false },
]);

const toggleableColumns = computed(() =>
  allColumns.value.filter((col) => !ALWAYS_VISIBLE_COLUMNS.has(col.key)),
);
const hiddenColumns = reactive<Set<string>>(new Set());
const showColumnDropdown = ref(false);
const columnDropdownRef = ref<HTMLElement | null>(null);

const getValidHiddenColumnKeys = () =>
  new Set(toggleableColumns.value.map((col) => col.key));

const loadSavedColumns = () => {
  hiddenColumns.clear();
  try {
    const saved = localStorage.getItem(HIDDEN_COLUMNS_KEY);
    const validKeys = getValidHiddenColumnKeys();

    if (saved) {
      const parsed = JSON.parse(saved);
      if (Array.isArray(parsed)) {
        parsed
          .filter(
            (key): key is string =>
              typeof key === "string" && validKeys.has(key),
          )
          .forEach((key) => hiddenColumns.add(key));
      }

      // Existing admins: auto-hide columns newly added as default-hidden.
      const storedVersion = Number(
        localStorage.getItem(COLUMN_SETTINGS_VERSION_KEY) ?? "1",
      );
      if (storedVersion < COLUMN_SETTINGS_VERSION) {
        let mutated = false;
        for (let v = storedVersion + 1; v <= COLUMN_SETTINGS_VERSION; v++) {
          for (const key of VERSION_NEW_HIDDEN_COLUMNS[v] ?? []) {
            if (validKeys.has(key) && !hiddenColumns.has(key)) {
              hiddenColumns.add(key);
              mutated = true;
            }
          }
        }
        if (mutated) {
          saveColumnsToStorage();
        } else {
          localStorage.setItem(
            COLUMN_SETTINGS_VERSION_KEY,
            String(COLUMN_SETTINGS_VERSION),
          );
        }
      }
    } else {
      DEFAULT_HIDDEN_COLUMNS.forEach((key) => {
        if (validKeys.has(key)) hiddenColumns.add(key);
      });
      saveColumnsToStorage();
    }
  } catch (error) {
    console.error("Failed to load group column settings:", error);
    DEFAULT_HIDDEN_COLUMNS.forEach((key) => hiddenColumns.add(key));
  }
};

const saveColumnsToStorage = () => {
  try {
    const validKeys = getValidHiddenColumnKeys();
    const keys = [...hiddenColumns].filter((key) => validKeys.has(key));
    localStorage.setItem(HIDDEN_COLUMNS_KEY, JSON.stringify(keys));
    localStorage.setItem(
      COLUMN_SETTINGS_VERSION_KEY,
      String(COLUMN_SETTINGS_VERSION),
    );
  } catch (error) {
    console.error("Failed to save group column settings:", error);
  }
};

const isColumnVisible = (key: string) => !hiddenColumns.has(key);
const hasVisibleUsageSummaryConsumer = computed(
  () => isColumnVisible("usage") || isColumnVisible("billing_type"),
);
const hasVisibleCapacityColumn = computed(() => isColumnVisible("capacity"));

const toggleColumn = (key: string) => {
  const validKeys = getValidHiddenColumnKeys();
  if (!validKeys.has(key)) return;

  const wasHidden = hiddenColumns.has(key);
  if (wasHidden) {
    hiddenColumns.delete(key);
  } else {
    hiddenColumns.add(key);
  }
  saveColumnsToStorage();

  if (wasHidden && (key === "usage" || key === "billing_type")) {
    loadUsageSummary();
  }
  if (wasHidden && key === "capacity") {
    loadCapacitySummary();
  }
};

const columns = computed<Column[]>(() =>
  allColumns.value.filter(
    (col) => ALWAYS_VISIBLE_COLUMNS.has(col.key) || !hiddenColumns.has(col.key),
  ),
);

if (typeof window !== "undefined") {
  loadSavedColumns();
}

// Filter options
const statusOptions = computed(() => [
  { value: "", label: t("admin.groups.allStatus") },
  { value: "active", label: t("admin.accounts.status.active") },
  { value: "inactive", label: t("admin.accounts.status.inactive") },
]);

const exclusiveOptions = computed(() => [
  { value: "", label: t("admin.groups.allGroups") },
  { value: "true", label: t("admin.groups.exclusive") },
  { value: "false", label: t("admin.groups.nonExclusive") },
]);

const platformOptions = computed(() => [
  { value: "anthropic", label: "Anthropic" },
  { value: "openai", label: "OpenAI" },
  { value: "gemini", label: "Gemini" },
  { value: "antigravity", label: "Antigravity" },
  { value: "grok", label: "Grok" },
  { value: "kimi", label: "Kimi" },
  { value: "zhipu", label: "Zhipu GLM" },
  { value: "deepseek", label: "DeepSeek" },
  { value: "composite", label: "Composite" },
]);

const platformFilterOptions = computed(() => [
  { value: "", label: t("admin.groups.allPlatforms") },
  { value: "anthropic", label: "Anthropic" },
  { value: "openai", label: "OpenAI" },
  { value: "gemini", label: "Gemini" },
  { value: "antigravity", label: "Antigravity" },
  { value: "grok", label: "Grok" },
  { value: "kimi", label: "Kimi" },
  { value: "zhipu", label: "Zhipu GLM" },
  { value: "deepseek", label: "DeepSeek" },
  { value: "composite", label: "Composite" },
]);

const selectPlatformFilter = (platform: string) => {
  if (filters.platform === platform) return;
  filters.platform = platform;
  pagination.page = 1;
  loadGroups();
};

const selectCreatePlatform = (platform: GroupPlatform) => {
  if (createForm.platform === platform) return;
  createForm.platform = platform;
  createForm.copy_accounts_from_group_ids = [];
};

const compositeRoutePlatformOptions = computed(() => [
  { value: "anthropic", label: "Anthropic" },
  { value: "openai", label: "OpenAI" },
  { value: "gemini", label: "Gemini" },
  { value: "antigravity", label: "Antigravity" },
  { value: "grok", label: "Grok" },
]);

const compositeRouteEndpointOptions = computed(() => [
  { value: "any", label: t("admin.groups.compositeRoutes.endpoints.any") },
  {
    value: "messages",
    label: t("admin.groups.compositeRoutes.endpoints.messages"),
  },
  {
    value: "count_tokens",
    label: t("admin.groups.compositeRoutes.endpoints.countTokens"),
  },
  {
    value: "responses",
    label: t("admin.groups.compositeRoutes.endpoints.responses"),
  },
  {
    value: "chat_completions",
    label: t("admin.groups.compositeRoutes.endpoints.chatCompletions"),
  },
  {
    value: "embeddings",
    label: t("admin.groups.compositeRoutes.endpoints.embeddings"),
  },
  { value: "images", label: t("admin.groups.compositeRoutes.endpoints.images") },
]);

const compositeRouteMatchOptions = computed(() => [
  { value: "exact", label: t("admin.groups.compositeRoutes.match.exact") },
  { value: "prefix", label: t("admin.groups.compositeRoutes.match.prefix") },
]);

const editStatusOptions = computed(() => [
  { value: "active", label: t("admin.accounts.status.active") },
  { value: "inactive", label: t("admin.accounts.status.inactive") },
]);

const subscriptionTypeOptions = computed(() => [
  { value: "standard", label: t("admin.groups.subscription.standard") },
  { value: "subscription", label: t("admin.groups.subscription.subscription") },
]);

// 降级分组选项（创建时）- 仅包含 anthropic 平台且未启用 claude_code_only 的分组
const fallbackGroupOptions = computed(() => {
  const options: { value: string | null; label: string }[] = [
    { value: null, label: t("admin.groups.claudeCode.noFallback") },
  ];
  const eligibleGroups = groups.value.filter(
    (g) =>
      g.platform === "anthropic" &&
      !g.claude_code_only &&
      g.status === "active",
  );
  eligibleGroups.forEach((g) => {
    options.push({ value: g.id, label: g.name });
  });
  return options;
});

// 降级分组选项（编辑时）- 排除自身
const fallbackGroupOptionsForEdit = computed(() => {
  const options: { value: string | null; label: string }[] = [
    { value: null, label: t("admin.groups.claudeCode.noFallback") },
  ];
  const currentId = editingGroup.value?.id;
  const eligibleGroups = groups.value.filter(
    (g) =>
      g.platform === "anthropic" &&
      !g.claude_code_only &&
      g.status === "active" &&
      g.id !== currentId,
  );
  eligibleGroups.forEach((g) => {
    options.push({ value: g.id, label: g.name });
  });
  return options;
});

// 无效请求兜底分组选项（创建时）- 仅包含 anthropic 平台、非订阅且未配置兜底的分组
const invalidRequestFallbackOptions = computed(() => {
  const options: { value: string | null; label: string }[] = [
    { value: null, label: t("admin.groups.invalidRequestFallback.noFallback") },
  ];
  const eligibleGroups = groups.value.filter(
    (g) =>
      g.platform === "anthropic" &&
      g.status === "active" &&
      g.subscription_type !== "subscription" &&
      g.fallback_group_id_on_invalid_request === null,
  );
  eligibleGroups.forEach((g) => {
    options.push({ value: g.id, label: g.name });
  });
  return options;
});

// 无效请求兜底分组选项（编辑时）- 排除自身
const invalidRequestFallbackOptionsForEdit = computed(() => {
  const options: { value: string | null; label: string }[] = [
    { value: null, label: t("admin.groups.invalidRequestFallback.noFallback") },
  ];
  const currentId = editingGroup.value?.id;
  const eligibleGroups = groups.value.filter(
    (g) =>
      g.platform === "anthropic" &&
      g.status === "active" &&
      g.subscription_type !== "subscription" &&
      g.fallback_group_id_on_invalid_request === null &&
      g.id !== currentId,
  );
  eligibleGroups.forEach((g) => {
    options.push({ value: g.id, label: g.name });
  });
  return options;
});

const canCopyAccountsFromGroup = (targetPlatform: GroupPlatform, sourcePlatform: GroupPlatform) =>
  targetPlatform === "composite" || sourcePlatform === targetPlatform;

const copyAccountsGroupLabel = (g: AdminGroup) => {
  const count = g.account_count || 0;
  const platform = t("admin.groups.platforms." + g.platform);
  return `${g.name} - ${platform} (${t("admin.groups.accountsCount", { count })})`;
};

// 复制账号的源分组选项（创建时）- 相同平台；composite 分组可汇总各平台账号
const copyAccountsGroupOptions = computed(() => {
  const eligibleGroups = groups.value.filter(
    (g) =>
      canCopyAccountsFromGroup(createForm.platform, g.platform) &&
      (g.account_count || 0) > 0,
  );
  return eligibleGroups.map((g) => ({
    value: g.id,
    label: copyAccountsGroupLabel(g),
  }));
});

// 复制账号的源分组选项（编辑时）- 相同平台；composite 分组可汇总各平台账号，排除自身
const copyAccountsGroupOptionsForEdit = computed(() => {
  const currentId = editingGroup.value?.id;
  const eligibleGroups = groups.value.filter(
    (g) =>
      canCopyAccountsFromGroup(editForm.platform, g.platform) &&
      (g.account_count || 0) > 0 &&
      g.id !== currentId,
  );
  return eligibleGroups.map((g) => ({
    value: g.id,
    label: copyAccountsGroupLabel(g),
  }));
});

const groups = ref<AdminGroup[]>([]);
const loading = ref(true);
type GroupUsageSummary = {
  today_cost: number;
  yesterday_cost: number;
  total_cost: number;
};

const usageMap = ref<Map<string, GroupUsageSummary>>(new Map());
const usageLoading = ref(false);
const capacityMap = ref<
  Map<
    string,
    {
      concurrencyUsed: number;
      concurrencyMax: number;
      sessionsUsed: number;
      sessionsMax: number;
      rpmUsed: number;
      rpmMax: number;
    }
  >
>(new Map());
const searchQuery = ref("");
const filters = reactive({
  platform: "",
  status: "",
  is_exclusive: "",
});
const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0,
});
const sortState = reactive({
  sort_by: "sort_order",
  sort_order: "asc" as "asc" | "desc",
});

let abortController: AbortController | null = null;

const showCreateModal = ref(false);
const showEditModal = ref(false);
const showDeleteDialog = ref(false);
const pendingLiveForm = ref<"create" | "edit" | null>(null);
const showUnsupportedLiveConfirm = computed(
  () => pendingLiveForm.value !== null,
);
const liveCapability = ref<{ supported: boolean; reason?: string } | null>(null);
let liveCapabilityRequest: Promise<{
  supported: boolean;
  reason?: string;
}> | null = null;
const showSortModal = ref(false);
const submitting = ref(false);
const sortSubmitting = ref(false);
const editingGroup = ref<AdminGroup | null>(null);
const deletingGroup = ref<AdminGroup | null>(null);
const duplicatingGroupIds = reactive(new Set<string>());
const showRateMultipliersModal = ref(false);
const showLevelRatesModal = ref(false);
const levelRatesGroup = ref<AdminGroup | null>(null);
const rateMultipliersGroup = ref<AdminGroup | null>(null);
const showRPMOverridesModal = ref(false);
const rpmOverridesGroup = ref<AdminGroup | null>(null);
const sortableGroups = ref<AdminGroup[]>([]);
type ConcreteGroupPlatform = Exclude<GroupPlatform, "composite">;
type CompositeRouteFormState = {
  public_model: string;
  match_type: CompositeRouteMatchType;
  target_platform: ConcreteGroupPlatform;
  upstream_model: string;
  endpoint: CompositeRouteEndpoint;
  priority: number;
  enabled: boolean;
  notes: string;
};

const showCompositeRoutesModal = ref(false);
const compositeRoutesGroup = ref<AdminGroup | null>(null);
const compositeRoutes = ref<CompositeModelRoute[]>([]);
const compositeRoutesLoading = ref(false);
const compositeRouteSaving = ref(false);
const compositeRouteEditingId = ref<string | null>(null);
const compositePreviewModel = ref("");
const compositePreviewEndpoint = ref<CompositeRouteEndpoint>("any");
const compositePreviewLoading = ref(false);
const compositePreviewDecision = ref<CompositeRouteDecision | null>(null);
const compositeRouteForm = reactive<CompositeRouteFormState>({
  public_model: "",
  match_type: "exact",
  target_platform: "openai",
  upstream_model: "",
  endpoint: "any",
  priority: 100,
  enabled: true,
  notes: "",
});
const createMessagesDispatchDefaults = createDefaultMessagesDispatchFormState();
const editMessagesDispatchDefaults = createDefaultMessagesDispatchFormState();
const createModelsListState = reactive(createInitialModelsListState());
const editModelsListState = reactive(createInitialModelsListState());
const createModelsListLoading = ref(false);
const editModelsListLoading = ref(false);
type ReasoningEffortPolicyFieldsExpose = {
  validate: () => boolean;
  resetValidation: () => void;
};
const createReasoningEffortPolicyRef = ref<ReasoningEffortPolicyFieldsExpose | null>(null);
const editReasoningEffortPolicyRef = ref<ReasoningEffortPolicyFieldsExpose | null>(null);
const modelsListCandidatesTracker = createModelsListCandidatesTracker();
const createModelsListSelectedCount = computed(
  () => createModelsListState.items.filter((item) => item.selected).length,
);
const editModelsListSelectedCount = computed(
  () => editModelsListState.items.filter((item) => item.selected).length,
);

const createForm = reactive({
  name: "",
  description: "",
  platform: "anthropic" as GroupPlatform,
  rate_multiplier: 1.0,
  is_exclusive: false,
  subscription_type: "standard" as SubscriptionType,
  daily_limit_points: null as number | null,
  weekly_limit_points: null as number | null,
  monthly_limit_points: null as number | null,
  long_context_pricing_enabled: true,
  model_pricing: [] as PricingFormEntry[],
  // 图片生成计费配置
  allow_image_generation: false,
  image_rate_independent: false,
  image_rate_multiplier: 1,
  image_price_1k: null as number | null,
  image_price_2k: null as number | null,
  image_price_4k: null as number | null,
  // 视频生成计费配置（仅 Grok 平台）
  video_rate_independent: false,
  video_rate_multiplier: 1,
  video_price_480p: null as number | null,
  video_price_720p: null as number | null,
  video_price_1080p: null as number | null,
  video_model_prices: createVideoModelPricesForm(),
  // Codex 网页搜索按次计费（仅 openai 平台使用）；null = 使用默认价 0.01
  web_search_price_per_call: null as number | null,
  search_price_per_1k: null as number | null,
  audio_realtime_price_per_min: null as number | null,
  audio_tts_price_per_million_chars: null as number | null,
  audio_stt_price_per_hour: null as number | null,
  // 高峰时段倍率配置
  peak_rate_enabled: false,
  peak_start: "",
  peak_end: "",
  peak_rate_multiplier: 1.0,
  // 分组利润控制；界面按百分比输入，提交时转小数
  profit_control_enabled: false,
  profit_min_margin_percent: 0,
  profit_safety_buffer_percent: 0,
  // Claude Code 客户端限制（仅 anthropic 平台使用）
  claude_code_only: false,
  fallback_group_id: null as string | null,
  fallback_group_id_on_invalid_request: null as string | null,
  // OpenAI Messages 调度配置（仅 openai 平台使用）
  allow_messages_dispatch: false,
  allow_live: false,
  opus_mapped_model: createMessagesDispatchDefaults.opus_mapped_model,
  sonnet_mapped_model: createMessagesDispatchDefaults.sonnet_mapped_model,
  haiku_mapped_model: createMessagesDispatchDefaults.haiku_mapped_model,
  exact_model_mappings: [] as MessagesDispatchMappingRow[],
  // 账号过滤控制
  require_oauth_only: false,
  require_privacy_set: false,
  // 模型路由开关
  model_routing_enabled: false,
  // 从分组复制账号
  copy_accounts_from_group_ids: [] as string[],
  // 分组级 RPM 限制（每用户每分钟最大请求数；0 = 不限制）
  rpm_limit: 0 as number,
  max_reasoning_effort: "",
  reasoning_effort_mappings: [] as ReasoningEffortMappingRow[],
});

// 简单账号类型（用于模型路由选择）
interface SimpleAccount {
  id: string;
  name: string;
}

// 模型路由规则类型
interface ModelRoutingRule {
  pattern: string;
  accounts: SimpleAccount[]; // 选中的账号对象数组
}

// 创建表单的模型路由规则
const createModelRoutingRules = ref<ModelRoutingRule[]>([]);

// 编辑表单的模型路由规则
const editModelRoutingRules = ref<ModelRoutingRule[]>([]);

// 规则对象稳定 key（避免使用 index 导致状态错位）
const resolveCreateRuleKey =
  createStableObjectKeyResolver<ModelRoutingRule>("create-rule");
const resolveEditRuleKey =
  createStableObjectKeyResolver<ModelRoutingRule>("edit-rule");
const resolveCreateMessagesDispatchRowKey =
  createStableObjectKeyResolver<MessagesDispatchMappingRow>(
    "create-messages-dispatch-row",
  );
const resolveEditMessagesDispatchRowKey =
  createStableObjectKeyResolver<MessagesDispatchMappingRow>(
    "edit-messages-dispatch-row",
  );

const getCreateRuleRenderKey = (rule: ModelRoutingRule) =>
  resolveCreateRuleKey(rule);
const getEditRuleRenderKey = (rule: ModelRoutingRule) =>
  resolveEditRuleKey(rule);
const getCreateMessagesDispatchRowKey = (row: MessagesDispatchMappingRow) =>
  resolveCreateMessagesDispatchRowKey(row);
const getEditMessagesDispatchRowKey = (row: MessagesDispatchMappingRow) =>
  resolveEditMessagesDispatchRowKey(row);

const getCreateRuleSearchKey = (rule: ModelRoutingRule) =>
  `create-${resolveCreateRuleKey(rule)}`;
const getEditRuleSearchKey = (rule: ModelRoutingRule) =>
  `edit-${resolveEditRuleKey(rule)}`;

const getRuleSearchKey = (rule: ModelRoutingRule, isEdit: boolean = false) => {
  return isEdit ? getEditRuleSearchKey(rule) : getCreateRuleSearchKey(rule);
};

// 账号搜索相关状态
const accountSearchKeyword = ref<Record<string, string>>({});
const accountSearchResults = ref<Record<string, SimpleAccount[]>>({});
const showAccountDropdown = ref<Record<string, boolean>>({});

const clearAccountSearchStateByKey = (key: string) => {
  delete accountSearchKeyword.value[key];
  delete accountSearchResults.value[key];
  delete showAccountDropdown.value[key];
};

const clearAllAccountSearchState = () => {
  accountSearchKeyword.value = {};
  accountSearchResults.value = {};
  showAccountDropdown.value = {};
};

const accountSearchRunner = useKeyedDebouncedSearch<SimpleAccount[]>({
  delay: 300,
  search: async (keyword, { signal }) => {
    const res = await adminAPI.accounts.list(
      1,
      20,
      {
        search: keyword,
        platform: "anthropic",
      },
      { signal },
    );
    return res.items.map((account) => ({ id: account.id, name: account.name }));
  },
  onSuccess: (key, result) => {
    accountSearchResults.value[key] = result;
  },
  onError: (key) => {
    accountSearchResults.value[key] = [];
  },
});

// 搜索账号（仅限 anthropic 平台）
const searchAccounts = (key: string) => {
  accountSearchRunner.trigger(key, accountSearchKeyword.value[key] || "");
};

const searchAccountsByRule = (
  rule: ModelRoutingRule,
  isEdit: boolean = false,
) => {
  searchAccounts(getRuleSearchKey(rule, isEdit));
};

// 选择账号
const selectAccount = (
  rule: ModelRoutingRule,
  account: SimpleAccount,
  isEdit: boolean = false,
) => {
  if (!rule) return;

  // 检查是否已选择
  if (!rule.accounts.some((a) => a.id === account.id)) {
    rule.accounts.push(account);
  }

  // 清空搜索
  const key = getRuleSearchKey(rule, isEdit);
  accountSearchKeyword.value[key] = "";
  showAccountDropdown.value[key] = false;
};

// 移除已选账号
const removeSelectedAccount = (
  rule: ModelRoutingRule,
  accountId: string,
  _isEdit: boolean = false,
) => {
  if (!rule) return;

  rule.accounts = rule.accounts.filter((a) => a.id !== accountId);
};

// 处理账号搜索输入框聚焦
const onAccountSearchFocus = (
  rule: ModelRoutingRule,
  isEdit: boolean = false,
) => {
  const key = getRuleSearchKey(rule, isEdit);
  showAccountDropdown.value[key] = true;
  // 如果没有搜索结果，触发一次搜索
  if (!accountSearchResults.value[key]?.length) {
    searchAccounts(key);
  }
};

// 添加创建表单的路由规则
const addCreateRoutingRule = () => {
  createModelRoutingRules.value.push({ pattern: "", accounts: [] });
};

// 删除创建表单的路由规则
const removeCreateRoutingRule = (rule: ModelRoutingRule) => {
  const index = createModelRoutingRules.value.indexOf(rule);
  if (index === -1) return;

  const key = getCreateRuleSearchKey(rule);
  accountSearchRunner.clearKey(key);
  clearAccountSearchStateByKey(key);
  createModelRoutingRules.value.splice(index, 1);
};

// 添加编辑表单的路由规则
const addEditRoutingRule = () => {
  editModelRoutingRules.value.push({ pattern: "", accounts: [] });
};

// 删除编辑表单的路由规则
const removeEditRoutingRule = (rule: ModelRoutingRule) => {
  const index = editModelRoutingRules.value.indexOf(rule);
  if (index === -1) return;

  const key = getEditRuleSearchKey(rule);
  accountSearchRunner.clearKey(key);
  clearAccountSearchStateByKey(key);
  editModelRoutingRules.value.splice(index, 1);
};

const resetModelsListState = (
  state: typeof createModelsListState,
  config?: Parameters<typeof createInitialModelsListState>[0],
) => {
  const fresh = createInitialModelsListState(config);
  state.enabled = fresh.enabled;
  state.savedModels = fresh.savedModels;
  state.items = fresh.items;
};

const loadModelsListCandidates = async (
  mode: "create" | "edit",
  groupID: string | null,
  platform: GroupPlatform,
) => {
  const request = { mode, groupID, platform };
  const requestID = modelsListCandidatesTracker.next(request);
  const state = mode === "create" ? createModelsListState : editModelsListState;
  const loadingRef = mode === "create" ? createModelsListLoading : editModelsListLoading;
  loadingRef.value = true;
  try {
    const models = await adminAPI.groups.getModelsListCandidates(groupID, platform);
    if (!modelsListCandidatesTracker.isCurrent(requestID, request)) {
      return;
    }
    setModelsListCandidates(state, models);
  } catch (error) {
    if (!modelsListCandidatesTracker.isCurrent(requestID, request)) {
      return;
    }
    console.error("Error loading group models list candidates:", error);
  } finally {
    if (modelsListCandidatesTracker.isCurrent(requestID, request)) {
      loadingRef.value = false;
    }
  }
};

const moveCreateModelsListItem = (fromIndex: number, toIndex: number) => {
  moveModelsListItem(createModelsListState, fromIndex, toIndex);
};

const moveEditModelsListItem = (fromIndex: number, toIndex: number) => {
  moveModelsListItem(editModelsListState, fromIndex, toIndex);
};

// 将 UI 格式的路由规则转换为 API 格式
const convertRoutingRulesToApiFormat = (
  rules: ModelRoutingRule[],
): Record<string, string[]> | null => {
  const result: Record<string, string[]> = {};
  let hasValidRules = false;

  for (const rule of rules) {
    const pattern = rule.pattern.trim();
    if (!pattern) continue;

    const accountIds = rule.accounts.map((a) => a.id).filter(Boolean);

    if (accountIds.length > 0) {
      result[pattern] = accountIds;
      hasValidRules = true;
    }
  }

  return hasValidRules ? result : null;
};

// 将 API 格式的路由规则转换为 UI 格式（需要加载账号名称）
const convertApiFormatToRoutingRules = async (
  apiFormat: Record<string, string[]> | null,
): Promise<ModelRoutingRule[]> => {
  if (!apiFormat) return [];

  const rules: ModelRoutingRule[] = [];
  for (const [pattern, accountIds] of Object.entries(apiFormat)) {
    // 加载账号信息
    const accounts: SimpleAccount[] = [];
    for (const id of accountIds) {
      try {
        const account = await adminAPI.accounts.getById(id);
        accounts.push({ id: account.id, name: account.name });
      } catch {
        // 如果账号不存在，仍然显示 ID
        accounts.push({ id, name: `#${id}` });
      }
    }
    rules.push({ pattern, accounts });
  }
  return rules;
};

const editForm = reactive({
  name: "",
  description: "",
  platform: "anthropic" as GroupPlatform,
  rate_multiplier: 1.0,
  is_exclusive: false,
  status: "active" as "active" | "inactive",
  subscription_type: "standard" as SubscriptionType,
  daily_limit_points: null as number | null,
  weekly_limit_points: null as number | null,
  monthly_limit_points: null as number | null,
  long_context_pricing_enabled: true,
  model_pricing: [] as PricingFormEntry[],
  // 图片生成计费配置
  allow_image_generation: false,
  image_rate_independent: false,
  image_rate_multiplier: 1,
  image_price_1k: null as number | null,
  image_price_2k: null as number | null,
  image_price_4k: null as number | null,
  // 视频生成计费配置（仅 Grok 平台）
  video_rate_independent: false,
  video_rate_multiplier: 1,
  video_price_480p: null as number | null,
  video_price_720p: null as number | null,
  video_price_1080p: null as number | null,
  video_model_prices: createVideoModelPricesForm(),
  // Codex 网页搜索按次计费（仅 openai 平台使用）；null = 使用默认价 0.01
  web_search_price_per_call: null as number | null,
  search_price_per_1k: null as number | null,
  audio_realtime_price_per_min: null as number | null,
  audio_tts_price_per_million_chars: null as number | null,
  audio_stt_price_per_hour: null as number | null,
  // 高峰时段倍率配置
  peak_rate_enabled: false,
  peak_start: "",
  peak_end: "",
  peak_rate_multiplier: 1.0,
  // 分组利润控制；界面按百分比输入，提交时转小数
  profit_control_enabled: false,
  profit_min_margin_percent: 0,
  profit_safety_buffer_percent: 0,
  // Claude Code 客户端限制（仅 anthropic 平台使用）
  claude_code_only: false,
  fallback_group_id: null as string | null,
  fallback_group_id_on_invalid_request: null as string | null,
  // OpenAI Messages 调度配置（仅 openai 平台使用）
  allow_messages_dispatch: false,
  allow_live: false,
  default_mapped_model: '',
  opus_mapped_model: editMessagesDispatchDefaults.opus_mapped_model,
  sonnet_mapped_model: editMessagesDispatchDefaults.sonnet_mapped_model,
  haiku_mapped_model: editMessagesDispatchDefaults.haiku_mapped_model,
  exact_model_mappings: [] as MessagesDispatchMappingRow[],
  // 账号过滤控制
  require_oauth_only: false,
  require_privacy_set: false,
  // 模型路由开关
  model_routing_enabled: false,
  // 从分组复制账号
  copy_accounts_from_group_ids: [] as string[],
  // 分组级 RPM 限制（每用户每分钟最大请求数；0 = 不限制）
  rpm_limit: 0 as number,
  max_reasoning_effort: "",
  reasoning_effort_mappings: [] as ReasoningEffortMappingRow[],
});

type ImagePricingFormState = {
  platform: GroupPlatform;
  allow_image_generation: boolean;
  rate_multiplier: number;
  image_rate_independent: boolean;
  image_rate_multiplier: number;
  image_price_1k: number | string | null;
  image_price_2k: number | string | null;
  image_price_4k: number | string | null;
  peak_rate_enabled: boolean;
  peak_start: string;
  peak_end: string;
  peak_rate_multiplier: number;
};

type VideoPricingFormState = {
  platform: GroupPlatform;
  rate_multiplier: number;
  video_rate_independent: boolean;
  video_rate_multiplier: number;
  video_price_480p: number | string | null;
  video_price_720p: number | string | null;
  video_price_1080p: number | string | null;
};

const imagePricingTiers = [
  { key: "image_price_1k", label: "1K" },
  { key: "image_price_2k", label: "2K" },
  { key: "image_price_4k", label: "4K" },
] as const;

const videoPricingTiers = [
  { key: "video_price_480p", label: "480p" },
  { key: "video_price_720p", label: "720p" },
  { key: "video_price_1080p", label: "1080p" },
] as const;

const normalizePreviewNumber = (value: number | string | null | undefined, fallback = 0) => {
  if (value === null || value === undefined || value === "") {
    return fallback;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parsePreviewPrice = (value: number | string | null | undefined) => {
  if (value === null || value === undefined || value === "") {
    return null;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : null;
};

const formatImagePricePreview = (value: number | string | null | undefined) => {
  if (value === null || value === undefined || value === "") {
    return t("admin.groups.imagePricing.notConfigured");
  }
  const price = Number(value);
  if (!Number.isFinite(price) || price < 0) {
    return t("admin.groups.imagePricing.notConfigured");
  }
  return `$${price.toFixed(6).replace(/0+$/, "").replace(/\.$/, "")}`;
};

const formatVideoPricePreview = (value: number | string | null | undefined) => {
  if (value === null || value === undefined || value === "") {
    return t("admin.groups.videoPricing.notConfigured");
  }
  const price = Number(value);
  if (!Number.isFinite(price) || price < 0) {
    return t("admin.groups.videoPricing.notConfigured");
  }
  return `$${price.toFixed(6).replace(/0+$/, "").replace(/\.$/, "")}`;
};

const buildImageFinalPricePreview = (form: ImagePricingFormState) => {
  const imageMultiplier = form.image_rate_independent
    ? normalizePreviewNumber(form.image_rate_multiplier, 1)
    : normalizePreviewNumber(form.rate_multiplier, 1);
  const multiplier = imageMultiplier;
  return imagePricingTiers.map((tier) => {
    const basePrice =
      parsePreviewPrice(form[tier.key]) ??
      getDefaultImagePreviewPrice(form.platform, tier.key);
    return {
      label: tier.label,
      value: basePrice !== null
        ? formatImagePricePreview(basePrice * multiplier)
        : t("admin.groups.imagePricing.notConfigured"),
    };
  });
};

const buildVideoFinalPricePreview = (form: VideoPricingFormState) => {
  const multiplier = form.video_rate_independent
    ? normalizePreviewNumber(form.video_rate_multiplier, 1)
    : normalizePreviewNumber(form.rate_multiplier, 1);
  return videoPricingTiers.map((tier) => {
    const basePrice =
      parsePreviewPrice(form[tier.key]) ??
      getDefaultVideoPreviewPrice(form.platform, tier.key);
    return {
      label: tier.label,
      value: basePrice !== null
        ? formatVideoPricePreview(basePrice * multiplier)
        : t("admin.groups.videoPricing.notConfigured"),
    };
  });
};

const createImageFinalPricePreview = computed(() =>
  buildImageFinalPricePreview(createForm),
);
const editImageFinalPricePreview = computed(() =>
  buildImageFinalPricePreview(editForm),
);
const createVideoFinalPricePreview = computed(() =>
  buildVideoFinalPricePreview(createForm),
);
const editVideoFinalPricePreview = computed(() =>
  buildVideoFinalPricePreview(editForm),
);

// Codex 网页搜索单次默认价（与后端 defaultWebSearchPricePerCall 一致，官方 $10/1000 次）
const DEFAULT_WEB_SEARCH_PRICE_PER_CALL = 0.01;

const buildWebSearchFinalPricePreview = (form: {
  web_search_price_per_call: number | string | null;
  rate_multiplier: number | string | null;
}) => {
  const basePrice =
    parsePreviewPrice(form.web_search_price_per_call) ??
    DEFAULT_WEB_SEARCH_PRICE_PER_CALL;
  const multiplier = normalizePreviewNumber(form.rate_multiplier, 1);
  return formatImagePricePreview(basePrice * multiplier);
};

const createWebSearchFinalPricePreview = computed(() =>
  buildWebSearchFinalPricePreview(createForm),
);
const editWebSearchFinalPricePreview = computed(() =>
  buildWebSearchFinalPricePreview(editForm),
);

// 根据分组类型返回不同的删除确认消息
const deleteConfirmMessage = computed(() => {
  if (!deletingGroup.value) {
    return "";
  }
  if (deletingGroup.value.subscription_type === "subscription") {
    return t("admin.groups.deleteConfirmSubscription", {
      name: deletingGroup.value.name,
    });
  }
  return t("admin.groups.deleteConfirm", { name: deletingGroup.value.name });
});

const loadLiveCapability = async () => {
  if (liveCapability.value) return liveCapability.value;
  if (!liveCapabilityRequest) {
    liveCapabilityRequest = adminAPI.groups
      .getLiveCapability()
      .catch(() => ({ supported: false }))
      .finally(() => {
        liveCapabilityRequest = null;
      });
  }
  liveCapability.value = await liveCapabilityRequest;
  return liveCapability.value ?? { supported: false };
};

const toggleLive = async (target: "create" | "edit") => {
  const form = target === "create" ? createForm : editForm;
  if (form.allow_live) {
    form.allow_live = false;
    return;
  }
  const capability = await loadLiveCapability();
  if (capability.supported) {
    form.allow_live = true;
    return;
  }
  pendingLiveForm.value = target;
};

const confirmUnsupportedLive = () => {
  if (pendingLiveForm.value === "create") createForm.allow_live = true;
  if (pendingLiveForm.value === "edit") editForm.allow_live = true;
  pendingLiveForm.value = null;
};

const cancelUnsupportedLive = () => {
  pendingLiveForm.value = null;
};

const loadGroups = async () => {
  if (abortController) {
    abortController.abort();
  }
  const currentController = new AbortController();
  abortController = currentController;
  const { signal } = currentController;
  loading.value = true;
  try {
    const response = await adminAPI.groups.list(
      pagination.page,
      pagination.page_size,
      {
        platform: (filters.platform as GroupPlatform) || undefined,
        status: filters.status as any,
        is_exclusive: filters.is_exclusive
          ? filters.is_exclusive === "true"
          : undefined,
        search: searchQuery.value.trim() || undefined,
        sort_by: sortState.sort_by,
        sort_order: sortState.sort_order,
      },
      { signal },
    );
    if (signal.aborted) return;
    groups.value = response.items;
    pagination.total = response.total;
    pagination.pages = response.pages;
    if (hasVisibleUsageSummaryConsumer.value) {
      loadUsageSummary();
    } else {
      usageLoading.value = false;
    }
    if (hasVisibleCapacityColumn.value) {
      loadCapacitySummary();
    }
  } catch (error: any) {
    if (
      signal.aborted ||
      error?.name === "AbortError" ||
      error?.code === "ERR_CANCELED"
    ) {
      return;
    }
    appStore.showError(t("admin.groups.failedToLoad"));
    console.error("Error loading groups:", error);
  } finally {
    if (abortController === currentController && !signal.aborted) {
      loading.value = false;
    }
  }
};

const groupLimitPoints = (
  group: AdminGroup,
  window: "daily" | "weekly" | "monthly",
): number => group[`${window}_limit_points`] ?? group[`${window}_limit_usd`] ?? 0;

const getQuotaUsageClass = (
  used: number,
  limit: number | null | undefined,
): string => {
  if (!limit || limit <= 0) {
    return "views-admin-groups-view__state";
  }
  const ratio = used / limit;
  if (ratio >= 1) {
    return "views-admin-groups-view__state-2";
  }
  if (ratio >= 0.8) {
    return "views-admin-groups-view__state-3";
  }
  return "views-admin-groups-view__state";
};

const loadUsageSummary = async () => {
  if (!hasVisibleUsageSummaryConsumer.value) {
    usageLoading.value = false;
    return;
  }
  usageLoading.value = true;
  try {
    const data = await adminAPI.groups.getUsageSummary();
    const map = new Map<string, GroupUsageSummary>();
    for (const item of data) {
      map.set(item.group_id, {
        today_cost: item.today_cost,
        yesterday_cost: item.yesterday_cost,
        total_cost: item.total_cost,
      });
    }
    usageMap.value = map;
  } catch (error) {
    console.error("Error loading group usage summary:", error);
  } finally {
    usageLoading.value = false;
  }
};

const isUsageSummaryPending = (groupId: string) =>
  usageLoading.value && !usageMap.value.has(groupId);

const loadCapacitySummary = async () => {
  if (!hasVisibleCapacityColumn.value) {
    return;
  }
  try {
    const data = await adminAPI.groups.getCapacitySummary();
    const map = new Map<
      string,
      {
        concurrencyUsed: number;
        concurrencyMax: number;
        sessionsUsed: number;
        sessionsMax: number;
        rpmUsed: number;
        rpmMax: number;
      }
    >();
    for (const item of data) {
      map.set(item.group_id, {
        concurrencyUsed: item.concurrency_used,
        concurrencyMax: item.concurrency_max,
        sessionsUsed: item.sessions_used,
        sessionsMax: item.sessions_max,
        rpmUsed: item.rpm_used,
        rpmMax: item.rpm_max,
      });
    }
    capacityMap.value = map;
  } catch (error) {
    console.error("Error loading group capacity summary:", error);
  }
};

let searchTimeout: ReturnType<typeof setTimeout>;
const handleSearch = () => {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    pagination.page = 1;
    loadGroups();
  }, 300);
};

const handleFilterChange = () => {
  pagination.page = 1;
  loadGroups();
};

const handlePageChange = (page: number) => {
  pagination.page = page;
  loadGroups();
};

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize;
  pagination.page = 1;
  loadGroups();
};

const handleSort = (key: string, order: 'asc' | 'desc') => {
  sortState.sort_by = key;
  sortState.sort_order = order;
  pagination.page = 1;
  loadGroups();
};

const openCreateModal = () => {
  showCreateModal.value = true;
  loadModelsListCandidates("create", null, createForm.platform);
};

const closeCreateModal = () => {
  showCreateModal.value = false;
  createModelRoutingRules.value.forEach((rule) => {
    accountSearchRunner.clearKey(getCreateRuleSearchKey(rule));
  });
  clearAllAccountSearchState();
  createForm.name = "";
  createForm.description = "";
  createForm.platform = "anthropic";
  createForm.rate_multiplier = 1.0;
  createForm.is_exclusive = false;
  createForm.subscription_type = "standard";
  createForm.daily_limit_points = null;
  createForm.weekly_limit_points = null;
  createForm.monthly_limit_points = null;
  createForm.allow_image_generation = false;
  createForm.image_rate_independent = false;
  createForm.image_rate_multiplier = 1;
  createForm.image_price_1k = null;
  createForm.image_price_2k = null;
  createForm.image_price_4k = null;
  createForm.video_rate_independent = false;
  createForm.video_rate_multiplier = 1;
  createForm.video_price_480p = null;
  createForm.video_price_720p = null;
  createForm.video_price_1080p = null;
  createForm.video_model_prices = createVideoModelPricesForm();
  createForm.long_context_pricing_enabled = true;
  createForm.model_pricing = [];
  createForm.web_search_price_per_call = null;
  createForm.search_price_per_1k = null;
  createForm.audio_realtime_price_per_min = null;
  createForm.audio_tts_price_per_million_chars = null;
  createForm.audio_stt_price_per_hour = null;
  createForm.peak_rate_enabled = false;
  createForm.peak_start = "";
  createForm.peak_end = "";
  createForm.peak_rate_multiplier = 1.0;
  createForm.profit_control_enabled = false;
  createForm.profit_min_margin_percent = 0;
  createForm.profit_safety_buffer_percent = 0;
  createForm.claude_code_only = false;
  createForm.fallback_group_id = null;
  createForm.fallback_group_id_on_invalid_request = null;
  resetMessagesDispatchFormState(createForm);
  createForm.allow_live = false;
  createForm.require_oauth_only = false;
  createForm.require_privacy_set = false;
  createForm.copy_accounts_from_group_ids = [];
  createForm.rpm_limit = 0;
  createForm.max_reasoning_effort = "";
  createForm.reasoning_effort_mappings = [];
  createReasoningEffortPolicyRef.value?.resetValidation();
  resetModelsListState(createModelsListState);
  createModelRoutingRules.value = [];
};

const normalizeOptionalLimit = (
  value: number | string | null | undefined,
): number | null => {
  if (value === null || value === undefined) {
    return null;
  }

  if (typeof value === "string") {
    const trimmed = value.trim();
    if (!trimmed) {
      return null;
    }
    const parsed = Number(trimmed);
    return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
  }

  return Number.isFinite(value) && value > 0 ? value : null;
};

const normalizeRateMultiplier = (
  value: number | string | null | undefined,
): number => {
  if (value === null || value === undefined || value === "") {
    return 1;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : 1;
};

// 利润控制表单辅助（换算与校验逻辑见 groupsProfitControl.ts，便于单测）。
const percentToDecimal = profitPercentToDecimal;
const decimalToPercent = profitDecimalToPercent;

const validateProfitControlForm = (form: ProfitControlFormState): boolean => {
  const errorKey = validateProfitControlFormState(form);
  if (errorKey) {
    appStore.showError(t(`admin.groups.profitControl.${errorKey}`));
    return false;
  }
  return true;
};

const handleCreateGroup = async () => {
  if (!createForm.name.trim()) {
    appStore.showError(t("admin.groups.nameRequired"));
    return;
  }
  if (
    supportsReasoningEffortPolicyPlatform(createForm.platform) &&
    createReasoningEffortPolicyRef.value &&
    !createReasoningEffortPolicyRef.value.validate()
  ) {
    return;
  }
  if (!validateProfitControlForm(createForm)) {
    return;
  }
  submitting.value = true;
  try {
    const {
      video_model_prices: _createFormVideoModelPrices,
      ...createGroupForm
    } = createForm;
    const videoModelPrices = serializeVideoModelPrices(
      createForm.video_model_prices,
    );
    // 构建请求数据，包含模型路由配置
    const requestData = {
      ...createGroupForm,
      model_pricing: groupPricingToAPI(
        createForm.model_pricing,
        createForm.platform,
      ),
      daily_limit_points: normalizeOptionalLimit(
        createForm.daily_limit_points as number | string | null,
      ),
      weekly_limit_points: normalizeOptionalLimit(
        createForm.weekly_limit_points as number | string | null,
      ),
      monthly_limit_points: normalizeOptionalLimit(
        createForm.monthly_limit_points as number | string | null,
      ),
      ...(Object.keys(videoModelPrices).length > 0
        ? { video_model_prices: videoModelPrices }
        : {}),
      model_routing: convertRoutingRulesToApiFormat(
        createModelRoutingRules.value,
      ),
      models_list_config: buildModelsListConfig(createModelsListState),
      messages_dispatch_model_config:
        createForm.platform === "openai"
          ? messagesDispatchFormStateToConfig({
              allow_messages_dispatch: createForm.allow_messages_dispatch,
              opus_mapped_model: createForm.opus_mapped_model,
              sonnet_mapped_model: createForm.sonnet_mapped_model,
              haiku_mapped_model: createForm.haiku_mapped_model,
              exact_model_mappings: createForm.exact_model_mappings,
            })
          : undefined,
      reasoning_effort_mappings: reasoningEffortMappingsToAPI(
        createForm.reasoning_effort_mappings,
      ),
      // 利润控制：界面百分比转小数提交
      profit_control_enabled:
        isProfitControlPlatform(createForm.platform) &&
        createForm.profit_control_enabled,
      profit_min_margin: percentToDecimal(createForm.profit_min_margin_percent),
      profit_safety_buffer: percentToDecimal(
        createForm.profit_safety_buffer_percent,
      ),
    };
    delete (requestData as Record<string, unknown>).profit_min_margin_percent;
    delete (requestData as Record<string, unknown>).profit_safety_buffer_percent;
    // v-model.number 清空输入框时产生 ""，转为 null 让后端设为无限制
    const emptyToNull = (v: any) => (v === "" ? null : v);
    requestData.daily_limit_points = emptyToNull(requestData.daily_limit_points);
    requestData.weekly_limit_points = emptyToNull(requestData.weekly_limit_points);
    requestData.monthly_limit_points = emptyToNull(requestData.monthly_limit_points);
    requestData.image_rate_multiplier = normalizeRateMultiplier(
      requestData.image_rate_multiplier,
    );
    requestData.video_rate_multiplier = normalizeRateMultiplier(
      requestData.video_rate_multiplier,
    );
    // 媒体价格输入清空时 v-model.number 产生 ""，直接提交会被后端 *float64 反序列化拒绝（400），
    // 创建时按"未配置"（null）处理。
    requestData.image_price_1k = emptyToNull(requestData.image_price_1k);
    requestData.image_price_2k = emptyToNull(requestData.image_price_2k);
    requestData.image_price_4k = emptyToNull(requestData.image_price_4k);
    requestData.video_price_480p = emptyToNull(requestData.video_price_480p);
    requestData.video_price_720p = emptyToNull(requestData.video_price_720p);
    requestData.video_price_1080p = emptyToNull(requestData.video_price_1080p);
    requestData.search_price_per_1k = emptyToNull(
      requestData.search_price_per_1k,
    );
    requestData.audio_realtime_price_per_min = emptyToNull(
      requestData.audio_realtime_price_per_min,
    );
    requestData.audio_tts_price_per_million_chars = emptyToNull(
      requestData.audio_tts_price_per_million_chars,
    );
    requestData.audio_stt_price_per_hour = emptyToNull(
      requestData.audio_stt_price_per_hour,
    );
    requestData.web_search_price_per_call = emptyToNull(
      requestData.web_search_price_per_call,
    );
    requestData.peak_rate_enabled = createForm.peak_rate_enabled;
    requestData.peak_start = createForm.peak_start;
    requestData.peak_end = createForm.peak_end;
    requestData.peak_rate_multiplier = normalizeRateMultiplier(
      createForm.peak_rate_multiplier,
    );
    await adminAPI.groups.create(requestData);
    appStore.showSuccess(t("admin.groups.groupCreated"));
    closeCreateModal();
    loadGroups();
    // Only advance tour if active, on submit step, and creation succeeded
    if (onboardingStore.isCurrentStep('[data-tour="group-form-submit"]')) {
      onboardingStore.nextStep(500);
    }
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail || t("admin.groups.failedToCreate"),
    );
    console.error("Error creating group:", error);
    // Don't advance tour on error
  } finally {
    submitting.value = false;
  }
};

const handleEdit = async (group: AdminGroup) => {
  editingGroup.value = group;
  editForm.name = group.name;
  editForm.description = group.description || "";
  editForm.platform = group.platform;
  editForm.rate_multiplier = group.rate_multiplier;
  editForm.is_exclusive = group.is_exclusive;
  editForm.status = group.status;
  editForm.subscription_type = group.subscription_type || "standard";
  editForm.daily_limit_points = group.daily_limit_points ?? group.daily_limit_usd;
  editForm.weekly_limit_points = group.weekly_limit_points ?? group.weekly_limit_usd;
  editForm.monthly_limit_points = group.monthly_limit_points ?? group.monthly_limit_usd;
  editForm.long_context_pricing_enabled =
    group.long_context_pricing_enabled ?? true;
  editForm.model_pricing = groupPricingFromAPI(group.model_pricing);
  editForm.allow_image_generation = group.allow_image_generation ?? false;
  editForm.image_rate_independent = group.image_rate_independent ?? false;
  editForm.image_rate_multiplier = group.image_rate_multiplier ?? 1;
  editForm.image_price_1k = group.image_price_1k;
  editForm.image_price_2k = group.image_price_2k;
  editForm.image_price_4k = group.image_price_4k;
  editForm.video_rate_independent = group.video_rate_independent ?? false;
  editForm.video_rate_multiplier = group.video_rate_multiplier ?? 1;
  editForm.video_price_480p = group.video_price_480p;
  editForm.video_price_720p = group.video_price_720p;
  editForm.video_price_1080p = group.video_price_1080p;
  editForm.video_model_prices = createVideoModelPricesForm(
    group.video_model_prices,
  );
  editForm.web_search_price_per_call = group.web_search_price_per_call ?? null;
  editForm.search_price_per_1k = group.search_price_per_1k ?? null;
  editForm.audio_realtime_price_per_min = group.audio_realtime_price_per_min ?? null;
  editForm.audio_tts_price_per_million_chars = group.audio_tts_price_per_million_chars ?? null;
  editForm.audio_stt_price_per_hour = group.audio_stt_price_per_hour ?? null;
  editForm.peak_rate_enabled = group.peak_rate_enabled ?? false;
  editForm.peak_start = group.peak_start ?? "";
  editForm.peak_end = group.peak_end ?? "";
  editForm.peak_rate_multiplier = group.peak_rate_multiplier ?? 1.0;
  editForm.profit_control_enabled = group.profit_control_enabled ?? false;
  editForm.profit_min_margin_percent = decimalToPercent(
    group.profit_min_margin ?? 0,
  );
  editForm.profit_safety_buffer_percent = decimalToPercent(
    group.profit_safety_buffer ?? 0,
  );
  editForm.claude_code_only = group.claude_code_only || false;
  editForm.fallback_group_id = group.fallback_group_id;
  editForm.fallback_group_id_on_invalid_request =
    group.fallback_group_id_on_invalid_request;
  const messagesDispatchFormState = messagesDispatchConfigToFormState(
    group.messages_dispatch_model_config,
  );
  editForm.allow_messages_dispatch =
    group.allow_messages_dispatch ||
    messagesDispatchFormState.allow_messages_dispatch;
  editForm.allow_live = group.allow_live ?? false;
  editForm.opus_mapped_model = messagesDispatchFormState.opus_mapped_model;
  editForm.sonnet_mapped_model = messagesDispatchFormState.sonnet_mapped_model;
  editForm.haiku_mapped_model = messagesDispatchFormState.haiku_mapped_model;
  editForm.exact_model_mappings =
    messagesDispatchFormState.exact_model_mappings;
  editForm.require_oauth_only = group.require_oauth_only ?? false;
  editForm.require_privacy_set = group.require_privacy_set ?? false;
  editForm.model_routing_enabled = group.model_routing_enabled || false;
  editForm.copy_accounts_from_group_ids = []; // 复制账号字段每次编辑时重置为空
  editForm.rpm_limit = group.rpm_limit ?? 0;
  editForm.max_reasoning_effort = normalizeReasoningEffortForPlatform(
    group.platform,
    group.max_reasoning_effort,
  );
  editForm.reasoning_effort_mappings = reasoningEffortMappingsToRows(
    group.reasoning_effort_mappings,
    group.platform,
  );
  resetModelsListState(editModelsListState, group.models_list_config);
  // 加载模型路由规则（异步加载账号名称）
  editModelRoutingRules.value = await convertApiFormatToRoutingRules(
    group.model_routing,
  );
  loadModelsListCandidates("edit", group.id, group.platform);
  showEditModal.value = true;
};

const closeEditModal = () => {
  editModelRoutingRules.value.forEach((rule) => {
    accountSearchRunner.clearKey(getEditRuleSearchKey(rule));
  });
  clearAllAccountSearchState();
  showEditModal.value = false;
  editingGroup.value = null;
  editForm.max_reasoning_effort = "";
  editForm.reasoning_effort_mappings = [];
  editReasoningEffortPolicyRef.value?.resetValidation();
  editModelRoutingRules.value = [];
  editForm.copy_accounts_from_group_ids = [];
  editForm.peak_rate_enabled = false;
  editForm.peak_start = "";
  editForm.peak_end = "";
  editForm.peak_rate_multiplier = 1.0;
  editForm.profit_control_enabled = false;
  editForm.profit_min_margin_percent = 0;
  editForm.profit_safety_buffer_percent = 0;
  editForm.video_rate_independent = false;
  editForm.video_rate_multiplier = 1;
  editForm.video_price_480p = null;
  editForm.video_price_720p = null;
  editForm.video_price_1080p = null;
  editForm.video_model_prices = createVideoModelPricesForm();
  editForm.long_context_pricing_enabled = true;
  editForm.model_pricing = [];
  editForm.web_search_price_per_call = null;
  editForm.search_price_per_1k = null;
  editForm.audio_realtime_price_per_min = null;
  editForm.audio_tts_price_per_million_chars = null;
  editForm.audio_stt_price_per_hour = null;
  resetMessagesDispatchFormState(editForm);
  editForm.allow_live = false;
  resetModelsListState(editModelsListState);
};

const handleUpdateGroup = async () => {
  if (!editingGroup.value) return;
  if (!editForm.name.trim()) {
    appStore.showError(t("admin.groups.nameRequired"));
    return;
  }
  if (
    supportsReasoningEffortPolicyPlatform(editForm.platform) &&
    editReasoningEffortPolicyRef.value &&
    !editReasoningEffortPolicyRef.value.validate()
  ) {
    return;
  }
  if (!validateProfitControlForm(editForm)) {
    return;
  }

  submitting.value = true;
  try {
    const payload = {
      ...editForm,
      model_pricing: groupPricingToAPI(
        editForm.model_pricing,
        editForm.platform,
      ),
      daily_limit_points: normalizeOptionalLimit(
        editForm.daily_limit_points as number | string | null,
      ),
      weekly_limit_points: normalizeOptionalLimit(
        editForm.weekly_limit_points as number | string | null,
      ),
      monthly_limit_points: normalizeOptionalLimit(
        editForm.monthly_limit_points as number | string | null,
      ),
      video_model_prices: serializeVideoModelPrices(
        editForm.video_model_prices,
      ),
      fallback_group_id: editForm.fallback_group_id,
      fallback_group_id_on_invalid_request:
        editForm.fallback_group_id_on_invalid_request,
      model_routing: convertRoutingRulesToApiFormat(
        editModelRoutingRules.value,
      ),
      models_list_config: buildModelsListConfig(editModelsListState),
      messages_dispatch_model_config:
        editForm.platform === "openai"
          ? messagesDispatchFormStateToConfig({
              allow_messages_dispatch: editForm.allow_messages_dispatch,
              opus_mapped_model: editForm.opus_mapped_model,
              sonnet_mapped_model: editForm.sonnet_mapped_model,
              haiku_mapped_model: editForm.haiku_mapped_model,
              exact_model_mappings: editForm.exact_model_mappings,
            })
          : undefined,
      reasoning_effort_mappings: reasoningEffortMappingsToAPI(
        editForm.reasoning_effort_mappings,
      ),
      // 利润控制：界面百分比转小数提交
      profit_control_enabled:
        isProfitControlPlatform(editForm.platform) &&
        editForm.profit_control_enabled,
      profit_min_margin: percentToDecimal(editForm.profit_min_margin_percent),
      profit_safety_buffer: percentToDecimal(
        editForm.profit_safety_buffer_percent,
      ),
    };
    delete (payload as Record<string, unknown>).profit_min_margin_percent;
    delete (payload as Record<string, unknown>).profit_safety_buffer_percent;
    // v-model.number 清空输入框时产生 ""，转为 null 让后端设为无限制
    const emptyToNull = (v: any) => (v === "" ? null : v);
    payload.daily_limit_points = emptyToNull(payload.daily_limit_points);
    payload.weekly_limit_points = emptyToNull(payload.weekly_limit_points);
    payload.monthly_limit_points = emptyToNull(payload.monthly_limit_points);
    payload.image_rate_multiplier = normalizeRateMultiplier(
      payload.image_rate_multiplier,
    );
    payload.video_rate_multiplier = normalizeRateMultiplier(
      payload.video_rate_multiplier,
    );
    // 媒体价格输入清空时 v-model.number 产生 ""，直接提交会被后端 *float64 反序列化拒绝（400）。
    // 更新语义中 null 表示"不修改"，因此清空后的字段发送 -1：后端 normalizePrice 将负价归一为
    // NULL，从而真正清除已配置的价格。
    const emptyPriceToClear = (v: any) => (v === "" || v === null ? -1 : v);
    payload.image_price_1k = emptyPriceToClear(payload.image_price_1k);
    payload.image_price_2k = emptyPriceToClear(payload.image_price_2k);
    payload.image_price_4k = emptyPriceToClear(payload.image_price_4k);
    payload.video_price_480p = emptyPriceToClear(payload.video_price_480p);
    payload.video_price_720p = emptyPriceToClear(payload.video_price_720p);
    payload.video_price_1080p = emptyPriceToClear(payload.video_price_1080p);
    payload.search_price_per_1k = emptyPriceToClear(
      payload.search_price_per_1k,
    );
    payload.audio_realtime_price_per_min = emptyPriceToClear(
      payload.audio_realtime_price_per_min,
    );
    payload.audio_tts_price_per_million_chars = emptyPriceToClear(
      payload.audio_tts_price_per_million_chars,
    );
    payload.audio_stt_price_per_hour = emptyPriceToClear(
      payload.audio_stt_price_per_hour,
    );
    payload.web_search_price_per_call = emptyPriceToClear(
      payload.web_search_price_per_call,
    );
    payload.peak_rate_enabled = editForm.peak_rate_enabled;
    payload.peak_start = editForm.peak_start;
    payload.peak_end = editForm.peak_end;
    payload.peak_rate_multiplier = normalizeRateMultiplier(
      editForm.peak_rate_multiplier,
    );
    await adminAPI.groups.update(editingGroup.value.id, payload);
    appStore.showSuccess(t("admin.groups.groupUpdated"));
    closeEditModal();
    loadGroups();
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail || t("admin.groups.failedToUpdate"),
    );
    console.error("Error updating group:", error);
  } finally {
    submitting.value = false;
  }
};

const addCreateMessagesDispatchMapping = () => {
  createForm.exact_model_mappings.push({ claude_model: "", target_model: "" });
};

const removeCreateMessagesDispatchMapping = (
  row: MessagesDispatchMappingRow,
) => {
  const index = createForm.exact_model_mappings.indexOf(row);
  if (index !== -1) {
    createForm.exact_model_mappings.splice(index, 1);
  }
};

const addEditMessagesDispatchMapping = () => {
  editForm.exact_model_mappings.push({ claude_model: "", target_model: "" });
};

const removeEditMessagesDispatchMapping = (row: MessagesDispatchMappingRow) => {
  const index = editForm.exact_model_mappings.indexOf(row);
  if (index !== -1) {
    editForm.exact_model_mappings.splice(index, 1);
  }
};

const handleRateMultipliers = (group: AdminGroup) => {
  rateMultipliersGroup.value = group;
  showRateMultipliersModal.value = true;
};

const handleLevelRates = (group: AdminGroup) => {
  levelRatesGroup.value = group;
  showLevelRatesModal.value = true;
};

const handleRPMOverrides = (group: AdminGroup) => {
  rpmOverridesGroup.value = group;
  showRPMOverridesModal.value = true;
};

const handleDuplicate = async (group: AdminGroup) => {
  if (duplicatingGroupIds.has(group.id)) return;

  duplicatingGroupIds.add(group.id);
  try {
    const duplicate = await adminAPI.groups.duplicate(group.id);
    appStore.showSuccess(
      t("admin.groups.duplicateSuccess", { name: duplicate.name }),
    );
    await loadGroups();
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(error, t("admin.groups.duplicateFailed")),
    );
  } finally {
    duplicatingGroupIds.delete(group.id);
  }
};

const compositeRouteMatchLabel = (matchType: CompositeRouteMatchType) =>
  compositeRouteMatchOptions.value.find((option) => option.value === matchType)
    ?.label || matchType;

const formatCompositeEndpoint = (endpoint: CompositeRouteEndpoint) =>
  compositeRouteEndpointOptions.value.find((option) => option.value === endpoint)
    ?.label || endpoint;

const formatCompositePlatform = (platform: string) => {
  if (!platform) return "—";
  return t(`admin.groups.platforms.${platform}`);
};

const compositeRouteSourceLabel = (source: string) => {
  if (source === "route") return t("admin.groups.compositeRoutes.sources.route");
  if (source === "detector") {
    return t("admin.groups.compositeRoutes.sources.detector");
  }
  return source || "—";
};

const resetCompositeRouteForm = () => {
  compositeRouteEditingId.value = null;
  compositeRouteForm.public_model = "";
  compositeRouteForm.match_type = "exact";
  compositeRouteForm.target_platform = "openai";
  compositeRouteForm.upstream_model = "";
  compositeRouteForm.endpoint = "any";
  compositeRouteForm.priority = 100;
  compositeRouteForm.enabled = true;
  compositeRouteForm.notes = "";
};

const toCompositeRouteInput = (): CompositeModelRouteInput => ({
  public_model: compositeRouteForm.public_model.trim(),
  match_type: compositeRouteForm.match_type,
  target_platform: compositeRouteForm.target_platform,
  upstream_model: compositeRouteForm.upstream_model.trim(),
  endpoint: compositeRouteForm.endpoint,
  priority: Number(compositeRouteForm.priority) || 100,
  enabled: compositeRouteForm.enabled,
  notes: compositeRouteForm.notes.trim(),
});

const loadCompositeRoutes = async () => {
  if (!compositeRoutesGroup.value) return;
  compositeRoutesLoading.value = true;
  try {
    const routes = await adminAPI.groups.listCompositeRoutes(
      compositeRoutesGroup.value.id,
    );
    compositeRoutes.value = routes.sort((a, b) => {
      if (a.priority !== b.priority) return a.priority - b.priority;
      return a.id.localeCompare(b.id);
    });
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail ||
        error.response?.data?.message ||
        t("admin.groups.compositeRoutes.failedToLoad"),
    );
    console.error("Error loading composite routes:", error);
  } finally {
    compositeRoutesLoading.value = false;
  }
};

const handleCompositeRoutes = async (group: AdminGroup) => {
  compositeRoutesGroup.value = group;
  compositePreviewModel.value = "";
  compositePreviewEndpoint.value = "any";
  compositePreviewDecision.value = null;
  resetCompositeRouteForm();
  showCompositeRoutesModal.value = true;
  await loadCompositeRoutes();
};

const closeCompositeRoutesModal = () => {
  showCompositeRoutesModal.value = false;
  compositeRoutesGroup.value = null;
  compositeRoutes.value = [];
  compositePreviewDecision.value = null;
  resetCompositeRouteForm();
};

const editCompositeRoute = (route: CompositeModelRoute) => {
  compositeRouteEditingId.value = route.id;
  compositeRouteForm.public_model = route.public_model;
  compositeRouteForm.match_type = route.match_type;
  compositeRouteForm.target_platform = route.target_platform;
  compositeRouteForm.upstream_model = route.upstream_model;
  compositeRouteForm.endpoint = route.endpoint;
  compositeRouteForm.priority = route.priority || 100;
  compositeRouteForm.enabled = route.enabled;
  compositeRouteForm.notes = route.notes || "";
};

const saveCompositeRoute = async () => {
  if (!compositeRoutesGroup.value) return;
  if (!compositeRouteForm.public_model.trim()) {
    appStore.showError(t("admin.groups.compositeRoutes.publicModelRequired"));
    return;
  }
  compositeRouteSaving.value = true;
  try {
    const payload = toCompositeRouteInput();
    if (compositeRouteEditingId.value) {
      await adminAPI.groups.updateCompositeRoute(
        compositeRoutesGroup.value.id,
        compositeRouteEditingId.value,
        payload,
      );
      appStore.showSuccess(t("admin.groups.compositeRoutes.routeUpdated"));
    } else {
      await adminAPI.groups.createCompositeRoute(
        compositeRoutesGroup.value.id,
        payload,
      );
      appStore.showSuccess(t("admin.groups.compositeRoutes.routeCreated"));
    }
    resetCompositeRouteForm();
    await loadCompositeRoutes();
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail ||
        error.response?.data?.message ||
        t("admin.groups.compositeRoutes.failedToSave"),
    );
    console.error("Error saving composite route:", error);
  } finally {
    compositeRouteSaving.value = false;
  }
};

const deleteCompositeRoute = async (route: CompositeModelRoute) => {
  if (!compositeRoutesGroup.value) return;
  if (!window.confirm(t("admin.groups.compositeRoutes.deleteConfirm"))) return;
  try {
    await adminAPI.groups.deleteCompositeRoute(
      compositeRoutesGroup.value.id,
      route.id,
    );
    if (compositeRouteEditingId.value === route.id) {
      resetCompositeRouteForm();
    }
    appStore.showSuccess(t("admin.groups.compositeRoutes.routeDeleted"));
    await loadCompositeRoutes();
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail ||
        error.response?.data?.message ||
        t("admin.groups.compositeRoutes.failedToDelete"),
    );
    console.error("Error deleting composite route:", error);
  }
};

const previewCompositeRoute = async () => {
  if (!compositeRoutesGroup.value || !compositePreviewModel.value.trim()) {
    return;
  }
  compositePreviewLoading.value = true;
  try {
    compositePreviewDecision.value = await adminAPI.groups.previewCompositeRoute(
      compositeRoutesGroup.value.id,
      {
        model: compositePreviewModel.value.trim(),
        endpoint: compositePreviewEndpoint.value,
      },
    );
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail ||
        error.response?.data?.message ||
        t("admin.groups.compositeRoutes.failedToPreview"),
    );
    console.error("Error previewing composite route:", error);
  } finally {
    compositePreviewLoading.value = false;
  }
};

const handleDelete = (group: AdminGroup) => {
  deletingGroup.value = group;
  showDeleteDialog.value = true;
};

const confirmDelete = async () => {
  if (!deletingGroup.value) return;

  try {
    await adminAPI.groups.delete(deletingGroup.value.id);
    appStore.showSuccess(t("admin.groups.groupDeleted"));
    showDeleteDialog.value = false;
    deletingGroup.value = null;
    loadGroups();
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail || t("admin.groups.failedToDelete"),
    );
    console.error("Error deleting group:", error);
  }
};

// 监听 subscription_type 变化，订阅模式时 is_exclusive 默认为 true；标准模式清空高峰配置
watch(
  () => createForm.subscription_type,
  (newVal) => {
    if (newVal === "subscription") {
      createForm.is_exclusive = true;
      createForm.fallback_group_id_on_invalid_request = null;
    } else {
      createForm.peak_rate_enabled = false;
      createForm.peak_start = "";
      createForm.peak_end = "";
      createForm.peak_rate_multiplier = 1.0;
    }
  },
);

// 编辑表单：切回标准模式时清空高峰配置，避免残留随更新请求提交被后端拒绝
watch(
  () => editForm.subscription_type,
  (newVal) => {
    if (newVal !== "subscription") {
      editForm.peak_rate_enabled = false;
      editForm.peak_start = "";
      editForm.peak_end = "";
      editForm.peak_rate_multiplier = 1.0;
    }
  },
);

watch(
  () => createForm.platform,
  (newVal) => {
    if (newVal !== "anthropic") {
      createForm.fallback_group_id_on_invalid_request = null;
    }
    if (newVal !== "openai") {
      resetMessagesDispatchFormState(createForm);
      createForm.allow_live = false;
    }
    if (!isProfitControlPlatform(newVal)) {
      createForm.profit_control_enabled = false;
      createForm.profit_min_margin_percent = 0;
      createForm.profit_safety_buffer_percent = 0;
    }
    createForm.max_reasoning_effort = normalizeReasoningEffortForPlatform(
      newVal,
      createForm.max_reasoning_effort,
    );
    createForm.reasoning_effort_mappings = reasoningEffortMappingsToRows(
      reasoningEffortMappingsToAPI(createForm.reasoning_effort_mappings),
      newVal,
    );
    createReasoningEffortPolicyRef.value?.resetValidation();
    if (!["openai", "anthropic"].includes(newVal)) {
      createForm.require_oauth_only = false;
      createForm.require_privacy_set = false;
    }
    resetModelsListState(createModelsListState);
    loadModelsListCandidates("create", null, newVal);
  },
);

watch(
  () => editForm.platform,
  (newVal) => {
    if (newVal !== "anthropic") {
      editForm.fallback_group_id_on_invalid_request = null;
    }
    if (newVal !== "openai") {
      resetMessagesDispatchFormState(editForm);
      editForm.allow_live = false;
    }
    if (!isProfitControlPlatform(newVal)) {
      editForm.profit_control_enabled = false;
      editForm.profit_min_margin_percent = 0;
      editForm.profit_safety_buffer_percent = 0;
    }
    editForm.max_reasoning_effort = normalizeReasoningEffortForPlatform(
      newVal,
      editForm.max_reasoning_effort,
    );
    editForm.reasoning_effort_mappings = reasoningEffortMappingsToRows(
      reasoningEffortMappingsToAPI(editForm.reasoning_effort_mappings),
      newVal,
    );
    editReasoningEffortPolicyRef.value?.resetValidation();
    if (!["openai", "anthropic"].includes(newVal)) {
      editForm.require_oauth_only = false;
      editForm.require_privacy_set = false;
    }
    if (editingGroup.value) {
      resetModelsListState(editModelsListState, editForm.platform === editingGroup.value.platform ? editingGroup.value.models_list_config : undefined);
      loadModelsListCandidates("edit", editingGroup.value.id, newVal);
    }
  },
);

watch(
  () => editForm.platform,
  (newVal) => {
    if (newVal !== 'anthropic') {
      editForm.fallback_group_id_on_invalid_request = null
    }
    if (newVal !== 'openai') {
      editForm.allow_messages_dispatch = false
      editForm.allow_live = false
      editForm.default_mapped_model = ''
    }
  }
)

// 点击外部关闭账号搜索下拉框
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement;
  // 检查是否点击在下拉框或输入框内
  if (!target.closest(".account-search-container")) {
    Object.keys(showAccountDropdown.value).forEach((key) => {
      showAccountDropdown.value[key] = false;
    });
  }
  if (columnDropdownRef.value && !columnDropdownRef.value.contains(target)) {
    showColumnDropdown.value = false;
  }
};

// 打开排序弹窗
const openSortModal = async () => {
  try {
    // 获取所有分组（不分页）
    const allGroups = await adminAPI.groups.getAll();
    // 按 sort_order 排序
    sortableGroups.value = [...allGroups].sort(
      (a, b) => a.sort_order - b.sort_order,
    );
    showSortModal.value = true;
  } catch (error) {
    appStore.showError(t("admin.groups.failedToLoad"));
    console.error("Error loading groups for sorting:", error);
  }
};

// 关闭排序弹窗
const closeSortModal = () => {
  showSortModal.value = false;
  sortableGroups.value = [];
};

// 保存排序
const saveSortOrder = async () => {
  sortSubmitting.value = true;
  try {
    const updates = sortableGroups.value.map((g, index) => ({
      id: g.id,
      sort_order: index * 10,
    }));
    await adminAPI.groups.updateSortOrder(updates);
    appStore.showSuccess(t("admin.groups.sortOrderUpdated"));
    closeSortModal();
    loadGroups();
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail || t("admin.groups.failedToUpdateSortOrder"),
    );
    console.error("Error updating sort order:", error);
  } finally {
    sortSubmitting.value = false;
  }
};

onMounted(() => {
  loadGroups();
  void loadLiveCapability();
  loadModelsListCandidates("create", null, createForm.platform);
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
  accountSearchRunner.clearAll();
  clearAllAccountSearchState();
});
</script>

<style scoped lang="scss">
.groups-toolbar {
  min-width: 0;
}

.groups-toolbar__filters {
  min-width: 0;
}

.groups-toolbar__actions {
  max-width: 100%;
}

.groups-toolbar__platforms {
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.views-admin-groups-view__panel-14 {
  overscroll-behavior-x: contain;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}
</style>
