<template>
  <BaseDialog :show="show" :title="t('admin.users.groupConfig')" width="wide" @close="$emit('close')">
    <div v-if="user" class="components-admin-user-user-allowed-groups-modal__panel">
      <!-- 用户信息头部 -->
      <div class="components-admin-user-user-allowed-groups-modal__panel-2">
        <div class="components-admin-user-user-allowed-groups-modal__panel-3">
          <span class="components-admin-user-user-allowed-groups-modal__text">{{ user.email.charAt(0).toUpperCase() }}</span>
        </div>
        <div class="components-admin-user-user-allowed-groups-modal__panel-4">
          <p class="components-admin-user-user-allowed-groups-modal__description">{{ user.email }}</p>
          <p class="components-admin-user-user-allowed-groups-modal__description-2">{{ t('admin.users.groupConfigHint', { email: user.email }) }}</p>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="components-admin-user-user-allowed-groups-modal__panel-5">
        <svg class="components-admin-user-user-allowed-groups-modal__icon" fill="none" viewBox="0 0 24 24">
          <circle class="components-admin-user-user-allowed-groups-modal__circle" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="components-admin-user-user-allowed-groups-modal__path" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>

      <div v-else class="components-admin-user-user-allowed-groups-modal__panel">
        <!-- 专属分组区域 -->
        <div v-if="exclusiveGroups.length > 0">
          <div class="components-admin-user-user-allowed-groups-modal__panel-6">
            <div class="components-admin-user-user-allowed-groups-modal__panel-7"></div>
            <h4 class="components-admin-user-user-allowed-groups-modal__heading">{{ t('admin.users.exclusiveGroups') }}</h4>
            <span class="components-admin-user-user-allowed-groups-modal__text-2">({{ exclusiveGroupConfigs.filter(c => c.isSelected).length }}/{{ exclusiveGroupConfigs.length }})</span>
          </div>
          <div class="components-admin-user-user-allowed-groups-modal__panel-8">
            <div
              v-for="config in exclusiveGroupConfigs"
              :key="config.groupId"
              class="components-admin-user-user-allowed-groups-modal__panel-9"
              :class="config.isSelected
                ? 'components-admin-user-user-allowed-groups-modal__panel-23'
                : 'components-admin-user-user-allowed-groups-modal__panel-24'"
            >
              <div class="components-admin-user-user-allowed-groups-modal__panel-10">
                <!-- 复选框 -->
                <div class="components-admin-user-user-allowed-groups-modal__panel-11">
                  <label class="components-admin-user-user-allowed-groups-modal__label">
                    <input
                      type="checkbox"
                      :checked="config.isSelected"
                      @change="toggleExclusiveGroup(config.groupId)"
                      class="components-admin-user-user-allowed-groups-modal__field"
                    />
                    <div class="components-admin-user-user-allowed-groups-modal__panel-12">
                      <svg v-if="config.isSelected" class="components-admin-user-user-allowed-groups-modal__icon-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                      </svg>
                    </div>
                  </label>
                </div>

                <!-- 分组信息 -->
                <div class="components-admin-user-user-allowed-groups-modal__panel-13">
                  <div class="components-admin-user-user-allowed-groups-modal__panel-14">
                    <span class="components-admin-user-user-allowed-groups-modal__text-3">{{ config.groupName }}</span>
                    <span class="components-admin-user-user-allowed-groups-modal__text-4">
                      {{ t('admin.groups.exclusive') }}
                    </span>
                  </div>
                  <div class="components-admin-user-user-allowed-groups-modal__panel-15">
                    <span class="components-admin-user-user-allowed-groups-modal__text-5">
                      <PlatformIcon :platform="config.platform" size="xs" />
                      <span>{{ config.platform }}</span>
                    </span>
                    <span class="components-admin-user-user-allowed-groups-modal__text-6">•</span>
                    <span class="components-admin-user-user-allowed-groups-modal__text-7">
                      {{ t('admin.users.defaultRate') }}: <span class="components-admin-user-user-allowed-groups-modal__text-8">{{ config.defaultRate }}x</span>
                    </span>
                  </div>
                </div>

                <!-- 专属倍率输入 -->
                <div class="components-admin-user-user-allowed-groups-modal__panel-16">
                  <label class="components-admin-user-user-allowed-groups-modal__label-2">{{ t('admin.users.customRate') }}</label>
                  <input
                    type="number"
                    step="0.001"
                    min="0.001"
                    :value="config.customRate ?? ''"
                    @input="updateCustomRate(config.groupId, ($event.target as HTMLInputElement).value)"
                    :placeholder="String(config.defaultRate)"
                    class="components-admin-user-user-allowed-groups-modal__field-2 hide-spinner"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 公开分组区域 -->
        <div v-if="publicGroups.length > 0">
          <div class="components-admin-user-user-allowed-groups-modal__panel-6">
            <div class="components-admin-user-user-allowed-groups-modal__panel-17"></div>
            <h4 class="components-admin-user-user-allowed-groups-modal__heading">{{ t('admin.users.publicGroups') }}</h4>
            <span class="components-admin-user-user-allowed-groups-modal__text-2">({{ publicGroupConfigs.length }})</span>
          </div>
          <div class="components-admin-user-user-allowed-groups-modal__panel-8">
            <div
              v-for="config in publicGroupConfigs"
              :key="config.groupId"
              class="components-admin-user-user-allowed-groups-modal__panel-18"
            >
              <div class="components-admin-user-user-allowed-groups-modal__panel-10">
                <!-- 复选框（禁用状态） -->
                <div class="components-admin-user-user-allowed-groups-modal__panel-11">
                  <div class="components-admin-user-user-allowed-groups-modal__panel-19">
                    <svg class="components-admin-user-user-allowed-groups-modal__icon-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                </div>

                <!-- 分组信息 -->
                <div class="components-admin-user-user-allowed-groups-modal__panel-13">
                  <div class="components-admin-user-user-allowed-groups-modal__panel-14">
                    <span class="components-admin-user-user-allowed-groups-modal__text-3">{{ config.groupName }}</span>
                  </div>
                  <div class="components-admin-user-user-allowed-groups-modal__panel-15">
                    <span class="components-admin-user-user-allowed-groups-modal__text-5">
                      <PlatformIcon :platform="config.platform" size="xs" />
                      <span>{{ config.platform }}</span>
                    </span>
                    <span class="components-admin-user-user-allowed-groups-modal__text-6">•</span>
                    <span class="components-admin-user-user-allowed-groups-modal__text-7">
                      {{ t('admin.users.defaultRate') }}: <span class="components-admin-user-user-allowed-groups-modal__text-8">{{ config.defaultRate }}x</span>
                    </span>
                  </div>
                </div>

                <!-- 专属倍率输入 -->
                <div class="components-admin-user-user-allowed-groups-modal__panel-16">
                  <label class="components-admin-user-user-allowed-groups-modal__label-2">{{ t('admin.users.customRate') }}</label>
                  <input
                    type="number"
                    step="0.001"
                    min="0.001"
                    :value="config.customRate ?? ''"
                    @input="updateCustomRate(config.groupId, ($event.target as HTMLInputElement).value)"
                    :placeholder="String(config.defaultRate)"
                    class="components-admin-user-user-allowed-groups-modal__field-2 hide-spinner"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 无分组提示 -->
        <div v-if="groups.length === 0" class="components-admin-user-user-allowed-groups-modal__panel-20">
          <div class="components-admin-user-user-allowed-groups-modal__panel-21">
            <svg class="components-admin-user-user-allowed-groups-modal__icon-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          </div>
          <p class="components-admin-user-user-allowed-groups-modal__text-7">{{ t('common.noGroupsAvailable') }}</p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="components-admin-user-user-allowed-groups-modal__panel-22">
        <button @click="$emit('close')" class="components-admin-user-user-allowed-groups-modal__action btn btn-secondary">{{ t('common.cancel') }}</button>
        <button @click="handleSave" :disabled="submitting" class="components-admin-user-user-allowed-groups-modal__action-2 btn btn-primary">
          <svg v-if="submitting" class="components-admin-user-user-allowed-groups-modal__icon-4" fill="none" viewBox="0 0 24 24">
            <circle class="components-admin-user-user-allowed-groups-modal__circle" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="components-admin-user-user-allowed-groups-modal__path" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ submitting ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { AdminUser, Group, GroupPlatform } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'

