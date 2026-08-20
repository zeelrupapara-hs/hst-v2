-- Per-login TradingView chart state: one widget.save() blob per chart slot, plus one
-- shared settings document per login for the library's settings_adapter.

CREATE TABLE IF NOT EXISTS hst.chart_layouts (
    login      BIGINT  NOT NULL,
    chart_id   INTEGER NOT NULL,
    content    JSONB   NOT NULL,
    updated_at BIGINT  NOT NULL DEFAULT 0,

    PRIMARY KEY (login, chart_id),
    CONSTRAINT chart_layouts_slot CHECK (chart_id BETWEEN 1 AND 4)
);

CREATE TABLE IF NOT EXISTS hst.chart_settings (
    login      BIGINT PRIMARY KEY,
    content    JSONB  NOT NULL DEFAULT '{}'::jsonb,
    updated_at BIGINT NOT NULL DEFAULT 0
);

COMMENT ON TABLE hst.chart_layouts IS 'per login TradingView widget.save() blobs, one per chart slot';
COMMENT ON COLUMN hst.chart_layouts.updated_at IS 'unix nanoseconds';
COMMENT ON TABLE hst.chart_settings IS 'per login TradingView settings_adapter key/value doc';
COMMENT ON COLUMN hst.chart_settings.updated_at IS 'unix nanoseconds';
