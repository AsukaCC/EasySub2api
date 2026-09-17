# Test Gates and Statistics Contracts

## Required CI gates

The main CI workflow builds the backend, compiles all Go tests with the `unit`
tag, and runs focused gateway/admin regression tests. The frontend gate runs
lint, typecheck, and `pnpm run test:ci` (router authorization, token refresh,
image workbench API, proxy API, key bulk updates, and user bulk deletion).
Failures in these gates fail CI; no `continue-on-error` or test skips are added.

These are incremental regression gates, not a claim of complete unit coverage.
The complete suites remain available through:

```sh
cd backend
go test -tags=unit ./... -count=1
```

```sh
cd frontend
pnpm run test:run
```

The 2026-09-17 frontend full-suite run found 56 failing files and 108 failing
tests (203 files and 1721 tests passed). Existing failures include numeric-ID
fixtures after the string-ID migration, removed CSS classes, changed platform
lists, and outdated payment/monitor fixtures. Reconcile these tests with the
current contracts before promoting the full suite to a required gate; do not
weaken production behavior merely to restore old expectations.

## Admin statistics

- Group key counts describe current non-deleted keys; active means status
  `active`. Usage counts and `total_cost` aggregate historical usage rows by
  their recorded group ID, including usage from subsequently deleted keys.
- Redeem-code active counts exclude unused codes whose expiration has passed,
  even if their stored status has not yet changed. `total_value_distributed`
  sums positive, used balance-code values only; concurrency and subscription
  days are not monetary values. `by_type` includes all stored code types.
- Proxy account counts describe current non-deleted account bindings. Historical
  `total_requests`, `success_rate`, and `average_latency` are `null`: usage logs
  do not snapshot the proxy used for each request, so current bindings cannot
  safely attribute historical traffic. Accurate metrics require persisted
  per-request proxy attribution before this contract can return measurements.
- Missing/deleted group and proxy IDs return 404; query failures return errors,
  not zero-filled successful responses.

Redeem CSV export reads all pages at the repository's 1000-row page limit and
only commits a successful response after all queries and CSV writes succeed.
Detected count changes or duplicate IDs abort with 409 so clients can retry.
This is not a transactional snapshot of concurrent edits to code values.

## Production audit enablement

Prompt Audit's real-traffic baseline, blocking-mode approval, alert ownership,
and rollback drill remain operational prerequisites. Local tests do not replace
the evidence and sign-off in
`openspec/changes/add-openai-compatible-prompt-audit/verification.md`.
