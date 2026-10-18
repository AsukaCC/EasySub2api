# Sub2API v0.2.13-v0.2.17 adaptation

Baseline: EasySub2api v0.0.73 (`7ff81e10b`).
Reference: https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.17

**PARTIAL IMPLEMENTATION. Full-plan acceptance has not been reached.**
Preserve UUID/string IDs, point wallets, refund reservations, model attribution,
multi-group pricing and global GPT-5.4/5.5 retirement. No later upstream releases
are included. No migrations, deployment, release or remote push were performed.

Status: tested = focused regression passed; adapted = code exists and compiles
but dedicated behavior coverage remains; partial = known implementation gaps;
pending = not implemented, not a claim of non-applicability.

## Phase 1

| Item | Upstream commits | Result and validation |
| --- | --- | --- |
| EasyPay whitelist, delimiters, merchant identity | d1aac6b98 | Tested in provider package; additionally rejects duplicate fields and nonfinite amounts |
| Query GET/form and numeric strings, key redaction | 1b92038ea | Query/confirmation tests pass; error bodies also redact merchant key |
| Paid state, amount, order and transaction confirmation | 1b92038ea | Provider tests pass; handler reuses verified merchant; DB concurrency/restart coverage pending |
| Return URL query cleanup | d1aac6b98, 1b92038ea | Adapted in payment_resume_service; complete resume regression pending |
| Stripe local snapshot and late async results | f996b0bfa, 1c996fe30 | Tested; string IDs and local secret; no secret in generated URL |
| Airwallex late SDK results | 795a7e278 | Tested |
| First-install administrator randomness/validation | 6e315616e | Setup package tests pass; EasySub2api domain retained |
| Refund freeze, stable IDs, reconciliation | 44e35fbf8, 7c6f70680, 83337a0c6, f37b69b77 | Existing local transactional point flow retained; isolated DB recovery tests pending |

## Phase 2

| Item | Upstream commits | Result and validation |
| --- | --- | --- |
| WS image input tokens | b6a157c19 | Relay package tests pass |
| WS per-turn billing refresh | 73381fc9c | Service regression covers selected group, changed rate, same timestamp/gate/settlement, and refresh failure; end-to-end WS coverage pending |
| Sol subscription long context | 43768680d | Adapted using actual billing credential/final upstream model; dedicated accounting regression pending |
| Gemini 3.6 / Antigravity Claude tier pricing | 4426593e9, 9d3cc4507 | Pending; local fallback pricing differs |
| Composite OAuth ambiguity | d19265ddb | HTTP/WS resolver and native/count_tokens guards adapted; resolver regression added; complete route matrix pending |
| Anthropic protocol base URL | 867563ae7 | Native and count_tokens adapted |
| Composite aliases / remote model constraints | 1888de135, 71e97d135 | Pending |
| Role, tool choice, legacy functions, refusal | 9a8ab4f88, f3d752693, 7a767d5ce, 775c1a3c1 | apicompat tests pass |
| Structured input recovery without unsafe retries | 983cf297d, 63800c50d, c15d858ec | Pending |
| Responses created_at | bdfe215cb | Partial: JSON/SSE/HTTP-WS paths adapted, stamper tests pass; synthetic terminal paths need audit |
| Thinking signatures, heartbeat, negative index | 412718edf, 9f84204c2, 15f469c86 | Adapted; conversion package tests pass, full stream matrix pending |
| Cache TTL ordering | da4419e62 | Adapted; dedicated gateway tests pending |
| Tool history, reasoning recovery, beta headers | dab3b87ea, 64caa9af8, a4ea8059b, 852f0f824 | Pending |
| Image bridge platform restriction | f48866958 | Adapted to OpenAI accounts only |
| Gemini error policy once per response | 14cd58717 | Native/Messages/Chat adapted; dedicated behavior tests pending |
| Vertex 429 same-account retry / missing SA | 89ad4c2ca, 20c4bf5ae | Pending |
| Model-scoped cooldown and all entrypoints | 5a25507dc, fc82e8310, 2023c290e, cfbbfd6b3 | Scope configuration/parser/state, credential-aware 401 handling, count_tokens/Bedrock/retry paths and local OpenAI runtime blocking adapted. Scope regression tests pass; complete ingress audit remains pending. |
| Gemini/Antigravity images, tools and native routing | ba2a71733, feab6d80d | Pending |
| Grok empty completion / consent | 3613174cf, 54810f8c8 | Empty completion adapted; consent update pending |
| OpenCode retry cap, neutral 403 and endpoints | aa3b04900, 480f668ba, df89228c2, 0dccd7b6f | Partial; entrypoint unsupported-model rejection pending |
| Kimi international / Zhipu classification | ce6242fca, 1774b79dc | Adapted; focused behavior tests pending |
| Codex import workspace isolation | 16611a8ee | Adapted with composite account/user identity key |
| Codex discovery, non-null tiers, 500k context | bbba01dae, c0e5a4670, 071858126 | Partial; OpenAI generated HTTP/WS config tested; mapping/other-platform audit pending |

