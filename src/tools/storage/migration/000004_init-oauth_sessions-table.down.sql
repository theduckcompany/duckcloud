DROP TABLE IF EXISTS oauth_sessions;

DROP INDEX IF EXISTS idx_oauth_sessions_expires_at;
DROP INDEX IF EXISTS idx_oauth_sessions_access;
DROP INDEX IF EXISTS idx_oauth_sessions_refresh;
