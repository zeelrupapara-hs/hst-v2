-- Mail server configurations, one row per SMTP account the broker sends from.
CREATE TABLE IF NOT EXISTS hst.mail_servers (
    mail_server_id BIGSERIAL PRIMARY KEY,
    enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    name           TEXT    NOT NULL DEFAULT '',
    sender_email   TEXT    NOT NULL DEFAULT '',
    sender_name    TEXT    NOT NULL DEFAULT '',
    smtp_server    TEXT    NOT NULL DEFAULT '',
    smtp_login     TEXT    NOT NULL DEFAULT '',
    smtp_password  TEXT    NOT NULL DEFAULT '',
    is_default     BOOLEAN NOT NULL DEFAULT FALSE,
    total_sent     BIGINT  NOT NULL DEFAULT 0,
    total_errors   BIGINT  NOT NULL DEFAULT 0,
    time_min_ms    BIGINT  NOT NULL DEFAULT 0,
    time_max_ms    BIGINT  NOT NULL DEFAULT 0,
    time_sum_ms    BIGINT  NOT NULL DEFAULT 0,
    created_at     BIGINT  NOT NULL DEFAULT 0,
    updated_at     BIGINT  NOT NULL DEFAULT 0
);

-- exactly one default, enforced by the database rather than by whoever writes next
CREATE UNIQUE INDEX IF NOT EXISTS mail_servers_one_default_idx
    ON hst.mail_servers ((is_default)) WHERE is_default;

COMMENT ON TABLE hst.mail_servers IS 'SMTP accounts the platform sends from; not hst.mails, which is the internal trader mailbox';
COMMENT ON COLUMN hst.mail_servers.smtp_server IS 'host:port, ports 25 465 and 587 are supported';
COMMENT ON COLUMN hst.mail_servers.time_sum_ms IS 'divided by total_sent for the average the manager list shows';

-- Outgoing queue. A welcome mail carries the only copy of a generated password, so it is
-- queued and retried rather than sent inline and lost.
CREATE TABLE IF NOT EXISTS hst.outbox (
    outbox_id      BIGSERIAL PRIMARY KEY,
    mail_server_id BIGINT NOT NULL DEFAULT 0,
    recipient      TEXT   NOT NULL,
    subject        TEXT   NOT NULL DEFAULT '',
    body           TEXT   NOT NULL DEFAULT '',
    state          INT    NOT NULL DEFAULT 0,
    attempts       INT    NOT NULL DEFAULT 0,
    last_error     TEXT   NOT NULL DEFAULT '',
    created_at     BIGINT NOT NULL DEFAULT 0,
    sent_at        BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS outbox_state_idx ON hst.outbox (state, outbox_id);

COMMENT ON COLUMN hst.outbox.state IS '0 queued, 1 sent, 2 failed';
