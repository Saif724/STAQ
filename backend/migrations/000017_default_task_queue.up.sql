-- Ensure the shared default queue exists.
INSERT INTO queues (
    id,
    name,
    description,
    is_active,
    created_at
)
VALUES (
    gen_random_uuid(),
    'default',
    'Default queue for automatically assigned tasks',
    TRUE,
    NOW()
)
ON CONFLICT (name) DO NOTHING;

-- Ensure the default queue is active.
UPDATE queues
SET is_active = TRUE
WHERE name = 'default';

-- Assign the default queue to existing settings that have no default.
UPDATE user_settings
SET
    default_queue_id = (
        SELECT id
        FROM queues
        WHERE name = 'default'
    ),
    updated_at = NOW()
WHERE default_queue_id IS NULL;

-- Create settings for existing users who don't have them.
INSERT INTO user_settings (
    id,
    user_id,
    timezone,
    default_queue_id,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid(),
    u.id,
    'UTC',
    q.id,
    NOW(),
    NOW()
FROM users u
CROSS JOIN (
    SELECT id
    FROM queues
    WHERE name = 'default'
) q
WHERE NOT EXISTS (
    SELECT 1
    FROM user_settings us
    WHERE us.user_id = u.id
);

-- A user must have at most one settings row.
CREATE UNIQUE INDEX uq_user_settings_user_id
ON user_settings(user_id);