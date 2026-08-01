-- hst.groups — client group config templates (path hierarchy).
-- Server column omitted: single-server.
--
-- There is no parent column and no child column. A group's name is its whole
-- path, demo\demo2\test, and a section is only a prefix that some group's path
-- passes through: sections come into being when a group is created below them
-- and vanish when the last one is deleted, because they were never stored. A
-- path can be both at once, so a group keeps its settings and its accounts
-- after another group appears beneath it.
--
-- The path is therefore the identity, and it is also the manager's access mask
-- and the websocket subject. Storing the hierarchy a second time in a parent
-- link would be a second copy of the same fact, free to disagree with it.

CREATE SEQUENCE IF NOT EXISTS hst.groups_group_id_seq AS BIGINT START WITH 1 CACHE 1;

CREATE TABLE IF NOT EXISTS hst.groups (
    group_id                BIGINT       PRIMARY KEY DEFAULT nextval('hst.groups_group_id_seq'),
    updated_at              BIGINT       NOT NULL DEFAULT 0,
    "group"                 VARCHAR(255) NOT NULL,

    permission_flags        INTEGER      NOT NULL DEFAULT 1,
    auth_mode               INTEGER      NOT NULL DEFAULT 0,
    auth_password_min       INTEGER      NOT NULL DEFAULT 0,

    company                 VARCHAR(255) NOT NULL DEFAULT '',
    company_page            TEXT         NOT NULL DEFAULT '',
    company_email           VARCHAR(255) NOT NULL DEFAULT '',
    company_support_page    TEXT         NOT NULL DEFAULT '',
    company_support_email   VARCHAR(255) NOT NULL DEFAULT '',
    company_catalog         VARCHAR(255) NOT NULL DEFAULT '',

    currency                VARCHAR(16)  NOT NULL DEFAULT 'USD',
    currency_digits         INTEGER      NOT NULL DEFAULT 2,

    reports_mode            INTEGER      NOT NULL DEFAULT 0,
    reports_flags           INTEGER      NOT NULL DEFAULT 0,
    reports_email           VARCHAR(255) NOT NULL DEFAULT '',
    reports_smtp            TEXT         NOT NULL DEFAULT '',
    reports_smtp_login      TEXT         NOT NULL DEFAULT '',

    news_mode               INTEGER      NOT NULL DEFAULT 0,
    news_category           TEXT         NOT NULL DEFAULT '',
    news_langs              INTEGER[]    NOT NULL DEFAULT '{}',
    mail_mode               INTEGER      NOT NULL DEFAULT 0,

    trade_flags             INTEGER      NOT NULL DEFAULT 0,
    trade_interest_rate     NUMERIC(20,8) NOT NULL DEFAULT 0,
    trade_virtual_credit    NUMERIC(20,8) NOT NULL DEFAULT 0,
    trade_transfer_mode     INTEGER      NOT NULL DEFAULT 0,

    margin_free_mode        INTEGER      NOT NULL DEFAULT 0,
    margin_so_mode          INTEGER      NOT NULL DEFAULT 0,
    margin_call             NUMERIC(20,8) NOT NULL DEFAULT 0,
    margin_stop_out         NUMERIC(20,8) NOT NULL DEFAULT 0,
    margin_free_profit_mode INTEGER      NOT NULL DEFAULT 0,
    margin_mode             INTEGER      NOT NULL DEFAULT 0,
    margin_flags            INTEGER      NOT NULL DEFAULT 0,

    -- both are genuinely unset when empty, which is not the same as zero: no deposit is a zero balance, no leverage is 1
    demo_leverage           INTEGER       NULL,
    demo_deposit            NUMERIC(20,8) NULL,

    limit_history           INTEGER      NOT NULL DEFAULT 0,
    limit_orders            INTEGER      NOT NULL DEFAULT 0,
    limit_symbols           INTEGER      NOT NULL DEFAULT 0,
    limit_positions         INTEGER      NOT NULL DEFAULT 0,
    limit_positions_volume  NUMERIC(20,8) NOT NULL DEFAULT 0,

    CONSTRAINT groups_path_set CHECK ("group" <> '')
);

COMMENT ON TABLE hst.groups IS 'Group config templates; the path in "group" is the identity and the whole hierarchy';
COMMENT ON COLUMN hst.groups."group" IS 'Hierarchical path (e.g. demo\forex\usd)';
COMMENT ON COLUMN hst.groups.permission_flags IS 'EnPermissionsFlags bitmask; 0 = disabled';
COMMENT ON COLUMN hst.groups.auth_mode IS 'EnAuthMode';
COMMENT ON COLUMN hst.groups.news_langs IS 'Windows LANGID values';
COMMENT ON COLUMN hst.groups.reports_smtp IS 'Obsolete';
COMMENT ON COLUMN hst.groups.reports_smtp_login IS 'Obsolete';
COMMENT ON COLUMN hst.groups.limit_positions_volume IS 'Currently unused';

