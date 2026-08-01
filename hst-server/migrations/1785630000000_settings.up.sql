-- Small runtime settings that belong to the platform rather than to a group or a symbol.
CREATE TABLE IF NOT EXISTS hst.settings (
    key        VARCHAR(64) PRIMARY KEY,
    value      VARCHAR(255) NOT NULL DEFAULT '',
    updated_at BIGINT       NOT NULL DEFAULT 0
);

COMMENT ON TABLE hst.settings IS
    'platform-wide runtime settings; the engine reads these at boot and on change';

INSERT INTO hst.settings (key, value, updated_at)
VALUES ('end_of_day_at', '23:59', 0)
ON CONFLICT (key) DO NOTHING;
