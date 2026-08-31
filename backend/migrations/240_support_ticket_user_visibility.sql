SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '5min';

-- Keep existing deployments published when the ticket runtime switch was
-- already enabled, while fresh installations remain hidden by default.
INSERT INTO settings (key, value, updated_at)
SELECT 'support_tickets_user_visible', COALESCE(
    (SELECT value FROM settings WHERE key = 'support_tickets_enabled'),
    'false'
), NOW()
ON CONFLICT (key) DO NOTHING;
