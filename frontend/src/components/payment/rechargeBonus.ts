import type { RechargeBonusTier } from '@/types/payment'

export function rechargeBonusPointsForAmount(
  tiers: RechargeBonusTier[] | null | undefined,
  amountCNY: number,
): number {
  if (!Number.isFinite(amountCNY) || amountCNY <= 0) return 0

  const qualifying = (tiers ?? [])
    .filter((tier) => Number.isFinite(tier.threshold_cny)
      && Number.isFinite(tier.bonus_percent)
      && tier.bonus_percent >= 0 && tier.bonus_percent <= 100
      && tier.threshold_cny <= amountCNY)
    .sort((a, b) => b.threshold_cny - a.threshold_cny)

  return Math.round(amountCNY * (qualifying[0]?.bonus_percent ?? 0) * 1e6) / 1e8
}