## Phase 3

| Item | Upstream commits | Result and validation |
| --- | --- | --- |
| Countdown and exponent formatting | c4224045d, 5ace0699b | Adapted |
| Zero multiplier | 7d0bf9c87 | Existing UsageTable already uses nullish fallback; retained |
| Long-context badge | d5f490e10 | Pending |
| IME, disabled selectors, topmost Escape | d01628975, 8167d85b3, e0019da46 | Adapted; browser verification pending |
| Group RPM/rate stale results | 0f189af8d, 8811c125e, 7b1aa9314 | Focused tests pass |
| Custom page stale Markdown | 4b6ed15a8 | Adapted; existing tests pass; sanitization retained |
| Cleanup, scheduled tests, templates, backup | 8a495b0a5, 8f636513f, 4551fb0d8, 00a922c42 | Adapted; remaining race tests pending |
| Verification, drafts, notification email | c64145312, 3f2f66f32, abbe42ac4, 636faa695 | Adapted; step-up tests pass, profile race tests pending |
| Compliance request isolation | 21d90fb5e | Existing fetch dedup retained; accept response version guard added |
| Errors, announcements, saved rule status | e180a939a, cee76030b, f022a0457, 5753552b9 | Adapted |
| Monitor samples/cache and loaded settings | 7cd3dc85c, 8e3d75b57, 81bd85ec5 | Adapted |
| Alert duration and pagination | b07b1fb5d, b47204251 | Adapted |
| Monitor platforms, timezone, error categories, group labels | 69d6d106f, 2bfc49477, 3abd64658, 66b7ce08e | Pending audit of differing local implementations |
| Deleted group moderation references | 27de27ac6 | Pending; transaction lock order and UUID JSON must be adapted together |
| Bare API fallback | abdfbb704 | Adapted; embedded-route tests pending |
| Final inline account filters | c2cca510f | Pending audit; existing layout retained |

## Phase 4

| Item | Upstream commits | Result and validation |
| --- | --- | --- |
| Profiles, platform catalog, application validation | c063a4016, 3f8eea080, 8ac2936a3, 3b89cc08d, c556aee8f, 7ffa38a8a | Pending |
| Database platform constraint migration | eefa7dd9d | Pending; no database constraints removed |
| Cline / Command Code | 874871a25, e2939cd8f, 926441461, 7ca08d521 | Pending across account/scheduling/billing/quota/monitoring |
| Haiku 5.5 | d56f5e787, 26f70c6a8, c4bdc4847 | Partial: registry/effort/pricing/validation/signed Responses history/whitelist implemented; request and billing tests pass; beta integration pending |
| Gemini 4 Argon / Antigravity catalogs | 373d43533, c6c9935ae | Pending |
| Vertex API Key, model discovery, AQ. key | 81e6ec534, 2588675ce, 1c68f9996, 6f9b02f12, daebb7a80 | API Key/Express URL/auth, account test and Gemini compatibility forwarding adapted; create/edit forms and AQ. redaction added. URL/auth/redaction tests pass. Model discovery and quota specialization remain pending. |
| Gemini thinking level/summary | 64211de8d | Pending |
| Minimum output tokens | 2a8cb088a | Partial: config default 128/minimum 16 and conversion helpers; full ingress/WS audit pending |
| API Key web search simulation | 26e55cf33, 7de15113c | Pending |
| Group translations | 9a5133310, 8b74741ef | Migration 284, Ent schema/generated code, admin CRUD/editor, duplicate descriptions, user key/subscription/redeem/usage/channel/plaza/payment localization implemented. Canonical locale collision rejection and fallback tests pass. Database upgrade and visual verification pending. |
| TPS, percentiles, request account names | 6f6a27f9c, 1817a8428, ba280b3d5 | Per-request output rate, raw-window P5/P10/P50/average/sample count and request account names implemented. String IDs retained; modal stale responses ignored. SQL mock and frontend output-rate tests pass. Live database/performance and visual checks pending. |

