# Account Protection

## Defaults

- New accounts receive server-owned protection defaults. Existing accounts are
  not migrated by deployment or ordinary edits.
- Independent OpenAI and Anthropic OAuth/setup-token accounts default to
  `legacy`. Other accounts, including credential shadows, receive a positive
  concurrency bound without an unsupported identity or TLS preset.
- Positive administrator concurrency limits are retained. Enabling legacy or
  mode1 on an unlimited account starts at 16; mode2 starts at 8.
- Disabling protection requires explicit confirmation. It restores managed
  identity/TLS settings but retains subsequent administrator concurrency edits.

| Strategy | Identity | TLS | Default OpenAI Integrity |
| --- | --- | --- | --- |
| legacy | session | nodejs24 | observe |
| mode1 (v3) | device | standard | enforce |
| mode2 | full | nodejs22 | observe |

The global TLS fingerprint switch remains authoritative. A configured profile
is not proof that a handshake has been observed. Compatibility presets do not
guarantee upstream model quality or account acceptance.

## Outbound Identity

The default Codex UA/originator is `codex-tui`. Valid explicit client identities
remain supported. Stable and prerelease versions share format and minimum-version
validation. The compiled fallback remains 0.154.0.

`codex_allow_prerelease_version` and
`disable_codex_originator_normalization` are deprecated compatibility keys and
have no runtime effect. `disable_codex_identity_enforcement` remains independent.

OAuth session/device metadata is scoped to the upstream credential namespace,
not the downstream API key. Identical client identifiers on the same upstream
account therefore produce identical projected identifiers across API keys.
Local authorization and scheduler cache scoping are unchanged. Credential shadows
resolve their identity source from the parent account.

Installation headers are no longer independently filled from raw device_id.
Identity projection and enabled convergence modes own their final values.
HTTP identity completion may add `responses=experimental`; compatibility
bridges omit originator, while WebSocket negotiation retains its own Beta token.

Deployment changes upstream identifiers. Existing upstream conversations and
connections may require a new session; do not clear persistent caches or volumes
as a deployment shortcut.

## Administrator API

All paths are under `/api/v1/admin/accounts` and require administrator access.
Account IDs remain strings.

- GET `/anti-degrade/strategies`: strategy registry.
- GET `/:id/anti-degrade?mode=legacy`: read-only preview and runtime plan.
- POST `/:id/anti-degrade/apply?mode=legacy`: apply or switch atomically.
- POST `/:id/anti-degrade/revert`: body `{"confirm_disable":true}`.
- PUT `/:id/anti-degrade`: body `{"enabled":false,"confirm_disable":true}`.
- POST `/anti-degrade/enable-batch`: body `{"account_ids":["..."]}`,
  at most 500 IDs; returns `success_ids` and per-account `failures`.
- PUT `/:id/anti-degrade/integrity`: body `{"mode":"observe"}`;
  accepts off, observe, or enforce for OpenAI accounts.

Protection metadata and snapshots live in account extra JSON. Ordinary edits,
refreshes and bulk updates preserve server-managed fields. A concurrent strategy
transition returns a configuration conflict rather than committing partial state.
Integrity diagnostics contain field names only, never conversation payloads.
