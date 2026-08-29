<template>
  <BaseDialog :show="show" :title="operation === 'add' ? t('admin.users.deposit') : t('admin.users.withdraw')" width="narrow" @close="$emit('close')">
    <form v-if="user" id="balance-form" @submit.prevent="handleBalanceSubmit" class="components-admin-user-user-balance-modal__form">
      <div class="components-admin-user-user-balance-modal__panel">
        <div class="components-admin-user-user-balance-modal__panel-2"><span class="components-admin-user-user-balance-modal__text">{{ user.email.charAt(0).toUpperCase() }}</span></div>
        <div class="components-admin-user-user-balance-modal__panel-3"><p class="components-admin-user-user-balance-modal__description">{{ user.email }}</p><p class="components-admin-user-user-balance-modal__description-2">{{ t('admin.users.currentBalance') }}: {{ formatPoints(user.balance) }}</p></div>
      </div>
      <div>
        <label class="input-label">{{ operation === 'add' ? t('admin.users.depositAmount') : t('admin.users.withdrawAmount') }}</label>
        <div class="components-admin-user-user-balance-modal__panel-4">
          <div class="components-admin-user-user-balance-modal__panel-5"><div class="components-admin-user-user-balance-modal__panel-6">{{ t('common.points') }}</div><input v-model.number="form.amount" type="number" step="any" min="0" required class="components-admin-user-user-balance-modal__field input" /></div>
          <button v-if="operation === 'subtract'" type="button" @click="fillAllBalance" class="components-admin-user-user-balance-modal__action btn btn-secondary">{{ t('admin.users.withdrawAll') }}</button>
        </div>
      </div>
	  <div v-if="operation === 'add'">
		<label class="input-label">{{ t('admin.users.balanceType') }}</label>
		<Select v-model="form.balanceType" :options="[
		  { value: 'recharge', label: t('admin.users.rechargeBalance') },
		  { value: 'bonus', label: t('admin.users.bonusBalance') }
		]" />
	  </div>
	  <div v-if="operation === 'add' && form.balanceType === 'bonus'">
		<label class="input-label">{{ t('admin.users.bonusValidityDays') }}</label>
		<input v-model.number="form.bonusValidityDays" type="number" min="1" max="3650" class="input" required />
	  </div>
      <div><label class="input-label">{{ t('admin.users.notes') }}</label><textarea v-model="form.notes" rows="3" class="input"></textarea></div>
      <div v-if="form.amount > 0" class="components-admin-user-user-balance-modal__panel-7"><div class="components-admin-user-user-balance-modal__panel-8"><span class="components-admin-user-user-balance-modal__text-2">{{ t('admin.users.newBalance') }}:</span><span class="components-admin-user-user-balance-modal__text-3">{{ formatPoints(calculateNewBalance()) }}</span></div></div>
    </form>
    <template #footer>
      <div class="components-admin-user-user-balance-modal__panel-9">
        <button @click="$emit('close')" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button type="submit" form="balance-form" :disabled="submitting || !form.amount" class="btn" :class="operation === 'add' ? 'components-admin-user-user-balance-modal__action-2' : 'btn-danger'">{{ submitting ? t('common.saving') : t('common.confirm') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import { formatPoints } from '@/utils/format'

const props = defineProps<{ show: boolean, user: AdminUser | null, operation: 'add' | 'subtract' }>()
const emit = defineEmits(['close', 'success']); const { t } = useI18n(); const appStore = useAppStore()

const submitting = ref(false); const form = reactive({ amount: 0, notes: '', balanceType: 'recharge' as 'recharge' | 'bonus', bonusValidityDays: 90 })
watch(() => props.show, (v) => { if(v) { form.amount = 0; form.notes = ''; form.balanceType = 'recharge'; form.bonusValidityDays = 90 } })

// 填入全部余额
const fillAllBalance = () => {
  if (props.user) {
    form.amount = props.user.balance
  }
}

const calculateNewBalance = () => {
  if (!props.user) return 0
  const result = props.operation === 'add' ? props.user.balance + form.amount : props.user.balance - form.amount
  // 避免浮点数精度问题导致的 -0.00 显示
  return Math.abs(result) < 1e-10 ? 0 : result
}
const handleBalanceSubmit = async () => {
  if (!props.user) return
  if (!form.amount || form.amount <= 0) {
    appStore.showError(t('admin.users.amountRequired'))
    return
  }
  // 退款时验证金额不超过实际余额
  if (props.operation === 'subtract' && form.amount > props.user.balance) {
    appStore.showError(t('admin.users.insufficientBalance'))
    return
  }
  submitting.value = true
  try {
	await adminAPI.users.updateBalance(props.user.id, form.amount, props.operation, form.notes, form.balanceType, form.bonusValidityDays)
    appStore.showSuccess(t('common.success')); emit('success'); emit('close')
  } catch (e: any) {
    console.error('Failed to update balance:', e)
    appStore.showError(e.response?.data?.detail || t('common.error'))
  } finally { submitting.value = false }
}
</script>
