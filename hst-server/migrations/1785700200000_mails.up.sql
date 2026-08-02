-- Internal messages between a trading account and support. Folders: 1 inbox, 2 outbox, 3 draft, 4 bin.
CREATE TABLE IF NOT EXISTS hst.mails (
    mail_id         BIGSERIAL PRIMARY KEY,
    tracking_id     TEXT      NOT NULL UNIQUE,
    sender_login    BIGINT    NOT NULL,
    recipient_login BIGINT    NOT NULL DEFAULT 0,
    subject         TEXT      NOT NULL DEFAULT '',
    body            TEXT      NOT NULL DEFAULT '',
    folder          INT       NOT NULL DEFAULT 1,
    read_at         BIGINT    NOT NULL DEFAULT 0,
    created_at      BIGINT    NOT NULL DEFAULT 0,
    updated_at      BIGINT    NOT NULL DEFAULT 0,
    deleted_at      BIGINT    NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS mails_recipient_login_idx ON hst.mails (recipient_login);
CREATE INDEX IF NOT EXISTS mails_sender_login_idx ON hst.mails (sender_login);

COMMENT ON TABLE hst.mails IS 'one row per copy: sending writes the sender outbox row and the recipient inbox row';
COMMENT ON COLUMN hst.mails.folder IS '1 inbox, 2 outbox, 3 draft, 4 bin';