interface GroupRateConfig {
  groupId: string
  groupName: string
  platform: GroupPlatform
  isExclusive: boolean
  defaultRate: number
  customRate: number | null
  isSelected: boolean
}

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits(['close', 'success'])
const { t } = useI18n()
const appStore = useAppStore()

const groups = ref<Group[]>([])
const groupConfigs = ref<GroupRateConfig[]>([])
const originalGroupRates = ref<Record<string, number>>({}) // 记录原始专属倍率，用于检测删除
const loading = ref(false)
const submitting = ref(false)

// 分离专属分组和公开分组
const exclusiveGroups = computed(() => groups.value.filter((g) => g.is_exclusive))
const publicGroups = computed(() => groups.value.filter((g) => !g.is_exclusive))

const exclusiveGroupConfigs = computed(() => groupConfigs.value.filter((c) => c.isExclusive))
const publicGroupConfigs = computed(() => groupConfigs.value.filter((c) => !c.isExclusive))

watch(
  () => props.show,
  (v) => {
    if (v && props.user) {
      load()
    }
  }
)

const load = async () => {
  loading.value = true
  try {
    const res = await adminAPI.groups.list(1, 1000)
    // 只显示标准类型且活跃的分组
    groups.value = res.items.filter((g) => g.subscription_type === 'standard' && g.status === 'active')

    // 初始化配置
    const userAllowedGroups = props.user?.allowed_groups || []
    const userGroupRates = props.user?.group_rates || {}

    // 保存原始专属倍率，用于检测删除操作
    originalGroupRates.value = { ...userGroupRates }

    groupConfigs.value = groups.value.map((g) => ({
      groupId: g.id,
      groupName: g.name,
      platform: g.platform,
      isExclusive: g.is_exclusive,
      defaultRate: g.rate_multiplier,
      customRate: userGroupRates[g.id] ?? null,
      // 专属分组：检查是否在 allowed_groups 中
      // 公开分组：始终选中
      isSelected: g.is_exclusive ? userAllowedGroups.includes(g.id) : true,
    }))
  } catch (error) {
    console.error('Failed to load groups:', error)
  } finally {
    loading.value = false
  }
}