## Phase 5

All explicit-cache items are pending. References: d5a802f84, 3bfd40f5a,
05d7946af, f0ee427f0, 26b31a83e, 43a3c197c.

- Collection/resource endpoints and user/key/group/account authorization.
- Gateway resource IDs, UUID migrations and generated Ent code.
- Stable operation IDs and point/key/subscription reservations.
- Serialized modifications, rollback, uncertain-result and restart recovery.
- Bound-account dispatch and missing/deleted/unsupported/capacity distinctions.
- Storage pricing fallback and creation-time billing snapshots on renewal.
- Concurrent, cross-user, precision and recovery regression tests.

## Verification and limitations

Passed: Go 1.27.2 `go build -o NUL ./cmd/server`; provider/setup/apicompat/
antigravity/openai_ws_v2 package tests; focused service timestamp/turn billing/
Haiku validation and pricing tests; frontend typecheck, production build and
read-only lint. Focused frontend tests cover payment, group modals, custom page,
step-up and Codex catalog: final rerun passed 7 files / 38 tests.
The final focused service/handler run also passed the OAuth ambiguity regression.
`git diff --check` passed with CRLF conversion notices only.

A broader service/handler regex run for Profit/CreatedAt/Composite/LongContext/
Payment/Refund failed. Examples: GPT-5.5 catalog expectations, auth snapshot
version 22 vs 26, GPT-5.4 long-context expectations, and Gemini/Antigravity token
invalidator expectations. Not all failures have been established as baseline
failures. This run is **not** counted as passing.

Go was absent from PATH; the cached bootstrap toolchain selected Go 1.27.2.
Docker engine was unavailable. Isolated PostgreSQL/Redis tests, schema upgrade
tests, desktop/narrow-screen visual checks and golangci-lint remain unperformed.
No existing application data was used as a test substitute. Build warnings
include chunk sizes, Sass legacy API, Browserslist age and pnpm overrides.

## Partial-batch upgrade notes

Migration 284 adds group translations with an empty-object default. It has not
been executed against a database. This is not a complete v0.2.17 feature upgrade.
EasyPay callback fulfillment now requires successful merchant order confirmation;
failures leave retry/reconciliation necessary. Stripe requires the matching local
recovery snapshot. Random administrator email applies only to first setup.
WS clients may need to reconnect when billing identity/status changes or fresh
billing configuration cannot be obtained. Historical bills and model mappings
are not rewritten. Refund business rules and retired models remain unchanged.

## Continuation verification (2026-10-18)

- Passed Go server build after the continuation changes.
- Passed focused domain/service/DTO locale tests, Vertex URL/auth tests,
  scope-aware 401 tests, AQ. redaction tests, repository TPS SQL mock tests.
- Passed frontend typecheck, production build and read-only lint; new tests
  for translation editing and per-request TPS passed (2 files / 3 tests).
- Ent generation required x/tools v0.52.0 for Go 1.27 export data. Windows
  mapped-file errors interrupted its formatting step; sequential goimports
  retries completed formatting, and generated code compiled. Existing
  generated runtime-import diagnostic wording was retained to avoid churn.
- A broader cooldown regex also selected existing repeated-Antigravity-401
  expectations and an error-policy mock missing SetModelRateLimit. That run
  failed and is not counted as passing. Local repeated-401 semantics remain.
- No live Vertex request, database migration test, browser layout check or
  production operation was performed. Cline/Command Code, Argon, Vertex model
  discovery/quota completion and all explicit-cache operations remain open.
