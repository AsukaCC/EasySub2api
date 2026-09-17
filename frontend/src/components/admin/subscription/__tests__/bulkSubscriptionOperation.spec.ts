import { beforeEach, describe, expect, it } from 'vitest'
import { completeBulkSubscriptionOperation, prepareBulkSubscriptionOperation } from '../bulkSubscriptionOperation'

let adminSequence = 1000

beforeEach(() => {
  localStorage.setItem('auth_user', JSON.stringify({ id: 'admin-' + ++adminSequence }))
  sessionStorage.clear()
})

describe('bulk subscription operation identity', () => {
  it('canonicalizes duplicate and reordered string IDs for storage and transport', () => {
    const first = prepareBulkSubscriptionOperation({ action: 'extend', subscription_ids: ['sub-3', 'sub-1', 'sub-3'], days: 7 })
    const retry = prepareBulkSubscriptionOperation({ subscription_ids: ['sub-1', 'sub-3'], days: 7, action: 'extend' })

    expect(retry.request).toEqual({ action: 'extend', subscription_ids: ['sub-1', 'sub-3'], days: 7 })
    expect(retry.key).toBe(first.key)
    expect(retry.outcomeUncertain).toBe(true)
  })

  it('isolates different parameters, actions, targets, and administrators', () => {
    const request = { action: 'extend' as const, subscription_ids: ['sub-1'], days: 7 }
    const first = prepareBulkSubscriptionOperation(request)
    const differentDays = prepareBulkSubscriptionOperation({ ...request, days: 14 })
    const differentAction = prepareBulkSubscriptionOperation({ action: 'revoke', subscription_ids: ['sub-1'] })
    const differentTargets = prepareBulkSubscriptionOperation({ ...request, subscription_ids: ['sub-2'] })
    localStorage.setItem('auth_user', JSON.stringify({ id: 'admin-' + ++adminSequence }))
    const differentAdmin = prepareBulkSubscriptionOperation(request)

    expect(new Set([first, differentDays, differentAction, differentTargets, differentAdmin].map(operation => operation.key)).size).toBe(5)
  })

  it('reuses a stored pending key and clears storage after completion', () => {
    const admin = JSON.parse(localStorage.getItem('auth_user') || '{}').id as string
    const request = { action: 'reset_quota' as const, subscription_ids: ['sub-1'], daily: true, weekly: false, monthly: false }
    const canonical = { subscription_ids: ['sub-1'], action: 'reset_quota', daily: true, weekly: false, monthly: false }
    const scope = 'sub2api:admin:subscription-bulk:' + admin + ':' + JSON.stringify(canonical)
    sessionStorage.setItem(scope, 'saved-operation-key')

    const resumed = prepareBulkSubscriptionOperation(request)
    expect(resumed.key).toBe('saved-operation-key')
    completeBulkSubscriptionOperation(resumed)
    expect(sessionStorage.getItem(scope)).toBeNull()
    expect(prepareBulkSubscriptionOperation(request).key).not.toBe(resumed.key)
  })
})
