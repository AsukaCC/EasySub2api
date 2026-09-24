# Account Model Fingerprinting

The account list's **Model Fingerprint Test** action selects a text model and
starts one background conversation with three sequential challenge/response
turns. Each turn includes the preceding turns; session identifiers are reused.
Different accounts can run concurrently (up to four jobs per server process).
There is never more than one active fingerprint job for the same account,
including across server replicas. Closing the dialog does not stop the job.

## Attribution

The Go scorer and bundled candidate bank are adapted from
[xqy2006/ModelTrace](https://github.com/xqy2006/ModelTrace), pinned at
`55a2e4a55170423b484d701e9a82ab62b268c811` (MIT). The original license is in
`backend/internal/pkg/modeltrace/LICENSE`.

Scoring retains upstream's nuisance-projected Hellinger features (75%), ordered
block features (25%), score averaging, and softmax. A golden test compares all
candidate probabilities against upstream's JavaScript implementation using
public GPT, Claude, and mixed reference samples.

The bank contains 16 GPT/Claude candidates. Other model families can be sampled,
but cannot be identified reliably by this bank. Percentages are closed-set
relative attribution, not proof of identity. In particular, this application's
requested **single conversation** mode differs from upstream's three independent
conversations: its samples are correlated, so upstream calibration accuracy does
not apply. The UI discloses this limitation. Do not automatically change account
model mappings, billing multipliers, or scheduling based on these shares.

## Resource Limits

- Exactly three sequential challenges of 292-332 integers; no automatic extra
  samples or failed-generation retries. Each sample must meet upstream's
  `max(80, ceil(expected_count * 0.55))` threshold; all three must pass.
- Native output limits are 1,536 tokens per turn. Codex OAuth does not accept that
  field, so the stream is cancelled when the requested integer count is received.
  All protocols also have a 70-second turn timeout and a four-minute job timeout.
  Provider-side reasoning tokens and billing after cancellation are controlled by
  the upstream service, so these are not guaranteed monetary limits.
- Recognized OpenAI reasoning models use `low` reasoning effort to reduce probe
  consumption. Other models retain provider defaults.
- No tools are offered. Antigravity probes bypass retry and optional overage
  fallback. ModelTrace scoring runs locally without another model call.
- Uses the existing account-test credentials, proxy and transport. This consumes
  upstream account quota like a connectivity test and is not a user API-key usage
  request. No business prompts or user conversations are sampled.

## Storage And API

Admin-authenticated routes:

- `POST /api/v1/admin/accounts/:id/model-fingerprint`, body `{"model_id":"..."}`.
- `GET /api/v1/admin/accounts/:id/model-fingerprint` returns the latest snapshot,
  or `null` before the first test.

Progress and the latest result are stored in the existing account `extra` JSONB
field under `model_fingerprint`; no schema migration or extra runtime is needed.
Raw prompts, generated sequences, and credentials are not stored in the result.
Claims are atomic and expire after five minutes. A stopped server's expired job
is displayed as interrupted; tests are not automatically restarted or charged
again. Writes compare the job ID so an expired worker cannot replace a newer job.
Progress updates refresh the account cache without rebuilding scheduler buckets.

Only the latest tested model/result is kept per account. Rerunning replaces it.
