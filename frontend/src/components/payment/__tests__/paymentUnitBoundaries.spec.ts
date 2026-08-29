import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const srcRoot = resolve(__dirname, '../../..')

function read(path: string): string {
  return readFileSync(resolve(srcRoot, path), 'utf8')
}

describe('payment unit display boundaries', () => {
  it('labels recharge input as CNY instead of USD', () => {
    const source = read('components/payment/AmountInput.vue')
    expect(source).toContain('¥')
    expect(source).not.toMatch(/>\s*\$\s*</)
  })

  it('keeps the recharge conversion and all amount components visible at zero', () => {
    const source = read('views/user/PaymentView.vue')
    expect(source).toContain("t('payment.rechargeRatePreview')")
    expect(source).toContain("t('payment.rechargePrincipal')")
    expect(source).toContain("t('payment.basePoints')")
    expect(source).toContain("t('payment.bonusPoints')")
    expect(source).toContain("t('payment.creditedPoints')")
    expect(source).not.toMatch(/<div v-if="feeRate > 0"[^>]*>[\s\S]*?t\('payment\.fee'\)/)
    expect(source).not.toMatch(/<div v-if="rechargeBonusPoints > 0"[^>]*>[\s\S]*?t\('payment\.bonusPoints'\)/)
  })

  it.each([
    'components/payment/PaymentQRDialog.vue',
    'components/payment/StripePaymentInline.vue',
  ])('%s displays credited amounts as points and payments as CNY', (path) => {
    const source = read(path)
    expect(source).toContain('formatPoints')
    expect(source).toContain('formatCNY')
    expect(source).not.toContain("currencySymbol('USD')")
  })

  it('keeps legacy admin order points and CNY fields separate', () => {
    const source = read('views/admin/orders/AdminOrdersView.vue')
    expect(source).toContain('formatPoints(orderPoints(selectedOrder))')
    expect(source).toContain('formatCNY(selectedOrder.pay_amount)')
    expect(source).toContain('formatCNY(selectedOrder.refund_amount)')
    expect(source).not.toContain("currencySymbol('USD')")
  })
})
