import type { SubscriptionPlan } from '@/types/payment'

function finiteOr(value: number | null | undefined, fallback: number): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : fallback
}

export function subscriptionPlanPricePoints(plan: SubscriptionPlan): number {
  return finiteOr(plan.price_points, finiteOr(plan.price, 0))
}

export function subscriptionPlanOriginalPricePoints(plan: SubscriptionPlan): number | null {
  const legacy = typeof plan.original_price === 'number' ? plan.original_price : null
  const value = plan.original_price_points ?? legacy
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}

export function subscriptionPlanLimitPoints(
  plan: SubscriptionPlan,
  period: 'daily' | 'weekly' | 'monthly',
): number | null {
  const pointValue = plan[`${period}_limit_points`]
  const legacyValue = plan[`${period}_limit_usd`]
  const value = pointValue !== undefined ? pointValue : legacyValue
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}

export function subscriptionPlanHasQuota(plan: SubscriptionPlan): boolean {
  return (['daily', 'weekly', 'monthly'] as const)
    .some((period) => subscriptionPlanLimitPoints(plan, period) !== null)
}
