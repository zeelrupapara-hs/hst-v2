-- Replies chain into threads; attachments hang off one shared id per logical message.
ALTER TABLE hst.mails
    ADD COLUMN thread_id UUID,
    ADD COLUMN attach_id UUID;

UPDATE hst.mails SET thread_id = tracking_id::uuid;

ALTER TABLE hst.mails ALTER COLUMN thread_id SET NOT NULL;

CREATE INDEX mails_thread_idx ON hst.mails (thread_id);

-- Staged rows carry a NULL attach_id until the send that references them stamps it.
CREATE TABLE hst.mail_attachments (
    attachment_id BIGSERIAL PRIMARY KEY,
    attach_id     UUID,
    owner_login   BIGINT       NOT NULL,
    name          VARCHAR(255) NOT NULL,
    size          BIGINT       NOT NULL,
    data          BYTEA        NOT NULL,
    created_at    BIGINT       NOT NULL DEFAULT 0
);

CREATE INDEX mail_attachments_attach_idx ON hst.mail_attachments (attach_id);
