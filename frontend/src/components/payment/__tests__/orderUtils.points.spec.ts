import { describe, expect, it } from 'vitest'
import {
  canDirectRefund,
  rechargeAffiliatePoints,
  rechargeBasePoints,
  rechargeBonusPoints,
  rechargeCreditedPoints,
  rechargeFeeAmount,
  rechargePrincipalAmount,
} from '../orderUtils'

describe('recharge order amount compatibility', () => {
  it('uses explicit money and point snapshots when present', () => {
    const order = {
      amount: 999,
      pay_amount: 108,
      gateway_base_amount: 77,
      principal_amount: 100,
      fee_amount: 8,
      base_points: 100,
      bonus_points: 20,
      credited_points: 120,
      affiliate_rebate_points: 10,
    }

    expect(rechargePrincipalAmount(order)).toBe(100)
    expect(rechargeFeeAmount(order)).toBe(8)
    expect(rechargeBasePoints(order)).toBe(100)
    expect(rechargeBonusPoints(order)).toBe(20)
    expect(rechargeCreditedPoints(order)).toBe(120)
    expect(rechargeAffiliatePoints(order)).toBe(10)
  })

  it('maps legacy order fields without mixing gateway money and credited points', () => {
    const order = { amount: 120, pay_amount: 108, gateway_base_amount: 100 }

    expect(rechargePrincipalAmount(order)).toBe(100)
    expect(rechargeFeeAmount(order)).toBe(8)
    expect(rechargeBasePoints(order)).toBe(100)
    expect(rechargeBonusPoints(order)).toBe(20)
    expect(rechargeCreditedPoints(order)).toBe(120)
    expect(rechargeAffiliatePoints(order)).toBe(0)
  })
})

describe('direct recharge refund window', () => {
  const completedAt = '2026-08-20T00:00:00Z'
  const refundDeadline = '2026-08-27T00:00:00Z'

  it('allows a recharge refund only inside the half-open seven-day window', () => {
    const order = {
      status: 'COMPLETED',
      order_type: 'balance',
      completed_at: completedAt,
      refund_deadline: refundDeadline,
    }
    expect(canDirectRefund(order, new Date('2026-08-26T23:59:59Z'))).toBe(true)
    expect(canDirectRefund(order, new Date(refundDeadline))).toBe(false)
  })

  it('rejects subscription and missing-completion orders', () => {
    expect(canDirectRefund({
      status: 'COMPLETED',
      order_type: 'subscription',
      completed_at: completedAt,
      refund_deadline: refundDeadline,
    }, new Date('2026-08-21T00:00:00Z'))).toBe(false)
    expect(canDirectRefund({ status: 'COMPLETED', order_type: 'balance' })).toBe(false)
  })
})
