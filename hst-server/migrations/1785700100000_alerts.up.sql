-- What a terminal asked to be told about: a price level, or a level of its own money.
CREATE TABLE IF NOT EXISTS hst.alerts (
    alert_id     BIGSERIAL PRIMARY KEY,
    login        BIGINT           NOT NULL,
    symbol       TEXT,
    kind         INT              NOT NULL,
    condition    INT              NOT NULL,
    value        DOUBLE PRECISION NOT NULL DEFAULT 0,
    enabled      BOOLEAN          NOT NULL DEFAULT TRUE,
    triggered_at BIGINT           NOT NULL DEFAULT 0,
    created_at   BIGINT           NOT NULL DEFAULT 0,
    updated_at   BIGINT           NOT NULL DEFAULT 0,
    comment      TEXT             NOT NULL DEFAULT ''
);

COMMENT ON TABLE hst.alerts IS 'trader alerts: kind 1 ask, 2 bid, 3 balance, 4 equity, 5 margin level';
COMMENT ON COLUMN hst.alerts.symbol IS 'set for price kinds only, null for account kinds';

CREATE INDEX IF NOT EXISTS alerts_login_idx ON hst.alerts (login);
-- the evaluator only ever loads the armed ones
CREATE INDEX IF NOT EXISTS alerts_enabled_idx ON hst.alerts (enabled) WHERE enabled;
