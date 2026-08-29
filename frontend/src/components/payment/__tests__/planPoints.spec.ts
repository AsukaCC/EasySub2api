import { describe, expect, it } from 'vitest'
import type { SubscriptionPlan } from '@/types/payment'
import {
  subscriptionPlanHasQuota,
  subscriptionPlanLimitPoints,
  subscriptionPlanOriginalPricePoints,
  subscriptionPlanPricePoints,
} from '../planPoints'

const legacyPlan = {
  price: 12,
  original_price: 15,
  daily_limit_usd: 30,
  weekly_limit_usd: null,
  monthly_limit_usd: null,
} as SubscriptionPlan

describe('subscription plan point aliases', () => {
  it('prefers point fields from the new contract', () => {
    const plan = {
      ...legacyPlan,
      price_points: 20,
      original_price_points: 25,
      daily_limit_points: 50,
    }

    expect(subscriptionPlanPricePoints(plan)).toBe(20)
    expect(subscriptionPlanOriginalPricePoints(plan)).toBe(25)
    expect(subscriptionPlanLimitPoints(plan, 'daily')).toBe(50)
  })

  it('falls back to legacy numeric fields during rollout', () => {
    expect(subscriptionPlanPricePoints(legacyPlan)).toBe(12)
    expect(subscriptionPlanOriginalPricePoints(legacyPlan)).toBe(15)
    expect(subscriptionPlanLimitPoints(legacyPlan, 'daily')).toBe(30)
    expect(subscriptionPlanHasQuota(legacyPlan)).toBe(true)
  })
})
