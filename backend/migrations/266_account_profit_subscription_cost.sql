-- Per-account subscription-cycle cost used by the expiry-window profit summary.
-- The table intentionally has no foreign key so the setting does not affect
-- account deletion or the durable account-profit history.
CREATE TABLE IF NOT EXISTS account_profit_settings (
    account_id TEXT PRIMARY KEY,
    subscription_cost_points NUMERIC(20, 10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT account_profit_settings_cost_nonnegative
        CHECK (subscription_cost_points >= 0)
);

COMMENT ON TABLE account_profit_settings IS
    'Per-account subscription-cycle cost used for the 30-day pre-expiry profit calculation.';
COMMENT ON COLUMN account_profit_settings.subscription_cost_points IS
    'Subscription-cycle cost in the platform-point numeric accounting basis.';
