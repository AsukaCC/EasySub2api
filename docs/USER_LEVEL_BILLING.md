# User Level Billing

Each live user, including administrators, belongs to exactly one enabled level
rule. The current tier is determined by that rule's rolling 7/14/30-day spend
window. An unset tier multiplier means 1.

Customer rates are group rate * current level rate * selected dynamic discount.
Overlapping discounts select the lowest coefficient, not their product. Dynamic
quota remains per user in account-side USD: the covered portion is discounted,
and the uncovered portion uses group rate * level rate. Wallet deductions, key
quotas, subscription usage and usage logs use the final eight-decimal amount.
Service-tier/model pricing and Free Fast keep their existing price bases.

Legacy user-group rates, group peak rates and independent media rates no longer
affect customer billing. Their data is retained; RPM overrides remain supported.
Legacy user-group rate write endpoints return USER_GROUP_RATE_DEPRECATED.

## Administration

- The default rule cannot be disabled or deleted. Rules with assigned users
  cannot be disabled or deleted. Other reference protections remain in place.
- Changing the default affects new users and reset-to-default operations only.
- Single/bulk assignment uses the existing rule_ids array with at most one ID.
  Empty replacement resets to default. Add and replace both reassign; removing
  the currently assigned rule resets to default. Removing another rule is a no-op.
- PUT /api/v1/admin/users/level-rules/:id/default changes the default atomically.
- Member queries support page, page_size and search (username/email substring).
- Rule responses include is_default. Default settings retain the array wire
  shape but require exactly one enabled rule ID.
- Key billing reports user_level_multiplier and dynamic_rate_multiplier in
  addition to the existing base/effective rates. Its legacy user_rate_multiplier
  is a compatibility alias of the level rate, not a saved user-group override.

## Upgrade

Pause ingress and drain in-flight usage before deploying. Do not run old and new
billing processes concurrently. Take a restorable database backup using the
existing deployment process. Do not edit or rerun historical migrations.

Migration 276 reuses the sole existing enabled rule and preserves its tiers and
prices. An empty database receives a default rule with a zero-threshold base tier
at 1x. Users without an assignment receive the default. Multiple assignments,
disabled assigned rules, or an ambiguous default fail the migration without
silently choosing prices. Existing historical consumption is never recalculated.

Existing 0.75 user-group overrides stop affecting future prices. For example,
group 0.2 with level 1 now bills at 0.2; a level multiplier of 0.75 bills at 0.15;
an eligible 0.5 discount reduces the covered portion to 0.075. Review the current
level tiers before rollout, then verify a normal request, discounted request,
subscription request, media request, and new-user binding after rollout.

Watch USER_LEVEL_RULES_UNAVAILABLE errors, assignment conflicts, billing failures
and wallet/usage reconciliation. A price lookup failure must not silently charge
the group default. Rollback requires a coordinated code/database restore, not
running an older application against the new binding constraints.

## Verification

Focused regression commands:

```text
cd backend
go test -tags=unit ./internal/service ./internal/repository ./internal/handler ./internal/handler/admin ./migrations -run 'TestSingleUserLevel|TestPreserveBaseLevel|TestUpdateLevelRule|TestDeleteLevelRule|TestListActiveDynamicRateOffers|TestUsageBillingCommandQuantizes|TestQuantizeUsageBillingAmount' -count=1
go test -tags=integration ./internal/repository -run TestSingleUserLevelAtomicLifecycle -count=1
go build ./cmd/server
```

The focused unit tests pass. The integration command is currently blocked by
pre-existing UUID/int64 fixture mismatches in account repository tests and an
outdated scheduler cache mock. Docker is also unavailable on this workstation.
Migration SQL was executed separately against isolated PGlite databases covering
empty/existing installations, default switching, binding constraints, protected
deletion/disable, user deletion, and rollback on multiple assignments. This does
not substitute for the multi-connection integration race test before release.

Three additional existing tests also fail on the unmodified HEAD:
TestBuildUsageBillingCommandUsesAccountSevenDayCostForDynamicQuota,
TestFinalizeLiveCallIsIdempotentAndWritesZeroUsage, and
TestFinalizeLiveCallUsageLogFallsBackToSyncCreate. They are not included in the
passing focused regression command above.

Frontend verification includes typecheck, production build, targeted ESLint,
35 focused Vitest cases, and desktop/mobile browser checks with synthetic API
fixtures. No production database or user data was accessed.
