-- Retire legacy registration access codes. Affiliate referral codes are now
-- the only invitation codes shown during registration.
INSERT INTO settings (key, value, updated_at)
VALUES
    ('invitation_code_enabled', 'false', NOW()),
    ('affiliate_rebate_recipient', 'inviter', NOW())
ON CONFLICT (key) DO UPDATE
SET value = CASE
        WHEN EXCLUDED.key = 'invitation_code_enabled' THEN 'false'
        ELSE settings.value
    END,
    updated_at = NOW();
