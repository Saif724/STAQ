-- Remove the uniqueness constraint without deleting user settings
-- or the default queue, since both may contain application data.
DROP INDEX IF EXISTS uq_user_settings_user_id;