const toggleExclusiveGroup = (groupId: string) => {
  const config = groupConfigs.value.find((c) => c.groupId === groupId)
  if (config && config.isExclusive) {
    config.isSelected = !config.isSelected
  }
}

const updateCustomRate = (groupId: string, value: string) => {
  const config = groupConfigs.value.find((c) => c.groupId === groupId)
  if (config) {
    if (value === '' || value === null || value === undefined) {
      config.customRate = null
    } else {
      const numValue = parseFloat(value)
      config.customRate = isNaN(numValue) ? null : numValue
    }
  }
}

const handleSave = async () => {
  if (!props.user) return
  submitting.value = true

  try {
    // 构建 allowed_groups（仅包含专属分组中被勾选的）
    const allowedGroups = groupConfigs.value.filter((c) => c.isExclusive && c.isSelected).map((c) => c.groupId)

    // 构建 group_rates
    // - 有新专属倍率: 设置为该值
    // - 原本有专属倍率但现在被清空: 设置为 null（表示删除）
    const groupRates: Record<string, number | null> = {}
    for (const c of groupConfigs.value) {
      const hadOriginalRate = originalGroupRates.value[c.groupId] !== undefined

      if (c.customRate !== null) {
        // 有专属倍率
        groupRates[c.groupId] = c.customRate
      } else if (hadOriginalRate) {
        // 原本有专属倍率，现在被清空，需要显式删除
        groupRates[c.groupId] = null
      }
    }

    await adminAPI.users.update(props.user.id, {
      allowed_groups: allowedGroups,
      group_rates: Object.keys(groupRates).length > 0 ? groupRates : undefined,
    })

    appStore.showSuccess(t('admin.users.groupConfigUpdated'))
    emit('success')
    emit('close')
  } catch (error) {
    console.error('Failed to update user group config:', error)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
/* 隐藏数字输入框的箭头按钮 */
.hide-spinner::-webkit-outer-spin-button,
.hide-spinner::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.hide-spinner {
  -moz-appearance: textfield;
}
</style>
