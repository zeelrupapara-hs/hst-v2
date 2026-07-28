CREATE SEQUENCE IF NOT EXISTS hst.users_login_seq AS BIGINT START WITH 1000000 CACHE 1;

CREATE TABLE IF NOT EXISTS hst.users (
    login               BIGINT       PRIMARY KEY DEFAULT nextval('hst.users_login_seq'),
    client_id           BIGINT,
    "group"             VARCHAR(128) NOT NULL,
    rights              BIGINT       NOT NULL DEFAULT 0,
    cert_serial_number  BIGINT       NOT NULL DEFAULT 0,

    name                VARCHAR(128) NOT NULL DEFAULT '',
    first_name          VARCHAR(64)  NOT NULL DEFAULT '',
    last_name           VARCHAR(64)  NOT NULL DEFAULT '',
    middle_name         VARCHAR(64)  NOT NULL DEFAULT '',
    company             VARCHAR(255) NOT NULL DEFAULT '',
    account             VARCHAR(64)  NOT NULL DEFAULT '',
    country             VARCHAR(64)  NOT NULL DEFAULT '',
    city                VARCHAR(64)  NOT NULL DEFAULT '',
    state               VARCHAR(64)  NOT NULL DEFAULT '',
    zip_code            VARCHAR(32)  NOT NULL DEFAULT '',
    address             TEXT         NOT NULL DEFAULT '',
    phone               VARCHAR(64)  NOT NULL DEFAULT '',
    email               VARCHAR(255) NOT NULL DEFAULT '',
    id_document         VARCHAR(64)  NOT NULL DEFAULT '',
    status              VARCHAR(64)  NOT NULL DEFAULT '',
    comment             TEXT         NOT NULL DEFAULT '',
    color               INTEGER      NOT NULL DEFAULT 0,
    language            INTEGER      NOT NULL DEFAULT 0,
    hsid                VARCHAR(64)  NOT NULL DEFAULT '',

    password_main       TEXT         NOT NULL DEFAULT '',
    password_investor   TEXT         NOT NULL DEFAULT '',
    password_api        TEXT         NOT NULL DEFAULT '',
    password_phone      TEXT         NOT NULL DEFAULT '',
    failed_attempts     INTEGER      NOT NULL DEFAULT 0,
    locked_until        BIGINT NOT NULL DEFAULT 0,

    leverage            INTEGER      NOT NULL DEFAULT 100,
    agent               BIGINT,
    limit_orders        INTEGER      NOT NULL DEFAULT 0,
    limit_positions     INTEGER      NOT NULL DEFAULT 0,

    balance             NUMERIC(20,8) NOT NULL DEFAULT 0,
    credit              NUMERIC(20,8) NOT NULL DEFAULT 0,
    interest_rate       NUMERIC(10,4) NOT NULL DEFAULT 0,
    commission_daily    NUMERIC(20,8) NOT NULL DEFAULT 0,
    commission_monthly  NUMERIC(20,8) NOT NULL DEFAULT 0,
    balance_prev_day    NUMERIC(20,8) NOT NULL DEFAULT 0,
    balance_prev_month  NUMERIC(20,8) NOT NULL DEFAULT 0,
    equity_prev_day     NUMERIC(20,8) NOT NULL DEFAULT 0,
    equity_prev_month   NUMERIC(20,8) NOT NULL DEFAULT 0,

    trade_accounts      TEXT         NOT NULL DEFAULT '',
    lead_campaign       VARCHAR(128) NOT NULL DEFAULT '',
    lead_source         VARCHAR(255) NOT NULL DEFAULT '',
    api_data            JSONB        NOT NULL DEFAULT '[]'::jsonb,

    registration        BIGINT  NOT NULL DEFAULT 0,
    last_access         BIGINT NOT NULL DEFAULT 0,
    last_pass_change    BIGINT NOT NULL DEFAULT 0,
    last_ip             INET,
    updated_at          BIGINT  NOT NULL DEFAULT 0,

    CONSTRAINT users_login_range CHECK (login > 0),
    CONSTRAINT users_group_set   CHECK ("group" <> '')
);

CREATE INDEX IF NOT EXISTS users_client_id_idx   ON hst.users (client_id);
CREATE INDEX IF NOT EXISTS users_group_idx       ON hst.users ("group");
CREATE INDEX IF NOT EXISTS users_agent_idx       ON hst.users (agent);
CREATE INDEX IF NOT EXISTS users_email_idx       ON hst.users (email);
CREATE INDEX IF NOT EXISTS users_last_access_idx ON hst.users (last_access DESC);
CREATE INDEX IF NOT EXISTS users_locked_until_idx ON hst.users (locked_until) WHERE locked_until > 0;

ALTER TABLE hst.users
    ADD CONSTRAINT users_client_id_fkey FOREIGN KEY (client_id)
    REFERENCES hst.clients (client_id) ON DELETE SET NULL;

ALTER TABLE hst.users
    ADD CONSTRAINT users_agent_fkey FOREIGN KEY (agent)
    REFERENCES hst.users (login) ON DELETE SET NULL;

ALTER TABLE hst.clients
    ADD CONSTRAINT clients_introducer_fkey FOREIGN KEY (introducer)
    REFERENCES hst.users (login) ON DELETE SET NULL;

ALTER TABLE hst.clients
    ADD CONSTRAINT clients_client_origin_login_fkey FOREIGN KEY (client_origin_login)
    REFERENCES hst.users (login) ON DELETE SET NULL;

COMMENT ON COLUMN hst.users.rights IS
    'EnUsersRights bitmask: 0x1=enabled 0x2=password 0x4=trade_disabled(INVERTED) 0x8=investor(internal) 0x10=confirmed 0x20=trailing 0x40=expert 0x100=reports 0x200=readonly(internal) 0x400=reset_pass 0x800=otp_enabled(unused) 0x2000=sponsored_hosting 0x4000=api_enabled 0x8000=push_notification 0x10000=technical 0x20000=exclude_reports. 0x1000 reserved.';
COMMENT ON COLUMN hst.users.language IS 'Windows LANGID';
COMMENT ON COLUMN hst.users.color IS 'COLORREF, request colour in the manager terminal';
COMMENT ON COLUMN hst.users.status IS 'free text, not an enum';
COMMENT ON COLUMN hst.users.password_main IS 'EnUsersPasswords slot 0, argon2id PHC string';
COMMENT ON COLUMN hst.users.password_investor IS 'EnUsersPasswords slot 1, read-only session';
COMMENT ON COLUMN hst.users.password_api IS 'EnUsersPasswords slot 2, gated by rights 0x4000';
COMMENT ON COLUMN hst.users.password_phone IS 'MT5 PhonePassword, support verification only, not a login credential';

COMMENT ON SEQUENCE hst.users_login_seq IS
    'CACHE 1: logins are user facing account numbers, gaps would confuse';

COMMENT ON TABLE hst.users IS 'all time columns are unix nanoseconds, 0 means unset';
