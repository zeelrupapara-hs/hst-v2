-- Group commission settings and tier ladders (mt5_commissions / mt5_commissions_tiers).

CREATE SEQUENCE IF NOT EXISTS hst.commissions_commission_id_seq AS BIGINT START WITH 1 CACHE 1;
CREATE SEQUENCE IF NOT EXISTS hst.commissions_tiers_tier_id_seq AS BIGINT START WITH 1 CACHE 1;

CREATE TABLE IF NOT EXISTS hst.commissions (
    commission_id     BIGINT PRIMARY KEY DEFAULT nextval('hst.commissions_commission_id_seq'),
    group_id          BIGINT       NOT NULL REFERENCES hst.groups (group_id) ON DELETE CASCADE,
    updated_at        BIGINT       NOT NULL DEFAULT 0,
    name              VARCHAR(64)  NOT NULL,
    description       VARCHAR(64)  NOT NULL DEFAULT '',
    path              VARCHAR(255) NOT NULL DEFAULT '',
    mode              INTEGER      NOT NULL DEFAULT 0,
    mode_range        INTEGER      NOT NULL DEFAULT 0,
    mode_charge       INTEGER      NOT NULL DEFAULT 0,
    turnover_currency VARCHAR(16)  NOT NULL DEFAULT '',
    mode_entry        INTEGER      NOT NULL DEFAULT 0,
    mode_action       INTEGER      NOT NULL DEFAULT 0,
    mode_profit       INTEGER      NOT NULL DEFAULT 0,
    mode_reason       INTEGER      NOT NULL DEFAULT 0,

    CONSTRAINT commissions_name_set CHECK (name <> '')
);

CREATE INDEX IF NOT EXISTS commissions_group_id_idx ON hst.commissions (group_id);
CREATE INDEX IF NOT EXISTS commissions_group_path_idx ON hst.commissions (group_id, path);

COMMENT ON TABLE hst.commissions IS 'Per-group commission headers; path is a symbol or path mask';
COMMENT ON COLUMN hst.commissions.mode IS '0=standard 1=agent';
COMMENT ON COLUMN hst.commissions.mode_range IS '0=volume 1=turnover_money 2=turnover_volume';
COMMENT ON COLUMN hst.commissions.mode_charge IS '0=daily 1=monthly 2=instant';
COMMENT ON COLUMN hst.commissions.mode_entry IS '0=all 1=in 2=out';
COMMENT ON COLUMN hst.commissions.mode_action IS '0=all 1=buy 2=sell';
COMMENT ON COLUMN hst.commissions.mode_profit IS '0=all 1=profit 2=loss';
COMMENT ON COLUMN hst.commissions.mode_reason IS 'deal-reason bitmask';

CREATE TABLE IF NOT EXISTS hst.commissions_tiers (
    tier_id       BIGINT PRIMARY KEY DEFAULT nextval('hst.commissions_tiers_tier_id_seq'),
    commission_id BIGINT        NOT NULL REFERENCES hst.commissions (commission_id) ON DELETE CASCADE,
    mode          INTEGER       NOT NULL DEFAULT 0,
    type          INTEGER       NOT NULL DEFAULT 0,
    value         NUMERIC(20,8) NOT NULL DEFAULT 0,
    range_from    NUMERIC(20,8) NOT NULL DEFAULT 0,
    range_to      NUMERIC(20,8) NOT NULL DEFAULT 0,
    minimal       NUMERIC(20,8) NOT NULL DEFAULT 0,
    currency      VARCHAR(16)   NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS commissions_tiers_commission_id_idx
    ON hst.commissions_tiers (commission_id);

COMMENT ON TABLE hst.commissions_tiers IS 'Commission level ladder under a commission header';
COMMENT ON COLUMN hst.commissions_tiers.mode IS '0=deposit 1=base 2=profit 3=margin 4=points 5=percent 6=specified currency';
COMMENT ON COLUMN hst.commissions_tiers.type IS '0=per_trade 1=per_volume';
