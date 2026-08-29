-- Add exact request-model allowlists to reusable account model restriction rules.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

ALTER TABLE account_model_rules
    ADD COLUMN IF NOT EXISTS whitelist JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE account_model_rules
SET whitelist = '[]'::jsonb
WHERE whitelist IS NULL;
