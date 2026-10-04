# Account Model Fingerprinting

The account list's **Model Fingerprint Test** action selects an API-key account,
a live upstream text model, a protocol, and a supported reasoning effort.
New tests collect three independent responses without previous assistant turns
or injected system prompts. Credentials remain on the server.
Different accounts can run concurrently (up to four jobs per server process).
There is never more than one active fingerprint job for the same account,
including across server replicas. Closing the dialog does not stop the job.

## Attribution

The Go scorer and bundled candidate bank are adapted from
[xqy2006/ModelTrace](https://github.com/xqy2006/ModelTrace), pinned at
`d4131b30243dfa05e70180b5eedde742103f1d73` (MIT). The original license is in
`backend/internal/pkg/modeltrace/LICENSE`.

Scoring retains upstream's nuisance-projected Hellinger features (75%), ordered
block features (25%), score averaging, and softmax. A golden test compares all
candidate probabilities against upstream's JavaScript implementation using
public GPT, Claude, and mixed reference samples.

The bank contains 17 GPT/Claude candidates, including Claude Opus 5.5. Other model families can be sampled,
but cannot be identified reliably by this bank. Percentages are closed-set
relative attribution, not proof of identity. Legacy single-conversation results
retain their original sampling mode. Do not automatically change account
model mappings, billing multipliers, or scheduling based on these shares.

## Resource Limits

- Three valid samples with at most six attempts, using 292-332 integers per
  challenge. Each sample must meet upstream's
  `max(80, ceil(expected_count * 0.55))` threshold; all three must pass.
- API tests use OpenAI Chat Completions or Anthropic Messages JSON responses.
  Automatic detection starts with Chat Completions and switches on
  400/404/405/422 before the first valid sample. Authentication and rate-limit
  failures stop the job. Every actual attempt is logged.
- Output budgets are 4,096 tokens for default/low effort and 16,384 for higher
  effort. Each attempt has a 90-second timeout; jobs end before their ten-minute
  lease expires. Provider-side billing remains controlled by the upstream.
- No tools are offered. ModelTrace scoring runs locally without another model call.
- Uses the existing account-test credentials, proxy and transport. This consumes
  upstream account quota like a connectivity test and is not a user API-key usage
  request. No business prompts or user conversations are sampled.

## Storage And API

Admin-authenticated routes:

- `POST /api/v1/admin/accounts/:id/model-fingerprint`, body
  `{"model_id":"...","protocol":"auto","reasoning_effort":"low"}`.
- `GET /api/v1/admin/accounts/:id/model-fingerprint` returns the latest snapshot,
  or `null` before the first test.

- `GET /api/v1/admin/accounts/:id/model-fingerprint/models`: live models and efforts.
- `GET /api/v1/admin/accounts/:id/model-fingerprint/history?page=1&page_size=10`.
- `GET /api/v1/admin/accounts/:id/model-fingerprint/schedule`.
- `PUT /api/v1/admin/accounts/:id/model-fingerprint/schedule`, body with `enabled`
  and `options` containing the same fields as a manual test.

Migration `283_model_fingerprint_history.sql` adds history and schedule tables.
The account `extra.model_fingerprint` field still holds the latest result.
Raw prompts, generated sequences, and credentials are not stored in the result.
Claims are atomic and expire after ten minutes. A stopped server's expired job
is displayed as interrupted; tests are not automatically restarted or charged
again. Writes compare the job ID so an expired worker cannot replace a newer job.
Progress updates refresh the account cache without rebuilding scheduler buckets.

Schedules are enabled explicitly and run at the next `:00` or `:30` slot, dispatched
by the existing minute scheduler. Missed slots are coalesced after restart.
Disabled/deleted/expired accounts and disabled/deleted/non-admin schedule owners
are not dispatched. Claiming the account lease, advancing the schedule and
inserting history share one database statement. Busy slots retry at the next tick.

New snapshots and history are visible for 24 hours from test start. Legacy snapshots
use their completion/lease timestamp. Expired history and fingerprint usage logs
are deleted in bounded batches every minute; list queries hide expired records
before cleanup. Ordinary usage records are not deleted.

Each probe writes the regular `usage_logs` table with `request_type=test`, the
administrator, account, model, effort, actual endpoint, duration, upstream tokens,
and `fingerprint:<job>:<attempt>` request ID. The dialog links to the normal
administrator usage page `/admin/usage/admin`, filtered by account and test type.
These calls consume upstream quota but do not debit a user's wallet or platform key.
Buffered JSON calls do not report a fabricated time-to-first-token.

Reproduce the bank import and golden fixtures from a checkout of the pinned commit:

```powershell
node frontend/scripts/sync-modeltrace.mjs <ModelTrace-checkout>
```
