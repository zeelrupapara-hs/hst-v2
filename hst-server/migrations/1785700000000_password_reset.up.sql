-- One time codes for trader password recovery. Only a hash of the code is kept.
CREATE TABLE IF NOT EXISTS hst.password_reset_codes (
    code_id    BIGSERIAL PRIMARY KEY,
    login      BIGINT NOT NULL,
    code_hash  BYTEA  NOT NULL,
    expires_at BIGINT NOT NULL DEFAULT 0,
    used_at    BIGINT NOT NULL DEFAULT 0,
    attempts   INT    NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS password_reset_codes_login_idx ON hst.password_reset_codes (login);

COMMENT ON TABLE hst.password_reset_codes IS
    'single use password recovery codes; expires_at, used_at and created_at are epoch nanoseconds';
