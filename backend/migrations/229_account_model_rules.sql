-- Reusable model mapping templates for administrator-managed accounts.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

CREATE TABLE IF NOT EXISTS account_model_rules (
    id          UUID        PRIMARY KEY DEFAULT public.uuid_v7(),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    platform    VARCHAR(50)  NOT NULL,
    mapping     JSONB       NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT account_model_rules_platform_name_key UNIQUE (platform, name)
);

CREATE INDEX IF NOT EXISTS account_model_rules_platform_idx
    ON account_model_rules (platform);

COMMENT ON TABLE account_model_rules IS 'Reusable platform-scoped account model mapping rules';
