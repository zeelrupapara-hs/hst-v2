CREATE TABLE IF NOT EXISTS hst.sessions (
    session_id      UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    login           BIGINT       NOT NULL REFERENCES hst.users (login) ON DELETE CASCADE,
    scope           SMALLINT     NOT NULL DEFAULT 0,
    connection_type SMALLINT     NOT NULL DEFAULT 0,
    token_hash      BYTEA        NOT NULL,
    ip              INET         NOT NULL,
    user_agent      TEXT         NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    last_seen_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ  NOT NULL,
    revoked_at      TIMESTAMPTZ,
    revoked_reason  VARCHAR(64)
);

CREATE INDEX IF NOT EXISTS sessions_login_idx      ON hst.sessions (login) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS sessions_token_hash_idx ON hst.sessions (token_hash);
CREATE INDEX IF NOT EXISTS sessions_expires_at_idx ON hst.sessions (expires_at);

COMMENT ON COLUMN hst.sessions.scope IS
    'EnUsersPasswords: 0=main 1=investor(read-only) 2=api';
COMMENT ON COLUMN hst.sessions.connection_type IS
    'EnUsersConnectionTypes: clients 0=terminal 3=api_web 4=iphone 5=android 11=web; staff 32=admin 33=manager 34=manager_api 36=admin_api 37=manager_api_web';
