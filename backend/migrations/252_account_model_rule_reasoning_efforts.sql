-- Add optional per-model reasoning effort overrides to reusable account model rules.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

ALTER TABLE account_model_rules
    ADD COLUMN IF NOT EXISTS reasoning_efforts JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE account_model_rules
SET reasoning_efforts = '{}'::jsonb
WHERE reasoning_efforts IS NULL;
