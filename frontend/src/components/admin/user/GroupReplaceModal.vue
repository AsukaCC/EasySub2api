<template>
  <BaseDialog :show="show" :title="t('admin.users.replaceGroupTitle')" width="narrow" @close="$emit('close')">
    <div v-if="oldGroup" class="components-admin-user-group-replace-modal__panel">
      <!-- 提示信息 -->
      <p class="components-admin-user-group-replace-modal__description">
        {{ t('admin.users.replaceGroupHint', { old: oldGroup.name }) }}
      </p>

      <!-- 当前分组 -->
      <div class="components-admin-user-group-replace-modal__panel-2">
        <div class="components-admin-user-group-replace-modal__panel-3">
          <Icon name="shield" size="sm" class="components-admin-user-group-replace-modal__icon" />
          <span class="components-admin-user-group-replace-modal__text">{{ oldGroup.name }}</span>
          <Icon name="arrowRight" size="sm" class="components-admin-user-group-replace-modal__icon-2" />
          <span v-if="selectedGroupId" class="components-admin-user-group-replace-modal__text-2">
            {{ availableGroups.find(g => g.id === selectedGroupId)?.name }}
          </span>
          <span v-else class="components-admin-user-group-replace-modal__text-3">?</span>
        </div>
      </div>

      <!-- 可选分组列表 -->
      <div v-if="availableGroups.length > 0" class="components-admin-user-group-replace-modal__panel-4">
        <label
          v-for="group in availableGroups"
          :key="group.id"
          class="components-admin-user-group-replace-modal__label"
          :class="selectedGroupId === group.id
            ? 'components-admin-user-group-replace-modal__label-2'
            : 'components-admin-user-group-replace-modal__label-3'"
        >
          <input
            type="radio"
            :value="group.id"
            v-model="selectedGroupId"
            class="components-admin-user-group-replace-modal__field"
          />
          <div
            class="components-admin-user-group-replace-modal__panel-5"
            :class="selectedGroupId === group.id
              ? 'components-admin-user-group-replace-modal__panel-10'
              : 'components-admin-user-group-replace-modal__panel-11'"
          >
            <div v-if="selectedGroupId === group.id" class="components-admin-user-group-replace-modal__panel-6"></div>
          </div>
          <div class="components-admin-user-group-replace-modal__panel-7">
            <span class="components-admin-user-group-replace-modal__text">{{ group.name }}</span>
            <span class="components-admin-user-group-replace-modal__text-4">{{ group.platform }}</span>
          </div>
        </label>
      </div>

      <!-- 无可选分组 -->
      <div v-else class="components-admin-user-group-replace-modal__panel-8">
        {{ t('admin.users.noOtherGroups') }}
      </div>
    </div>

    <template #footer>
      <div class="components-admin-user-group-replace-modal__panel-9">
        <button @click="$emit('close')" class="components-admin-user-group-replace-modal__action btn btn-secondary">{{ t('common.cancel') }}</button>
        <button :aria-busy="submitting"
          @click="handleReplace"
          :disabled="!selectedGroupId || submitting"
          class="components-admin-user-group-replace-modal__action-2 btn btn-primary"
        >
<LoadingButtonContent :loading="submitting" :loading-text="t('common.saving')">
{{ t('admin.users.replaceGroupConfirm') }}
</LoadingButtonContent>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import LoadingButtonContent from '@/components/common/LoadingButtonContent.vue'

import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { AdminUser, AdminGroup } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

interface Props {
  show: boolean
  user: AdminUser | null
  oldGroup: { id: string; name: string } | null
  allGroups: AdminGroup[]
}

const props = defineProps<Props>()
const emit = defineEmits(['close', 'success'])
const { t } = useI18n()
const appStore = useAppStore()

const selectedGroupId = ref<string | null>(null)
const submitting = ref(false)

// 可选的专属标准分组（排除当前 oldGroup）
const availableGroups = computed(() => {
  if (!props.oldGroup) return []
  return props.allGroups.filter(
    g => g.status === 'active' && g.is_exclusive && g.subscription_type === 'standard' && g.id !== props.oldGroup!.id
  )
})

watch(() => props.show, (v) => {
  if (v) {
    selectedGroupId.value = null
  }
})

const handleReplace = async () => {
  if (!props.user || !props.oldGroup || !selectedGroupId.value) return
  submitting.value = true

  try {
    const result = await adminAPI.users.replaceGroup(props.user.id, props.oldGroup.id, selectedGroupId.value)
    appStore.showSuccess(t('admin.users.replaceGroupSuccess', { count: result.migrated_keys }))
    emit('success')
    emit('close')
  } catch (error) {
    console.error('Failed to replace group:', error)
  } finally {
    submitting.value = false
  }
}
</script>
