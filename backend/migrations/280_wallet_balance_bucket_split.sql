-- Split the legacy aggregate wallet into independent recharge and bonus buckets.
-- Historical balance is total available points, so subtract existing bonus points
-- before dropping the legacy columns. Negative recharge balances are preserved.
-- users carries deferred constraint triggers (276); PostgreSQL rejects ALTER TABLE
-- while their events are still queued, so fire them per statement here.
SET CONSTRAINTS ALL IMMEDIATE;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS recharge_balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS frozen_recharge_balance DECIMAL(20,8) NOT NULL DEFAULT 0;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'balance') THEN
        EXECUTE 'UPDATE users SET recharge_balance = balance - bonus_balance, frozen_recharge_balance = frozen_balance - frozen_bonus_balance';
    END IF;
END $$;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_bonus_balance_nonnegative;
ALTER TABLE users
    ADD CONSTRAINT users_bonus_balance_nonnegative
    CHECK (bonus_balance >= 0);

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_frozen_bonus_balance_valid;
ALTER TABLE users
    ADD CONSTRAINT users_frozen_bonus_balance_valid
    CHECK (frozen_bonus_balance >= 0);

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_recharge_balance_bucket_valid;
ALTER TABLE users
    ADD CONSTRAINT users_recharge_balance_bucket_valid
    CHECK (frozen_recharge_balance >= 0);

ALTER TABLE users
    DROP COLUMN IF EXISTS balance,
    DROP COLUMN IF EXISTS frozen_balance;
