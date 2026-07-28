CREATE TABLE IF NOT EXISTS hst.accounts (
    login               BIGINT        PRIMARY KEY REFERENCES hst.users (login) ON DELETE CASCADE,
    currency_digits     SMALLINT      NOT NULL DEFAULT 2,
    balance             NUMERIC(20,8) NOT NULL DEFAULT 0,
    credit              NUMERIC(20,8) NOT NULL DEFAULT 0,
    margin              NUMERIC(20,8) NOT NULL DEFAULT 0,
    margin_free         NUMERIC(20,8) NOT NULL DEFAULT 0,
    margin_level        NUMERIC(18,4) NOT NULL DEFAULT 0,
    margin_leverage     INTEGER       NOT NULL DEFAULT 100,
    margin_initial      NUMERIC(20,8) NOT NULL DEFAULT 0,
    margin_maintenance  NUMERIC(20,8) NOT NULL DEFAULT 0,
    profit              NUMERIC(20,8) NOT NULL DEFAULT 0,
    storage             NUMERIC(20,8) NOT NULL DEFAULT 0,
    floating            NUMERIC(20,8) NOT NULL DEFAULT 0,
    equity              NUMERIC(20,8) NOT NULL DEFAULT 0,
    blocked_commission  NUMERIC(20,8) NOT NULL DEFAULT 0,
    blocked_profit      NUMERIC(20,8) NOT NULL DEFAULT 0,
    assets              NUMERIC(20,8) NOT NULL DEFAULT 0,
    liabilities         NUMERIC(20,8) NOT NULL DEFAULT 0,
    updated_at          BIGINT   NOT NULL DEFAULT 0
);

COMMENT ON TABLE hst.accounts IS 'all time columns are unix nanoseconds, 0 means unset';