CREATE UNIQUE INDEX IF NOT EXISTS groups_group_uidx ON hst.groups ("group");
CREATE INDEX IF NOT EXISTS groups_updated_at_idx ON hst.groups (updated_at DESC);

-- Per-group symbol overrides. NULL from trade_mode onward = inherit base symbol.
CREATE SEQUENCE IF NOT EXISTS hst.groups_symbols_symbol_id_seq AS BIGINT START WITH 1 CACHE 1;

CREATE TABLE IF NOT EXISTS hst.groups_symbols (
    symbol_id                          BIGINT PRIMARY KEY DEFAULT nextval('hst.groups_symbols_symbol_id_seq'),
    group_id                           BIGINT NOT NULL REFERENCES hst.groups (group_id) ON DELETE CASCADE,
    updated_at                         BIGINT NOT NULL DEFAULT 0,
    path                               VARCHAR(255) NOT NULL,
    config_index                       INTEGER NOT NULL DEFAULT 0,

    trade_mode                         INTEGER,
    exec_mode                          INTEGER,
    fill_flags                         INTEGER,
    expir_flags                        INTEGER,
    spread_diff                        INTEGER,
    spread_diff_balance                INTEGER,
    stops_level                        INTEGER,
    freeze_level                       INTEGER,
    volume_min                         BIGINT,
    volume_min_ext                     BIGINT,
    volume_max                         BIGINT,
    volume_max_ext                     BIGINT,
    volume_step                        BIGINT,
    volume_step_ext                    BIGINT,
    volume_limit                       BIGINT,
    volume_limit_ext                   BIGINT,
    margin_flags                       INTEGER,
    margin_initial                     NUMERIC(20,8),
    margin_maintenance                 NUMERIC(20,8),
    margin_initial_buy                 NUMERIC(20,8),
    margin_initial_sell                NUMERIC(20,8),
    margin_initial_buy_limit           NUMERIC(20,8),
    margin_initial_sell_limit          NUMERIC(20,8),
    margin_initial_buy_stop            NUMERIC(20,8),
    margin_initial_sell_stop           NUMERIC(20,8),
    margin_initial_buy_stop_limit      NUMERIC(20,8),
    margin_initial_sell_stop_limit     NUMERIC(20,8),
    margin_maintenance_buy             NUMERIC(20,8),
    margin_maintenance_sell            NUMERIC(20,8),
    margin_maintenance_buy_limit       NUMERIC(20,8),
    margin_maintenance_sell_limit      NUMERIC(20,8),
    margin_maintenance_buy_stop        NUMERIC(20,8),
    margin_maintenance_sell_stop       NUMERIC(20,8),
    margin_maintenance_buy_stop_limit  NUMERIC(20,8),
    margin_maintenance_sell_stop_limit NUMERIC(20,8),
    margin_currency                    VARCHAR(16),
    margin_liquidity                   NUMERIC(20,8),
    margin_hedged                      NUMERIC(20,8),
    swap_mode                          INTEGER,
    swap_long                          NUMERIC(20,8),
    swap_short                         NUMERIC(20,8),
    swap_year_day                      INTEGER,
    swap_flags                         INTEGER,
    swap_rate_sunday                   NUMERIC(20,8),
    swap_rate_monday                   NUMERIC(20,8),
    swap_rate_tuesday                  NUMERIC(20,8),
    swap_rate_wednesday                NUMERIC(20,8),
    swap_rate_thursday                 NUMERIC(20,8),
    swap_rate_friday                   NUMERIC(20,8),
    swap_rate_saturday                 NUMERIC(20,8),
    re_timeout                         INTEGER,
    ie_check_mode                      INTEGER,
    ie_timeout                         INTEGER,
    ie_slip_profit                     INTEGER,
    ie_slip_losing                     INTEGER,
    ie_volume_max                      BIGINT,
    ie_volume_max_ext                  BIGINT,
    ie_flags                           INTEGER,
    order_flags                        INTEGER,
    permissions_flags                  INTEGER,
    permissions_book_depth             INTEGER,
    re_flags                           INTEGER,

    CONSTRAINT groups_symbols_path_set CHECK (path <> '')
);

COMMENT ON TABLE hst.groups_symbols IS 'Per-group symbol overrides; NULL = inherit base symbol';
COMMENT ON COLUMN hst.groups_symbols.path IS 'Symbol path mask (EURUSD, Forex\*, *)';
COMMENT ON COLUMN hst.groups_symbols.trade_mode IS 'NULL = inherit (and following override fields)';

CREATE INDEX IF NOT EXISTS groups_symbols_group_id_idx ON hst.groups_symbols (group_id);
CREATE INDEX IF NOT EXISTS groups_symbols_group_path_idx ON hst.groups_symbols (group_id, path);
CREATE INDEX IF NOT EXISTS groups_symbols_updated_at_idx ON hst.groups_symbols (updated_at DESC);
