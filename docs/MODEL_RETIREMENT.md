# GPT-5.4 / GPT-5.5 Retirement

The entire GPT-5.4 and GPT-5.5 families are withdrawn from new traffic.
This includes mini, nano, pro, dated versions, reasoning/context suffixes,
case variants, compact spelling, and provider-prefixed IDs.

## Active Traffic

- Public model lists, Codex manifests, channel availability and plaza pricing
  omit retired IDs. Upstream discovery cannot reintroduce them.
- HTTP requests using retired IDs return a local model-retired error.
- Passthrough does not bypass retirement. Ordinary account mappings pointing
  to a retired target cannot schedule that target.
- WebSocket initial and subsequent model declarations are checked, including
  session.update and nested response.model.
- Retired IDs are never silently converted into an active replacement model.
- New default account tests, compact requests and client configuration examples
  use gpt-5.6-sol. New lightweight Messages dispatch defaults use gpt-5.6-luna.

## Existing Data

No database migration, record deletion, balance change, or historical cost
recalculation is performed. Existing usage rows, model names, bills and exports
are retained. Legacy pricing metadata remains available internally for historical
and in-flight accounting; it is not a new-traffic allowlist or public offer.

Stored account/group mappings and local environment overrides are not rewritten.
Operators must update any explicit retired targets, especially
GATEWAY_OPENAI_COMPACT_MODEL, before using those routes again.

The shared Codex instruction text remains available to active models. Its source
filename is historical provenance, not an indication that GPT-5.5 is still served.
