import { describe, expect, it } from 'vitest'
import { rechargeBonusPointsForAmount } from '../rechargeBonus'

describe('recharge bonus tiers', () => {
  const tiers = [
    { id: 'small', threshold_cny: 50, bonus_percent: 3 },
    { id: 'large', threshold_cny: 200, bonus_percent: 20 },
    { id: 'medium', threshold_cny: 100, bonus_percent: 8 },
  ]

  it('applies only the highest qualifying percentage to the full principal', () => {
    expect(rechargeBonusPointsForAmount(tiers, 49.99)).toBe(0)
    expect(rechargeBonusPointsForAmount(tiers, 100)).toBe(8)
    expect(rechargeBonusPointsForAmount(tiers, 500)).toBe(100)
    expect(rechargeBonusPointsForAmount(tiers, 150)).toBe(12)
  })

  it('never returns invalid or negative bonus points', () => {
    expect(rechargeBonusPointsForAmount([
      { threshold_cny: 10, bonus_percent: Number.NaN },
      { threshold_cny: 5, bonus_percent: -4 },
    ], 20)).toBe(0)
  })
})
