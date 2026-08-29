/**
 * Shared utility functions for payment order display.
 * Used by AdminOrderDetail, AdminOrderTable, AdminRefundDialog, AdminOrdersView, etc.
 */

export interface RechargeOrderAmounts {
  amount?: number | null
  pay_amount?: number | null
  gateway_base_amount?: number | null
  principal_amount?: number | null
  fee_amount?: number | null
  base_points?: number | null
  bonus_points?: number | null
  credited_points?: number | null
  affiliate_rebate_points?: number | null
}

export interface DirectRefundOrder {
  status: string
  order_type?: string | null
  completed_at?: string | null
  refund_deadline?: string | null
}

function validAmount(value: number | null | undefined): number | undefined {
  return typeof value === 'number' && Number.isFinite(value) ? value : undefined
}

export function rechargePrincipalAmount(order: RechargeOrderAmounts): number {
  return validAmount(order.principal_amount) ?? validAmount(order.gateway_base_amount) ?? 0
}

export function rechargeFeeAmount(order: RechargeOrderAmounts): number {
  const explicit = validAmount(order.fee_amount)
  if (explicit !== undefined) return explicit
  return Math.max((validAmount(order.pay_amount) ?? 0) - rechargePrincipalAmount(order), 0)
}

export function rechargeBasePoints(order: RechargeOrderAmounts): number {
  return validAmount(order.base_points) ?? rechargePrincipalAmount(order)
}

export function rechargeCreditedPoints(order: RechargeOrderAmounts): number {
  return validAmount(order.credited_points) ?? validAmount(order.amount) ?? 0
}

export function rechargeBonusPoints(order: RechargeOrderAmounts): number {
  const explicit = validAmount(order.bonus_points)
  if (explicit !== undefined) return explicit
  return Math.max(rechargeCreditedPoints(order) - rechargeBasePoints(order), 0)
}

export function rechargeAffiliatePoints(order: RechargeOrderAmounts): number {
  return validAmount(order.affiliate_rebate_points) ?? 0
}

const STATUS_BADGE_MAP: Record<string, string> = {
  PENDING: 'badge-warning',
  PAID: 'badge-info',
  RECHARGING: 'badge-info',
  COMPLETED: 'badge-success',
  EXPIRED: 'badge-secondary',
  CANCELLED: 'badge-secondary',
  FAILED: 'badge-danger',
  REFUND_REQUESTED: 'badge-warning',
  REFUNDING: 'badge-warning',
  REFUND_PENDING: 'badge-warning',
  PARTIALLY_REFUNDED: 'badge-warning',
  REFUNDED: 'badge-info',
  REFUND_FAILED: 'badge-danger',
}

const REFUNDABLE_STATUSES = ['COMPLETED', 'PARTIALLY_REFUNDED', 'REFUND_REQUESTED', 'REFUND_FAILED']

export function statusBadgeClass(status: string): string {
  return STATUS_BADGE_MAP[status] || 'badge-secondary'
}

export function canRefund(status: string): boolean {
  return REFUNDABLE_STATUSES.includes(status)
}

export function canDirectRefund(order: DirectRefundOrder, now: Date | number = Date.now()): boolean {
  if (order.order_type !== 'balance' || !canRefund(order.status)) return false
  const nowMs = typeof now === 'number' ? now : now.getTime()
  const completedAtMs = order.completed_at ? Date.parse(order.completed_at) : Number.NaN
  const explicitDeadlineMs = order.refund_deadline ? Date.parse(order.refund_deadline) : Number.NaN
  const deadlineMs = Number.isFinite(explicitDeadlineMs)
    ? explicitDeadlineMs
    : Number.isFinite(completedAtMs)
      ? completedAtMs + 168 * 60 * 60 * 1000
      : Number.NaN
  return Number.isFinite(completedAtMs)
    && Number.isFinite(deadlineMs)
    && nowMs >= completedAtMs
    && nowMs < deadlineMs
}

export function formatOrderDateTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}
