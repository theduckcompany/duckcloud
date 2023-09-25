DROP TABLE IF EXISTS oauth_sessions;

DROP INDEX IF EXISTS idx_oauth_sessions_access_token;
DROP INDEX IF EXISTS idx_oauth_sessions_refresh_token;
DROP INDEX IF EXISTS idx_oauth_sessions_user_id;
