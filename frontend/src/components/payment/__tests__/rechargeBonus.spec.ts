import { describe, expect, it } from 'vitest'
import { rechargeBonusPointsForAmount } from '../rechargeBonus'

describe('recharge bonus tiers', () => {
  const tiers = [
    { id: 'small', threshold_cny: 50, bonus_points: 3 },
    { id: 'large', threshold_cny: 200, bonus_points: 20 },
    { id: 'medium', threshold_cny: 100, bonus_points: 8 },
  ]

  it('uses the fixed bonus from the highest qualifying CNY threshold', () => {
    expect(rechargeBonusPointsForAmount(tiers, 49.99)).toBe(0)
    expect(rechargeBonusPointsForAmount(tiers, 100)).toBe(8)
    expect(rechargeBonusPointsForAmount(tiers, 500)).toBe(20)
  })

  it('never returns invalid or negative bonus points', () => {
    expect(rechargeBonusPointsForAmount([
      { threshold_cny: 10, bonus_points: Number.NaN },
      { threshold_cny: 5, bonus_points: -4 },
    ], 20)).toBe(0)
  })
})
