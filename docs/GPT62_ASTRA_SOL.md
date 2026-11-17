# GPT-6.2 Astra / Sol

Verified against the official OpenAI API documentation on 2026-11-18:

- https://developers.openai.com/api/docs/models/gpt-6.2-astra
- https://developers.openai.com/api/docs/models/gpt-6.2-sol
- https://developers.openai.com/api/docs/pricing
- https://developers.openai.com/api/docs/guides/latest-model

## Pricing and points

USD per million tokens, Standard processing:

| Model | Input | Cache read | 30m cache write | 12h cache write | Output |
| --- | ---: | ---: | ---: | ---: | ---: |
| gpt-6.2-astra | 8 | 0.40 | 10 | 20 | 40 |
| gpt-6.2-sol | 2 | 0.10 | 2.50 | 5 | 10 |

Fast / priority is 2x. Ultrafast is 3x before 2027-03-01 00:00 Pacific
(2027-03-01 08:00 UTC), and 6x from that instant. The billing timestamp controls
this boundary, including when using the unified billing resolver. Prices cached
before the cutoff cannot retain the promotional multiplier indefinitely.

For requests above 272,000 total input tokens (ordinary input plus cache reads
and writes), input and all cache rates double; output is 1.5x for the entire
request. Exactly 272,000 remains short context. Existing group/account controls
still determine whether long-context pricing is enabled.

The resulting token cost uses the existing points settlement pipeline: group
rate x user-level rate x selected dynamic discount. Existing configured prices,
including explicit zero prices, remain authoritative. No recharge conversion,
historical usage, wallet, reservation, or refund rules are changed.

Batch, discounted background completion windows, and regional surcharges are
not added by this update. Batch discounts must not be inferred from a Flex tier.

## Model and export support

Both models are available in the OpenAI model list, whitelist, mapping presets,
Codex manifest, and OpenCode export catalog. Their context window is 1,050,000
tokens; maximum input is 922,000 and maximum output is 128,000. Input modalities
are text and image; output is text.

Tool calls require Responses. Disabled reasoning (`none` / `minimal`) is rejected
at the compatibility boundary. Model aliases preserve their family for routing
and billing rather than falling back to GPT-6 Astra.

OpenAI API Key Codex and CC Switch exports default to `gpt-6.2-sol`. Users can
edit the main model, set a separate subagent model, and toggle WebSocket support
before export. Omitting a subagent override lets it follow the main model.
Explicit `image_generation` tools retain the existing forwarding behavior;
the gateway does not add the tool automatically.
