-- Compose templates for the manager mail dialog, one row per manager per name.
CREATE TABLE IF NOT EXISTS hst.mail_templates (
    template_id   BIGSERIAL    PRIMARY KEY,
    manager_login BIGINT       NOT NULL,
    name          VARCHAR(128) NOT NULL,
    subject       VARCHAR(128) NOT NULL DEFAULT '',
    body          TEXT         NOT NULL DEFAULT '',
    created_at    BIGINT       NOT NULL DEFAULT 0,
    updated_at    BIGINT       NOT NULL DEFAULT 0,
    UNIQUE (manager_login, name)
);

COMMENT ON TABLE hst.mail_templates IS 'per-manager compose templates for the mail dialog; saving under a taken name overwrites';
