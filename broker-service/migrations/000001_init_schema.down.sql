-- Drop indexes first
DROP INDEX IF EXISTS idx_csrf_tokens_token;
DROP INDEX IF EXISTS idx_csrf_tokens_session_id;
DROP INDEX IF EXISTS idx_session_tokens_token;
DROP INDEX IF EXISTS idx_session_tokens_user_id;

-- Drop tables in correct order (reverse of creation)
DROP TABLE IF EXISTS csrf_tokens;
DROP TABLE IF EXISTS session_tokens;
DROP TABLE IF EXISTS users;

-- Drop extensions last
DROP EXTENSION IF EXISTS "uuid-ossp";

